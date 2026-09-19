# GoreeCloud Backup — Competitive Objectives

## Purpose

This record defines the products, projects, and recovery experiences GoreeCloud Backup uses as benchmarks and the capabilities it intends to match, exceed, or deliberately approach differently.

Competitive objectives are product-development targets. They are not claims that every objective is already implemented, validated, Stable, or production accepted.

For the detailed working analysis and implementation backlog, see:

- [`docs/goreecloud/COMPETITIVE_POSITIONING.md`](docs/goreecloud/COMPETITIVE_POSITIONING.md)
- [`docs/goreecloud/COMPETITIVE_WORK_ITEMS.md`](docs/goreecloud/COMPETITIVE_WORK_ITEMS.md)
- [`FEATURES.md`](FEATURES.md)
- [`BENEFITS.md`](BENEFITS.md)

## Primary competitors and benchmarks

### Kopia

Kopia is the current upstream foundation and therefore the most important compatibility and engine-quality benchmark during the maintained-fork phase.

GoreeCloud Backup should preserve or improve the mature capabilities that materially support recoverability, including encrypted repositories and snapshots, deduplication, compression, storage-backend flexibility, repository maintenance, CLI/server operation, and restoration behavior.

Kopia is not treated as something to replace merely for product identity. Recovery-critical replacement requires measurable benefit and representative recovery evidence.

### restic with Backrest-style management

restic is a major engine benchmark for portability, simplicity, repository discipline, encryption, verification, and broad storage support. Backrest is a particularly useful management-layer benchmark because it demonstrates how scheduling, repository health, snapshot browsing, restore workflows, and administration can be layered over a mature engine.

GoreeCloud Backup should exceed the simple engine-plus-dashboard model by making protection state, restore verification, recovery evidence, application consistency, and GoreeCloud integration first-class product concepts.

### BorgBackup with borgmatic-style orchestration

BorgBackup and borgmatic are benchmarks for deduplication, authenticated encryption, efficient repositories, declarative configuration, retention, checks, hooks, database-aware protection, monitoring, and multi-repository administration.

GoreeCloud Backup should match the operational discipline of this model while providing clearer recovery-state semantics and a simpler first-party administration experience through Glaze UI and stable APIs.

### Duplicati

Duplicati is a benchmark for approachable backup administration, broad storage support, scheduling, retention, encryption, multi-machine operation, and centralized policy concepts.

GoreeCloud Backup should aim for equally understandable administration without introducing advertising, tracking, mandatory vendor cloud dependencies, or ambiguous recoverability claims.

### UrBackup

UrBackup is a benchmark for self-hosted client/server backup, multi-system administration, web-based restore workflows, file and image backup concepts, and practical household or small-organization deployment.

GoreeCloud Backup should preserve a narrower recovery-engine boundary where that improves portability and maintainability while providing stronger evidence that protected data can actually be restored.

## Adjacent recovery benchmarks

The following technologies solve complementary failure classes and are not replacement targets:

- Proxmox Backup Server and Proxmox virtual-machine backups;
- Proxmox snapshots;
- TrueNAS and ZFS snapshots;
- application-native database backup and export mechanisms;
- offline recovery media;
- off-site repositories and provider-level disaster-recovery mechanisms.

GoreeCloud Backup should integrate with or coexist with these layers where they provide stronger whole-machine, storage-level, application-consistent, or disaster-recovery protection.

## Capabilities worth matching

GoreeCloud Backup should match mature competitors where the capability materially improves backup or recovery:

- encrypted and authenticated backup repositories;
- deduplication and compression;
- efficient incremental backup behavior;
- broad self-hosted and object-storage repository support;
- multiple repositories and independent copies;
- understandable scheduling and retention;
- repository verification and maintenance;
- snapshot and recovery-point browsing;
- granular restore workflows;
- database- and application-consistent protection;
- hooks or adapters with strict safety boundaries;
- useful CLI and API administration;
- multi-system status visibility;
- actionable monitoring and notifications;
- exportable and portable configuration where practical;
- clear documentation and recovery procedures.

## Capabilities GoreeCloud intends to exceed

### Provable recovery state

The primary competitive objective is to make the difference between "a backup ran" and "recovery was proven" explicit.

The product should provide deterministic states:

- Unprotected;
- Configured;
- Backing Up;
- Protected;
- Restore Verified;
- Degraded.

A successful snapshot must never by itself imply `Restore Verified`.

### Recovery evidence

GoreeCloud Backup should record bounded, privacy-conscious evidence showing why protection or restore verification is considered healthy or degraded without storing private restored content merely to prove that a test occurred.

### Guided and repeatable restore verification

Restore testing should become an ordinary administrative workflow rather than an exceptional manual exercise. Representative restores should support validation of content, metadata, ownership, permissions, and application behavior where appropriate.

### Application Protection Profiles

GoreeCloud applications should eventually be able to describe their own protection contract: protected paths, exports or checkpoints, exclusions, consistency steps, schedule, retention, repository assignment, verification requirements, restore procedure, and validation requirements.

### Independent protection monitoring

The backup engine should not be the only authority reporting its own health. GoreeCloud Monitor and GoreeCloud Notify integrations should provide independent, structured signals without gaining unnecessary repository access.

### Engine independence

GoreeCloud product logic, APIs, UI, and integrations should not become unnecessarily coupled to one backup engine. Kopia remains the mature foundation while it serves GoreeCloud well, but product-layer boundaries should preserve a controlled future replacement path when evidence justifies one.

## Capabilities GoreeCloud intentionally rejects

GoreeCloud Backup should not adopt competitor behavior that conflicts with GoreeCloud principles or recovery safety. It intentionally rejects:

- advertising or sponsorship-driven product behavior;
- mandatory telemetry or user tracking;
- mandatory vendor-hosted control planes when self-hosting is practical;
- proprietary repository formats created only for branding or lock-in;
- claims of recoverability based only on successful job execution;
- hiding failed, stale, skipped, or unavailable validation evidence;
- weakening encryption, verification, retention, or restore safety for interface simplicity;
- silently executing privileged or destructive application hooks;
- coupling unrelated GoreeCloud services directly to recovery-critical repository internals;
- replacing mature recovery-critical formats or algorithms merely to make the product more "native";
- treating whole-machine backup, filesystem snapshots, and application-native exports as redundant when they protect different failure domains.

## Privacy and security objectives

GoreeCloud Backup should compete through privacy and security as operational properties rather than marketing labels:

- self-hosting by default where practical;
- client-side encryption where appropriate;
- least-privilege repository and integration access;
- sensitive-information separation;
- no routine storage of restored private contents in recovery evidence;
- no secrets, tokens, keys, cookies, authorization material, or protected contents in ordinary telemetry;
- bounded structured failure categories instead of raw sensitive backend output;
- independent credentials and repositories where appropriate for ransomware resistance;
- offline, immutable, or geographically separated copies where required by policy;
- explicit recovery-access planning so stronger security does not make legitimate disaster recovery impossible;
- Wardveil Security presentation that remains tied to the technical evidence actually available.

## Ownership, self-hosting, and independence objectives

The long-term product should allow GoreeCloud to:

- choose and change storage providers;
- move repositories without surrendering control of protected information;
- retain recovery capability if an upstream project or service disappears;
- preserve required historical recovery points during migrations;
- operate without a mandatory third-party account or control plane;
- replace supporting components when technically justified without losing protected information;
- preserve open-source licensing, attribution, and transparent development.

## User-experience and accessibility objectives

GoreeCloud Backup should make complex recovery concepts understandable without hiding advanced controls.

Glaze UI should provide:

- simple, polished defaults;
- advanced controls when needed;
- visible protection and degradation state;
- clear distinction between backup success and verified restore;
- safe destructive-action review;
- useful empty, loading, warning, error, denied, and recovery states;
- keyboard accessibility;
- practical pointer/touch targets;
- light and dark presentation;
- reduced-motion and reduced-transparency behavior;
- increased-contrast and forced-colors support;
- responsive Compact, Medium, Expanded, and Wide layouts.

Visual polish must never obscure recovery risk.

## Performance and reliability objectives

GoreeCloud Backup should preserve mature engine efficiency while adding product-layer assurance with minimal unnecessary overhead.

Objectives include:

- incremental operation that avoids retransmitting unchanged data where the engine supports it;
- efficient deduplication and compression;
- bounded background verification and maintenance;
- clear behavior under slow or unavailable repositories;
- bounded retries and visible persistent failure;
- deterministic protection-state evaluation;
- representative recovery testing that can run against isolated destinations;
- continued compatibility with existing recovery points until a separately validated migration is approved.

## Interoperability and administrative-control objectives

The product should expose explicit, versioned, least-privilege interfaces suitable for:

- GoreeCloud Manager;
- GoreeCloud Monitor;
- GoreeCloud Notify;
- future GoreeCloud Identity integration where appropriate;
- CLI administration;
- automation and orchestration;
- future Application Protection Profiles.

These consumers should not require direct access to Kopia repository internals to understand product-layer protection state.

## Data portability and recovery objectives

A successful GoreeCloud Backup implementation should make it possible to determine:

- what is protected;
- what is not protected;
- where recovery copies exist;
- whether each required repository is healthy;
- whether backups are current;
- whether retention and maintenance are healthy;
- whether recovery credentials can be obtained;
- whether representative restoration has succeeded;
- when restore verification last occurred;
- what failed when a protection state is degraded.

Storage or engine migration must preserve protected information, required historical recovery points, and a documented rollback path.

## GoreeCloud differentiators

The intended GoreeCloud-specific combination is:

- recovery-first protection-state semantics;
- explicit `Restore Verified` evidence;
- privacy-conscious recovery records;
- application-aware protection contracts;
- independent monitoring and notification integration;
- engine abstraction and migration discipline;
- Glaze UI operational clarity;
- Wardveil Security evidence-scoped presentation;
- Everkeep alignment for broader resilience and preservation;
- privacy by default;
- no advertising or sponsorships;
- self-hosting and open-source transparency;
- strong ownership and vendor-independence goals.

These differentiators are objectives only where their supporting functionality is not yet implemented. Current implementation state is authoritative in [`FEATURES.md`](FEATURES.md).

## Review rule

This record must be reviewed when:

- a material competitor adds a relevant capability;
- a major GoreeCloud Backup feature is implemented, redesigned, removed, or deprecated;
- the maintained fork moves materially closer to a native product layer;
- a major release is prepared;
- competitive research changes an architectural decision;
- `FEATURES.md` or `BENEFITS.md` becomes inconsistent with this record.

Competitive review is planning evidence. It does not replace security, compatibility, release, or representative recovery acceptance.
