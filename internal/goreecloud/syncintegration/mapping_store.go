package syncintegration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	datasetScopeStoreVersion = 1
	maxDatasetScopeMappings  = 4096
)

// ErrDatasetScopeStoreNotInitialized indicates that the configured durable
// mapping file does not yet exist. This is kept distinct from an initialized
// store that simply has no mapping for the requested dataset.
var ErrDatasetScopeStoreNotInitialized = errors.New("Backup dataset-scope mapping store is not initialized")

type datasetScopeStoreFile struct {
	Version  int                   `json:"version"`
	Mappings []DatasetScopeMapping `json:"mappings"`
}

// FileDatasetScopeStore is a small durable Backup-owned implementation of
// DatasetScopeResolver. The caller supplies the storage path; this package does
// not select a production location or grant Sync authority to change mappings.
//
// Replacement writes validate the complete new mapping set before touching the
// existing file, then publish through a private temporary file and atomic rename
// in the same directory. Reads are strict and fail closed on unknown fields,
// trailing data, unsupported versions, invalid mappings, duplicate dataset IDs,
// non-regular files, and loose Unix permissions.
type FileDatasetScopeStore struct {
	path string
	mu   sync.RWMutex
}

// NewFileDatasetScopeStore creates a store handle for an explicit file path. It
// does not create the file or its parent directory.
func NewFileDatasetScopeStore(path string) (*FileDatasetScopeStore, error) {
	if path == "" {
		return nil, fmt.Errorf("dataset-scope mapping store path must not be empty")
	}
	clean := filepath.Clean(path)
	if clean == "." || filepath.Base(clean) == "." || filepath.Base(clean) == string(filepath.Separator) {
		return nil, fmt.Errorf("dataset-scope mapping store path must identify a file")
	}
	return &FileDatasetScopeStore{path: clean}, nil
}

// ReplaceMappings atomically replaces the complete durable mapping set after
// validating it. This is a Backup administration primitive; it is intentionally
// not part of the Backup-to-Sync Operation allowlist.
func (s *FileDatasetScopeStore) ReplaceMappings(ctx context.Context, mappings []DatasetScopeMapping) error {
	if s == nil || s.path == "" {
		return fmt.Errorf("dataset-scope mapping store is not initialized")
	}
	if ctx == nil {
		return fmt.Errorf("context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateDatasetScopeMappings(mappings); err != nil {
		return err
	}

	payload, err := json.MarshalIndent(datasetScopeStoreFile{
		Version:  datasetScopeStoreVersion,
		Mappings: mappings,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode dataset-scope mappings: %w", err)
	}
	payload = append(payload, '\n')

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}
	return writePrivateAtomicFile(s.path, payload)
}

// ResolveBackupScope resolves one Sync dataset through the current durable
// Backup-owned mapping state. It returns inactive mappings too; CheckpointService
// separately rejects inactive mappings before execution so administrative state
// remains distinguishable from an absent mapping.
func (s *FileDatasetScopeStore) ResolveBackupScope(ctx context.Context, datasetID string) (DatasetScopeMapping, error) {
	if s == nil || s.path == "" {
		return DatasetScopeMapping{}, fmt.Errorf("dataset-scope mapping store is not initialized")
	}
	if ctx == nil {
		return DatasetScopeMapping{}, fmt.Errorf("context is required")
	}
	if err := ctx.Err(); err != nil {
		return DatasetScopeMapping{}, err
	}
	if err := validateOpaqueIdentifier("dataset ID", datasetID); err != nil {
		return DatasetScopeMapping{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	mappings, err := s.loadMappings()
	if err != nil {
		return DatasetScopeMapping{}, err
	}
	for _, mapping := range mappings {
		if mapping.DatasetID == datasetID {
			return mapping, nil
		}
	}
	return DatasetScopeMapping{}, ErrDatasetScopeMappingNotFound
}

func (s *FileDatasetScopeStore) loadMappings() ([]DatasetScopeMapping, error) {
	info, err := os.Lstat(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrDatasetScopeStoreNotInitialized
		}
		return nil, fmt.Errorf("stat dataset-scope mapping store: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("dataset-scope mapping store must be a regular file")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("dataset-scope mapping store permissions must not grant group or other access")
	}

	payload, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("read dataset-scope mapping store: %w", err)
	}
	var stored datasetScopeStoreFile
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&stored); err != nil {
		return nil, fmt.Errorf("decode dataset-scope mapping store: %w", err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return nil, fmt.Errorf("decode dataset-scope mapping store: %w", err)
	}
	if stored.Version != datasetScopeStoreVersion {
		return nil, fmt.Errorf("unsupported dataset-scope mapping store version %d", stored.Version)
	}
	if err := validateDatasetScopeMappings(stored.Mappings); err != nil {
		return nil, fmt.Errorf("invalid dataset-scope mapping store: %w", err)
	}
	return append([]DatasetScopeMapping(nil), stored.Mappings...), nil
}

func validateDatasetScopeMappings(mappings []DatasetScopeMapping) error {
	if len(mappings) > maxDatasetScopeMappings {
		return fmt.Errorf("dataset-scope mapping count exceeds %d", maxDatasetScopeMappings)
	}
	seenDatasets := make(map[string]struct{}, len(mappings))
	for i, mapping := range mappings {
		if err := mapping.Validate(); err != nil {
			return fmt.Errorf("mapping %d: %w", i, err)
		}
		if _, exists := seenDatasets[mapping.DatasetID]; exists {
			return fmt.Errorf("duplicate dataset-scope mapping for dataset %q", mapping.DatasetID)
		}
		seenDatasets[mapping.DatasetID] = struct{}{}
	}
	return nil
}

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("trailing JSON value is not permitted")
		}
		return err
	}
	return nil
}

func writePrivateAtomicFile(path string, payload []byte) error {
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, ".goreecloud-backup-sync-mapping-*")
	if err != nil {
		return fmt.Errorf("create temporary dataset-scope mapping file: %w", err)
	}
	tempPath := temp.Name()
	cleanup := func() {
		_ = temp.Close()
		_ = os.Remove(tempPath)
	}

	if err := temp.Chmod(0o600); err != nil {
		cleanup()
		return fmt.Errorf("protect temporary dataset-scope mapping file: %w", err)
	}
	if _, err := temp.Write(payload); err != nil {
		cleanup()
		return fmt.Errorf("write temporary dataset-scope mapping file: %w", err)
	}
	if err := temp.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync temporary dataset-scope mapping file: %w", err)
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("close temporary dataset-scope mapping file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("publish dataset-scope mapping file: %w", err)
	}
	return nil
}

var _ DatasetScopeResolver = (*FileDatasetScopeStore)(nil)
