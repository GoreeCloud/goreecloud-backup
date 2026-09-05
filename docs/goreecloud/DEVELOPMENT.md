# GoreeCloud Backup — Development and Recovery Safety

## Status

GoreeCloud Backup is an active GoreeCloud-maintained fork of Kopia. Development may improve product identity, Glaze UI, Wardveil Security presentation, diagnostics, packaging, integration contracts, and operational readiness while preserving the mature recovery engine unless a separately reviewed recovery-critical change is justified.

This document does not authorize production replacement of an existing Kopia deployment.

## Governing principle

The application exists to preserve recoverability. A successful build, process start, security scan, UI render, snapshot, or monitoring heartbeat is evidence for a specific layer only. None of those events alone proves that protected information can be restored.

Representative restoration remains mandatory recovery evidence.

## Development boundaries

### Product-layer work

Normal GoreeCloud product-layer development may include:

- GoreeCloud Backup identity and packaging;
- Glaze UI implementation and accessibility/resilience;
- Wardveil Security status and protection presentation;
- diagnostics and privacy-conscious observability;
- policy/status/verification presentation;
- explicit API and integration contracts;
- desktop-shell security and usability;
- build/release engineering;
- source validation and evidence retention.

### Recovery-critical work

Changes affecting any of the following require explicit compatibility, destructive/failure-path, and restoration evidence before they may influence production:

- repository format;
- encryption or key derivation;
- content addressing;
- deduplication;
- snapshot serialization;
- storage-provider writes;
- retention deletion;
- garbage collection;
- repository maintenance/compaction;
- restore semantics;
- credential/key migration that could make existing recovery points unreadable.

Branding or product independence is not sufficient justification for changing a recovery-critical format or algorithm.

## Planned protection states

The GoreeCloud product layer is expected to distinguish at least:

- **Unprotected**;
- **Configured**;
- **Backing Up**;
- **Protected**;
- **Restore Verified**;
- **Degraded**.

A snapshot-success event alone must not imply `Restore Verified`. A security scanner pass or Wardveil presentation must not imply `Protected` or verified recoverability either.

## Development data

Use synthetic or disposable repositories and data whenever possible. Do not use the only production backup repository as a development test target. Do not write tests that require real production credentials, private keys, backup passwords, or protected file contents.

When a representative recovery test must eventually use target-environment data, minimize exposed information, use approved credentials, and follow the backup/recovery and privacy policies.

## Source-control and secret boundary

Reusable secrets must remain outside ordinary source control. Source examples use placeholders. The GoreeCloud security validator rejects selected secret-bearing paths and high-confidence reusable-secret patterns in GoreeCloud-owned changes, but automated scanning is an additional control rather than permission to commit secrets.

A secret accidentally committed must be treated as exposed and rotated or revoked according to the applicable GoreeCloud credential policy.

## Glaze UI development

GoreeCloud-controlled UI must follow the canonical Glaze UI source and `docs/goreecloud/GLAZE_UI_CONFORMANCE.md`.

Source conformance does not replace runtime review. Release acceptance must include light/dark, adaptive ranges, keyboard focus, reduced motion/transparency, increased contrast, forced colors, loading, success, empty, warning, degraded, denied-access, error, destructive, and recovery-sensitive states.

The current server-side overlay is a compatibility bridge. Direct GoreeCloud frontend-source ownership remains the long-term mechanism for full product-wide presentation control.

## Wardveil Security development

Wardveil Security by GoreeCloud is the official security/protection identity. It is not a replacement security implementation or a second design system.

Wardveil-facing UI must:

- remain Glaze UI;
- identify the underlying technical authority producing security state;
- avoid invented or unapproved Wardveil module names;
- avoid generalized `Protected by Wardveil` claims when evidence supports only a narrower control;
- avoid treating process health, successful authentication, successful snapshot creation, or a clean scanner result as verified recovery;
- fail visibly and safely when security state is unavailable or degraded.

## Observability and error handling

Development must follow `docs/goreecloud/OBSERVABILITY.md`.

Errors should preserve safe diagnostic context without exposing credentials, cookies, authorization headers, protected backup contents, or untrusted raw backend data to routine logs or user-facing messages.

Retries must be bounded and used only when retrying is safe. Persistent failures must become visible. Loading or long-running operations must not silently stall.

## Integration development

GoreeCloud Manager, Monitor, Notify, Identity, API, CLI, and other integrations must be explicit maintained contracts. They must not obtain privileged access merely because they are GoreeCloud components.

Integration work should:

- use the least privilege required for the documented purpose;
- avoid direct coupling to recovery-critical repository internals when an API/event boundary is practical;
- define authentication, authorization, timeout, retry, failure, compatibility, and privacy behavior;
- expose only the data needed by the consumer;
- preserve the source system as technical authority;
- include synthetic or disposable integration tests before target acceptance.

## Known source blockers

The previously recorded HTTP authentication logging and short-term authentication-cookie source blockers have been remediated in the maintained branch and now have focused regression coverage plus Wardveil source-security enforcement.

The current source/repository blocker is:

- Dependency Review cannot operate until GitHub Dependency Graph is enabled for the fork.

Separate release blockers remain outside that source defect: approved canonical GoreeCloud Backup artwork has not yet been supplied, representative packaged-runtime/Glaze accessibility acceptance remains outstanding, and target-environment backup/restore/monitoring acceptance is still required.

Any new source blocker must remain visible in `SECURITY.md`, `docs/goreecloud/OBSERVABILITY.md`, and `docs/goreecloud/PRODUCTION_READINESS.md` until corrected or covered by an explicit approved exception. Convenience is not an exception.

## Pull-request acceptance

A GoreeCloud Backup pull request should remain draft while material source or release blockers are unresolved or exact-head validation is incomplete.

Applicable acceptance evidence includes:

- intentional scope and recovery boundary documented;
- required upstream attribution/provenance preserved;
- formatter/lint compliance;
- unit/integration/end-to-end tests appropriate to the change;
- compatibility tests for inherited behavior;
- race/platform/build checks where applicable;
- module integrity and vulnerability/dependency scanning;
- GoreeCloud UI conformance checks for presentation work;
- GoreeCloud/Wardveil security checks for security-sensitive work;
- explicit documentation of any remaining target/manual gate;
- rollback/recovery plan for material changes.

A workflow that is unavailable, skipped unexpectedly, cancelled by a newer head, or unable to evaluate required evidence is not automatically a pass. Release evidence must correspond to the exact candidate commit whose readiness is being assessed.

## Production-readiness authority

The complete release evidence model is defined in `docs/goreecloud/PRODUCTION_READINESS.md`.

Stable classification requires more than source readiness. Target-environment acceptance must prove applicable authentication/authorization, repository access, backup scheduling, retention, monitoring, notifications, integrity/maintenance, representative restoration, ownership/permission recovery, and rollback/recovery behavior.

Production Kopia remains authoritative until the replacement is proven rather than merely implemented.
