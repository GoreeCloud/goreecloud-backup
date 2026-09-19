package syncintegration

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrDatasetScopeMappingNotFound indicates that Backup has no approved mapping
// for the requested Sync dataset. Callers must not infer a Backup scope from a
// path, name, or other Sync-owned metadata when this occurs.
var ErrDatasetScopeMappingNotFound = errors.New("backup scope mapping not found for Sync dataset")

// ErrDatasetScopeMappingInactive indicates that a mapping exists as historical
// or administrative state but is not currently eligible for checkpoint use.
var ErrDatasetScopeMappingInactive = errors.New("backup scope mapping is inactive")

// DatasetScopeMapping binds one opaque GoreeCloud Sync dataset identifier to a
// Backup-owned protection scope identifier. The mapping contains identifiers
// only; it intentionally carries no filesystem path, protected content,
// credentials, repository location, encryption material, or Sync policy state.
//
// MappingRevision is a Backup-owned opaque revision used to bind an authorized
// checkpoint request to the exact mapping that was resolved for it.
type DatasetScopeMapping struct {
	ContractVersion string    `json:"contractVersion"`
	DatasetID       string    `json:"datasetId"`
	BackupScopeID   string    `json:"backupScopeId"`
	MappingRevision string    `json:"mappingRevision"`
	Active          bool      `json:"active"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// Validate checks the shape of a mapping. It does not make an inactive mapping
// active and does not authorize access to either product.
func (m DatasetScopeMapping) Validate() error {
	if m.ContractVersion != ContractVersion {
		return fmt.Errorf("unsupported contract version %q", m.ContractVersion)
	}
	if err := validateOpaqueIdentifier("dataset ID", m.DatasetID); err != nil {
		return err
	}
	if err := validateOpaqueIdentifier("Backup scope ID", m.BackupScopeID); err != nil {
		return err
	}
	if err := validateOpaqueIdentifier("mapping revision", m.MappingRevision); err != nil {
		return err
	}
	if m.UpdatedAt.IsZero() {
		return fmt.Errorf("mapping update time must not be zero")
	}
	return nil
}

func (m DatasetScopeMapping) validateForDataset(datasetID string) error {
	if err := validateOpaqueIdentifier("dataset ID", datasetID); err != nil {
		return err
	}
	if err := m.Validate(); err != nil {
		return err
	}
	if m.DatasetID != datasetID {
		return fmt.Errorf("resolved mapping dataset ID does not match checkpoint request")
	}
	if !m.Active {
		return ErrDatasetScopeMappingInactive
	}
	return nil
}

// DatasetScopeResolver is the Backup-owned seam that resolves a Sync dataset
// into the exact Backup protection scope permitted for checkpoint execution.
//
// The runtime implementation must use authoritative Backup mapping state. Sync
// must not supply BackupScopeID or derive it from a filesystem path. Durable
// storage for that authoritative mapping is intentionally a separate runtime
// concern from this contract type.
type DatasetScopeResolver interface {
	ResolveBackupScope(context.Context, string) (DatasetScopeMapping, error)
}
