// Package protection contains GoreeCloud Backup product-layer recovery-assurance
// state and evidence types. It intentionally does not depend on Kopia repository
// internals so the product layer can evolve without changing recovery-critical
// repository behavior.
package protection

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

var errInvalidProtectionAssessment = errors.New("invalid protection assessment")

// State is the user-facing protection state for a protected system or dataset.
type State string

// StateUnprotected and related values describe user-facing protection state.
const (
	StateUnprotected     State = "unprotected"
	StateConfigured      State = "configured"
	StateBackingUp       State = "backing_up"
	StateProtected       State = "protected"
	StateRestoreVerified State = "restore_verified"
	StateDegraded        State = "degraded"
)

// EvidenceStatus describes the current status of one recovery-assurance check.
type EvidenceStatus string

// EvidenceUnknown and related values describe bounded evidence status.
const (
	EvidenceUnknown       EvidenceStatus = "unknown"
	EvidencePassing       EvidenceStatus = "passing"
	EvidenceFailing       EvidenceStatus = "failing"
	EvidenceStale         EvidenceStatus = "stale"
	EvidenceNotApplicable EvidenceStatus = "not_applicable"
)

func normalizeEvidenceStatus(s EvidenceStatus) EvidenceStatus {
	if s == "" {
		return EvidenceUnknown
	}

	return s
}

func (s EvidenceStatus) valid() bool {
	switch s {
	case EvidenceUnknown, EvidencePassing, EvidenceFailing, EvidenceStale, EvidenceNotApplicable:
		return true
	default:
		return false
	}
}

// EvidenceKind identifies a bounded product-layer recovery-assurance check.
type EvidenceKind string

// EvidenceRepositoryAvailable and related values identify recovery-assurance checks.
const (
	EvidenceRepositoryAvailable    EvidenceKind = "repository_available"
	EvidenceCredentialsRecoverable EvidenceKind = "credentials_recoverable"
	EvidenceRecoveryPointAvailable EvidenceKind = "recovery_point_available"
	EvidenceBackupCurrent          EvidenceKind = "backup_current"
	EvidenceIntegrity              EvidenceKind = "integrity"
	EvidenceScope                  EvidenceKind = "scope"
	EvidenceApplicationConsistency EvidenceKind = "application_consistency"
	EvidenceRetention              EvidenceKind = "retention"
	EvidenceMaintenance            EvidenceKind = "maintenance"
	EvidenceMonitoring             EvidenceKind = "monitoring"
	EvidenceNotification           EvidenceKind = "notification"
)

var baselineRequiredEvidence = []EvidenceKind{
	EvidenceRepositoryAvailable,
	EvidenceCredentialsRecoverable,
	EvidenceRecoveryPointAvailable,
	EvidenceBackupCurrent,
	EvidenceIntegrity,
	EvidenceScope,
	EvidenceApplicationConsistency,
	EvidenceRetention,
	EvidenceMaintenance,
	EvidenceMonitoring,
	EvidenceNotification,
}

func (k EvidenceKind) valid() bool {
	switch k {
	case EvidenceRepositoryAvailable,
		EvidenceCredentialsRecoverable,
		EvidenceRecoveryPointAvailable,
		EvidenceBackupCurrent,
		EvidenceIntegrity,
		EvidenceScope,
		EvidenceApplicationConsistency,
		EvidenceRetention,
		EvidenceMaintenance,
		EvidenceMonitoring,
		EvidenceNotification:
		return true
	default:
		return false
	}
}

// BaselineRequiredEvidence returns the default evidence gates for Protected.
// The returned slice is a copy and may be safely modified by the caller.
func BaselineRequiredEvidence() []EvidenceKind {
	return append([]EvidenceKind(nil), baselineRequiredEvidence...)
}

// EvidenceItem is one policy-selected recovery-assurance check.
// Baseline evidence always gates Protected and Degraded state. Required may be
// used by future policies to make an otherwise optional bounded check a gate;
// it never disables a baseline requirement.
type EvidenceItem struct {
	Kind     EvidenceKind   `json:"kind"`
	Status   EvidenceStatus `json:"status"`
	Required bool           `json:"required"`
}

// Assessment is the product-layer input used to derive a protection state.
// RestoreVerification is separate from operational evidence because Protected
// and Restore Verified are deliberately distinct states.
type Assessment struct {
	Configured          bool           `json:"configured"`
	BackupInProgress    bool           `json:"backupInProgress"`
	Evidence            []EvidenceItem `json:"evidence"`
	RestoreVerification EvidenceStatus `json:"restoreVerification"`
}

// ReasonCode is a bounded explanation suitable for API, Manager, Monitor, and
// Glaze UI consumers without exposing private backup contents.
type ReasonCode string

// ReasonNotConfigured and related values explain evaluated protection state.
const (
	ReasonNotConfigured              ReasonCode = "not_configured"
	ReasonBackupInProgress           ReasonCode = "backup_in_progress"
	ReasonRequiredEvidenceMissing    ReasonCode = "required_evidence_missing"
	ReasonRequiredEvidenceFailed     ReasonCode = "required_evidence_failed"
	ReasonRequiredEvidenceStale      ReasonCode = "required_evidence_stale"
	ReasonOperationalEvidencePassing ReasonCode = "operational_evidence_passing"
	ReasonRestoreVerificationPassing ReasonCode = "restore_verification_passing"
	ReasonRestoreVerificationFailed  ReasonCode = "restore_verification_failed"
	ReasonRestoreVerificationStale   ReasonCode = "restore_verification_stale"
)

// Evaluation is a deterministic, privacy-conscious state result.
type Evaluation struct {
	State   State          `json:"state"`
	Reasons []ReasonCode   `json:"reasons,omitempty"`
	Missing []EvidenceKind `json:"missing,omitempty"`
	Failed  []EvidenceKind `json:"failed,omitempty"`
	Stale   []EvidenceKind `json:"stale,omitempty"`
}

// Evaluate derives protection state from policy-selected evidence.
//
// Precedence is intentionally conservative:
//   - not configured is Unprotected;
//   - failed or stale required evidence is Degraded;
//   - an active backup is Backing Up;
//   - missing required evidence remains Configured;
//   - passing required operational evidence is Protected;
//   - only an explicit passing restore verification can produce Restore Verified.
//
// Baseline evidence is always required even when a caller omits it from the
// assessment, so a successful snapshot alone cannot produce Protected or
// Restore Verified.
func Evaluate(a Assessment) (Evaluation, error) {
	restoreVerification := normalizeEvidenceStatus(a.RestoreVerification)
	if !restoreVerification.valid() {
		return Evaluation{}, fmt.Errorf("%w: invalid restore verification status %q", errInvalidProtectionAssessment, a.RestoreVerification)
	}

	var out Evaluation

	seen := map[EvidenceKind]struct{}{}
	statusByKind := map[EvidenceKind]EvidenceStatus{}

	required := map[EvidenceKind]struct{}{}
	for _, kind := range baselineRequiredEvidence {
		required[kind] = struct{}{}
	}

	for _, item := range a.Evidence {
		if strings.TrimSpace(string(item.Kind)) == "" {
			return Evaluation{}, errors.New("evidence kind must not be empty")
		}

		if !item.Kind.valid() {
			return Evaluation{}, fmt.Errorf("%w: invalid evidence kind %q", errInvalidProtectionAssessment, item.Kind)
		}

		status := normalizeEvidenceStatus(item.Status)
		if !status.valid() {
			return Evaluation{}, fmt.Errorf("%w: invalid status %q for evidence %q", errInvalidProtectionAssessment, item.Status, item.Kind)
		}

		if _, ok := seen[item.Kind]; ok {
			return Evaluation{}, fmt.Errorf("%w: duplicate evidence kind %q", errInvalidProtectionAssessment, item.Kind)
		}

		seen[item.Kind] = struct{}{}

		statusByKind[item.Kind] = status
		if item.Required {
			required[item.Kind] = struct{}{}
		}
	}

	for kind := range required {
		status, ok := statusByKind[kind]
		if !ok || status == EvidenceUnknown {
			out.Missing = append(out.Missing, kind)
			continue
		}

		if status == EvidenceNotApplicable {
			return Evaluation{}, fmt.Errorf("%w: required evidence %q cannot be not applicable", errInvalidProtectionAssessment, kind)
		}

		switch status {
		case EvidenceFailing:
			out.Failed = append(out.Failed, kind)
		case EvidenceStale:
			out.Stale = append(out.Stale, kind)
		}
	}

	sortEvidenceKinds(out.Missing)
	sortEvidenceKinds(out.Failed)
	sortEvidenceKinds(out.Stale)

	if !a.Configured {
		out.State = StateUnprotected
		out.Reasons = []ReasonCode{ReasonNotConfigured}

		return out, nil
	}

	if len(out.Failed) > 0 || len(out.Stale) > 0 || restoreVerification == EvidenceFailing || restoreVerification == EvidenceStale {
		out.State = StateDegraded
		if len(out.Failed) > 0 {
			out.Reasons = append(out.Reasons, ReasonRequiredEvidenceFailed)
		}

		if len(out.Stale) > 0 {
			out.Reasons = append(out.Reasons, ReasonRequiredEvidenceStale)
		}

		if restoreVerification == EvidenceFailing {
			out.Reasons = append(out.Reasons, ReasonRestoreVerificationFailed)
		}

		if restoreVerification == EvidenceStale {
			out.Reasons = append(out.Reasons, ReasonRestoreVerificationStale)
		}

		return out, nil
	}

	if a.BackupInProgress {
		out.State = StateBackingUp
		out.Reasons = []ReasonCode{ReasonBackupInProgress}

		return out, nil
	}

	if len(out.Missing) > 0 {
		out.State = StateConfigured
		out.Reasons = []ReasonCode{ReasonRequiredEvidenceMissing}

		return out, nil
	}

	out.State = StateProtected

	out.Reasons = []ReasonCode{ReasonOperationalEvidencePassing}
	if restoreVerification == EvidencePassing {
		out.State = StateRestoreVerified
		out.Reasons = append(out.Reasons, ReasonRestoreVerificationPassing)
	}

	return out, nil
}

func sortEvidenceKinds(v []EvidenceKind) {
	slices.Sort(v)
}

// VerificationType describes the bounded kind of representative restore test.
type VerificationType string

// VerificationFileSample and related values identify representative restore tests.
const (
	VerificationFileSample            VerificationType = "file_sample"
	VerificationMetadataSample        VerificationType = "metadata_sample"
	VerificationApplicationDataset    VerificationType = "application_dataset"
	VerificationApplicationBehavior   VerificationType = "application_behavior"
	VerificationRepresentativeRestore VerificationType = "representative_restore"
)

func (v VerificationType) valid() bool {
	switch v {
	case VerificationFileSample, VerificationMetadataSample, VerificationApplicationDataset, VerificationApplicationBehavior, VerificationRepresentativeRestore:
		return true
	default:
		return false
	}
}

// ValidationCheck identifies what a restore test actually validated.
type ValidationCheck string

// ValidationContentHash and related values identify restore-validation checks.
const (
	ValidationContentHash         ValidationCheck = "content_hash"
	ValidationMetadata            ValidationCheck = "metadata"
	ValidationOwnership           ValidationCheck = "ownership"
	ValidationPermissions         ValidationCheck = "permissions"
	ValidationApplicationStart    ValidationCheck = "application_start"
	ValidationApplicationBehavior ValidationCheck = "application_behavior"
)

func (v ValidationCheck) valid() bool {
	switch v {
	case ValidationContentHash, ValidationMetadata, ValidationOwnership, ValidationPermissions, ValidationApplicationStart, ValidationApplicationBehavior:
		return true
	default:
		return false
	}
}

// FailureCategory is deliberately bounded so evidence can explain failures
// without storing restored content, credentials, raw backend responses, or
// other sensitive information.
type FailureCategory string

// FailureNone and related values classify bounded restore-test failures.
const (
	FailureNone                  FailureCategory = "none"
	FailureRepositoryUnavailable FailureCategory = "repository_unavailable"
	FailureAuthentication        FailureCategory = "authentication"
	FailureIntegrity             FailureCategory = "integrity"
	FailureRestore               FailureCategory = "restore"
	FailureValidation            FailureCategory = "validation"
	FailureCleanup               FailureCategory = "cleanup"
	FailureTimeout               FailureCategory = "timeout"
	FailureCapacity              FailureCategory = "capacity"
	FailurePolicy                FailureCategory = "policy"
	FailureUnknown               FailureCategory = "unknown"
)

func (f FailureCategory) valid() bool {
	switch f {
	case FailureNone, FailureRepositoryUnavailable, FailureAuthentication, FailureIntegrity, FailureRestore, FailureValidation, FailureCleanup, FailureTimeout, FailureCapacity, FailurePolicy, FailureUnknown:
		return true
	default:
		return false
	}
}

// RestoreTestEvidence records only the bounded facts needed to prove and
// explain a representative restore test. It intentionally contains no field
// for restored file contents, credentials, tokens, or raw backend output.
type RestoreTestEvidence struct {
	Type            VerificationType  `json:"type"`
	Status          EvidenceStatus    `json:"status"`
	CompletedAt     time.Time         `json:"completedAt"`
	Checks          []ValidationCheck `json:"checks,omitempty"`
	FailureCategory FailureCategory   `json:"failureCategory"`
}

// RecoveryEvidence records recovery-assurance evidence for one recovery point.
// Policy and procedure identifiers are references only; sensitive secret data
// must remain outside this record.
type RecoveryEvidence struct {
	DatasetID           string               `json:"datasetId"`
	RepositoryID        string               `json:"repositoryId"`
	RecoveryPointID     string               `json:"recoveryPointId"`
	ObservedAt          time.Time            `json:"observedAt"`
	SoftwareVersion     string               `json:"softwareVersion,omitempty"`
	ProtectionPolicyID  string               `json:"protectionPolicyId,omitempty"`
	RecoveryProcedureID string               `json:"recoveryProcedureId,omitempty"`
	BackupStatus        EvidenceStatus       `json:"backupStatus"`
	IntegrityStatus     EvidenceStatus       `json:"integrityStatus"`
	RestoreTest         *RestoreTestEvidence `json:"restoreTest,omitempty"`
}

// Validate checks that a recovery-evidence record is internally coherent.
func (r RecoveryEvidence) Validate() error {
	if strings.TrimSpace(r.DatasetID) == "" {
		return errors.New("dataset ID must not be empty")
	}

	if strings.TrimSpace(r.RepositoryID) == "" {
		return errors.New("repository ID must not be empty")
	}

	if strings.TrimSpace(r.RecoveryPointID) == "" {
		return errors.New("recovery point ID must not be empty")
	}

	if r.ObservedAt.IsZero() {
		return errors.New("observed time must not be zero")
	}

	if !r.BackupStatus.valid() || r.BackupStatus == EvidenceNotApplicable {
		return fmt.Errorf("%w: invalid backup status %q", errInvalidProtectionAssessment, r.BackupStatus)
	}

	if !r.IntegrityStatus.valid() || r.IntegrityStatus == EvidenceNotApplicable {
		return fmt.Errorf("%w: invalid integrity status %q", errInvalidProtectionAssessment, r.IntegrityStatus)
	}

	if r.RestoreTest == nil {
		return nil
	}

	t := r.RestoreTest
	if !t.Type.valid() {
		return fmt.Errorf("%w: invalid restore verification type %q", errInvalidProtectionAssessment, t.Type)
	}

	if t.Status != EvidencePassing && t.Status != EvidenceFailing {
		return fmt.Errorf("%w: restore test status must be passing or failing, got %q", errInvalidProtectionAssessment, t.Status)
	}

	if t.CompletedAt.IsZero() {
		return errors.New("restore test completion time must not be zero")
	}

	if !t.FailureCategory.valid() {
		return fmt.Errorf("%w: invalid failure category %q", errInvalidProtectionAssessment, t.FailureCategory)
	}

	if t.Status == EvidencePassing && t.FailureCategory != FailureNone {
		return fmt.Errorf("%w: passing restore test cannot have failure category %q", errInvalidProtectionAssessment, t.FailureCategory)
	}

	if t.Status == EvidenceFailing && t.FailureCategory == FailureNone {
		return errors.New("failing restore test must have a failure category")
	}

	seenChecks := map[ValidationCheck]struct{}{}
	for _, check := range t.Checks {
		if !check.valid() {
			return fmt.Errorf("%w: invalid validation check %q", errInvalidProtectionAssessment, check)
		}

		if _, ok := seenChecks[check]; ok {
			return fmt.Errorf("%w: duplicate validation check %q", errInvalidProtectionAssessment, check)
		}

		seenChecks[check] = struct{}{}
	}

	return nil
}
