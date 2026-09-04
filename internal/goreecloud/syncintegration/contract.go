// Package syncintegration defines the GoreeCloud Backup product-layer contract
// exposed to GoreeCloud Sync. It intentionally models coordination only: it
// does not implement network transport, GoreeCloud Identity authorization,
// GoreeCloud Mesh delivery, backup execution, repository mutation, or Sync
// runtime behavior.
//
// The boundary is deliberately narrow. GoreeCloud Backup remains authoritative
// for independent recovery state, recovery points, retention, repositories,
// verification, and restore behavior. GoreeCloud Sync remains authoritative for
// ongoing synchronization and reconciliation of Sync-managed datasets.
package syncintegration

import (
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/kopia/kopia/internal/goreecloud/protection"
)

// ContractVersion is the source-level version of this pre-stabilization
// Backup-to-Sync contract. It is not a public compatibility promise.
const ContractVersion = "goreecloud.backup-sync/v1alpha1"

const maxOpaqueIdentifierBytes = 256

// Operation identifies a bounded action that the Backup product layer may
// expose to an authorized Sync integration adapter.
type Operation string

const (
	// OperationReadProtection allows a consumer to obtain Backup-owned,
	// privacy-conscious protection state for an explicitly scoped dataset.
	OperationReadProtection Operation = "read_protection"
	// OperationRequestCheckpoint allows an authorized caller to request a
	// pre-change or pre-migration recovery checkpoint. The contract does not
	// itself authorize or execute the backup.
	OperationRequestCheckpoint Operation = "request_checkpoint"
	// OperationCoordinateRestore allows the two products to exchange the
	// minimum state required to restore safely around a Sync-managed path.
	OperationCoordinateRestore Operation = "coordinate_restore"
)

// ValidateOperation fails closed for every operation outside the narrow
// Backup-to-Sync contract. In particular, Sync receives no authority here to
// delete recovery points or repositories, change retention, rotate encryption
// material, or perform repository maintenance.
func ValidateOperation(op Operation) error {
	switch op {
	case OperationReadProtection, OperationRequestCheckpoint, OperationCoordinateRestore:
		return nil
	default:
		return fmt.Errorf("operation %q is not permitted by the Backup-to-Sync contract", op)
	}
}

// AllowedOperations returns a copy of the operations intentionally exposed by
// this contract.
func AllowedOperations() []Operation {
	return []Operation{
		OperationReadProtection,
		OperationRequestCheckpoint,
		OperationCoordinateRestore,
	}
}

// ProtectionView is the bounded Backup-owned state that may be surfaced to
// Sync. It contains reason/evidence categories only and no backed-up content,
// credentials, repository secrets, file names, paths, or reusable tokens.
type ProtectionView struct {
	ContractVersion string                    `json:"contractVersion"`
	DatasetID       string                    `json:"datasetId"`
	EvaluatedAt     time.Time                 `json:"evaluatedAt"`
	State           protection.State          `json:"state"`
	Reasons         []protection.ReasonCode   `json:"reasons,omitempty"`
	Missing         []protection.EvidenceKind `json:"missing,omitempty"`
	Failed          []protection.EvidenceKind `json:"failed,omitempty"`
	Stale           []protection.EvidenceKind `json:"stale,omitempty"`
}

// NewProtectionView converts an already-derived Backup protection evaluation
// into a Sync-safe view. This function does not collect evidence and does not
// allow Sync to influence the evaluation.
func NewProtectionView(datasetID string, evaluatedAt time.Time, evaluation protection.Evaluation) (ProtectionView, error) {
	if err := validateOpaqueIdentifier("dataset ID", datasetID); err != nil {
		return ProtectionView{}, err
	}
	if evaluatedAt.IsZero() {
		return ProtectionView{}, fmt.Errorf("evaluation time must not be zero")
	}
	if !validProtectionState(evaluation.State) {
		return ProtectionView{}, fmt.Errorf("invalid protection state %q", evaluation.State)
	}

	return ProtectionView{
		ContractVersion: ContractVersion,
		DatasetID:       datasetID,
		EvaluatedAt:     evaluatedAt,
		State:           evaluation.State,
		Reasons:         append([]protection.ReasonCode(nil), evaluation.Reasons...),
		Missing:         append([]protection.EvidenceKind(nil), evaluation.Missing...),
		Failed:          append([]protection.EvidenceKind(nil), evaluation.Failed...),
		Stale:           append([]protection.EvidenceKind(nil), evaluation.Stale...),
	}, nil
}

func validProtectionState(state protection.State) bool {
	switch state {
	case protection.StateUnprotected,
		protection.StateConfigured,
		protection.StateBackingUp,
		protection.StateProtected,
		protection.StateRestoreVerified,
		protection.StateDegraded:
		return true
	default:
		return false
	}
}

// CheckpointPurpose describes the only Sync-originated checkpoint purposes
// recognized by this contract.
type CheckpointPurpose string

const (
	CheckpointPreChange    CheckpointPurpose = "pre_change"
	CheckpointPreMigration CheckpointPurpose = "pre_migration"
)

// CheckpointRequest is a structural request envelope for a pre-change or
// pre-migration backup checkpoint.
//
// AuthorizationDecisionRef is an opaque reference only. Its presence is not
// proof of authorization. A future authenticated integration adapter must
// validate the referenced GoreeCloud Identity/security decision before invoking
// any Backup operation. This package intentionally cannot turn a caller-supplied
// string into authorization.
type CheckpointRequest struct {
	ContractVersion          string            `json:"contractVersion"`
	RequestID                string            `json:"requestId"`
	DatasetID                string            `json:"datasetId"`
	Purpose                  CheckpointPurpose `json:"purpose"`
	AuthorizationDecisionRef string            `json:"authorizationDecisionRef"`
}

// Validate checks only the shape and bounded values of a checkpoint request.
// It does not authorize the caller or execute a backup.
func (r CheckpointRequest) Validate() error {
	if r.ContractVersion != ContractVersion {
		return fmt.Errorf("unsupported contract version %q", r.ContractVersion)
	}
	if err := validateOpaqueIdentifier("request ID", r.RequestID); err != nil {
		return err
	}
	if err := validateOpaqueIdentifier("dataset ID", r.DatasetID); err != nil {
		return err
	}
	if err := validateOpaqueIdentifier("authorization decision reference", r.AuthorizationDecisionRef); err != nil {
		return err
	}

	switch r.Purpose {
	case CheckpointPreChange, CheckpointPreMigration:
		return nil
	default:
		return fmt.Errorf("checkpoint purpose %q is not permitted by the Backup-to-Sync contract", r.Purpose)
	}
}

// SyncAvailability describes whether Sync coordination can currently be
// reached. Backup recovery must remain possible when Sync is unavailable.
type SyncAvailability string

const (
	SyncAvailabilityAvailable   SyncAvailability = "available"
	SyncAvailabilityUnavailable SyncAvailability = "unavailable"
	SyncAvailabilityUnknown     SyncAvailability = "unknown"
)

func (a SyncAvailability) valid() bool {
	switch a {
	case SyncAvailabilityAvailable, SyncAvailabilityUnavailable, SyncAvailabilityUnknown:
		return true
	default:
		return false
	}
}

// CoordinationAction is a bounded action required before or after promotion of
// a restore into a Sync-managed production path.
type CoordinationAction string

const (
	ActionStageRestore         CoordinationAction = "stage_restore"
	ActionPauseOrMaintenance   CoordinationAction = "pause_or_maintenance"
	ActionValidateRestoredData CoordinationAction = "validate_restored_data"
	ActionReconcileWithSync    CoordinationAction = "reconcile_with_sync"
)

// ReconciliationState records only the Sync-related post-restore state.
type ReconciliationState string

const (
	ReconciliationNotRequired ReconciliationState = "not_required"
	ReconciliationRequired    ReconciliationState = "required"
	ReconciliationPending     ReconciliationState = "pending"
)

// RestoreCoordination is a safety plan for the Sync boundary. It is not a
// restore authorization or a restore execution result.
type RestoreCoordination struct {
	ContractVersion    string              `json:"contractVersion"`
	DatasetID          string              `json:"datasetId"`
	SyncManaged        bool                `json:"syncManaged"`
	SyncAvailability   SyncAvailability    `json:"syncAvailability"`
	RequiredActions    []CoordinationAction `json:"requiredActions,omitempty"`
	StagingRequired    bool                `json:"stagingRequired"`
	DirectWriteAllowed bool                `json:"directWriteAllowed"`
	Reconciliation     ReconciliationState `json:"reconciliation"`
}

// PlanRestoreCoordination derives the safe Sync-boundary requirements for a
// restore target.
//
// A Sync-managed target always requires staging and never receives direct-write
// permission from this source-level contract. Promotion into the production
// path must be separately authorized and coordinated after restored data is
// validated. If Sync is unavailable, Backup may still restore to a controlled
// staging location; reconciliation remains pending rather than blocking access
// to independent recovery data.
func PlanRestoreCoordination(datasetID string, syncManaged bool, availability SyncAvailability) (RestoreCoordination, error) {
	if err := validateOpaqueIdentifier("dataset ID", datasetID); err != nil {
		return RestoreCoordination{}, err
	}
	if !availability.valid() {
		return RestoreCoordination{}, fmt.Errorf("invalid Sync availability %q", availability)
	}

	plan := RestoreCoordination{
		ContractVersion:    ContractVersion,
		DatasetID:          datasetID,
		SyncManaged:        syncManaged,
		SyncAvailability:   availability,
		DirectWriteAllowed: !syncManaged,
		Reconciliation:     ReconciliationNotRequired,
	}

	if !syncManaged {
		return plan, nil
	}

	plan.StagingRequired = true
	plan.DirectWriteAllowed = false
	plan.RequiredActions = []CoordinationAction{
		ActionStageRestore,
		ActionPauseOrMaintenance,
		ActionValidateRestoredData,
		ActionReconcileWithSync,
	}
	if availability == SyncAvailabilityAvailable {
		plan.Reconciliation = ReconciliationRequired
	} else {
		plan.Reconciliation = ReconciliationPending
	}

	return plan, nil
}

func validateOpaqueIdentifier(name, value string) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s must be valid UTF-8", name)
	}
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must not be empty", name)
	}
	if len(value) > maxOpaqueIdentifierBytes {
		return fmt.Errorf("%s exceeds %d bytes", name, maxOpaqueIdentifierBytes)
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return fmt.Errorf("%s must not contain control characters", name)
		}
	}
	return nil
}
