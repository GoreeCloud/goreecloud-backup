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

type fakeDatasetScopeResolver struct {
	mapping DatasetScopeMapping
	err     error
	calls   int
	dataset string
}

func (f *fakeDatasetScopeResolver) ResolveBackupScope(_ context.Context, datasetID string) (DatasetScopeMapping, error) {
	f.calls++
	f.dataset = datasetID
	return f.mapping, f.err
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
		RequestID:     "request-123",
		OperationID:   "backup-operation-789",
		DatasetID:     "family-documents",
		BackupScopeID: "backup-scope-family-documents",
		AcceptedAt:    time.Date(2026, time.September, 4, 12, 30, 0, 0, time.UTC),
	}
}

func validCheckpointRuntimeSeams() (*fakeCheckpointAuthorizer, *fakeDatasetScopeResolver, *fakeCheckpointExecutor) {
	request := validCheckpointRequest()
	return &fakeCheckpointAuthorizer{
		decision: AuthorizationDecision{
			DecisionRef: request.AuthorizationDecisionRef,
			Allowed:     true,
		},
	}, &fakeDatasetScopeResolver{
		mapping: validDatasetScopeMapping(),
	}, &fakeCheckpointExecutor{
		submission: validCheckpointSubmission(),
	}
}

func TestCheckpointServiceRequiresAllRuntimeSeams(t *testing.T) {
	authorizer, resolver, executor := validCheckpointRuntimeSeams()

	if _, err := NewCheckpointService(nil, resolver, executor); err == nil {
		t.Fatal("NewCheckpointService() accepted nil authorizer")
	}
	if _, err := NewCheckpointService(authorizer, nil, executor); err == nil {
		t.Fatal("NewCheckpointService() accepted nil resolver")
	}
	if _, err := NewCheckpointService(authorizer, resolver, nil); err == nil {
		t.Fatal("NewCheckpointService() accepted nil executor")
	}
	if _, err := NewCheckpointService(authorizer, resolver, executor); err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}
}

func TestCheckpointServiceAuthorizesThenResolvesScopeBeforeExecution(t *testing.T) {
	request := validCheckpointRequest()
	authorizer, resolver, executor := validCheckpointRuntimeSeams()
	service, err := NewCheckpointService(authorizer, resolver, executor)
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
	if authorizer.calls != 1 || resolver.calls != 1 || executor.calls != 1 {
		t.Fatalf("calls authorizer=%d resolver=%d executor=%d, want 1/1/1", authorizer.calls, resolver.calls, executor.calls)
	}
	if resolver.dataset != request.DatasetID {
		t.Fatalf("resolver dataset = %q, want %q", resolver.dataset, request.DatasetID)
	}
	if executor.request.AuthorizationDecisionRef != request.AuthorizationDecisionRef {
		t.Fatalf("executor authorization ref = %q, want %q", executor.request.AuthorizationDecisionRef, request.AuthorizationDecisionRef)
	}
	if executor.request.DatasetID != request.DatasetID || executor.request.Purpose != request.Purpose {
		t.Fatalf("executor request = %#v", executor.request)
	}
	if executor.request.BackupScopeID != resolver.mapping.BackupScopeID {
		t.Fatalf("executor BackupScopeID = %q, want %q", executor.request.BackupScopeID, resolver.mapping.BackupScopeID)
	}
	if executor.request.MappingRevision != resolver.mapping.MappingRevision {
		t.Fatalf("executor MappingRevision = %q, want %q", executor.request.MappingRevision, resolver.mapping.MappingRevision)
	}
}

func TestCheckpointServiceRejectsMalformedRequestBeforeAuthorization(t *testing.T) {
	authorizer, resolver, executor := validCheckpointRuntimeSeams()
	service, err := NewCheckpointService(authorizer, resolver, executor)
	if err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}

	request := validCheckpointRequest()
	request.Purpose = CheckpointPurpose("routine_sync")
	if _, err := service.RequestCheckpoint(context.Background(), request); err == nil {
		t.Fatal("RequestCheckpoint() accepted invalid purpose")
	}
	if authorizer.calls != 0 || resolver.calls != 0 || executor.calls != 0 {
		t.Fatalf("malformed request reached adapters: authorizer=%d resolver=%d executor=%d", authorizer.calls, resolver.calls, executor.calls)
	}
}

func TestCheckpointServiceFailsClosedOnAuthorizationDenialBeforeMapping(t *testing.T) {
	request := validCheckpointRequest()
	authorizer, resolver, executor := validCheckpointRuntimeSeams()
	authorizer.decision.Allowed = false
	service, err := NewCheckpointService(authorizer, resolver, executor)
	if err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}

	_, err = service.RequestCheckpoint(context.Background(), request)
	if !errors.Is(err, ErrAuthorizationDenied) {
		t.Fatalf("RequestCheckpoint() error = %v, want ErrAuthorizationDenied", err)
	}
	if resolver.calls != 0 || executor.calls != 0 {
		t.Fatalf("denied request reached later adapters: resolver=%d executor=%d", resolver.calls, executor.calls)
	}
}

func TestCheckpointServiceFailsClosedOnMismatchedAuthorizationDecision(t *testing.T) {
	request := validCheckpointRequest()
	authorizer, resolver, executor := validCheckpointRuntimeSeams()
	authorizer.decision.DecisionRef = "different-decision"
	service, err := NewCheckpointService(authorizer, resolver, executor)
	if err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}

	if _, err := service.RequestCheckpoint(context.Background(), request); err == nil {
		t.Fatal("RequestCheckpoint() accepted mismatched authorization decision")
	}
	if resolver.calls != 0 || executor.calls != 0 {
		t.Fatalf("mismatched decision reached later adapters: resolver=%d executor=%d", resolver.calls, executor.calls)
	}
}

func TestCheckpointServiceFailsClosedOnAuthorizerError(t *testing.T) {
	request := validCheckpointRequest()
	authorizer, resolver, executor := validCheckpointRuntimeSeams()
	authorizer.err = errors.New("identity unavailable")
	service, err := NewCheckpointService(authorizer, resolver, executor)
	if err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}

	if _, err := service.RequestCheckpoint(context.Background(), request); err == nil {
		t.Fatal("RequestCheckpoint() ignored authorizer error")
	}
	if resolver.calls != 0 || executor.calls != 0 {
		t.Fatalf("authorizer error reached later adapters: resolver=%d executor=%d", resolver.calls, executor.calls)
	}
}

func TestCheckpointServiceFailsClosedOnScopeResolutionFailure(t *testing.T) {
	request := validCheckpointRequest()
	authorizer, resolver, executor := validCheckpointRuntimeSeams()
	resolver.err = ErrDatasetScopeMappingNotFound
	service, err := NewCheckpointService(authorizer, resolver, executor)
	if err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}

	if _, err := service.RequestCheckpoint(context.Background(), request); !errors.Is(err, ErrDatasetScopeMappingNotFound) {
		t.Fatalf("RequestCheckpoint() error = %v, want ErrDatasetScopeMappingNotFound", err)
	}
	if authorizer.calls != 1 || resolver.calls != 1 || executor.calls != 0 {
		t.Fatalf("calls authorizer=%d resolver=%d executor=%d, want 1/1/0", authorizer.calls, resolver.calls, executor.calls)
	}
}

func TestCheckpointServiceRejectsInactiveOrMismatchedScopeMapping(t *testing.T) {
	request := validCheckpointRequest()

	for _, mutate := range []func(*DatasetScopeMapping){
		func(m *DatasetScopeMapping) { m.Active = false },
		func(m *DatasetScopeMapping) { m.DatasetID = "different-dataset" },
		func(m *DatasetScopeMapping) { m.BackupScopeID = "" },
	} {
		authorizer, resolver, executor := validCheckpointRuntimeSeams()
		mutate(&resolver.mapping)
		service, err := NewCheckpointService(authorizer, resolver, executor)
		if err != nil {
			t.Fatalf("NewCheckpointService() error = %v", err)
		}
		if _, err := service.RequestCheckpoint(context.Background(), request); err == nil {
			t.Fatal("RequestCheckpoint() accepted invalid scope mapping")
		}
		if executor.calls != 0 {
			t.Fatalf("invalid scope mapping reached executor %d times", executor.calls)
		}
	}
}

func TestCheckpointServicePropagatesExecutorFailureWithoutInventingSuccess(t *testing.T) {
	request := validCheckpointRequest()
	authorizer, resolver, executor := validCheckpointRuntimeSeams()
	executor.err = errors.New("backup engine unavailable")
	service, err := NewCheckpointService(authorizer, resolver, executor)
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

	for _, submission := range []CheckpointSubmission{
		{RequestID: "other-request", OperationID: "operation-1", DatasetID: request.DatasetID, BackupScopeID: "backup-scope-family-documents", AcceptedAt: time.Now().UTC()},
		{RequestID: request.RequestID, OperationID: "", DatasetID: request.DatasetID, BackupScopeID: "backup-scope-family-documents", AcceptedAt: time.Now().UTC()},
		{RequestID: request.RequestID, OperationID: "operation-1", DatasetID: "other-dataset", BackupScopeID: "backup-scope-family-documents", AcceptedAt: time.Now().UTC()},
		{RequestID: request.RequestID, OperationID: "operation-1", DatasetID: request.DatasetID, BackupScopeID: "other-scope", AcceptedAt: time.Now().UTC()},
		{RequestID: request.RequestID, OperationID: "operation-1", DatasetID: request.DatasetID, BackupScopeID: "backup-scope-family-documents"},
	} {
		authorizer, resolver, executor := validCheckpointRuntimeSeams()
		executor.submission = submission
		service, err := NewCheckpointService(authorizer, resolver, executor)
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
	authorizer, resolver, executor := validCheckpointRuntimeSeams()
	service, err := NewCheckpointService(authorizer, resolver, executor)
	if err != nil {
		t.Fatalf("NewCheckpointService() error = %v", err)
	}

	if _, err := service.RequestCheckpoint(nil, request); err == nil {
		t.Fatal("RequestCheckpoint() accepted nil context")
	}
	if authorizer.calls != 0 || resolver.calls != 0 || executor.calls != 0 {
		t.Fatalf("nil context reached adapters: authorizer=%d resolver=%d executor=%d", authorizer.calls, resolver.calls, executor.calls)
	}
}
