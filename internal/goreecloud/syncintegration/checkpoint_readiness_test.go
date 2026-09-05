package syncintegration

import (
	"testing"
	"time"
)

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

func TestReadyForSubmissionWithinFailsClosedOnStaleOrInvalidFreshness(t *testing.T) {
	submission := validCheckpointSubmission()
	status := validCompletedCheckpointStatus()
	now := status.ObservedAt.Add(30 * time.Minute)

	if !status.ReadyForSubmissionWithin(submission, now, 30*time.Minute) {
		t.Fatal("checkpoint at the exact policy freshness boundary was rejected")
	}
	if status.ReadyForSubmissionWithin(submission, now.Add(time.Nanosecond), 30*time.Minute) {
		t.Fatal("stale checkpoint was accepted past the policy freshness boundary")
	}
	if status.ReadyForSubmissionWithin(submission, status.ObservedAt.Add(-time.Second), 30*time.Minute) {
		t.Fatal("future-observed checkpoint was accepted")
	}
	if status.ReadyForSubmissionWithin(submission, now, 0) {
		t.Fatal("zero max age was accepted")
	}
	if status.ReadyForSubmissionWithin(submission, time.Time{}, 30*time.Minute) {
		t.Fatal("zero evaluation time was accepted")
	}

	unverified := status
	unverified.IntegrityVerified = false
	if unverified.ReadyForSubmissionWithin(submission, now, time.Hour) {
		t.Fatal("fresh but unverified checkpoint was accepted")
	}
}
