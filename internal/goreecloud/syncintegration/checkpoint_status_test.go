package syncintegration

import (
	"context"
	"testing"
	"time"
)

func validCompletedCheckpointStatus() CheckpointStatus {
	return CheckpointStatus{
		ContractVersion:     ContractVersion,
		RequestID:           "request-123",
		OperationID:         "backup-operation-789",
		DatasetID:           "family-documents",
		BackupScopeID:       "backup-scope-family-documents",
		ObservedAt:          time.Date(2026, time.September, 4, 13, 15, 0, 0, time.UTC),
		State:               CheckpointStateCompleted,
		RecoveryPointID:     "recovery-point-456",
		RecoveryPointUsable: true,
		IntegrityVerified:   true,
	}
}

func TestCheckpointStatusReadyForProtectedChangeRequiresRecoveryEvidence(t *testing.T) {
	ready := validCompletedCheckpointStatus()
	if err := ready.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !ready.ReadyForProtectedChange() {
		t.Fatal("integrity-verified usable recovery point was not ready")
	}

	for _, tc := range []struct {
		name   string
		mutate func(*CheckpointStatus)
	}{
		{name: "accepted only", mutate: func(s *CheckpointStatus) {
			s.State = CheckpointStateAccepted
			s.RecoveryPointID = ""
			s.RecoveryPointUsable = false
			s.IntegrityVerified = false
		}},
		{name: "running", mutate: func(s *CheckpointStatus) {
			s.State = CheckpointStateRunning
			s.RecoveryPointID = ""
			s.RecoveryPointUsable = false
			s.IntegrityVerified = false
		}},
		{name: "completed but unusable", mutate: func(s *CheckpointStatus) {
			s.RecoveryPointUsable = false
		}},
		{name: "completed without integrity", mutate: func(s *CheckpointStatus) {
			s.IntegrityVerified = false
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status := validCompletedCheckpointStatus()
			tc.mutate(&status)
			if err := status.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if status.ReadyForProtectedChange() {
				t.Fatal("status unexpectedly allowed protected change")
			}
		})
	}
}

func TestCheckpointStatusRejectsContradictoryEvidence(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status CheckpointStatus
	}{
		{
			name: "accepted with recovery point",
			status: CheckpointStatus{
				ContractVersion: ContractVersion,
				RequestID:       "request-123",
				OperationID:     "operation-1",
				DatasetID:       "dataset-1",
				BackupScopeID:   "scope-1",
				ObservedAt:      time.Now().UTC(),
				State:           CheckpointStateAccepted,
				RecoveryPointID: "recovery-point-1",
			},
		},
		{
			name: "failed without bounded category",
			status: CheckpointStatus{
				ContractVersion: ContractVersion,
				RequestID:       "request-123",
				OperationID:     "operation-1",
				DatasetID:       "dataset-1",
				BackupScopeID:   "scope-1",
				ObservedAt:      time.Now().UTC(),
				State:           CheckpointStateFailed,
			},
		},
		{
			name: "completed without recovery point",
			status: CheckpointStatus{
				ContractVersion: ContractVersion,
				RequestID:       "request-123",
				OperationID:     "operation-1",
				DatasetID:       "dataset-1",
				BackupScopeID:   "scope-1",
				ObservedAt:      time.Now().UTC(),
				State:           CheckpointStateCompleted,
			},
		},
		{
			name: "restore verified without usable integrity-verified point",
			status: CheckpointStatus{
				ContractVersion: ContractVersion,
				RequestID:       "request-123",
				OperationID:     "operation-1",
				DatasetID:       "dataset-1",
				BackupScopeID:   "scope-1",
				ObservedAt:      time.Now().UTC(),
				State:           CheckpointStateCompleted,
				RecoveryPointID: "recovery-point-1",
				RestoreVerified: true,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.status.Validate(); err == nil {
				t.Fatal("Validate() unexpectedly succeeded")
			}
			if tc.status.ReadyForProtectedChange() {
				t.Fatal("contradictory status unexpectedly allowed protected change")
			}
		})
	}
}

type compileOnlyCheckpointStatusProvider struct{}

func (compileOnlyCheckpointStatusProvider) CheckpointStatus(context.Context, string) (CheckpointStatus, error) {
	return CheckpointStatus{}, nil
}

var _ CheckpointStatusProvider = compileOnlyCheckpointStatusProvider{}
