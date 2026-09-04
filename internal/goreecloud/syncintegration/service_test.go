package syncintegration

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeCheckpointAuthorizer struct {
	decision AuthorizationDecision
	err      error
	calls    int
}

func (f *fakeCheckpointAuthorizer) AuthorizeCheckpoint(_ context.Context, _ CheckpointRequest) (AuthorizationDecision, error) {
	f.calls++
	return f.decision, f.err
}

type fakeCheckpointExecutor struct {
	submission CheckpointSubmission
	err        error
	calls      int
	request    AuthorizedCheckpointRequest
}

func (f *fakeCheckpointExecutor) RequestCheckpoint(_ context.Context, request AuthorizedCheckpointRequest) (CheckpointSubmission, error) {
	f.calls++
	f.request = request
	return f.submission, f.err
}

func validCheckpointRequest() CheckpointRequest {
	return CheckpointRequest{
		ContractVersion:          ContractVersion,
		RequestID:                "request-123",
		DatasetID:                "family-documents",
		Purpose:                  CheckpointPreMigration,
		AuthorizationDecisionRef: "identity-decision-456",
	}
}

func validCheckpointSubmission() CheckpointSubmission {
	return CheckpointSubmission{
		RequestID:   "request-123",
		OperationID: "backup-operation-789",
		AcceptedAt:  time.Date(2026, time.September, 4, 12, 30, 0, 0, time.UTC),
	}
}

func TestCheckpointServiceRequiresBothRuntimeSeams(t *testing.T) {
	authorizer := &fakeCheckpointAuthorizer{}
	executor := &fakeCheckpointExecutor{}

	if _, err := NewCheckpointService(nil, executor); err == nil {
		t.Fatal("NewCheckpointService() accepted nil authorizer")
	}
	if _, err := NewCheckpointService(authorizer, nil); err == nil {
		t.Fatal("NewCheckpointService() accepted nil executor")
	}
	if _, err := NewCheckpointService(authorizer, executor); err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}
}

func TestCheckpointServiceAuthorizesBeforeExecution(t *testing.T) {
	request := validCheckpointRequest()
	authorizer := &fakeCheckpointAuthorizer{
		decision: AuthorizationDecision{
			DecisionRef: request.AuthorizationDecisionRef,
			Allowed:     true,
		},
	}
	executor := &fakeCheckpointExecutor{submission: validCheckpointSubmission()}
	service, err := NewCheckpointService(authorizer, executor)
	if err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}

	submission, err := service.RequestCheckpoint(context.Background(), request)
	if err != nil {
		t.Fatalf("RequestCheckpoint() error = %v", err)
	}
	if submission.OperationID != "backup-operation-789" {
		t.Fatalf("OperationID = %q", submission.OperationID)
	}
	if authorizer.calls != 1 {
		t.Fatalf("authorizer calls = %d, want 1", authorizer.calls)
	}
	if executor.calls != 1 {
		t.Fatalf("executor calls = %d, want 1", executor.calls)
	}
	if executor.request.AuthorizationDecisionRef != request.AuthorizationDecisionRef {
		t.Fatalf("executor authorization ref = %q, want %q", executor.request.AuthorizationDecisionRef, request.AuthorizationDecisionRef)
	}
	if executor.request.DatasetID != request.DatasetID || executor.request.Purpose != request.Purpose {
		t.Fatalf("executor request = %#v", executor.request)
	}
}

func TestCheckpointServiceRejectsMalformedRequestBeforeAuthorization(t *testing.T) {
	authorizer := &fakeCheckpointAuthorizer{
		decision: AuthorizationDecision{DecisionRef: "identity-decision-456", Allowed: true},
	}
	executor := &fakeCheckpointExecutor{submission: validCheckpointSubmission()}
	service, err := NewCheckpointService(authorizer, executor)
	if err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}

	request := validCheckpointRequest()
	request.Purpose = CheckpointPurpose("routine_sync")
	if _, err := service.RequestCheckpoint(context.Background(), request); err == nil {
		t.Fatal("RequestCheckpoint() accepted invalid purpose")
	}
	if authorizer.calls != 0 {
		t.Fatalf("authorizer calls = %d, want 0", authorizer.calls)
	}
	if executor.calls != 0 {
		t.Fatalf("executor calls = %d, want 0", executor.calls)
	}
}

func TestCheckpointServiceFailsClosedOnAuthorizationDenial(t *testing.T) {
	request := validCheckpointRequest()
	authorizer := &fakeCheckpointAuthorizer{
		decision: AuthorizationDecision{
			DecisionRef: request.AuthorizationDecisionRef,
			Allowed:     false,
		},
	}
	executor := &fakeCheckpointExecutor{submission: validCheckpointSubmission()}
	service, err := NewCheckpointService(authorizer, executor)
	if err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}

	_, err = service.RequestCheckpoint(context.Background(), request)
	if !errors.Is(err, ErrAuthorizationDenied) {
		t.Fatalf("RequestCheckpoint() error = %v, want ErrAuthorizationDenied", err)
	}
	if executor.calls != 0 {
		t.Fatalf("executor calls = %d, want 0", executor.calls)
	}
}

func TestCheckpointServiceFailsClosedOnMismatchedAuthorizationDecision(t *testing.T) {
	request := validCheckpointRequest()
	authorizer := &fakeCheckpointAuthorizer{
		decision: AuthorizationDecision{
			DecisionRef: "different-decision",
			Allowed:     true,
		},
	}
	executor := &fakeCheckpointExecutor{submission: validCheckpointSubmission()}
	service, err := NewCheckpointService(authorizer, executor)
	if err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}

	if _, err := service.RequestCheckpoint(context.Background(), request); err == nil {
		t.Fatal("RequestCheckpoint() accepted mismatched authorization decision")
	}
	if executor.calls != 0 {
		t.Fatalf("executor calls = %d, want 0", executor.calls)
	}
}

func TestCheckpointServiceFailsClosedOnAuthorizerError(t *testing.T) {
	request := validCheckpointRequest()
	authorizer := &fakeCheckpointAuthorizer{err: errors.New("identity unavailable")}
	executor := &fakeCheckpointExecutor{submission: validCheckpointSubmission()}
	service, err := NewCheckpointService(authorizer, executor)
	if err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}

	if _, err := service.RequestCheckpoint(context.Background(), request); err == nil {
		t.Fatal("RequestCheckpoint() ignored authorizer error")
	}
	if executor.calls != 0 {
		t.Fatalf("executor calls = %d, want 0", executor.calls)
	}
}

func TestCheckpointServicePropagatesExecutorFailureWithoutInventingSuccess(t *testing.T) {
	request := validCheckpointRequest()
	authorizer := &fakeCheckpointAuthorizer{
		decision: AuthorizationDecision{
			DecisionRef: request.AuthorizationDecisionRef,
			Allowed:     true,
		},
	}
	executor := &fakeCheckpointExecutor{err: errors.New("backup engine unavailable")}
	service, err := NewCheckpointService(authorizer, executor)
	if err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}

	if _, err := service.RequestCheckpoint(context.Background(), request); err == nil {
		t.Fatal("RequestCheckpoint() invented a successful submission")
	}
	if executor.calls != 1 {
		t.Fatalf("executor calls = %d, want 1", executor.calls)
	}
}

func TestCheckpointServiceRejectsMalformedSubmission(t *testing.T) {
	request := validCheckpointRequest()
	authorizer := &fakeCheckpointAuthorizer{
		decision: AuthorizationDecision{
			DecisionRef: request.AuthorizationDecisionRef,
			Allowed:     true,
		},
	}

	for _, submission := range []CheckpointSubmission{
		{RequestID: "other-request", OperationID: "operation-1", AcceptedAt: time.Now().UTC()},
		{RequestID: request.RequestID, OperationID: "", AcceptedAt: time.Now().UTC()},
		{RequestID: request.RequestID, OperationID: "operation-1"},
	} {
		executor := &fakeCheckpointExecutor{submission: submission}
		service, err := NewCheckpointService(authorizer, executor)
		if err != nil {
			t.Fatalf("NewCheckpointService() error = %v", err)
		}
		if _, err := service.RequestCheckpoint(context.Background(), request); err == nil {
			t.Fatalf("RequestCheckpoint() accepted malformed submission %#v", submission)
		}
	}
}

func TestCheckpointServiceRejectsNilContext(t *testing.T) {
	request := validCheckpointRequest()
	authorizer := &fakeCheckpointAuthorizer{
		decision: AuthorizationDecision{DecisionRef: request.AuthorizationDecisionRef, Allowed: true},
	}
	executor := &fakeCheckpointExecutor{submission: validCheckpointSubmission()}
	service, err := NewCheckpointService(authorizer, executor)
	if err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}

	if _, err := service.RequestCheckpoint(nil, request); err == nil {
		t.Fatal("RequestCheckpoint() accepted nil context")
	}
	if authorizer.calls != 0 || executor.calls != 0 {
		t.Fatalf("nil context reached adapters: authorizer=%d executor=%d", authorizer.calls, executor.calls)
	}
}
