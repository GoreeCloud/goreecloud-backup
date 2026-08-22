# GoreeCloud Backup — Features

## Status and interpretation

This file is the repository-level capability inventory for GoreeCloud Backup.

A feature listed here as **implemented in source** exists on the current development branch, but that does not make it Stable or production accepted. GoreeCloud Backup remains an active maintained fork and is not approved to replace the existing production Kopia deployment.

A feature listed as **partial** has implemented foundations but is not yet a complete user-facing or operational capability. A feature listed as **planned** is a product objective and must not be represented as current behavior.

## Current implemented source capabilities

### Maintained-fork foundation

- GoreeCloud Backup product identity on top of the Kopia upstream foundation.
- Preserved upstream licensing, attribution, provenance, and synchronization boundaries.
- Documented recovery-critical change boundary separating product-layer work from repository-format, encryption/KDF, content-addressing, deduplication, snapshot-serialization, provider-write, retention-deletion, garbage-collection, maintenance, and restore-semantic changes.
- Controlled upstream synchronization guidance.
- Exact-candidate release-evidence policy: older or superseded CI results are not treated as validation for a newer source head.

### Inherited backup-engine capabilities

The maintained fork currently inherits Kopia's mature backup-engine foundation, including:

- encrypted backup repositories and snapshots;
- content-addressed storage and deduplication;
- compression;
- incremental snapshot behavior;
- repository maintenance and retention mechanisms;
- granular file and directory backup and restore;
- CLI and server foundations;
- multiple local, network, object-storage, and supported remote storage backends inherited from Kopia;
- repository-server and current inherited frontend foundations.

These are inherited engine capabilities, not proof that the GoreeCloud product layer has completed target-environment acceptance.

### Glaze UI source foundation

- GoreeCloud-controlled presentation layer around the inherited HTML UI boundary.
- Locally vendored Glaze UI 1.0 source foundation.
- Canvas, Solid, Raised, Glaze, and Overlay surface hierarchy.
- Semantic design roles for spacing, radius, focus, motion, status, and presentation.
- Compact, Medium, Expanded, and Wide source-level adaptive rules.
- Keyboard-focus handling.
- Practical target sizing and coarse-pointer support.
- Reduced-motion and reduced-transparency handling.
- Increased-contrast and forced-colors handling.
- No-blur fallback behavior.
- Print resilience.
- Source-level treatment for navigation, forms, cards, tables, dialogs, menus, status feedback, loading, and destructive actions.
- Fail-closed GoreeCloud UI source validator and dedicated GitHub Actions gate.
- No new GoreeCloud-owned analytics, advertising technology, remote-font dependency, or remote UI runtime introduced by the Glaze layer.

Representative packaged runtime and full visual/accessibility acceptance are still required.

### Wardveil Security and source hardening

- Wardveil Security by GoreeCloud presentation boundary tied to underlying technical evidence.
- Hardened Electron context isolation, renderer Node boundary, sandboxing, web-security posture, IPC validation, certificate handling, and external-navigation controls.
- Hardened authentication-cookie attributes and JWT validation in the maintained branch.
- Privacy-conscious authentication and CSRF failure logging using bounded structured reason categories instead of secret-bearing token/session output.
- Normal request diagnostics avoid logging raw query strings through the GoreeCloud-owned hardening path.
- Focused GoreeCloud server security regression tests.
- Source-secret and sensitive-path checks.
- `go mod verify` and reachable-vulnerability scanning in the GoreeCloud security workflow.
- Production Electron dependency auditing.
- Deterministic dependency/toolchain evidence retention through GitHub Actions.

A security workflow pass does not imply a backup is Protected or Restore Verified.

### Observability and production-readiness contracts

- Privacy-conscious structured-observability contract.
- Bounded event and failure-category direction for GoreeCloud-owned telemetry.
- Explicit separation between source validation, packaging/runtime acceptance, target-environment acceptance, and representative recovery evidence.
- Production-readiness documentation preserving unresolved blockers instead of silently waiving them.

### Protection-state and recovery-evidence domain foundation

The branch contains a GoreeCloud-owned product-layer package at `internal/goreecloud/protection` that implements:

- `Unprotected` state;
- `Configured` state;
- `Backing Up` state;
- `Protected` state;
- `Restore Verified` state;
- `Degraded` state;
- deterministic bounded reason codes;
- a model-owned baseline required-evidence set;
- missing, failed, and stale evidence reporting;
- conservative state precedence;
- explicit separation of operational protection evidence from restore-verification evidence;
- prevention of snapshot/recovery-point success alone producing `Protected` or `Restore Verified`;
- bounded recovery-evidence records;
- bounded restore-test types, validation checks, and failure categories;
- recovery-evidence validation that intentionally has no field for restored private contents, reusable credentials, tokens, encryption keys, or raw backend output;
- automated unit tests for the core state and evidence rules.

This is currently a source-level domain foundation. It is not yet persisted, exposed through a stable product API, connected to the UI, or fed by production backup/repository/monitoring systems.

## Experimental or partial features

### GoreeCloud-owned frontend transition

The current server-side presentation overlay is an interim compatibility bridge. Direct GoreeCloud ownership of the editable frontend source/build output remains incomplete.

### Protection-state runtime integration

The protection-state evaluator is implemented and tested at the domain layer, but runtime adapters that collect authoritative repository, backup, integrity, retention, monitoring, notification, credential-recovery, and application-consistency evidence are not yet complete.

### Recovery-evidence persistence and query

The recovery-evidence types and validation contract exist in source. Durable privacy-conscious storage, retention, querying, audit behavior, migration, API exposure, and administrator workflows remain to be implemented.

### Canonical visual identity

The repository contains the icon contract and release validator, but approved canonical GoreeCloud Backup artwork and all required derivatives are not yet supplied. Inherited compatibility icons must not be represented as approved GoreeCloud Backup artwork.

## Planned features

### Protection management

- Protected-system and dataset inventory.
- Protection policy model.
- Policy-specific evidence requirements and freshness thresholds.
- Repository assignment and multi-repository policy.
- Central scheduling.
- Retention administration.
- Activity and protection history.
- Repository-health and capacity status.

### Guided restore and verification

- Guided recovery-point selection.
- Browsable recovery-point contents.
- Safe alternate or isolated restore destinations.
- Representative file restore tests.
- Metadata, ownership, and permission validation.
- Application-dataset recovery tests.
- Application-behavior validation where appropriate.
- Temporary recovery-data cleanup.
- Recorded recovery evidence.
- Scheduled restore verification after guided/manual behavior is proven safe.

### Application Protection Profiles

- Declarative protected paths.
- Database exports or checkpoints.
- Application-native export requirements.
- Exclusions.
- Pre- and post-backup consistency actions.
- Backup frequency.
- Retention.
- Repository assignment.
- Verification requirements.
- Restore procedure references.
- Restore-validation requirements.
- Monitoring and notification requirements.

### Application-consistency adapters

- PostgreSQL export support.
- SQLite-safe backup patterns.
- Application-native exports.
- Filesystem freeze/thaw where justified.
- Controlled container or service lifecycle coordination.
- Strictly validated least-privilege custom hooks where unavoidable.

### Backup-engine abstraction

Stable GoreeCloud product-layer interfaces are planned for:

- repository discovery and health;
- recovery-point listing;
- backup execution state;
- retention state;
- integrity verification;
- restore planning and execution;
- diagnostics.

The abstraction is intended to reduce unnecessary Kopia-specific coupling without hiding recovery-critical behavior from validation.

### GoreeCloud platform integrations

- GoreeCloud Manager protection summaries and administrative deep links.
- GoreeCloud Monitor independent health and execution signals.
- GoreeCloud Notify structured backup/recovery events.
- Future GoreeCloud Identity authentication/authorization where appropriate to backup sensitivity and recovery access.
- Versioned GoreeCloud API.
- `goreecloud-backup` CLI product interface.

### Packaging and clients

- Approved GoreeCloud Backup application icon and derivatives.
- Representative Linux desktop/package acceptance.
- Representative Windows desktop acceptance.
- Representative macOS desktop acceptance.
- Future Android client only if justified by the product role; no Android source tree exists in the current checkpoint.

### Recovery resilience

- Multiple independent repositories where policy requires them.
- Offline or immutable recovery copies where appropriate.
- Ransomware-resistant credential and permission separation.
- Recovery-credential preservation and validation.
- Repository/provider migration workflows.
- Engine-migration planning and validation if future evidence justifies replacing inherited subsystems.

## Deprecated or removed features

No GoreeCloud Backup product-layer feature has yet been formally deprecated as a released GoreeCloud capability.

Inherited Kopia behavior may be replaced or removed only through the controlled maintained-fork process with applicable compatibility, migration, rollback, and recovery evidence.

## Production-acceptance boundary

The current branch is not production acceptance.

Before GoreeCloud Backup may replace production Kopia protection, the project still requires applicable target-environment evidence for repository access, encryption, credentials, schedules, retention, maintenance, monitoring, notifications, integrity, multiple recovery points, representative restoration, restored-data validation, ownership and permissions, rollback, and recovery documentation.

Current product objectives are tracked in [`COMPETITIVE-OBJECTIVES.md`](COMPETITIVE-OBJECTIVES.md). Supportable benefits are tracked in [`BENEFITS.md`](BENEFITS.md).
