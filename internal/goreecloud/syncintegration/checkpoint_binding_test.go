package syncintegration

import (
	"testing"
	"time"
)

func TestCheckpointStatusBindsToExactSubmission(t *testing.T) {
	submission := validCheckpointSubmission()
	status := validCompletedCheckpointStatus()
	if err := status.ValidateForSubmission(submission); err != nil {
		t.Fatalf("ValidateForSubmission() error = %v", err)
	}

	for _, tc := range []struct {
		name   string
		mutate func(*CheckpointStatus)
	}{
		{name: "request mismatch", mutate: func(s *CheckpointStatus) { s.RequestID = "other-request" }},
		{name: "operation mismatch", mutate: func(s *CheckpointStatus) { s.OperationID = "other-operation" }},
		{name: "dataset mismatch", mutate: func(s *CheckpointStatus) { s.DatasetID = "other-dataset" }},
		{name: "scope mismatch", mutate: func(s *CheckpointStatus) { s.BackupScopeID = "other-scope" }},
		{name: "status predates acceptance", mutate: func(s *CheckpointStatus) { s.ObservedAt = submission.AcceptedAt.Add(-time.Second) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := validCompletedCheckpointStatus()
			tc.mutate(&candidate)
			if err := candidate.ValidateForSubmission(submission); err == nil {
				t.Fatal("ValidateForSubmission() unexpectedly succeeded")
			}
		})
	}
}

func TestSubmissionDoesNotBecomeRecoveryEvidenceByItself(t *testing.T) {
	submission := validCheckpointSubmission()
	accepted := CheckpointStatus{
		ContractVersion: ContractVersion,
		RequestID:       submission.RequestID,
		OperationID:     submission.OperationID,
		DatasetID:       submission.DatasetID,
		BackupScopeID:   submission.BackupScopeID,
		ObservedAt:      submission.AcceptedAt,
		State:           CheckpointStateAccepted,
	}
	if err := accepted.ValidateForSubmission(submission); err != nil {
		t.Fatalf("ValidateForSubmission() error = %v", err)
	}
	if accepted.ReadyForProtectedChange() {
		t.Fatal("accepted submission was incorrectly treated as recovery-ready")
	}
}
