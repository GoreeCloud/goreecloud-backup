# GoreeCloud Backup ↔ GoreeCloud Sync Integration

## Status

**Development status:** Source-level authority, authorization-ordering, Backup-owned dataset-scope mapping with a durable local store, durable checkpoint lifecycle/recovery-evidence storage, and restore-safety contracts are implemented on the GoreeCloud Backup development branch.

**Production status:** Not implemented or accepted for production use.

This document describes the repository-local implementation boundary between GoreeCloud Backup and GoreeCloud Sync. The authoritative product requirements remain in the GoreeCloud project records. Source code and tests in this repository are authoritative for what the current branch actually implements.

## Authority boundary

GoreeCloud Backup and GoreeCloud Sync have different responsibilities and must remain independently useful.

GoreeCloud Backup owns recovery-domain truth, including independent protection state, recovery points, protection scopes, retention, repository state, verification, restore evidence, and backup-domain restore behavior.

GoreeCloud Sync owns synchronization-domain truth, including ongoing state convergence, synchronization direction, Sync-managed dataset state, conflicts, deletion propagation, and post-restore synchronization reconciliation.

Synchronization is not backup. A synchronized duplicate must not be represented as independent recovery evidence merely because another copy exists.

The Backup-to-Sync contract therefore exposes only three product-layer operations:

1. Read bounded Backup-owned protection state for an explicitly scoped dataset.
2. Request a pre-change or pre-migration checkpoint through an authorization-gated Backup service boundary.
3. Coordinate the safety boundary around a restore that targets a Sync-managed dataset.

The contract does not expose authority to delete recovery points, delete repositories, change retention, rotate encryption material, run repository maintenance, disable protection, administer dataset-to-scope mappings, or perform general synchronization.

## Source package

The source-level contract is implemented in:

`internal/goreecloud/syncintegration`

It depends on the GoreeCloud-owned recovery-assurance model in:

`internal/goreecloud/protection`

It intentionally does not depend directly on Kopia repository internals. This keeps GoreeCloud product integration separate from recovery-critical repository behavior and reduces the risk that a Sync integration changes repository semantics.

The current contract identifier is:

`goreecloud.backup-sync/v1alpha1`

This is a pre-stabilization internal identifier, not a public compatibility promise.

## Protection awareness

`ProtectionView` provides a privacy-conscious projection of Backup-owned protection state. It can carry the scoped dataset identifier, evaluation time, derived protection state, bounded reason codes, and bounded missing, failed, and stale evidence categories.

It does not contain backed-up file contents, file paths, credentials, repository passwords, encryption keys, access tokens, raw backend output, or other reusable secret material.

GoreeCloud Sync may display or react to this state, but it does not derive, override, or become authoritative for the state.

## Pre-change and pre-migration checkpoints

The source contract recognizes only two Sync-originated checkpoint purposes:

- `pre_change`;
- `pre_migration`.

The request contains bounded opaque identifiers and an authorization-decision reference. That reference is metadata only. Its presence is not proof of authorization.

`CheckpointService` fails closed unless a `CheckpointAuthorizer`, Backup-owned `DatasetScopeResolver`, and Backup-owned `CheckpointExecutor` are configured. Before execution, the service:

1. validates the operation and request structure;
2. asks the authorizer to independently evaluate the request;
3. requires an explicit allowed decision whose decision reference exactly matches the request;
4. only after authorization, resolves the Sync dataset through Backup-owned mapping state;
5. requires the resolved mapping to match the requested dataset and be active;
6. inserts the Backup-owned scope ID and mapping revision into the reduced authorized request;
7. invokes the Backup-owned executor; and
8. validates that the executor receipt is bound to the exact request, dataset, and Backup scope.

Authorization occurs before scope resolution so an unauthorized caller cannot use this service to probe whether a Backup protection scope exists for a dataset.

Authorization errors, denials, mismatched decision references, mapping lookup failures, inactive or mismatched mappings, malformed requests, executor failures, and malformed execution receipts all fail closed.

## Backup-owned dataset-to-scope mapping

`DatasetScopeMapping` maps one opaque GoreeCloud Sync dataset ID to a Backup-owned protection scope ID. It also carries a Backup-owned mapping revision, active state, and update time.

The mapping intentionally carries no filesystem path, protected content, credentials, repository location, encryption material, or Sync policy state.

GoreeCloud Sync does not submit `BackupScopeID`. The Backup integration resolves that value through `DatasetScopeResolver` only after the checkpoint request has passed authorization. This prevents Sync from selecting or escalating into a broader Backup scope by changing a request field.

The source now includes `FileDatasetScopeStore`, a Backup-owned durable implementation of `DatasetScopeResolver`. Its storage path must be supplied explicitly by the Backup runtime or administrative configuration; the integration package does not invent a production location. The store:

- uses a versioned JSON envelope;
- accepts at most a bounded number of mappings;
- validates every mapping before publication;
- rejects duplicate Sync dataset IDs;
- snapshots caller-provided mapping data before validation/encoding;
- publishes through a private temporary file and same-directory atomic rename;
- uses private file permissions and rejects group/other-readable mapping files on Unix-like systems;
- rejects non-regular files, unknown JSON fields, trailing JSON values, unsupported store versions, and malformed mappings; and
- distinguishes an uninitialized store from an initialized store that lacks a requested mapping.

`ReplaceMappings` is a Backup administration primitive and is not part of the Backup-to-Sync operation allowlist. The current source does not expose any Sync operation that can create, replace, activate, or deactivate these mappings.

A runtime-selected canonical storage location, mapping administration API/UX, authorization for mapping administration, audit history, migration policy for future store versions, deployment integration, and production acceptance are still incomplete.

## Checkpoint lifecycle and recovery evidence

`CheckpointSubmission` means only that the Backup-owned execution adapter accepted a request. Acceptance is not backup completion, integrity verification, restore verification, or proof that a usable recovery point exists.

`CheckpointStatus` adds a Backup-owned lifecycle/evidence model with four states:

- `accepted`;
- `running`;
- `failed`;
- `completed`.

Accepted and running states are prohibited from claiming a recovery point, integrity verification, restore verification, or failure evidence. Failed status must use a bounded failure category and cannot claim a usable recovery point. Completed status must identify a recovery point, but completion by itself is still not sufficient for a protected-change decision.

`ReadyForProtectedChange()` returns true only when the status is structurally valid, the operation is completed, Backup reports the recovery point as usable, and integrity verification has passed. `ReadyForSubmission()` additionally requires that the status is bound to the exact accepted submission before returning true, and is the safer helper for a concrete pre-change or pre-migration decision. Restore verification is stronger evidence and is represented separately; it is not automatically required for every individual pre-change checkpoint because a controlling runtime policy may impose a stronger requirement for a particular change.

Cross-product failure evidence is deliberately limited to bounded categories rather than raw repository, path, credential, backend, or protected-content errors.

`ValidateForSubmission` binds lifecycle evidence to the exact request ID, operation ID, dataset ID, and Backup scope ID in the accepted submission and rejects status observations that predate acceptance. This prevents evidence from a different checkpoint or protection scope from satisfying the request.

`CheckpointStatusProvider` is only a Backup-owned source seam for authoritative runtime evidence. A submission or operation ID is correlation data, not a bearer credential; authenticated and authorized runtime access to status remains required.

### Durable checkpoint lifecycle/evidence store

The source now includes `FileCheckpointStatusStore`, a Backup-owned durable implementation of `CheckpointStatusProvider`. Its path is explicitly supplied by Backup runtime or administrative configuration; the integration package does not select a production location and does not expose a Sync mutation operation for this store.

Each checkpoint operation is initialized from one accepted `CheckpointSubmission`. The store persists that immutable correlation record together with a bounded chronological history of validated `CheckpointStatus` observations. It:

- uses a versioned JSON envelope;
- limits the number of operations and observations retained in one store file;
- records the accepted submission as the first lifecycle observation;
- treats exact submission and latest-status replays as idempotent;
- rejects a different submission for an already recorded operation;
- rejects duplicate operation IDs and duplicate request IDs in durable state;
- requires every later status to remain bound to the exact request, operation, Sync dataset, and Backup scope;
- requires observation time to advance;
- prevents lifecycle regressions back to `accepted` or `running` after later states;
- permits completed recovery evidence to strengthen without changing the recovery-point ID or discarding previously established usable, integrity, or restore-verification evidence;
- permits an explicit later failure observation after completion so a subsequently discovered verification or repository failure cannot remain represented as recovery-ready;
- treats a failed operation as terminal so a distinct retry requires a new operation ID;
- validates the complete persisted history every time it is loaded or rewritten;
- publishes through the same private atomic-file boundary used by the Backup-owned mapping store; and
- rejects malformed JSON, unknown fields, trailing JSON values, unsupported versions, non-regular files, loose Unix permissions, invalid histories, and uninitialized state.

`CheckpointStatus()` returns only the latest validated Backup-owned observation for an operation while the durable record retains the earlier lifecycle observations. A stored `accepted` or merely `completed` status still does not satisfy `ReadyForSubmission()` without a usable integrity-verified recovery point.

This store is a source-level persistence foundation, not a production engine observer. The current `CheckpointService` does not silently turn executor acceptance into durable recovery evidence, and no real Backup engine or repository adapter currently writes authoritative completion/integrity observations into this store. Runtime wiring must handle accepted-execution/persistence failure and retry semantics without creating duplicate backup work before production use.

## Restore safety for Sync-managed paths

A restore targeting a Sync-managed production dataset is treated conservatively.

The current source contract requires:

1. Restore into a controlled staging boundary.
2. Place the affected Sync relationship into an approved pause or maintenance state before production promotion where applicable.
3. Validate the restored information.
4. Reconcile the promoted authoritative dataset through GoreeCloud Sync.

The source-level contract never grants direct-write permission into a Sync-managed production path. A later runtime layer must separately prove the required authorization, maintenance state, restore validation, and promotion decision.

This protects against a restore unintentionally causing stale state, deletion, corruption, or conflicting versions to propagate through synchronization before the recovered dataset is intentionally made authoritative.

## Sync unavailability and disaster recovery

GoreeCloud Backup must not require GoreeCloud Sync to be available merely to access independent recovery information.

If Sync is unavailable or its state is unknown, the contract still allows the recovery workflow to proceed to a controlled staging location. Direct production promotion remains outside this contract, and the Sync reconciliation state is recorded as pending.

This prevents a circular dependency in which a failed Sync service blocks recovery of the information required to restore Sync or its surrounding environment.

## Authorization and security boundary

This package is deliberately not an authentication or identity system.

It provides an authorization-gated service seam for checkpoint submission, but network admission, authenticated service identity, user/device authentication, policy evaluation, replay protection, transport security, and concrete authorization-decision verification remain responsibilities of explicit platform adapters and the applicable GoreeCloud Identity, Mesh, Wardveil Security, and Privacy Shield contracts.

No caller-supplied field is treated as proof of permission by itself. A future concrete `CheckpointAuthorizer` must verify the external decision rather than echoing the caller's reference.

Backup repository credentials and destructive backup authority must remain separately protected. Access to a synchronized dataset must not automatically grant authority to erase its independent recovery history or administer its Backup scope mapping.

## Validation expectations

The repository includes focused unit tests intended to prove that:

- the exposed operation set remains narrow;
- destructive Backup-domain operations fail closed at this boundary;
- Backup protection state is copied into a bounded Sync-facing view;
- invalid contract versions, identifiers, protection states, and checkpoint purposes fail closed;
- malformed checkpoint requests are rejected before authorization;
- an explicit matching allowed authorization decision is required before mapping or executor invocation;
- denied or failed authorization cannot be used to probe dataset-to-scope mappings through `CheckpointService`;
- Sync cannot supply a Backup protection scope to the executor;
- missing, inactive, malformed, or mismatched dataset mappings prevent execution;
- durable mapping replacement is atomic at the file-publication boundary and invalid replacements do not replace prior valid state;
- durable mapping files fail closed for malformed structure, unsupported versions, loose Unix permissions, and uninitialized state;
- executor receipts must match the exact request, dataset, and resolved Backup scope;
- a checkpoint submission alone never becomes recovery-ready evidence;
- lifecycle evidence rejects contradictory states and binds to the exact accepted submission;
- the durable checkpoint store distinguishes uninitialized and missing-operation state, persists the accepted submission, preserves validated chronological evidence, and treats exact replays idempotently;
- the durable checkpoint store rejects rebinding, stale observations, lifecycle regressions, weakened completed evidence, malformed store state, and loose Unix permissions;
- an explicitly failed checkpoint does not become ready and cannot be silently rewritten as a successful retry under the same operation ID;
- only a completed, usable, integrity-verified recovery point satisfies the source-level protected-change predicate;
- submission-bound readiness rejects evidence from another checkpoint;
- Sync-managed restores require staging and reconciliation safeguards;
- loss of Sync availability does not eliminate access to controlled staged recovery; and
- non-Sync-managed restores do not become dependent on Sync.

Repository CI remains required for the exact source revision. A source-level test pass does not establish production integration, target-environment authorization, production restore safety, or Stable qualification.

## Remaining implementation work

The following remain intentionally incomplete:

- authenticated Backup↔Sync transport;
- concrete GoreeCloud Identity service/user authorization adapter and decision verification;
- GoreeCloud Mesh discovery, event, or evidence transport;
- Wardveil Security acceptance for the runtime integration;
- Privacy Shield acceptance for runtime metadata and data flows;
- runtime configuration, administration, authorization, audit, and migration lifecycle for the durable dataset-to-scope store;
- runtime-selected canonical location, retention/migration policy, and authorized administration for the durable checkpoint status store;
- a concrete Backup checkpoint executor;
- runtime wiring that durably records accepted submissions without creating duplicate checkpoint work on persistence/retry failures;
- an authoritative engine/repository evidence adapter that records real checkpoint completion, recovery-point usability, integrity, and applicable restore evidence;
- Sync pause or maintenance orchestration;
- restore staging and promotion implementation;
- post-restore reconciliation execution;
- audit integration for cross-product operations;
- UI integration in Backup and Sync;
- target-environment restore testing;
- failure-mode testing with one product unavailable; and
- production and Stable acceptance.

Until those items are implemented and verified, this integration must be represented as a source-level contract, authorization-ordering, durable mapping-store, durable checkpoint-evidence-store, and restore-safety foundation only.
