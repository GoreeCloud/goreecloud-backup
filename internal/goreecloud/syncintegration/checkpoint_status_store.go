package syncintegration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	checkpointStatusStoreVersion = 1
	maxCheckpointStatusRecords   = 4096
	maxCheckpointStatusHistory   = 32
)

var (
	// ErrCheckpointStatusStoreNotInitialized distinguishes a configured store
	// path that has never been initialized from an initialized store that does
	// not contain the requested operation.
	ErrCheckpointStatusStoreNotInitialized = errors.New("Backup checkpoint status store is not initialized")

	// ErrCheckpointStatusNotFound indicates that no Backup-owned checkpoint
	// submission has been recorded for the requested operation ID.
	ErrCheckpointStatusNotFound = errors.New("Backup checkpoint status not found")
)

type checkpointStatusStoreFile struct {
	Version int                      `json:"version"`
	Records []checkpointStatusRecord `json:"records"`
}

type checkpointStatusRecord struct {
	Submission CheckpointSubmission `json:"submission"`
	History    []CheckpointStatus   `json:"history"`
}

// FileCheckpointStatusStore is a durable Backup-owned checkpoint lifecycle and
// recovery-evidence store. The caller supplies the path; this package does not
// choose a production location or grant GoreeCloud Sync authority to mutate
// evidence.
//
// Each operation is seeded from one accepted CheckpointSubmission and retains a
// bounded chronological status history. Writes publish atomically through a
// private temporary file. Reads fail closed on malformed JSON, unknown fields,
// unsupported versions, duplicate operation or request IDs, invalid lifecycle
// histories, non-regular files, and loose Unix permissions.
type FileCheckpointStatusStore struct {
	path string
	mu   sync.RWMutex
}

// NewFileCheckpointStatusStore creates a store handle for an explicit file
// path. It does not create the file or its parent directory.
func NewFileCheckpointStatusStore(path string) (*FileCheckpointStatusStore, error) {
	if path == "" {
		return nil, fmt.Errorf("checkpoint status store path must not be empty")
	}
	clean := filepath.Clean(path)
	if clean == "." || filepath.Base(clean) == "." || filepath.Base(clean) == string(filepath.Separator) {
		return nil, fmt.Errorf("checkpoint status store path must identify a file")
	}
	return &FileCheckpointStatusStore{path: clean}, nil
}

// RecordSubmission initializes one operation with Backup's accepted receipt.
// Exact replays are idempotent. A different submission cannot replace an
// existing operation because the request, dataset, scope, and acceptance time
// are part of the immutable correlation boundary.
func (s *FileCheckpointStatusStore) RecordSubmission(ctx context.Context, submission CheckpointSubmission) error {
	if s == nil || s.path == "" {
		return fmt.Errorf("checkpoint status store is not initialized")
	}
	if ctx == nil {
		return fmt.Errorf("context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	submission.AcceptedAt = submission.AcceptedAt.UTC()
	accepted, err := acceptedStatusForSubmission(submission)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}
	records, err := s.loadRecords()
	if err != nil && !errors.Is(err, ErrCheckpointStatusStoreNotInitialized) {
		return err
	}
	if errors.Is(err, ErrCheckpointStatusStoreNotInitialized) {
		records = nil
	}

	for _, record := range records {
		if record.Submission.OperationID != submission.OperationID {
			continue
		}
		if sameCheckpointSubmission(record.Submission, submission) {
			return nil
		}
		return fmt.Errorf("checkpoint operation %q already has a different submission", submission.OperationID)
	}
	if len(records) >= maxCheckpointStatusRecords {
		return fmt.Errorf("checkpoint status record count exceeds %d", maxCheckpointStatusRecords)
	}

	records = append(records, checkpointStatusRecord{
		Submission: submission,
		History:    []CheckpointStatus{accepted},
	})
	return s.writeRecords(records)
}

// RecordStatus appends one authoritative Backup-owned observation for an
// existing checkpoint operation. Exact replays of the latest observation are
// idempotent; all other observations must advance time, remain bound to the
// original submission, and follow the allowed lifecycle progression.
func (s *FileCheckpointStatusStore) RecordStatus(ctx context.Context, status CheckpointStatus) error {
	if s == nil || s.path == "" {
		return fmt.Errorf("checkpoint status store is not initialized")
	}
	if ctx == nil {
		return fmt.Errorf("context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	status.ObservedAt = status.ObservedAt.UTC()
	if err := status.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}
	records, err := s.loadRecords()
	if err != nil {
		return err
	}
	for i := range records {
		record := &records[i]
		if record.Submission.OperationID != status.OperationID {
			continue
		}
		if err := status.ValidateForSubmission(record.Submission); err != nil {
			return fmt.Errorf("checkpoint status does not match submission: %w", err)
		}
		last := record.History[len(record.History)-1]
		if sameCheckpointStatus(last, status) {
			return nil
		}
		if !status.ObservedAt.After(last.ObservedAt) {
			return fmt.Errorf("checkpoint status observation time must advance")
		}
		if err := validateCheckpointStatusTransition(last, status); err != nil {
			return err
		}
		if len(record.History) >= maxCheckpointStatusHistory {
			return fmt.Errorf("checkpoint status history exceeds %d observations", maxCheckpointStatusHistory)
		}
		record.History = append(record.History, status)
		return s.writeRecords(records)
	}
	return ErrCheckpointStatusNotFound
}

// CheckpointStatus implements CheckpointStatusProvider and returns only the
// latest validated Backup-owned observation for one operation. The persisted
// history remains available to Backup for future audit/evidence adapters; Sync
// receives no mutation authority through this provider.
func (s *FileCheckpointStatusStore) CheckpointStatus(ctx context.Context, operationID string) (CheckpointStatus, error) {
	if s == nil || s.path == "" {
		return CheckpointStatus{}, fmt.Errorf("checkpoint status store is not initialized")
	}
	if ctx == nil {
		return CheckpointStatus{}, fmt.Errorf("context is required")
	}
	if err := ctx.Err(); err != nil {
		return CheckpointStatus{}, err
	}
	if err := validateOpaqueIdentifier("checkpoint operation ID", operationID); err != nil {
		return CheckpointStatus{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	records, err := s.loadRecords()
	if err != nil {
		return CheckpointStatus{}, err
	}
	for _, record := range records {
		if record.Submission.OperationID == operationID {
			return record.History[len(record.History)-1], nil
		}
	}
	return CheckpointStatus{}, ErrCheckpointStatusNotFound
}

func acceptedStatusForSubmission(submission CheckpointSubmission) (CheckpointStatus, error) {
	accepted := CheckpointStatus{
		ContractVersion: ContractVersion,
		RequestID:       submission.RequestID,
		OperationID:     submission.OperationID,
		DatasetID:       submission.DatasetID,
		BackupScopeID:   submission.BackupScopeID,
		ObservedAt:      submission.AcceptedAt.UTC(),
		State:           CheckpointStateAccepted,
	}
	if err := accepted.ValidateForSubmission(submission); err != nil {
		return CheckpointStatus{}, fmt.Errorf("invalid checkpoint submission: %w", err)
	}
	return accepted, nil
}

func validateCheckpointStatusTransition(previous, next CheckpointStatus) error {
	switch previous.State {
	case CheckpointStateAccepted:
		switch next.State {
		case CheckpointStateRunning, CheckpointStateFailed, CheckpointStateCompleted:
			return nil
		}
	case CheckpointStateRunning:
		switch next.State {
		case CheckpointStateFailed, CheckpointStateCompleted:
			return nil
		}
	case CheckpointStateCompleted:
		switch next.State {
		case CheckpointStateFailed:
			return nil
		case CheckpointStateCompleted:
			if previous.RecoveryPointID != next.RecoveryPointID {
				return fmt.Errorf("completed checkpoint recovery point ID must not change")
			}
			if previous.RecoveryPointUsable && !next.RecoveryPointUsable {
				return fmt.Errorf("completed checkpoint usable evidence must not regress")
			}
			if previous.IntegrityVerified && !next.IntegrityVerified {
				return fmt.Errorf("completed checkpoint integrity evidence must not regress")
			}
			if previous.RestoreVerified && !next.RestoreVerified {
				return fmt.Errorf("completed checkpoint restore evidence must not regress")
			}
			return nil
		}
	case CheckpointStateFailed:
		// A failed operation is terminal. A retry must receive a new Backup
		// operation ID so evidence from distinct executions cannot be merged.
	}
	return fmt.Errorf("checkpoint lifecycle transition %q -> %q is not permitted", previous.State, next.State)
}

func sameCheckpointSubmission(a, b CheckpointSubmission) bool {
	return a.RequestID == b.RequestID &&
		a.OperationID == b.OperationID &&
		a.DatasetID == b.DatasetID &&
		a.BackupScopeID == b.BackupScopeID &&
		a.AcceptedAt.Equal(b.AcceptedAt)
}

func sameCheckpointStatus(a, b CheckpointStatus) bool {
	return a.ContractVersion == b.ContractVersion &&
		a.RequestID == b.RequestID &&
		a.OperationID == b.OperationID &&
		a.DatasetID == b.DatasetID &&
		a.BackupScopeID == b.BackupScopeID &&
		a.ObservedAt.Equal(b.ObservedAt) &&
		a.State == b.State &&
		a.RecoveryPointID == b.RecoveryPointID &&
		a.RecoveryPointUsable == b.RecoveryPointUsable &&
		a.IntegrityVerified == b.IntegrityVerified &&
		a.RestoreVerified == b.RestoreVerified &&
		a.FailureCategory == b.FailureCategory
}

func (s *FileCheckpointStatusStore) loadRecords() ([]checkpointStatusRecord, error) {
	info, err := os.Lstat(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrCheckpointStatusStoreNotInitialized
		}
		return nil, fmt.Errorf("stat checkpoint status store: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("checkpoint status store must be a regular file")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("checkpoint status store permissions must not grant group or other access")
	}

	payload, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("read checkpoint status store: %w", err)
	}
	var stored checkpointStatusStoreFile
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&stored); err != nil {
		return nil, fmt.Errorf("decode checkpoint status store: %w", err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return nil, fmt.Errorf("decode checkpoint status store: %w", err)
	}
	if stored.Version != checkpointStatusStoreVersion {
		return nil, fmt.Errorf("unsupported checkpoint status store version %d", stored.Version)
	}
	if err := validateCheckpointStatusRecords(stored.Records); err != nil {
		return nil, fmt.Errorf("invalid checkpoint status store: %w", err)
	}
	return cloneCheckpointStatusRecords(stored.Records), nil
}

func (s *FileCheckpointStatusStore) writeRecords(records []checkpointStatusRecord) error {
	if err := validateCheckpointStatusRecords(records); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(checkpointStatusStoreFile{
		Version: checkpointStatusStoreVersion,
		Records: records,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode checkpoint status store: %w", err)
	}
	payload = append(payload, '\n')
	return writePrivateAtomicFile(s.path, payload)
}

func validateCheckpointStatusRecords(records []checkpointStatusRecord) error {
	if len(records) > maxCheckpointStatusRecords {
		return fmt.Errorf("checkpoint status record count exceeds %d", maxCheckpointStatusRecords)
	}
	seenOperations := make(map[string]struct{}, len(records))
	seenRequests := make(map[string]struct{}, len(records))
	for i, record := range records {
		accepted, err := acceptedStatusForSubmission(record.Submission)
		if err != nil {
			return fmt.Errorf("record %d submission: %w", i, err)
		}
		if _, exists := seenOperations[record.Submission.OperationID]; exists {
			return fmt.Errorf("duplicate checkpoint operation ID %q", record.Submission.OperationID)
		}
		seenOperations[record.Submission.OperationID] = struct{}{}
		if _, exists := seenRequests[record.Submission.RequestID]; exists {
			return fmt.Errorf("duplicate checkpoint request ID %q", record.Submission.RequestID)
		}
		seenRequests[record.Submission.RequestID] = struct{}{}
		if len(record.History) == 0 {
			return fmt.Errorf("record %d checkpoint history must not be empty", i)
		}
		if len(record.History) > maxCheckpointStatusHistory {
			return fmt.Errorf("record %d checkpoint status history exceeds %d observations", i, maxCheckpointStatusHistory)
		}
		if !sameCheckpointStatus(record.History[0], accepted) {
			return fmt.Errorf("record %d first checkpoint status must be the accepted submission", i)
		}
		for j := range record.History {
			status := record.History[j]
			if err := status.ValidateForSubmission(record.Submission); err != nil {
				return fmt.Errorf("record %d status %d: %w", i, j, err)
			}
			if j == 0 {
				continue
			}
			previous := record.History[j-1]
			if !status.ObservedAt.After(previous.ObservedAt) {
				return fmt.Errorf("record %d status %d observation time must advance", i, j)
			}
			if err := validateCheckpointStatusTransition(previous, status); err != nil {
				return fmt.Errorf("record %d status %d: %w", i, j, err)
			}
		}
	}
	return nil
}

func cloneCheckpointStatusRecords(records []checkpointStatusRecord) []checkpointStatusRecord {
	cloned := make([]checkpointStatusRecord, len(records))
	for i, record := range records {
		cloned[i] = checkpointStatusRecord{
			Submission: record.Submission,
			History:    append([]CheckpointStatus(nil), record.History...),
		}
	}
	return cloned
}

var _ CheckpointStatusProvider = (*FileCheckpointStatusStore)(nil)
