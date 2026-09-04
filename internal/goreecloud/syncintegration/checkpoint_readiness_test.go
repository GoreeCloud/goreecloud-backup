package syncintegration

import "testing"

func TestReadyForSubmissionRequiresExactBindingAndRecoveryEvidence(t *testing.T) {
	submission := validCheckpointSubmission()
	status := validCompletedCheckpointStatus()

	if !status.ReadyForSubmission(submission) {
		t.Fatal("valid completed checkpoint was not ready for its exact submission")
	}

	mismatched := submission
	mismatched.OperationID = "different-operation"
	if status.ReadyForSubmission(mismatched) {
		t.Fatal("status was ready for a different submission")
	}

	unverified := status
	unverified.IntegrityVerified = false
	if unverified.ReadyForSubmission(submission) {
		t.Fatal("unverified recovery point was ready for protected change")
	}
}
