package syncintegration

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func newTestDatasetScopeStore(t *testing.T) *FileDatasetScopeStore {
	t.Helper()
	store, err := NewFileDatasetScopeStore(filepath.Join(t.TempDir(), "dataset-scopes.json"))
	if err != nil {
		t.Fatalf("NewFileDatasetScopeStore() error = %v", err)
	}
	return store
}

func TestFileDatasetScopeStoreRequiresExplicitFilePath(t *testing.T) {
	if _, err := NewFileDatasetScopeStore(""); err == nil {
		t.Fatal("NewFileDatasetScopeStore() accepted empty path")
	}
}

func TestFileDatasetScopeStoreDistinguishesUninitializedAndMissingMapping(t *testing.T) {
	store := newTestDatasetScopeStore(t)
	if _, err := store.ResolveBackupScope(context.Background(), "family-documents"); !errors.Is(err, ErrDatasetScopeStoreNotInitialized) {
		t.Fatalf("ResolveBackupScope() error = %v, want ErrDatasetScopeStoreNotInitialized", err)
	}

	if err := store.ReplaceMappings(context.Background(), nil); err != nil {
		t.Fatalf("ReplaceMappings() error = %v", err)
	}
	if _, err := store.ResolveBackupScope(context.Background(), "family-documents"); !errors.Is(err, ErrDatasetScopeMappingNotFound) {
		t.Fatalf("ResolveBackupScope() error = %v, want ErrDatasetScopeMappingNotFound", err)
	}
}

func TestFileDatasetScopeStoreRoundTrip(t *testing.T) {
	store := newTestDatasetScopeStore(t)
	mapping := validDatasetScopeMapping()
	if err := store.ReplaceMappings(context.Background(), []DatasetScopeMapping{mapping}); err != nil {
		t.Fatalf("ReplaceMappings() error = %v", err)
	}

	got, err := store.ResolveBackupScope(context.Background(), mapping.DatasetID)
	if err != nil {
		t.Fatalf("ResolveBackupScope() error = %v", err)
	}
	if got != mapping {
		t.Fatalf("ResolveBackupScope() = %#v, want %#v", got, mapping)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(store.path)
		if err != nil {
			t.Fatalf("Stat() error = %v", err)
		}
		if info.Mode().Perm()&0o077 != 0 {
			t.Fatalf("mapping store permissions = %o, want no group/other access", info.Mode().Perm())
		}
	}
}

func TestFileDatasetScopeStoreRejectsInvalidReplacementBeforePublishing(t *testing.T) {
	store := newTestDatasetScopeStore(t)
	mapping := validDatasetScopeMapping()
	if err := store.ReplaceMappings(context.Background(), []DatasetScopeMapping{mapping}); err != nil {
		t.Fatalf("initial ReplaceMappings() error = %v", err)
	}

	duplicate := mapping
	duplicate.BackupScopeID = "different-scope"
	duplicate.MappingRevision = "mapping-revision-8"
	if err := store.ReplaceMappings(context.Background(), []DatasetScopeMapping{mapping, duplicate}); err == nil {
		t.Fatal("ReplaceMappings() accepted duplicate dataset mapping")
	}

	got, err := store.ResolveBackupScope(context.Background(), mapping.DatasetID)
	if err != nil {
		t.Fatalf("ResolveBackupScope() after rejected replacement error = %v", err)
	}
	if got != mapping {
		t.Fatalf("rejected replacement changed durable mapping: got %#v want %#v", got, mapping)
	}
}

func TestFileDatasetScopeStoreRejectsMalformedOrUntrustedFileState(t *testing.T) {
	for _, tc := range []struct {
		name    string
		content string
	}{
		{name: "unknown field", content: `{"version":1,"mappings":[],"extra":true}`},
		{name: "trailing value", content: `{"version":1,"mappings":[]} {}`},
		{name: "unsupported version", content: `{"version":2,"mappings":[]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := newTestDatasetScopeStore(t)
			content := strings.ReplaceAll(tc.content, `\"`, `"`)
			if err := os.WriteFile(store.path, []byte(content), 0o600); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}
			if _, err := store.ResolveBackupScope(context.Background(), "family-documents"); err == nil {
				t.Fatal("ResolveBackupScope() accepted malformed store")
			}
		})
	}
}

func TestFileDatasetScopeStoreRejectsLooseUnixPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission semantics do not apply on Windows")
	}
	store := newTestDatasetScopeStore(t)
	payload := strings.ReplaceAll(`{"version":1,"mappings":[]}`, `\"`, `"`)
	if err := os.WriteFile(store.path, []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := store.ResolveBackupScope(context.Background(), "family-documents"); err == nil || !strings.Contains(err.Error(), "permissions") {
		t.Fatalf("ResolveBackupScope() error = %v, want permission failure", err)
	}
}

func TestFileDatasetScopeStoreRejectsNilOrCancelledContext(t *testing.T) {
	store := newTestDatasetScopeStore(t)
	if err := store.ReplaceMappings(nil, nil); err == nil {
		t.Fatal("ReplaceMappings() accepted nil context")
	}
	if _, err := store.ResolveBackupScope(nil, "family-documents"); err == nil {
		t.Fatal("ResolveBackupScope() accepted nil context")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := store.ReplaceMappings(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("ReplaceMappings() error = %v, want context.Canceled", err)
	}
	if _, err := store.ResolveBackupScope(ctx, "family-documents"); !errors.Is(err, context.Canceled) {
		t.Fatalf("ResolveBackupScope() error = %v, want context.Canceled", err)
	}
}
