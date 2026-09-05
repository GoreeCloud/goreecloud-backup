# GoreeCloud Backup — Competitive Positioning

## Purpose

This document defines how GoreeCloud Backup should be evaluated against other backup and recovery products without allowing competitive pressure to weaken recovery safety, privacy, portability, or independence.

GoreeCloud Backup is not intended to win by creating the largest possible feature list. Its competitive value must come from making recovery status understandable, testable, and demonstrable while retaining a mature backup-engine foundation.

## Benchmark set

### Primary direct benchmarks

The primary direct benchmark products and stacks are:

- Kopia;
- restic with management layers such as Backrest;
- BorgBackup with orchestration layers such as borgmatic;
- Duplicati;
- UrBackup.

These products are useful comparison points for repository architecture, encryption, deduplication, scheduling, retention, administration, restore workflows, multi-host operation, and user experience.

### Adjacent recovery benchmarks

Proxmox Backup Server, Proxmox snapshots, TrueNAS/ZFS snapshots, and application-native backup/export mechanisms are adjacent recovery benchmarks rather than products GoreeCloud Backup must replace.

GoreeCloud Backup should remain complementary to whole-machine, storage-level, and application-native recovery mechanisms when those mechanisms solve a different failure class more effectively.

## Closest architectural benchmark

Backrest/restic is an especially important architectural benchmark because it demonstrates the value of placing a management, scheduling, repository-health, and restore layer over a mature backup engine.

That comparison establishes an important product constraint: GoreeCloud Backup must become more than a re-skinned Kopia interface.

Its independent value must come from GoreeCloud-controlled recovery policy, protection-state evaluation, verification, restore evidence, application-aware protection, integration, and operational clarity.

## Primary competitive differentiators

GoreeCloud Backup should differentiate primarily through recovery confidence rather than backup-job count or interface appearance.

### Explicit protection states

The product layer should distinguish at least:

- Unprotected;
- Configured;
- Backing Up;
- Protected;
- Restore Verified;
- Degraded.

A successful snapshot must not automatically imply verified recoverability.

### Recovery evidence

GoreeCloud Backup should record evidence that representative restoration and validation actually succeeded without unnecessarily retaining restored private information.

Recovery evidence should be available to the product itself and to approved GoreeCloud management integrations without becoming a second copy of protected data.

### Restore testing

Guided restore testing should be a first-class product workflow. Long term, selected restore tests may be scheduled or automated where that can be done safely.

The purpose is to prove recovery paths before a real failure requires them.

### Application Protection Profiles

Reusable Application Protection Profiles should be able to define a complete protection contract including:

- protected paths;
- database exports or checkpoints;
- application-native exports;
- consistency actions;
- exclusions;
- pre-backup and post-backup actions;
- backup frequency;
- retention;
- repository assignment;
- verification requirements;
- restore procedure;
- restore-validation requirements;
- monitoring requirements;
- notification behavior.

This should allow GoreeCloud applications and common self-hosted services to describe how they are safely protected and recovered rather than treating every workload as an arbitrary directory tree.

### Application-consistent recovery

The product must recognize that a filesystem copy is not automatically an application-consistent backup.

Where appropriate, protection should coordinate database dumps, checkpoints, native exports, filesystem freeze operations, service pauses, container lifecycle actions, or other documented consistency methods.

### Multi-repository recovery architecture

GoreeCloud Backup should support multiple repositories when a real recovery, security, ownership, retention, ransomware-resistance, or failure-isolation requirement exists.

Relevant repository roles may include:

- local repositories;
- off-site repositories;
- restricted repositories;
- offline copies;
- immutable storage where appropriate;
- geographically separated repositories.

Repository complexity should exist only when the recovery model justifies it.

### Backup-engine independence

GoreeCloud policy, API, protection-state evaluation, health, verification, restore orchestration, evidence, and integration contracts should remain separated from the underlying engine where practical.

Kopia-derived components may remain permanently when they continue to provide the strongest safe implementation. Engine replacement is an option when justified, not a product objective by itself.

### Independent monitoring and notifications

The backup system should not be the only authority reporting its own health.

GoreeCloud Monitor and other approved independent monitoring paths should eventually be able to observe service availability, repository reachability, scheduled-job completion, missed backups, verification failures, and capacity thresholds.

GoreeCloud Notify should receive structured failure events without requiring privileged repository access.

### Privacy-first self-hosting

GoreeCloud Backup must preserve self-hosted operation and avoid:

- advertising;
- unnecessary telemetry;
- tracking;
- mandatory vendor-cloud dependencies;
- unnecessary third-party runtime services.

Protected filenames, contents, credentials, and recovery material should be exposed only when operationally required.

### Glaze UI operational clarity

Glaze UI should make backup and recovery administration simpler and more approachable without hiding risk.

Visual polish must not convert unknown, degraded, unverified, skipped, unavailable, or stale recovery evidence into a healthy-looking state.

## Competitive review dimensions

Product reviews should compare GoreeCloud Backup across the following dimensions:

- backup-engine reliability and repository integrity;
- encryption and secret handling;
- deduplication and storage efficiency;
- repository portability and backend choice;
- scheduling and retention;
- application consistency;
- restore usability;
- restore testing;
- recovery evidence;
- protection-state accuracy;
- multi-repository support;
- ransomware and destructive-event resistance;
- monitoring and notifications;
- API and CLI quality;
- accessibility and adaptive interface quality;
- privacy and telemetry behavior;
- self-hosting and offline operability;
- vendor and provider independence;
- migration and engine-replacement capability;
- documentation and repeatable recovery procedures.

## Competitive safety boundary

Competitive benchmarking does not authorize recovery-critical change.

The project must not change repository formats, encryption/key derivation, content addressing, deduplication, snapshot serialization, storage writes, retention deletion, garbage collection, maintenance/compaction, provider behavior, restore semantics, or recovery credential compatibility merely to match a competitor.

Any recovery-critical change must independently satisfy the compatibility, failure-path, destructive-path, restoration, rollback, and migration requirements defined by the GoreeCloud Backup recovery-safety model.

## Product target

The long-term competitive target is to combine:

- a mature backup-engine foundation;
- centralized management;
- simple administration;
- application-aware protection;
- verifiable restoration;
- recovery evidence;
- portable repositories;
- independent monitoring;
- privacy-first operation;
- independently controlled recovery architecture.

GoreeCloud Backup succeeds competitively when it can demonstrate that protected information is recoverable, portable, understandable, and independently controlled—not merely that backup jobs completed.
