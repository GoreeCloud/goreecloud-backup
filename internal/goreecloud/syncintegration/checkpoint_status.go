package syncintegration

import (
	"context"
	"fmt"
	"time"
)

// CheckpointLifecycleState distinguishes request acceptance and engine activity
// from the later existence of an independently usable recovery point.
type CheckpointLifecycleState string

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
		return fmt.Errorf("unsupported contract version %q", s.ContractVersion)
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
		return fmt.Errorf("checkpoint status observation time must not be zero")
	}
	if !s.State.valid() {
		return fmt.Errorf("invalid checkpoint lifecycle state %q", s.State)
	}

	switch s.State {
	case CheckpointStateAccepted, CheckpointStateRunning:
		if s.RecoveryPointID != "" || s.RecoveryPointUsable || s.IntegrityVerified || s.RestoreVerified || s.FailureCategory != CheckpointFailureNone {
			return fmt.Errorf("checkpoint state %q must not claim recovery success or failure evidence", s.State)
		}
	case CheckpointStateFailed:
		if !s.FailureCategory.validFailure() {
			return fmt.Errorf("failed checkpoint requires a bounded failure category")
		}
		if s.RecoveryPointID != "" || s.RecoveryPointUsable || s.IntegrityVerified || s.RestoreVerified {
			return fmt.Errorf("failed checkpoint must not claim a usable or verified recovery point")
		}
	case CheckpointStateCompleted:
		if s.FailureCategory != CheckpointFailureNone {
			return fmt.Errorf("completed checkpoint must not carry a failure category")
		}
		if err := validateOpaqueIdentifier("recovery point ID", s.RecoveryPointID); err != nil {
			return fmt.Errorf("completed checkpoint requires a recovery point: %w", err)
		}
		if s.RestoreVerified && (!s.RecoveryPointUsable || !s.IntegrityVerified) {
			return fmt.Errorf("restore verification requires a usable integrity-verified recovery point")
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
		return fmt.Errorf("checkpoint submission acceptance time must not be zero")
	}
	if s.RequestID != submission.RequestID {
		return fmt.Errorf("checkpoint status request ID does not match submission")
	}
	if s.OperationID != submission.OperationID {
		return fmt.Errorf("checkpoint status operation ID does not match submission")
	}
	if s.DatasetID != submission.DatasetID {
		return fmt.Errorf("checkpoint status dataset ID does not match submission")
	}
	if s.BackupScopeID != submission.BackupScopeID {
		return fmt.Errorf("checkpoint status Backup scope ID does not match submission")
	}
	if s.ObservedAt.Before(submission.AcceptedAt) {
		return fmt.Errorf("checkpoint status predates checkpoint submission")
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
	CheckpointStatus(context.Context, string) (CheckpointStatus, error)
}
