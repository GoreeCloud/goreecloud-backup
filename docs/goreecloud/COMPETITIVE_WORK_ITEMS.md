# GoreeCloud Backup — Competitive Work Items

This backlog converts the competitive-positioning requirements into implementation-sized GoreeCloud Backup work while preserving the recovery-critical safety boundary.

## Priority 1 — Protection-state domain model

Implement a GoreeCloud-controlled product-layer state model for:

- Unprotected;
- Configured;
- Backing Up;
- Protected;
- Restore Verified;
- Degraded.

Required behavior:

- snapshot success alone cannot produce `Restore Verified`;
- unavailable, stale, skipped, or failed required evidence must not be treated as passing;
- state transitions must be deterministic and testable;
- the model must remain separate from Wardveil presentation and from inherited Kopia repository internals;
- future API consumers must be able to read state without direct repository access.

## Priority 2 — Recovery-evidence model

Create a privacy-conscious evidence model capable of recording:

- protected system or dataset identity;
- repository identity;
- snapshot/recovery-point identity;
- backup completion status;
- integrity result;
- restore-test date and type;
- validation result;
- bounded failure category;
- software version;
- applicable protection policy;
- applicable recovery procedure.

Do not store restored private contents merely to prove that a restore succeeded.

## Priority 3 — Guided restore verification

Design and implement a guided restore-test workflow that can:

1. select a protected system and recovery point;
2. restore selected representative data to an isolated or alternate destination;
3. validate expected hashes, metadata, ownership, permissions, or application behavior where appropriate;
4. record recovery evidence;
5. clean up temporary recovery data after approved validation;
6. update protection state only from authoritative results.

Long-term automation may be added only after the guided workflow and destructive/failure behavior are proven safe.

## Priority 4 — Application Protection Profiles

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

Build a controlled adapter boundary for common consistency mechanisms such as:

- PostgreSQL exports;
- SQLite-safe backup patterns;
- application-native exports;
- filesystem freeze/thaw;
- container or service lifecycle hooks when required;
- custom pre/post commands with strict validation and least privilege.

The adapter model must make the consistency method visible to the administrator and to recovery evidence.

## Priority 6 — Engine abstraction boundary

Define stable GoreeCloud interfaces around the inherited backup engine for:

- repository discovery and health;
- snapshot listing;
- backup execution status;
- retention status;
- verification;
- restore planning and execution;
- diagnostics.

The abstraction must not intentionally hide recovery-critical behavior from validation. Its purpose is to prevent GoreeCloud integrations and product logic from becoming unnecessarily coupled to Kopia-specific command or internal implementation details.

## Priority 7 — Independent observability integration

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

Create a repeatable internal review process using the dimensions in `COMPETITIVE_POSITIONING.md`.

At each material product milestone, record whether GoreeCloud Backup has improved, regressed, or remains intentionally different in:

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

## Recovery-critical exclusion

None of these work items authorizes changes to repository format, encryption/key derivation, content addressing, deduplication, snapshot serialization, provider writes, retention deletion, garbage collection, maintenance/compaction, or restore semantics.

Any change to those areas requires a separate recovery-critical design and acceptance path with compatibility, migration, rollback, failure-path, and representative-restoration evidence.
