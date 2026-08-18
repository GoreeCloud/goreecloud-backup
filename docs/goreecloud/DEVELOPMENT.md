# GoreeCloud Backup Development Model

## Purpose

This document defines the initial development boundary for the GoreeCloud-maintained Kopia fork.

The first development objective is not to maximize visible divergence from upstream. It is to create a controlled, reviewable foundation that can evolve without weakening backup or restore reliability.

## Development principles

1. **Recovery before branding.** User-facing identity may change, but changes must not obscure whether the underlying recovery behavior is still inherited from Kopia.
2. **No production cutover by implication.** A successful build, test, or UI change does not authorize migration of an existing production repository.
3. **Preserve repository compatibility where practical.** Changes to repository-critical code require explicit compatibility review.
4. **Inspect before replacing.** Mature upstream subsystems are replaced only when GoreeCloud gains a documented benefit.
5. **Separate product and engine layers.** GoreeCloud-specific management, verification, API, UI, and integration behavior should be isolated from recovery-critical engine code where practical.
6. **Record upstream provenance.** Upstream baseline and synchronization decisions remain documented.
7. **Keep secrets out of source.** Repository passwords, encryption keys, tokens, signing credentials, and provider credentials must never be committed.
8. **Use pull requests for material changes.** Development should remain reviewable and reversible.

## Planned protection states

The GoreeCloud product layer is expected to distinguish at least:

- Unprotected
- Configured
- Backing Up
- Protected
- Restore Verified
- Degraded

A snapshot-success event alone must not automatically imply `Restore Verified`.

## Initial development sequence

### Foundation

- establish GoreeCloud repository identity;
- record upstream baseline and licensing;
- preserve upstream buildability;
- introduce GoreeCloud-specific documentation and safety boundaries;
- identify UI, CLI, server, repository, and storage-backend boundaries.

### Product shell

- introduce GoreeCloud Backup product metadata;
- begin Glaze UI integration;
- establish accessibility and appearance foundations;
- preserve required upstream attribution;
- keep recovery-critical engine behavior unchanged unless separately justified.

### Protection model

- protected-system inventory;
- repository health;
- snapshot recency;
- retention status;
- integrity verification;
- restore-test evidence;
- degraded-state reasons.

### Recovery workflows

- guided file restore;
- alternate-location restore;
- test restore;
- application-consistent recovery workflows;
- ownership and permission validation;
- recovery-evidence recording.

### GoreeCloud integrations

- GoreeCloud Manager summary integration;
- GoreeCloud Monitor independent health integration;
- GoreeCloud Notify structured failure events;
- GoreeCloud Identity integration only where its authorization model is suitable for highly sensitive backup administration.

## Change-risk classes

### Low risk

Examples:

- documentation;
- non-runtime metadata;
- tests that do not alter backup behavior;
- isolated design tokens;
- additive diagnostics.

### Medium risk

Examples:

- UI workflows;
- CLI wrappers;
- API layers;
- scheduler changes;
- policy interpretation;
- packaging metadata.

These require functional validation and regression testing.

### High risk

Examples:

- repository format;
- encryption or key derivation;
- content addressing;
- deduplication;
- snapshot serialization;
- storage writes;
- retention deletion;
- garbage collection;
- maintenance compaction;
- restore semantics.

High-risk changes require explicit design review, compatibility testing, failure testing, and representative restoration before they may influence production.

## Pull-request acceptance gate

A development pull request should state:

- what subsystem is affected;
- whether repository compatibility may change;
- whether existing recovery data is touched;
- expected security/privacy impact;
- tests performed;
- rollback approach;
- unresolved limitations.

No pull request should claim production readiness unless target-environment recovery acceptance has actually been completed.
