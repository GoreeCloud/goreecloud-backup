# GoreeCloud Backup

GoreeCloud Backup is the GoreeCloud-maintained backup and recovery platform built on the mature [Kopia](https://github.com/kopia/kopia) codebase.

> **Development status:** active maintained-fork development. The GoreeCloud product and Glaze UI source boundary is established and automated, but representative runtime visual acceptance, unique product artwork, and target-environment recovery acceptance remain required before a stable release or production cutover.

## Product role

GoreeCloud Backup is designed to protect GoreeCloud information while making recoverability—not merely successful snapshot creation—the primary operational outcome.

The long-term product layer is intended to provide clear protection state, repository health, retention visibility, restore verification, recovery evidence, guided recovery workflows, and controlled integrations with the wider GoreeCloud platform while preserving a narrowly bounded recovery engine.

The planned protection vocabulary is:

- **Unprotected** — no approved protection is configured;
- **Configured** — protection is configured but required successful evidence is incomplete;
- **Backing Up** — an approved backup operation is active;
- **Protected** — current backup and policy evidence meet the defined protection threshold;
- **Restore Verified** — representative recovery has been successfully validated;
- **Degraded** — required protection or recovery evidence is incomplete, stale, or failing.

A successful snapshot does not by itself imply **Restore Verified**.

## Current implementation

The maintained-fork foundation currently provides:

- explicit GoreeCloud Backup product identity and fork provenance;
- controlled upstream synchronization rules in [UPSTREAM.md](UPSTREAM.md);
- maintained-fork development and recovery-safety rules in [docs/goreecloud/DEVELOPMENT.md](docs/goreecloud/DEVELOPMENT.md);
- exact inherited frontend provenance and a controlled transition path in [docs/goreecloud/FRONTEND.md](docs/goreecloud/FRONTEND.md);
- a GoreeCloud-owned presentation overlay around the inherited HTML UI boundary;
- locally vendored canonical Glaze UI 1.0 web and accessibility foundations;
- application-specific Glaze treatment across navigation, forms, cards, tables, dialogs, menus, status feedback, loading indicators, destructive actions, and responsive data surfaces;
- Compact, Medium, Expanded, and Wide adaptive source rules;
- keyboard focus, practical target sizing, coarse-pointer support, reduced motion, reduced transparency, increased contrast, forced-colors handling, no-blur fallbacks, and print resilience;
- no new remote UI dependency, analytics, tracker, advertising, or third-party-font requirement in the GoreeCloud-owned Glaze layer;
- a fail-closed Glaze source conformance validator and dedicated GitHub Actions validation gate;
- an inherited end-to-end title/security assertion reconciled to the GoreeCloud Backup product identity while preserving HTML-escaping coverage.

The current Glaze implementation and remaining visual acceptance gates are recorded in [docs/goreecloud/GLAZE_UI_CONFORMANCE.md](docs/goreecloud/GLAZE_UI_CONFORMANCE.md).

## Glaze UI

GoreeCloud Backup targets **Glaze UI 1.0** using the canonical GoreeCloud design-system revision recorded in the conformance document.

The application follows the Glaze surface hierarchy of **Canvas, Solid, Raised, Glaze, and Overlay** and uses semantic design roles for color, depth, radius, spacing, target sizing, focus, motion, and status presentation.

Current automated source conformance does **not** replace representative runtime acceptance. Before visual completion, the application still requires:

- representative light and dark runtime review;
- Compact, Medium, Expanded, and Wide runtime review;
- keyboard, reduced-motion, reduced-transparency, increased-contrast, and forced-colors acceptance;
- runtime review of loading, empty, warning, error, destructive, and recovery-oriented states;
- a unique canonical GoreeCloud Backup application icon and derived favicon/launcher assets;
- removal or approved documentation of remaining production-visible upstream branding;
- completion of the controlled frontend-source ownership transition where required for full presentation control.

## Upstream foundation

This repository is forked from [`kopia/kopia`](https://github.com/kopia/kopia). Kopia provides the underlying encrypted snapshot, deduplication, compression, repository, storage-backend, CLI, server, and current frontend foundations that GoreeCloud Backup initially inherits.

The exact fork baseline and upstream-maintenance rules are recorded in [UPSTREAM.md](UPSTREAM.md).

Required upstream copyright, license, notice, provenance, and attribution remain preserved. GoreeCloud product identity does not rewrite upstream authorship.

## Architecture and product boundaries

Recovery-critical behavior and GoreeCloud product behavior are deliberately separated where practical.

### Recovery engine boundary

Repository format, encryption and key derivation, content addressing, deduplication, snapshot serialization, storage writes, retention deletion, garbage collection, maintenance compaction, provider behavior, and restore semantics are treated as high-risk recovery-engine concerns.

Changes in that boundary require explicit compatibility, failure, and restoration evidence before they may influence production.

### GoreeCloud product boundary

GoreeCloud-owned development is expected to concentrate on:

- product identity and Glaze UI;
- protection-state evaluation;
- repository and snapshot health presentation;
- integrity and retention evidence;
- guided restore and restore-testing workflows;
- recovery-evidence recording;
- diagnostics and operational readiness;
- controlled integration APIs and events;
- packaging, release engineering, and long-term maintainability.

## Planned GoreeCloud integrations

Integrations are intentionally service-specific rather than a shared-database design.

- **GoreeCloud Manager** — protection summary, repository health, restore-verification state, and administrative deep links.
- **GoreeCloud Monitor** — independent availability and backup/verification health signals; Monitor must not become the backup authority.
- **GoreeCloud Notify** — structured backup failure, degradation, verification, and recovery-related events with no secret-bearing payloads.
- **GoreeCloud Identity** — future authentication/authorization integration only when its security model is appropriate for highly sensitive backup administration and a recovery-access path remains documented.

These integrations remain implementation targets unless explicitly identified elsewhere as completed. They must use governed interfaces rather than direct coupling to recovery-critical internal storage.

## Recovery safety boundary

Development in this repository must not be treated as authorization to modify, remove, migrate, or replace an existing production Kopia repository.

A future production cutover requires separate target-environment acceptance, including representative restoration and recovery-evidence collection. Existing recovery points remain authoritative until a replacement has been proven.

No Glaze UI, branding, build, or source-conformance success can substitute for restore validation.

## Validation

The repository retains inherited Kopia validation and adds a GoreeCloud-owned presentation gate.

The GoreeCloud UI workflow currently performs:

1. exact source checkout through SHA-pinned GitHub Actions;
2. Go module verification;
3. fail-closed Glaze UI source-contract validation;
4. server-package tests for the presentation boundary.

The broader inherited build, lint, test, coverage, compatibility, race-detector, HTML UI, licensing, and platform workflows remain important acceptance evidence and are not weakened merely to make a fork-specific change pass.

Dependency Review is also intentionally retained. If GitHub repository configuration prevents it from running, that configuration issue remains a readiness item rather than a reason to remove the security gate.

## Building

The inherited Kopia build infrastructure remains in place during the controlled fork phase. See [BUILD.md](BUILD.md) for the current upstream-derived build process.

GoreeCloud-specific build, packaging, artifact identity, and release changes will be added incrementally and validated before they are relied upon for production artifacts.

## Licensing and attribution

The inherited Kopia code is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).

The locally vendored Glaze UI reference foundation retains its canonical MIT license in the application source boundary.

GoreeCloud Backup preserves required upstream copyright, license, notice, provenance, and attribution information. Product branding may change where permitted, but upstream authorship and licensing must not be obscured.

## Upstream Kopia project

For upstream Kopia documentation, releases, support resources, and contribution guidance, use the official Kopia project resources:

- [Kopia repository](https://github.com/kopia/kopia)
- [Kopia documentation](https://kopia.io/docs/)
- [Kopia website](https://kopia.io/)

GoreeCloud Backup is independently maintained by GoreeCloud and is not the official Kopia distribution.
