# GoreeCloud Backup — Benefits

## Purpose and claim boundary

This record explains why GoreeCloud Backup's current capabilities matter.

Benefits are limited to what the current source, architecture, and validated controls can support. Planned capabilities are listed separately and must not be represented as current user benefits until their supporting functionality exists and is validated.

GoreeCloud Backup remains in active maintained-fork development and is not approved to replace the existing production Kopia deployment.

## Current supportable benefits

### Recovery safety is protected from branding-driven change

The maintained-fork architecture explicitly separates GoreeCloud product-layer work from recovery-critical repository behavior. This reduces the risk that rebranding, UI redesign, or integration work casually changes repository format, encryption/KDF, content addressing, deduplication, snapshot serialization, provider writes, retention deletion, garbage collection, maintenance, or restore semantics.

For administrators, that means product evolution has a defined boundary around the mechanisms that existing recovery points depend on.

### Mature backup-engine capability is retained while GoreeCloud builds its own product layer

GoreeCloud Backup inherits Kopia's established encrypted-repository, snapshot, deduplication, compression, storage-backend, maintenance, CLI/server, and restore foundations rather than reimplementing a recovery engine solely for ownership or branding.

This provides a safer path to product independence: GoreeCloud can add first-party management and assurance concepts while preserving mature recovery behavior until a replacement is demonstrably better.

### Upstream dependence is visible and governable

The repository records provenance and synchronization boundaries instead of hiding its Kopia origin.

This benefits long-term maintainability by making it possible to review upstream security fixes, repository changes, dependency updates, storage changes, restore behavior, licensing, and other compatibility-sensitive changes deliberately.

### Security-facing claims are evidence scoped

Wardveil Security presentation is explicitly separated from the technical systems that produce security or recovery evidence.

This reduces the risk of misleading statements such as treating a clean source scan, successful login, healthy process, or successful snapshot as proof that protected information is recoverable.

### Authentication and desktop-shell hardening reduce avoidable attack surface

The maintained branch includes GoreeCloud-specific hardening for Electron isolation and navigation boundaries, authentication cookies, JWT validation, CSRF handling, and privacy-conscious authentication/request logging.

These controls reduce unnecessary exposure of browser/desktop privileges and sensitive authentication material while preserving the inherited backup engine.

### Privacy-conscious observability limits unnecessary sensitive data exposure

GoreeCloud-owned observability contracts favor structured bounded events and sanitized failure categories rather than raw secret-bearing or content-bearing diagnostic output.

This helps administrators diagnose failures without making routine logs a secondary repository of credentials, cookies, authorization material, backup contents, or other reusable secrets.

### Source-level protection semantics prevent an important false-positive class

The current protection domain model deliberately prevents a successful recovery point from automatically becoming `Protected` or `Restore Verified`.

The evaluator owns a baseline evidence set and treats omitted required evidence as missing. Failed or stale required evidence becomes `Degraded`, and only explicit passing restore verification can produce `Restore Verified`.

This provides a safer foundation for future UI, API, Manager, Monitor, and Notify consumers because they can be built around deterministic recovery semantics rather than assuming that job success equals recoverability.

### Recovery confidence can expire instead of remaining permanently green

The source-level protection policy can apply explicit positive freshness limits to required operational evidence and restore-verification results. When a previously passing result exceeds the applicable policy age, it becomes stale and the evaluated protection state becomes `Degraded` rather than remaining indefinitely `Protected` or `Restore Verified`.

This prevents a past success from being treated as current recovery confidence forever. The baseline policy does not invent arbitrary freshness durations: a positive age limit must come from an applicable policy, while evidence producers may also report stale state directly.

### Freshness evaluation is deterministic and replayable

Freshness-aware state calculation receives an explicit evaluation timestamp and does not read the system clock directly.

That makes the same evidence and policy produce the same result for a selected point in time. This is useful for tests, future API behavior, incident review, audit history, and replay of historical recovery evidence without time-dependent ambiguity.

### Policies cannot silently weaken the baseline recovery contract

The current source-level policy validator rejects a policy that removes any GoreeCloud baseline evidence requirement. Policies may tighten freshness requirements, but they cannot turn missing core recovery evidence into an intentional pass merely through configuration.

This creates a safer foundation for future administrator policy editing and workload assignment, although those user-facing administration workflows are not yet implemented.

### Recovery-evidence types minimize private-content retention by design

The current recovery-evidence model records bounded identifiers, statuses, validation checks, timestamps, and failure categories without providing fields for restored private file contents, reusable credentials, tokens, encryption keys, or raw backend responses.

This creates a privacy-preserving foundation for proving recovery outcomes without encouraging unnecessary inspection or retention of personal and family information.

### Deterministic state results improve future interoperability

Protection evaluation returns stable states, reason codes, and sorted missing/failed/stale evidence identifiers.

Although the stable product API is not yet implemented, this deterministic source contract reduces ambiguity for future integrations and makes the domain behavior straightforward to test.

### Focused recovery-assurance CI catches domain drift quickly

The dedicated GoreeCloud Protection workflow checks formatting and runs the recovery-assurance package tests whenever the protection domain changes.

This provides a narrow, fast validation path for state, evidence, and policy semantics without replacing the broader inherited test, compatibility, security, packaging, or target-recovery gates.

### Glaze UI provides a consistent accessibility and presentation foundation

The current GoreeCloud-owned UI layer establishes adaptive source rules, visible focus, practical target sizing, reduced-motion/transparency behavior, increased contrast, forced colors, no-blur fallbacks, and consistent application-state presentation.

This provides a common foundation for an administration interface that can make protection and degradation status understandable across devices and accessibility preferences.

Runtime visual/accessibility acceptance is still required before those source-level foundations may be treated as fully validated packaged-product behavior.

### No advertising or tracking dependency was introduced by the GoreeCloud product layer

The GoreeCloud-owned Glaze and product work adds no advertising technology, analytics requirement, remote font dependency, or new remote UI runtime.

This supports the product's privacy and independence goals without claiming that every inherited or future dependency has completed a final privacy audit.

### Exact-candidate validation reduces false release confidence

The repository explicitly distinguishes current-candidate evidence from older workflow results. Cancelled, skipped, unavailable, superseded, or older-head checks are not silently counted as passing validation for a newer candidate.

This makes release decisions more trustworthy and preserves visibility into repository-configuration blockers such as the currently disabled GitHub Dependency Graph.

## Administrator benefits

Current source and governance provide administrators with:

- a documented recovery-critical boundary;
- visible unresolved readiness blockers;
- upstream provenance and synchronization guidance;
- source-level Glaze UI and Wardveil Security contracts;
- privacy-conscious observability rules;
- deterministic protection-state semantics ready for future integration;
- a freshness-aware source policy that can expire old recovery confidence;
- bounded recovery-evidence structures ready for future persistence and API work;
- focused recovery-assurance CI plus automated GoreeCloud UI and security validation in addition to inherited upstream checks.

These are development and governance benefits. They do not yet provide a complete GoreeCloud-native backup-management console.

## Privacy benefits

Current privacy benefits include:

- no GoreeCloud advertising or sponsorship mechanism;
- no new GoreeCloud analytics/tracking requirement;
- no new remote-font requirement in the GoreeCloud UI layer;
- explicit sensitive-data exclusions for routine logging;
- bounded recovery-evidence structures that avoid restored-content fields;
- evidence-scoped security wording;
- minimum-necessary-data direction for future restore testing.

## Security benefits

Current security benefits include:

- least-privilege integration direction;
- hardened Electron boundaries;
- hardened authentication and request-integrity paths in the maintained branch;
- source-secret checks;
- module integrity validation;
- reachable Go vulnerability analysis;
- production Electron dependency auditing;
- deterministic security evidence retention;
- explicit separation between security evidence and recovery evidence;
- a policy-validation boundary that prevents removal of baseline recovery evidence requirements.

## Ownership and independence benefits

Current architecture and governance reduce future lock-in by:

- keeping GoreeCloud product logic conceptually separate from one engine's recovery internals;
- preserving open-source provenance and licensing;
- maintaining an upstream synchronization process;
- defining a controlled fork-to-native path instead of requiring permanent dependence or premature rewrite;
- rejecting proprietary formats created only for product differentiation;
- treating repository/provider/engine migration as recovery-sensitive operations rather than ordinary configuration changes.

A complete engine-independent runtime is still planned, not current.

## Reliability and recovery benefits

The strongest current recovery benefit is architectural discipline: source and documentation consistently require representative restoration rather than successful backup execution as the ultimate evidence of recoverability.

The implemented protection-state and freshness-policy domain layers turn part of that principle into deterministic code. Old passing evidence can become stale, but representative restore orchestration, authoritative runtime evidence collection, durable evidence history, and production recovery acceptance are still future work.

## Accessibility benefits

The Glaze UI source foundation includes accommodations for keyboard use, practical targets, multiple adaptive widths, reduced motion, reduced transparency, increased contrast, forced colors, and no-blur environments.

These source-level accommodations reduce the likelihood that recovery administration becomes dependent on one input method, display size, motion preference, or transparency capability. Packaged runtime acceptance remains outstanding.

## Integration benefits

The product-layer state, policy, and evidence structures are designed so future GoreeCloud consumers can read bounded product state instead of directly coupling to recovery-critical repository internals.

This is expected to make Manager, Monitor, Notify, API, CLI, and future Identity integration safer and easier to evolve. Those integrations are not yet complete.

## Planned benefits not yet claimable as current

The following benefits depend on planned functionality and must remain future-facing until implemented and validated:

- administrator-facing protection-policy creation, editing, assignment, persistence, and versioning;
- routine guided restore testing from the administration interface;
- scheduled automated representative recovery tests;
- application-aware backup contracts through Application Protection Profiles;
- centralized multi-system protection inventory;
- durable recovery-evidence history and audit views;
- independent GoreeCloud Monitor health evaluation;
- GoreeCloud Notify backup/recovery alerts;
- Manager recovery-readiness summaries;
- stable GoreeCloud Backup API and `goreecloud-backup` CLI;
- policy-driven multi-repository orchestration;
- first-party migration between backup engines;
- full packaged desktop acceptance;
- production replacement of Kopia;
- any superiority claim over Kopia, restic/Backrest, Borg/borgmatic, Duplicati, UrBackup, Proxmox Backup Server, or another benchmark that has not been demonstrated through current comparative evidence.

## Relationship to other product records

- [`FEATURES.md`](FEATURES.md) answers what is implemented, partial, or planned.
- [`COMPETITIVE-OBJECTIVES.md`](COMPETITIVE-OBJECTIVES.md) answers what GoreeCloud Backup is trying to learn from, match, exceed, or deliberately approach differently.
- [`docs/goreecloud/COMPETITIVE_POSITIONING.md`](docs/goreecloud/COMPETITIVE_POSITIONING.md) contains detailed working competitive context.

Benefits must be revised whenever feature implementation state changes materially.
