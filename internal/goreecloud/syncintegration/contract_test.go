package syncintegration

import (
	"strings"
	"testing"
	"time"

	"github.com/kopia/kopia/internal/goreecloud/protection"
)

func TestAllowedOperationsStayNarrow(t *testing.T) {
	allowed := AllowedOperations()
	want := []Operation{
		OperationReadProtection,
		OperationRequestCheckpoint,
		OperationCoordinateRestore,
	}
	if len(allowed) != len(want) {
		t.Fatalf("AllowedOperations() length = %d, want %d", len(allowed), len(want))
	}
	for i := range want {
		if allowed[i] != want[i] {
			t.Fatalf("AllowedOperations()[%d] = %q, want %q", i, allowed[i], want[i])
		}
		if err := ValidateOperation(allowed[i]); err != nil {
			t.Fatalf("ValidateOperation(%q) returned error: %v", allowed[i], err)
		}
	}

	for _, forbidden := range []Operation{
		"delete_recovery_point",
		"delete_repository",
		"change_retention",
		"rotate_encryption_key",
		"run_repository_maintenance",
		"sync_folder",
	} {
		if err := ValidateOperation(forbidden); err == nil {
			t.Fatalf("ValidateOperation(%q) unexpectedly allowed authority outside the contract", forbidden)
		}
	}
}

func TestNewProtectionViewCopiesBoundedBackupState(t *testing.T) {
	evaluatedAt := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	evaluation := protection.Evaluation{
		State:   protection.StateDegraded,
		Reasons: []protection.ReasonCode{protection.ReasonRequiredEvidenceStale},
		Stale:   []protection.EvidenceKind{protection.EvidenceIntegrity},
	}

	view, err := NewProtectionView("family-documents", evaluatedAt, evaluation)
	if err != nil {
		t.Fatalf("NewProtectionView() error = %v", err)
	}
	if view.ContractVersion != ContractVersion {
		t.Fatalf("ContractVersion = %q, want %q", view.ContractVersion, ContractVersion)
	}
	if view.State != protection.StateDegraded {
		t.Fatalf("State = %q, want %q", view.State, protection.StateDegraded)
	}
	if len(view.Reasons) != 1 || view.Reasons[0] != protection.ReasonRequiredEvidenceStale {
		t.Fatalf("Reasons = %#v", view.Reasons)
	}
	if len(view.Stale) != 1 || view.Stale[0] != protection.EvidenceIntegrity {
		t.Fatalf("Stale = %#v", view.Stale)
	}

	// Mutating the original evaluation must not mutate the exported view.
	evaluation.Reasons[0] = protection.ReasonNotConfigured
	evaluation.Stale[0] = protection.EvidenceScope
	if view.Reasons[0] != protection.ReasonRequiredEvidenceStale {
		t.Fatalf("view aliases evaluation reasons")
	}
	if view.Stale[0] != protection.EvidenceIntegrity {
		t.Fatalf("view aliases evaluation stale evidence")
	}
}

func TestNewProtectionViewRejectsInvalidInput(t *testing.T) {
	now := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		name      string
		datasetID string
		at        time.Time
		state     protection.State
	}{
		{name: "empty dataset", datasetID: " ", at: now, state: protection.StateProtected},
		{name: "zero time", datasetID: "dataset-1", state: protection.StateProtected},
		{name: "invalid state", datasetID: "dataset-1", at: now, state: protection.State("green")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewProtectionView(tc.datasetID, tc.at, protection.Evaluation{State: tc.state})
			if err == nil {
				t.Fatal("NewProtectionView() unexpectedly succeeded")
			}
		})
	}
}

func TestCheckpointRequestValidationIsStructuralNotAuthorization(t *testing.T) {
	for _, purpose := range []CheckpointPurpose{CheckpointPreChange, CheckpointPreMigration} {
		req := CheckpointRequest{
			ContractVersion:          ContractVersion,
			RequestID:                "request-123",
			DatasetID:                "family-documents",
			Purpose:                  purpose,
			AuthorizationDecisionRef: "identity-decision-456",
		}
		if err := req.Validate(); err != nil {
			t.Fatalf("Validate() for purpose %q error = %v", purpose, err)
		}
	}

	invalid := []CheckpointRequest{
		{
			ContractVersion:          "v2",
			RequestID:                "request-123",
			DatasetID:                "family-documents",
			Purpose:                  CheckpointPreChange,
			AuthorizationDecisionRef: "identity-decision-456",
		},
		{
			ContractVersion:          ContractVersion,
			RequestID:                "request-123",
			DatasetID:                "family-documents",
			Purpose:                  CheckpointPurpose("routine_sync"),
			AuthorizationDecisionRef: "identity-decision-456",
		},
		{
			ContractVersion: ContractVersion,
			RequestID:       "request-123",
			DatasetID:       "family-documents",
			Purpose:         CheckpointPreChange,
		},
	}
	for i, req := range invalid {
		if err := req.Validate(); err == nil {
			t.Fatalf("invalid request %d unexpectedly validated", i)
		}
	}
}

func TestSyncManagedRestoreRequiresSafeCoordination(t *testing.T) {
	plan, err := PlanRestoreCoordination("family-documents", true, SyncAvailabilityAvailable)
	if err != nil {
		t.Fatalf("PlanRestoreCoordination() error = %v", err)
	}
	if !plan.StagingRequired {
		t.Fatal("Sync-managed restore did not require staging")
	}
	if plan.DirectWriteAllowed {
		t.Fatal("Sync-managed restore unexpectedly allowed direct write")
	}
	if plan.Reconciliation != ReconciliationRequired {
		t.Fatalf("Reconciliation = %q, want %q", plan.Reconciliation, ReconciliationRequired)
	}

	wantActions := []CoordinationAction{
		ActionStageRestore,
		ActionPauseOrMaintenance,
		ActionValidateRestoredData,
		ActionReconcileWithSync,
	}
	if len(plan.RequiredActions) != len(wantActions) {
		t.Fatalf("RequiredActions length = %d, want %d", len(plan.RequiredActions), len(wantActions))
	}
	for i := range wantActions {
		if plan.RequiredActions[i] != wantActions[i] {
			t.Fatalf("RequiredActions[%d] = %q, want %q", i, plan.RequiredActions[i], wantActions[i])
		}
	}
}

func TestBackupRecoveryDoesNotDependOnSyncAvailability(t *testing.T) {
	for _, availability := range []SyncAvailability{SyncAvailabilityUnavailable, SyncAvailabilityUnknown} {
		plan, err := PlanRestoreCoordination("family-documents", true, availability)
		if err != nil {
			t.Fatalf("PlanRestoreCoordination(%q) error = %v", availability, err)
		}
		if !plan.StagingRequired {
			t.Fatalf("availability %q did not preserve staging recovery", availability)
		}
		if plan.DirectWriteAllowed {
			t.Fatalf("availability %q unexpectedly allowed direct production write", availability)
		}
		if plan.Reconciliation != ReconciliationPending {
			t.Fatalf("availability %q reconciliation = %q, want %q", availability, plan.Reconciliation, ReconciliationPending)
		}
	}
}

func TestNonSyncManagedRestoreAddsNoSyncAuthority(t *testing.T) {
	plan, err := PlanRestoreCoordination("backup-only-dataset", false, SyncAvailabilityUnknown)
	if err != nil {
		t.Fatalf("PlanRestoreCoordination() error = %v", err)
	}
	if plan.StagingRequired {
		t.Fatal("non-Sync-managed restore unexpectedly requires Sync staging")
	}
	if !plan.DirectWriteAllowed {
		t.Fatal("non-Sync-managed restore was blocked by the Sync boundary")
	}
	if plan.Reconciliation != ReconciliationNotRequired {
		t.Fatalf("Reconciliation = %q, want %q", plan.Reconciliation, ReconciliationNotRequired)
	}
	if len(plan.RequiredActions) != 0 {
		t.Fatalf("RequiredActions = %#v, want none", plan.RequiredActions)
	}
}

func TestOpaqueIdentifiersAreBoundedAndNonControl(t *testing.T) {
	valid := CheckpointRequest{
		ContractVersion:          ContractVersion,
		RequestID:                "request-123",
		DatasetID:                strings.Repeat("a", maxOpaqueIdentifierBytes),
		Purpose:                  CheckpointPreMigration,
		AuthorizationDecisionRef: "decision-1",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("max-size identifier rejected: %v", err)
	}

	for _, datasetID := range []string{
		strings.Repeat("a", maxOpaqueIdentifierBytes+1),
		"dataset\nsecret",
	} {
		req := valid
		req.DatasetID = datasetID
		if err := req.Validate(); err == nil {
			t.Fatalf("dataset ID %q unexpectedly validated", datasetID)
		}
	}
}
