# GoreeCloud Backup ↔ GoreeCloud Sync Integration

## Status

**Development status:** Source-level contract foundation implemented on the GoreeCloud Backup development branch.

**Production status:** Not implemented or accepted for production use.

This document describes the repository-local implementation boundary between GoreeCloud Backup and GoreeCloud Sync. The authoritative product requirements remain in the GoreeCloud project records. Source code and tests in this repository are authoritative for what the current branch actually implements.

## Authority boundary

GoreeCloud Backup and GoreeCloud Sync have different responsibilities and must remain independently useful.

GoreeCloud Backup owns recovery-domain truth, including independent protection state, recovery points, retention, repository state, verification, restore evidence, and backup-domain restore behavior.

GoreeCloud Sync owns synchronization-domain truth, including ongoing state convergence, synchronization direction, Sync-managed folder state, conflicts, deletion propagation, and post-restore synchronization reconciliation.

Synchronization is not backup. A synchronized duplicate must not be represented as independent recovery evidence merely because another copy exists.

The Backup-to-Sync contract therefore exposes only three product-layer operations:

1. Read bounded Backup-owned protection state for an explicitly scoped dataset.
2. Request a pre-change or pre-migration checkpoint through a future authenticated adapter.
3. Coordinate the safety boundary around a restore that targets a Sync-managed dataset.

The contract does not expose authority to delete recovery points, delete repositories, change retention, rotate encryption material, run repository maintenance, disable protection, or perform general synchronization.

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

`ProtectionView` provides a privacy-conscious projection of Backup-owned protection state. It can carry:

- the scoped dataset identifier;
- the evaluation time;
- the derived protection state;
- bounded reason codes;
- bounded missing, failed, and stale evidence categories.

It does not contain backed-up file contents, file paths, credentials, repository passwords, encryption keys, access tokens, raw backend output, or other reusable secret material.

GoreeCloud Sync may display or react to this state, but it does not derive, override, or become authoritative for the state.

## Pre-change and pre-migration checkpoints

The source contract recognizes only two Sync-originated checkpoint purposes:

- `pre_change`;
- `pre_migration`.

The request contains bounded opaque identifiers and an authorization-decision reference. That reference is metadata only. Its presence is not proof of authorization.

A future runtime adapter must authenticate the caller and validate applicable GoreeCloud Identity, Wardveil Security, Privacy Shield, and other authorization requirements before invoking a real Backup operation. The current package does not implement that runtime authorization and must not be described as doing so.

A checkpoint request also does not prove that a backup completed, was independently stored, passed integrity verification, or can be restored. Those remain Backup-owned recovery facts.

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

This package is deliberately not an authentication or authorization system.

It provides structural validation only. Network admission, authenticated service identity, user/device authorization, policy evaluation, replay protection, transport security, and authorization-decision verification remain responsibilities of future explicit platform adapters and the applicable GoreeCloud Identity, Mesh, Wardveil Security, and Privacy Shield contracts.

No caller-supplied field should be treated as proof of permission by itself.

Backup repository credentials and destructive backup authority must remain separately protected. Access to a synchronized dataset must not automatically grant authority to erase its independent recovery history.

## Validation

The repository includes focused unit tests intended to prove that:

- the exposed operation set remains narrow;
- destructive Backup-domain operations fail closed at this boundary;
- Backup protection state is copied into a bounded Sync-facing view;
- invalid contract versions, identifiers, protection states, and checkpoint purposes fail closed;
- Sync-managed restores require staging and reconciliation safeguards;
- loss of Sync availability does not eliminate access to controlled staged recovery;
- non-Sync-managed restores do not become dependent on Sync.

Repository CI remains required for the exact source revision. A source-level test pass does not establish production integration, target-environment authorization, production restore safety, or Stable qualification.

## Remaining implementation work

The following remain intentionally incomplete:

- authenticated Backup↔Sync transport;
- GoreeCloud Identity service/user authorization integration;
- GoreeCloud Mesh discovery, event, or evidence transport;
- Wardveil Security acceptance for the runtime integration;
- Privacy Shield acceptance for runtime metadata and data flows;
- durable mapping between Sync dataset identifiers and Backup protection scopes;
- runtime checkpoint execution and completion evidence;
- Sync pause or maintenance orchestration;
- restore staging and promotion implementation;
- post-restore reconciliation execution;
- audit history for cross-product operations;
- UI integration in Backup and Sync;
- target-environment restore testing;
- failure-mode testing with one product unavailable;
- production and Stable acceptance.

Until those items are implemented and verified, this integration must be represented as a source-level contract foundation only.
