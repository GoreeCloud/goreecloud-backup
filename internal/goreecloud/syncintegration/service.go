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
// structurally validated the original request and obtained an explicit matching
// allowed decision from CheckpointAuthorizer.
type AuthorizedCheckpointRequest struct {
	ContractVersion          string            `json:"contractVersion"`
	RequestID                string            `json:"requestId"`
	DatasetID                string            `json:"datasetId"`
	Purpose                  CheckpointPurpose `json:"purpose"`
	AuthorizationDecisionRef string            `json:"authorizationDecisionRef"`
}

// CheckpointSubmission records that the Backup-owned execution adapter accepted
// a checkpoint request. Acceptance is not backup completion, integrity
// verification, restore verification, or proof that a usable recovery point
// exists.
type CheckpointSubmission struct {
	RequestID   string    `json:"requestId"`
	OperationID string    `json:"operationId"`
	AcceptedAt  time.Time `json:"acceptedAt"`
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

// CheckpointService enforces the source-level authorization boundary before a
// Sync-originated checkpoint request can reach a Backup execution adapter.
type CheckpointService struct {
	authorizer CheckpointAuthorizer
	executor   CheckpointExecutor
}

// NewCheckpointService fails closed unless both an authorization adapter and a
// Backup-owned checkpoint executor are provided.
func NewCheckpointService(authorizer CheckpointAuthorizer, executor CheckpointExecutor) (*CheckpointService, error) {
	if authorizer == nil {
		return nil, fmt.Errorf("checkpoint authorizer is required")
	}
	if executor == nil {
		return nil, fmt.Errorf("checkpoint executor is required")
	}
	return &CheckpointService{authorizer: authorizer, executor: executor}, nil
}

// RequestCheckpoint structurally validates the request, obtains an explicit
// matching authorization decision, and only then passes a reduced authorized
// request to the Backup-owned executor.
//
// The caller-supplied AuthorizationDecisionRef is never treated as permission
// by itself. Failure to authorize, mismatched decision references, adapter
// errors, or malformed executor receipts all fail closed.
func (s *CheckpointService) RequestCheckpoint(ctx context.Context, request CheckpointRequest) (CheckpointSubmission, error) {
	if s == nil || s.authorizer == nil || s.executor == nil {
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

	authorized := AuthorizedCheckpointRequest{
		ContractVersion:          request.ContractVersion,
		RequestID:                request.RequestID,
		DatasetID:                request.DatasetID,
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
