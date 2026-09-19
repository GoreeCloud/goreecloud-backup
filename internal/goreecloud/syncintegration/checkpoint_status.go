package syncintegration

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var errInvalidCheckpointStatus = errors.New("invalid checkpoint status")

// CheckpointLifecycleState distinguishes request acceptance and engine activity
// from the later existence of an independently usable recovery point.
type CheckpointLifecycleState string

// CheckpointStateAccepted and related values describe checkpoint lifecycle state.
const (
	CheckpointStateAccepted  CheckpointLifecycleState = "accepted"
	CheckpointStateRunning   CheckpointLifecycleState = "running"
	CheckpointStateFailed    CheckpointLifecycleState = "failed"
	CheckpointStateCompleted CheckpointLifecycleState = "completed"
)

func (s CheckpointLifecycleState) valid() bool {
	switch s {
	case CheckpointStateAccepted, CheckpointStateRunning, CheckpointStateFailed, CheckpointStateCompleted:
		return true
	default:
		return false
	}
}

// CheckpointFailureCategory is deliberately bounded so cross-product status
// does not leak raw engine, repository, credential, path, or protected-content
// errors into GoreeCloud Sync.
type CheckpointFailureCategory string

// CheckpointFailureNone and related values classify bounded checkpoint failures.
const (
	CheckpointFailureNone         CheckpointFailureCategory = ""
	CheckpointFailureExecution    CheckpointFailureCategory = "execution"
	CheckpointFailureVerification CheckpointFailureCategory = "verification"
	CheckpointFailureRepository   CheckpointFailureCategory = "repository"
	CheckpointFailurePolicy       CheckpointFailureCategory = "policy"
	CheckpointFailureUnavailable  CheckpointFailureCategory = "unavailable"
	CheckpointFailureUnknown      CheckpointFailureCategory = "unknown"
)

func (c CheckpointFailureCategory) validFailure() bool {
	switch c {
	case CheckpointFailureExecution,
		CheckpointFailureVerification,
		CheckpointFailureRepository,
		CheckpointFailurePolicy,
		CheckpointFailureUnavailable,
		CheckpointFailureUnknown:
		return true
	default:
		return false
	}
}

// CheckpointStatus is Backup-owned checkpoint lifecycle and recovery evidence
// that may be surfaced to the authorized Sync integration. It deliberately
// distinguishes execution completion from a checkpoint that is safe to rely on
// before a destructive or migration-related Sync change.
type CheckpointStatus struct {
	ContractVersion     string                    `json:"contractVersion"`
	RequestID           string                    `json:"requestId"`
	OperationID         string                    `json:"operationId"`
	DatasetID           string                    `json:"datasetId"`
	BackupScopeID       string                    `json:"backupScopeId"`
	ObservedAt          time.Time                 `json:"observedAt"`
	State               CheckpointLifecycleState  `json:"state"`
	RecoveryPointID     string                    `json:"recoveryPointId,omitempty"`
	RecoveryPointUsable bool                      `json:"recoveryPointUsable"`
	IntegrityVerified   bool                      `json:"integrityVerified"`
	RestoreVerified     bool                      `json:"restoreVerified"`
	FailureCategory     CheckpointFailureCategory `json:"failureCategory,omitempty"`
}

// Validate enforces internally consistent lifecycle evidence. In particular,
// Accepted or Running can never carry recovery-point success, and Completed by
// itself does not imply that Sync may safely proceed.
func (s CheckpointStatus) Validate() error {
	if s.ContractVersion != ContractVersion {
		return fmt.Errorf("%w: unsupported contract version %q", errInvalidCheckpointStatus, s.ContractVersion)
	}

	if err := validateOpaqueIdentifier("checkpoint status request ID", s.RequestID); err != nil {
		return err
	}

	if err := validateOpaqueIdentifier("checkpoint status operation ID", s.OperationID); err != nil {
		return err
	}

	if err := validateOpaqueIdentifier("dataset ID", s.DatasetID); err != nil {
		return err
	}

	if err := validateOpaqueIdentifier("Backup scope ID", s.BackupScopeID); err != nil {
		return err
	}

	if s.ObservedAt.IsZero() {
		return errCheckpointStatusObservedTimeZero
	}

	if !s.State.valid() {
		return fmt.Errorf("%w: invalid checkpoint lifecycle state %q", errInvalidCheckpointStatus, s.State)
	}

	switch s.State {
	case CheckpointStateAccepted, CheckpointStateRunning:
		if s.RecoveryPointID != "" || s.RecoveryPointUsable || s.IntegrityVerified || s.RestoreVerified || s.FailureCategory != CheckpointFailureNone {
			return fmt.Errorf("%w: checkpoint state %q must not claim recovery success or failure evidence", errInvalidCheckpointStatus, s.State)
		}
	case CheckpointStateFailed:
		if !s.FailureCategory.validFailure() {
			return errCheckpointFailureCategoryRequired
		}

		if s.RecoveryPointID != "" || s.RecoveryPointUsable || s.IntegrityVerified || s.RestoreVerified {
			return errFailedCheckpointRecoveryEvidence
		}
	case CheckpointStateCompleted:
		if s.FailureCategory != CheckpointFailureNone {
			return errCompletedCheckpointFailureCategory
		}

		if err := validateOpaqueIdentifier("recovery point ID", s.RecoveryPointID); err != nil {
			return fmt.Errorf("completed checkpoint requires a recovery point: %w", err)
		}

		if s.RestoreVerified && (!s.RecoveryPointUsable || !s.IntegrityVerified) {
			return errRestoreVerificationEvidenceIncomplete
		}
	}

	return nil
}

// ValidateForSubmission binds Backup-owned lifecycle evidence to the exact
// accepted checkpoint receipt. This prevents a recovery point or status from
// another request, dataset, or Backup scope from satisfying the checkpoint.
func (s CheckpointStatus) ValidateForSubmission(submission CheckpointSubmission) error {
	if err := s.Validate(); err != nil {
		return err
	}

	if err := validateOpaqueIdentifier("checkpoint submission request ID", submission.RequestID); err != nil {
		return err
	}

	if err := validateOpaqueIdentifier("checkpoint submission operation ID", submission.OperationID); err != nil {
		return err
	}

	if err := validateOpaqueIdentifier("checkpoint submission dataset ID", submission.DatasetID); err != nil {
		return err
	}

	if err := validateOpaqueIdentifier("checkpoint submission Backup scope ID", submission.BackupScopeID); err != nil {
		return err
	}

	if submission.AcceptedAt.IsZero() {
		return errCheckpointSubmissionAcceptedAtZero
	}

	if s.RequestID != submission.RequestID {
		return errCheckpointStatusRequestMismatch
	}

	if s.OperationID != submission.OperationID {
		return errCheckpointStatusOperationMismatch
	}

	if s.DatasetID != submission.DatasetID {
		return errCheckpointStatusDatasetMismatch
	}

	if s.BackupScopeID != submission.BackupScopeID {
		return errCheckpointStatusScopeMismatch
	}

	if s.ObservedAt.Before(submission.AcceptedAt) {
		return errCheckpointStatusPredatesSubmission
	}

	return nil
}

// ReadyForProtectedChange reports whether Backup itself has supplied enough
// evidence for Sync to treat the requested pre-change/pre-migration checkpoint
// as usable. Engine completion alone is intentionally insufficient.
//
// RestoreVerified is stronger evidence but is not required for every individual
// pre-change checkpoint; applicable policy may impose a stronger gate at the
// runtime orchestration layer.
func (s CheckpointStatus) ReadyForProtectedChange() bool {
	return s.Validate() == nil &&
		s.State == CheckpointStateCompleted &&
		s.RecoveryPointUsable &&
		s.IntegrityVerified
}

// CheckpointStatusProvider is the Backup-owned runtime seam for authoritative
// checkpoint lifecycle/recovery evidence. The provider, not Sync, determines
// whether a recovery point exists and is usable. Calling the provider remains
// subject to the separately authenticated and authorized runtime boundary; a
// CheckpointSubmission is correlation evidence, not a bearer credential.
type CheckpointStatusProvider interface {
	CheckpointStatus(ctx context.Context, operationID string) (CheckpointStatus, error)
}
