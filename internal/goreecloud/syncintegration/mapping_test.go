package syncintegration

import (
	"context"
	"errors"
	"testing"
	"time"
)

func validDatasetScopeMapping() DatasetScopeMapping {
	return DatasetScopeMapping{
		ContractVersion: ContractVersion,
		DatasetID:       "family-documents",
		BackupScopeID:   "backup-scope-family-documents",
		MappingRevision: "mapping-revision-7",
		Active:          true,
		UpdatedAt:       time.Date(2026, time.September, 4, 13, 0, 0, 0, time.UTC),
	}
}

func TestDatasetScopeMappingValidation(t *testing.T) {
	mapping := validDatasetScopeMapping()
	if err := mapping.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if err := mapping.validateForDataset(mapping.DatasetID); err != nil {
		t.Fatalf("validateForDataset() error = %v", err)
	}
}

func TestDatasetScopeMappingRejectsInvalidOrMismatchedState(t *testing.T) {
	valid := validDatasetScopeMapping()

	for _, tc := range []struct {
		name    string
		mapping DatasetScopeMapping
		dataset string
		wantErr error
	}{
		{
			name: "unsupported version",
			mapping: func() DatasetScopeMapping {
				m := valid
				m.ContractVersion = "goreecloud.backup-sync/v2"
				return m
			}(),
			dataset: valid.DatasetID,
		},
		{
			name: "empty backup scope",
			mapping: func() DatasetScopeMapping {
				m := valid
				m.BackupScopeID = ""
				return m
			}(),
			dataset: valid.DatasetID,
		},
		{
			name: "zero update time",
			mapping: func() DatasetScopeMapping {
				m := valid
				m.UpdatedAt = time.Time{}
				return m
			}(),
			dataset: valid.DatasetID,
		},
		{
			name:    "mismatched dataset",
			mapping: valid,
			dataset: "different-dataset",
		},
		{
			name: "inactive mapping",
			mapping: func() DatasetScopeMapping {
				m := valid
				m.Active = false
				return m
			}(),
			dataset: valid.DatasetID,
			wantErr: ErrDatasetScopeMappingInactive,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.mapping.validateForDataset(tc.dataset)
			if err == nil {
				t.Fatal("validateForDataset() unexpectedly succeeded")
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("validateForDataset() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// Compile-time contract check: runtime resolvers receive only the opaque Sync
// dataset ID and return Backup-owned mapping state.
type compileOnlyDatasetScopeResolver struct{}

func (compileOnlyDatasetScopeResolver) ResolveBackupScope(context.Context, string) (DatasetScopeMapping, error) {
	return DatasetScopeMapping{}, ErrDatasetScopeMappingNotFound
}

var _ DatasetScopeResolver = compileOnlyDatasetScopeResolver{}
