# GoreeCloud Backup — Competitive Work Items

This backlog converts the competitive-positioning requirements into implementation-sized GoreeCloud Backup work while preserving the recovery-critical safety boundary.

Implementation status in this file is source-level status unless explicitly stated otherwise. Source implementation does not imply Stable classification, packaged-runtime acceptance, target-environment acceptance, or production cutover.

## Priority 1 — Protection-state and policy domain

**Status: source foundation implemented; runtime evidence and policy administration remain incomplete.**

The GoreeCloud-controlled product-layer state model now exists in `internal/goreecloud/protection` for:

- Unprotected;
- Configured;
- Backing Up;
- Protected;
- Restore Verified;
- Degraded.

Implemented state behavior includes:

- recovery-point or snapshot success alone cannot produce `Protected` or `Restore Verified`;
- the domain owns a baseline required-evidence set instead of trusting callers to enumerate every required check;
- omitted required evidence is treated as missing;
- stale or failed required evidence produces `Degraded`;
- only explicit passing restore verification can produce `Restore Verified` after required operational evidence passes;
- state results expose bounded reason codes and sorted missing/failed/stale evidence identifiers;
- evidence kinds and statuses are bounded and validated;
- state transitions are deterministic and unit tested;
- the model remains separate from Wardveil presentation and inherited Kopia repository internals.

Implemented policy/freshness behavior includes:

- `BaselinePolicy()` defines the minimum GoreeCloud recovery-evidence requirements without inventing arbitrary time limits;
- policy validation rejects removal of any baseline evidence requirement;
- positive `MaxAge` values can tighten freshness for individual required evidence checks;
- a positive restore-verification maximum age can expire a previously passing representative-restore result;
- expired passing evidence is converted to stale evidence and drives `Degraded` state;
- producer-reported stale evidence remains stale even when a policy does not independently calculate age;
- missing timestamps become stale when a positive freshness threshold requires a trustworthy observation time;
- future-dated observations are rejected;
- evaluation receives an explicit evaluation timestamp and does not call the wall clock, making results deterministic and replayable;
- focused tests cover baseline preservation, invalid freshness values, duplicate requirements, operational evidence expiration, restore-verification expiration, missing timestamps, stale producer state, and future observation rejection;
- the dedicated GoreeCloud Protection workflow validates gofmt and runs the recovery-assurance package tests.

Remaining work:

- authoritative runtime collectors for repository, credential, recovery-point, currency, integrity, scope, consistency, retention, maintenance, monitoring, and notification evidence;
- durable policy storage, assignment, versioning, and migration;
- administrator-facing policy editing and validation;
- stable API representation for policy durations and versions rather than exposing internal Go `time.Duration` serialization by accident;
- workload-specific applicability rules that preserve the baseline recovery contract;
- persistent current-state materialization where appropriate;
- stable API and UI exposure;
- target-environment acceptance.

## Priority 2 — Recovery-evidence model

**Status: source schema and validation implemented; persistence and operational evidence collection remain incomplete.**

The product layer now contains privacy-conscious evidence types capable of representing:

- protected system or dataset identity;
- repository identity;
- snapshot/recovery-point identity;
- backup completion status;
- integrity result;
- restore-test date and type;
- bounded validation checks;
- bounded failure category;
- software version;
- applicable protection policy reference;
- applicable recovery procedure reference.

Implemented privacy boundary:

- the model intentionally has no field for restored private contents;
- it has no field for reusable credentials, access tokens, encryption keys, authorization material, or raw backend output;
- passing and failing restore evidence is validated for internal consistency;
- duplicate or unknown validation checks are rejected.

Remaining work:

- durable evidence storage and schema-version/migration policy;
- retention and pruning rules;
- evidence query and history interfaces;
- authorization boundaries;
- Manager/API/UI views;
- collection from real backup, verification, and restore workflows;
- target-environment validation.

## Priority 3 — Guided restore verification

**Status: planned.**

Design and implement a guided restore-test workflow that can:

1. select a protected system and recovery point;
2. restore selected representative data to an isolated or alternate destination;
3. validate expected hashes, metadata, ownership, permissions, or application behavior where appropriate;
4. record recovery evidence;
5. clean up temporary recovery data after approved validation;
6. update protection state only from authoritative results.

Long-term automation may be added only after the guided workflow and destructive/failure behavior are proven safe.

## Priority 4 — Application Protection Profiles

**Status: planned.**

Introduce reusable profiles that can describe a workload protection contract including:

- protected paths;
- database dumps/checkpoints;
- application-native exports;
- exclusions;
- pre-backup actions;
- post-backup actions;
- backup frequency;
- retention;
- repository assignment;
- verification requirements;
- restore procedure;
- restore-validation requirements;
- monitoring requirements;
- notification behavior.

Profiles must be declarative where practical and must not silently execute privileged or destructive hooks without explicit authorization and validation.

## Priority 5 — Application-consistency adapters

**Status: planned.**

Build a controlled adapter boundary for common consistency mechanisms such as:

- PostgreSQL exports;
- SQLite-safe backup patterns;
- application-native exports;
- filesystem freeze/thaw;
- container or service lifecycle hooks when required;
- custom pre/post commands with strict validation and least privilege.

The adapter model must make the consistency method visible to the administrator and to recovery evidence.

## Priority 6 — Engine abstraction boundary

**Status: planned.**

Define stable GoreeCloud interfaces around the inherited backup engine for:

- repository discovery and health;
- snapshot/recovery-point listing;
- backup execution status;
- retention status;
- verification;
- restore planning and execution;
- diagnostics.

The abstraction must not intentionally hide recovery-critical behavior from validation. Its purpose is to prevent GoreeCloud integrations and product logic from becoming unnecessarily coupled to Kopia-specific commands or internal implementation details.

## Priority 7 — Independent observability integration

**Status: observability contract exists; product integration is planned.**

Expose structured, least-privilege events and status suitable for future GoreeCloud Monitor and GoreeCloud Notify integrations.

Required event families should eventually cover:

- backup completed;
- backup failed;
- backup missed;
- repository unreachable;
- repository integrity failed;
- repository maintenance failed;
- unsafe capacity threshold;
- restore verification passed;
- restore verification failed;
- protection state changed.

No event may contain reusable credentials, encryption keys, authorization material, or protected backup contents.

## Priority 8 — Competitive review harness

**Status: repository product records and CI enforcement established; automated comparative review remains planned.**

The repository now maintains:

- root `COMPETITIVE-OBJECTIVES.md`;
- root `FEATURES.md`;
- root `BENEFITS.md`;
- detailed `docs/goreecloud/COMPETITIVE_POSITIONING.md`;
- this implementation backlog;
- a GoreeCloud Product Records CI gate that prevents the required repository records and implementation-state boundaries from disappearing silently.

At each material product milestone, review whether GoreeCloud Backup has improved, regressed, or remains intentionally different in:

- recovery confidence;
- restore usability;
- repository portability;
- application consistency;
- monitoring;
- privacy;
- accessibility;
- independence;
- migration capability.

Competitive review results are planning evidence, not production-readiness evidence.

## Next implementation sequence

With the state, evidence-schema, and source-level freshness-policy foundations now established, the next product-layer sequence should be:

1. define the durable recovery-evidence persistence boundary, including privacy, authorization, retention, schema versioning, migration, and failure behavior;
2. define a narrow read-only engine-adapter interface for repository availability, recovery-point visibility, integrity, and other authoritative evidence before adding mutation or restore execution;
3. define durable protection-policy storage and assignment plus an explicit stable duration/API representation;
4. expose a small internal/status API for protection evaluation without granting consumers direct repository access;
5. begin guided restore-verification design against disposable repositories and isolated destinations;
6. only then connect Manager, Monitor, Notify, and UI consumers to authoritative product-layer state.

This order keeps evidence authority, persistence, and API contracts stable before multiple consumers depend on them.

## Recovery-critical exclusion

None of these work items authorizes changes to repository format, encryption/key derivation, content addressing, deduplication, snapshot serialization, provider writes, retention deletion, garbage collection, maintenance/compaction, or restore semantics.

Any change to those areas requires a separate recovery-critical design and acceptance path with compatibility, migration, rollback, failure-path, and representative-restoration evidence.
