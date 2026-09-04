package syncintegration

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrAuthorizationDenied is returned when an authoritative authorization
// adapter explicitly denies a Backup operation requested through the Sync
// integration boundary.
var ErrAuthorizationDenied = errors.New("backup-to-sync operation authorization denied")

// AuthorizationDecision is the minimum result that a trusted runtime
// authorization adapter must return after independently validating the caller,
// the referenced GoreeCloud Identity/security decision, and the requested
// Backup operation.
//
// DecisionRef must identify the same decision reference supplied in the
// structurally validated request. Allowed is never inferred from the presence
// of a reference.
type AuthorizationDecision struct {
	DecisionRef string `json:"decisionRef"`
	Allowed     bool   `json:"allowed"`
}

func (d AuthorizationDecision) validateForCheckpoint(request CheckpointRequest) error {
	if err := validateOpaqueIdentifier("authorization decision reference", d.DecisionRef); err != nil {
		return fmt.Errorf("invalid authorization decision: %w", err)
	}
	if d.DecisionRef != request.AuthorizationDecisionRef {
		return fmt.Errorf("authorization decision reference does not match checkpoint request")
	}
	if !d.Allowed {
		return ErrAuthorizationDenied
	}
	return nil
}

// CheckpointAuthorizer is the required authorization seam for a runtime
// Backup-to-Sync checkpoint integration. Its implementation is expected to be
// backed by the applicable GoreeCloud Identity/security contracts.
//
// This repository-local interface deliberately does not prescribe a transport,
// credential format, or token type. Those concerns belong to the authorized
// platform adapter rather than the recovery-domain contract.
type CheckpointAuthorizer interface {
	AuthorizeCheckpoint(context.Context, CheckpointRequest) (AuthorizationDecision, error)
}

// AuthorizedCheckpointRequest is produced only after CheckpointService has
// structurally validated the original request, obtained an explicit matching
// allowed decision from CheckpointAuthorizer, and resolved the Sync dataset
// through Backup-owned mapping state.
//
// BackupScopeID and MappingRevision are never accepted from GoreeCloud Sync.
// They are inserted by Backup after authorization and mapping resolution.
type AuthorizedCheckpointRequest struct {
	ContractVersion          string            `json:"contractVersion"`
	RequestID                string            `json:"requestId"`
	DatasetID                string            `json:"datasetId"`
	BackupScopeID            string            `json:"backupScopeId"`
	MappingRevision          string            `json:"mappingRevision"`
	Purpose                  CheckpointPurpose `json:"purpose"`
	AuthorizationDecisionRef string            `json:"authorizationDecisionRef"`
}

// CheckpointSubmission records that the Backup-owned execution adapter accepted
// a checkpoint request. Acceptance is not backup completion, integrity
// verification, restore verification, or proof that a usable recovery point
// exists.
//
// DatasetID and BackupScopeID bind the receipt to the exact resolved scope and
// prevent a receipt for another dataset or scope from being reused.
type CheckpointSubmission struct {
	RequestID     string    `json:"requestId"`
	OperationID   string    `json:"operationId"`
	DatasetID     string    `json:"datasetId"`
	BackupScopeID string    `json:"backupScopeId"`
	AcceptedAt    time.Time `json:"acceptedAt"`
}

func (s CheckpointSubmission) validateForRequest(request AuthorizedCheckpointRequest) error {
	if err := validateOpaqueIdentifier("checkpoint submission request ID", s.RequestID); err != nil {
		return err
	}
	if s.RequestID != request.RequestID {
		return fmt.Errorf("checkpoint submission request ID does not match request")
	}
	if err := validateOpaqueIdentifier("checkpoint operation ID", s.OperationID); err != nil {
		return err
	}
	if err := validateOpaqueIdentifier("checkpoint submission dataset ID", s.DatasetID); err != nil {
		return err
	}
	if s.DatasetID != request.DatasetID {
		return fmt.Errorf("checkpoint submission dataset ID does not match request")
	}
	if err := validateOpaqueIdentifier("checkpoint submission Backup scope ID", s.BackupScopeID); err != nil {
		return err
	}
	if s.BackupScopeID != request.BackupScopeID {
		return fmt.Errorf("checkpoint submission Backup scope ID does not match resolved scope")
	}
	if s.AcceptedAt.IsZero() {
		return fmt.Errorf("checkpoint submission acceptance time must not be zero")
	}
	return nil
}

// CheckpointExecutor is the Backup-owned runtime seam that may accept an
// already-authorized checkpoint request. Implementations must still preserve
// Backup's own policy, repository, verification, and recovery semantics.
type CheckpointExecutor interface {
	RequestCheckpoint(context.Context, AuthorizedCheckpointRequest) (CheckpointSubmission, error)
}

// CheckpointService enforces the source-level authorization and scope-mapping
// boundaries before a Sync-originated checkpoint request can reach a Backup
// execution adapter.
type CheckpointService struct {
	authorizer CheckpointAuthorizer
	resolver   DatasetScopeResolver
	executor   CheckpointExecutor
}

// NewCheckpointService fails closed unless authorization, Backup-owned dataset
// mapping, and a Backup-owned checkpoint executor are all provided.
func NewCheckpointService(authorizer CheckpointAuthorizer, resolver DatasetScopeResolver, executor CheckpointExecutor) (*CheckpointService, error) {
	if authorizer == nil {
		return nil, fmt.Errorf("checkpoint authorizer is required")
	}
	if resolver == nil {
		return nil, fmt.Errorf("dataset scope resolver is required")
	}
	if executor == nil {
		return nil, fmt.Errorf("checkpoint executor is required")
	}
	return &CheckpointService{authorizer: authorizer, resolver: resolver, executor: executor}, nil
}

// RequestCheckpoint structurally validates the request, obtains an explicit
// matching authorization decision, resolves the Sync dataset through
// Backup-owned mapping state, and only then passes a reduced authorized request
// to the Backup-owned executor.
//
// Authorization runs before mapping resolution so an unauthorized caller cannot
// use this service to probe whether a Backup scope exists for a dataset. The
// caller-supplied AuthorizationDecisionRef is never treated as permission by
// itself, and the caller never supplies BackupScopeID. Authorization failure,
// mapping failure or mismatch, adapter errors, and malformed executor receipts
// all fail closed.
func (s *CheckpointService) RequestCheckpoint(ctx context.Context, request CheckpointRequest) (CheckpointSubmission, error) {
	if s == nil || s.authorizer == nil || s.resolver == nil || s.executor == nil {
		return CheckpointSubmission{}, fmt.Errorf("checkpoint service is not initialized")
	}
	if ctx == nil {
		return CheckpointSubmission{}, fmt.Errorf("context is required")
	}
	if err := ValidateOperation(OperationRequestCheckpoint); err != nil {
		return CheckpointSubmission{}, err
	}
	if err := request.Validate(); err != nil {
		return CheckpointSubmission{}, err
	}

	decision, err := s.authorizer.AuthorizeCheckpoint(ctx, request)
	if err != nil {
		return CheckpointSubmission{}, fmt.Errorf("authorize checkpoint: %w", err)
	}
	if err := decision.validateForCheckpoint(request); err != nil {
		return CheckpointSubmission{}, err
	}

	mapping, err := s.resolver.ResolveBackupScope(ctx, request.DatasetID)
	if err != nil {
		return CheckpointSubmission{}, fmt.Errorf("resolve Backup scope: %w", err)
	}
	if err := mapping.validateForDataset(request.DatasetID); err != nil {
		return CheckpointSubmission{}, fmt.Errorf("invalid Backup scope mapping: %w", err)
	}

	authorized := AuthorizedCheckpointRequest{
		ContractVersion:          request.ContractVersion,
		RequestID:                request.RequestID,
		DatasetID:                request.DatasetID,
		BackupScopeID:            mapping.BackupScopeID,
		MappingRevision:          mapping.MappingRevision,
		Purpose:                  request.Purpose,
		AuthorizationDecisionRef: decision.DecisionRef,
	}

	submission, err := s.executor.RequestCheckpoint(ctx, authorized)
	if err != nil {
		return CheckpointSubmission{}, fmt.Errorf("request checkpoint: %w", err)
	}
	if err := submission.validateForRequest(authorized); err != nil {
		return CheckpointSubmission{}, fmt.Errorf("invalid checkpoint submission: %w", err)
	}
	return submission, nil
}
