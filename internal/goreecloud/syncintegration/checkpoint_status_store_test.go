package syncintegration

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func newTestCheckpointStatusStore(t *testing.T) *FileCheckpointStatusStore {
	t.Helper()
	store, err := NewFileCheckpointStatusStore(filepath.Join(t.TempDir(), "checkpoint-status.json"))
	if err != nil {
		t.Fatalf("NewFileCheckpointStatusStore() error = %v", err)
	}
	return store
}

func validCheckpointSubmission() CheckpointSubmission {
	return CheckpointSubmission{
		RequestID:     "request-123",
		OperationID:   "backup-operation-789",
		DatasetID:     "family-documents",
		BackupScopeID: "backup-scope-family-documents",
		AcceptedAt:    time.Date(2026, time.September, 5, 14, 0, 0, 0, time.UTC),
	}
}

func checkpointStatusForSubmission(submission CheckpointSubmission, observedAt time.Time, state CheckpointLifecycleState) CheckpointStatus {
	return CheckpointStatus{
		ContractVersion: ContractVersion,
		RequestID:       submission.RequestID,
		OperationID:     submission.OperationID,
		DatasetID:       submission.DatasetID,
		BackupScopeID:   submission.BackupScopeID,
		ObservedAt:      observedAt,
		State:           state,
	}
}

func TestFileCheckpointStatusStoreRequiresExplicitFilePath(t *testing.T) {
	if _, err := NewFileCheckpointStatusStore(""); err == nil {
		t.Fatal("NewFileCheckpointStatusStore() accepted empty path")
	}
}

func TestFileCheckpointStatusStoreDistinguishesUninitializedAndMissingOperation(t *testing.T) {
	store := newTestCheckpointStatusStore(t)
	if _, err := store.CheckpointStatus(context.Background(), "operation-1"); !errors.Is(err, ErrCheckpointStatusStoreNotInitialized) {
		t.Fatalf("CheckpointStatus() error = %v, want ErrCheckpointStatusStoreNotInitialized", err)
	}

	submission := validCheckpointSubmission()
	if err := store.RecordSubmission(context.Background(), submission); err != nil {
		t.Fatalf("RecordSubmission() error = %v", err)
	}
	if _, err := store.CheckpointStatus(context.Background(), "operation-other"); !errors.Is(err, ErrCheckpointStatusNotFound) {
		t.Fatalf("CheckpointStatus() error = %v, want ErrCheckpointStatusNotFound", err)
	}
}

func TestFileCheckpointStatusStoreRecordsAcceptedSubmissionAndStrengthenedRecoveryEvidence(t *testing.T) {
	store := newTestCheckpointStatusStore(t)
	submission := validCheckpointSubmission()
	if err := store.RecordSubmission(context.Background(), submission); err != nil {
		t.Fatalf("RecordSubmission() error = %v", err)
	}
	if err := store.RecordSubmission(context.Background(), submission); err != nil {
		t.Fatalf("idempotent RecordSubmission() error = %v", err)
	}

	accepted, err := store.CheckpointStatus(context.Background(), submission.OperationID)
	if err != nil {
		t.Fatalf("CheckpointStatus() error = %v", err)
	}
	if accepted.State != CheckpointStateAccepted {
		t.Fatalf("initial state = %q, want accepted", accepted.State)
	}
	if accepted.ReadyForSubmission(submission) {
		t.Fatal("accepted submission unexpectedly satisfied protected-change readiness")
	}

	running := checkpointStatusForSubmission(submission, submission.AcceptedAt.Add(time.Minute), CheckpointStateRunning)
	if err := store.RecordStatus(context.Background(), running); err != nil {
		t.Fatalf("RecordStatus(running) error = %v", err)
	}

	completed := checkpointStatusForSubmission(submission, submission.AcceptedAt.Add(2*time.Minute), CheckpointStateCompleted)
	completed.RecoveryPointID = "recovery-point-456"
	if err := store.RecordStatus(context.Background(), completed); err != nil {
		t.Fatalf("RecordStatus(completed) error = %v", err)
	}
	if completed.ReadyForSubmission(submission) {
		t.Fatal("unverified completed status unexpectedly satisfied protected-change readiness")
	}

	verified := completed
	verified.ObservedAt = submission.AcceptedAt.Add(3 * time.Minute)
	verified.RecoveryPointUsable = true
	verified.IntegrityVerified = true
	if err := store.RecordStatus(context.Background(), verified); err != nil {
		t.Fatalf("RecordStatus(verified) error = %v", err)
	}
	if err := store.RecordStatus(context.Background(), verified); err != nil {
		t.Fatalf("idempotent RecordStatus() error = %v", err)
	}

	latest, err := store.CheckpointStatus(context.Background(), submission.OperationID)
	if err != nil {
		t.Fatalf("CheckpointStatus() error = %v", err)
	}
	if !sameCheckpointStatus(latest, verified) {
		t.Fatalf("latest status = %#v, want %#v", latest, verified)
	}
	if !latest.ReadyForSubmission(submission) {
		t.Fatal("usable integrity-verified recovery point did not satisfy readiness")
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(store.path)
		if err != nil {
			t.Fatalf("Stat() error = %v", err)
		}
		if info.Mode().Perm()&0o077 != 0 {
			t.Fatalf("status store permissions = %o, want no group/other access", info.Mode().Perm())
		}
	}
}

func TestFileCheckpointStatusStoreRejectsRebindingStaleObservationsAndLifecycleRegression(t *testing.T) {
	store := newTestCheckpointStatusStore(t)
	submission := validCheckpointSubmission()
	if err := store.RecordSubmission(context.Background(), submission); err != nil {
		t.Fatalf("RecordSubmission() error = %v", err)
	}

	completed := checkpointStatusForSubmission(submission, submission.AcceptedAt.Add(2*time.Minute), CheckpointStateCompleted)
	completed.RecoveryPointID = "recovery-point-456"
	completed.RecoveryPointUsable = true
	completed.IntegrityVerified = true
	if err := store.RecordStatus(context.Background(), completed); err != nil {
		t.Fatalf("RecordStatus(completed) error = %v", err)
	}

	rebound := completed
	rebound.ObservedAt = submission.AcceptedAt.Add(3 * time.Minute)
	rebound.DatasetID = "different-dataset"
	if err := store.RecordStatus(context.Background(), rebound); err == nil {
		t.Fatal("RecordStatus() accepted status rebound to another dataset")
	}

	stale := completed
	stale.ObservedAt = submission.AcceptedAt.Add(time.Minute)
	if err := store.RecordStatus(context.Background(), stale); err == nil {
		t.Fatal("RecordStatus() accepted stale observation")
	}

	regressed := checkpointStatusForSubmission(submission, submission.AcceptedAt.Add(4*time.Minute), CheckpointStateRunning)
	if err := store.RecordStatus(context.Background(), regressed); err == nil {
		t.Fatal("RecordStatus() accepted completed-to-running regression")
	}

	weakened := completed
	weakened.ObservedAt = submission.AcceptedAt.Add(5 * time.Minute)
	weakened.IntegrityVerified = false
	if err := store.RecordStatus(context.Background(), weakened); err == nil {
		t.Fatal("RecordStatus() accepted weakened completed evidence")
	}

	latest, err := store.CheckpointStatus(context.Background(), submission.OperationID)
	if err != nil {
		t.Fatalf("CheckpointStatus() error = %v", err)
	}
	if !sameCheckpointStatus(latest, completed) {
		t.Fatalf("rejected update changed latest status: got %#v want %#v", latest, completed)
	}
}

func TestFileCheckpointStatusStoreAllowsCompletedCheckpointToBecomeExplicitlyFailed(t *testing.T) {
	store := newTestCheckpointStatusStore(t)
	submission := validCheckpointSubmission()
	if err := store.RecordSubmission(context.Background(), submission); err != nil {
		t.Fatalf("RecordSubmission() error = %v", err)
	}

	completed := checkpointStatusForSubmission(submission, submission.AcceptedAt.Add(time.Minute), CheckpointStateCompleted)
	completed.RecoveryPointID = "recovery-point-456"
	if err := store.RecordStatus(context.Background(), completed); err != nil {
		t.Fatalf("RecordStatus(completed) error = %v", err)
	}

	failed := checkpointStatusForSubmission(submission, submission.AcceptedAt.Add(2*time.Minute), CheckpointStateFailed)
	failed.FailureCategory = CheckpointFailureVerification
	if err := store.RecordStatus(context.Background(), failed); err != nil {
		t.Fatalf("RecordStatus(failed) error = %v", err)
	}
	if failed.ReadyForSubmission(submission) {
		t.Fatal("failed verification unexpectedly satisfied readiness")
	}

	retried := completed
	retried.ObservedAt = submission.AcceptedAt.Add(3 * time.Minute)
	if err := store.RecordStatus(context.Background(), retried); err == nil {
		t.Fatal("RecordStatus() accepted a new completed state after terminal failure")
	}
}

func TestFileCheckpointStatusStoreRejectsConflictingSubmission(t *testing.T) {
	store := newTestCheckpointStatusStore(t)
	submission := validCheckpointSubmission()
	if err := store.RecordSubmission(context.Background(), submission); err != nil {
		t.Fatalf("RecordSubmission() error = %v", err)
	}

	conflict := submission
	conflict.DatasetID = "different-dataset"
	if err := store.RecordSubmission(context.Background(), conflict); err == nil {
		t.Fatal("RecordSubmission() accepted conflicting operation binding")
	}
}

func TestFileCheckpointStatusStoreRejectsMalformedOrUntrustedFileState(t *testing.T) {
	for _, tc := range []struct {
		name    string
		content string
	}{
		{name: "unknown field", content: "{\"version\":1,\"records\":[],\"extra\":true}"},
		{name: "trailing value", content: "{\"version\":1,\"records\":[]} {}"},
		{name: "unsupported version", content: "{\"version\":2,\"records\":[]}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestCheckpointStatusStore(t)
			if err := os.WriteFile(store.path, []byte(tc.content), 0o600); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}
			if _, err := store.CheckpointStatus(context.Background(), "operation-1"); err == nil {
				t.Fatal("CheckpointStatus() accepted malformed store")
			}
		})
	}
}

func TestFileCheckpointStatusStoreRejectsLooseUnixPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission semantics do not apply on Windows")
	}
	store := newTestCheckpointStatusStore(t)
	payload := "{\"version\":1,\"records\":[]}"
	if err := os.WriteFile(store.path, []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := store.CheckpointStatus(context.Background(), "operation-1"); err == nil || !strings.Contains(err.Error(), "permissions") {
		t.Fatalf("CheckpointStatus() error = %v, want permission failure", err)
	}
}

func TestFileCheckpointStatusStoreRejectsNilOrCancelledContext(t *testing.T) {
	store := newTestCheckpointStatusStore(t)
	submission := validCheckpointSubmission()
	if err := store.RecordSubmission(nil, submission); err == nil {
		t.Fatal("RecordSubmission() accepted nil context")
	}
	if err := store.RecordStatus(nil, checkpointStatusForSubmission(submission, submission.AcceptedAt, CheckpointStateAccepted)); err == nil {
		t.Fatal("RecordStatus() accepted nil context")
	}
	if _, err := store.CheckpointStatus(nil, submission.OperationID); err == nil {
		t.Fatal("CheckpointStatus() accepted nil context")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := store.RecordSubmission(ctx, submission); !errors.Is(err, context.Canceled) {
		t.Fatalf("RecordSubmission() error = %v, want context.Canceled", err)
	}
	if _, err := store.CheckpointStatus(ctx, submission.OperationID); !errors.Is(err, context.Canceled) {
		t.Fatalf("CheckpointStatus() error = %v, want context.Canceled", err)
	}
}
