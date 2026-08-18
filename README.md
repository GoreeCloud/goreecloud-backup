# GoreeCloud Backup

GoreeCloud Backup is the GoreeCloud-maintained backup and recovery platform built on the mature [Kopia](https://github.com/kopia/kopia) codebase.

> **Development status:** active maintained-fork development. This candidate is **not Stable and is not approved to replace an existing production Kopia deployment**. The GoreeCloud product, Glaze UI, Wardveil Security, security-evidence, and operational-governance boundaries are established, but source blockers, representative runtime acceptance, unique product artwork, integrations, and target-environment recovery acceptance remain required.

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
- a GoreeCloud/Wardveil security policy and fail-closed source validator;
- deterministic dependency/toolchain evidence retention in the GoreeCloud Security workflow;
- hardened Electron renderer, IPC, certificate, and navigation boundaries;
- a privacy-conscious observability/audit contract in [docs/goreecloud/OBSERVABILITY.md](docs/goreecloud/OBSERVABILITY.md);
- a fail-closed production-readiness evidence contract in [docs/goreecloud/PRODUCTION_READINESS.md](docs/goreecloud/PRODUCTION_READINESS.md);
- an inherited end-to-end title/security assertion reconciled to the GoreeCloud Backup product identity while preserving HTML-escaping coverage.

The current Glaze implementation and remaining visual acceptance gates are recorded in [docs/goreecloud/GLAZE_UI_CONFORMANCE.md](docs/goreecloud/GLAZE_UI_CONFORMANCE.md).

## Glaze UI

GoreeCloud Backup targets **Glaze UI 1.0** using the canonical GoreeCloud design-system revision recorded in the conformance document.

The application follows the Glaze surface hierarchy of **Canvas, Solid, Raised, Glaze, and Overlay** and uses semantic design roles for color, depth, radius, spacing, target sizing, focus, motion, and status presentation.

Current automated source conformance does **not** replace representative runtime acceptance. Before visual completion, the application still requires:

- representative light and dark runtime review;
- Compact, Medium, Expanded, and Wide runtime review;
- keyboard, reduced-motion, reduced-transparency, increased-contrast, and forced-colors acceptance;
- runtime review of loading, empty, warning, error, denied, destructive, and recovery-oriented states;
- a unique canonical GoreeCloud Backup application icon and derived favicon/launcher/PWA/packaging assets;
- removal or approved documentation of remaining production-visible upstream branding;
- completion of the controlled frontend-source ownership transition where required for full presentation control.

## Wardveil Security

**Wardveil Security by GoreeCloud** is the platform-wide security and protection identity used by GoreeCloud Backup when presenting security posture, safeguards, security alerts, or protection controls.

Wardveil does not replace the technical authority that produced a state. Repository encryption, authentication, authorization, integrity checks, vulnerability scanners, monitoring, policy, and restore evidence remain authoritative for their own subjects.

The approved `Protected by Wardveil` phrase must be evidence-scoped. The application shell deliberately does not use it as a blanket claim merely because the source passes a security workflow. A scanner pass, healthy process, successful login, successful snapshot, or Glaze UI presentation is not proof of verified recovery.

Wardveil-facing surfaces must use Glaze UI and preserve the originating component/control so security branding never hides technical authority.

## Security, privacy, and observability

Security policy is defined in [SECURITY.md](SECURITY.md). Operational and audit logging rules are defined in [docs/goreecloud/OBSERVABILITY.md](docs/goreecloud/OBSERVABILITY.md).

The GoreeCloud-owned security gate currently includes:

- `go mod verify`;
- `govulncheck` reachable-vulnerability analysis;
- production Electron dependency auditing;
- source-secret/path checks and policy-drift validation;
- Electron security-boundary checks;
- Glaze/Wardveil shell privacy checks;
- deterministic Go/Node/npm and dependency-input evidence retained through GitHub Actions.

Routine logs must not contain passwords, backup passwords, access tokens, private keys, raw authorization/cookie values, protected file contents, or other reusable secret material. Authentication logging must also avoid echoing untrusted submitted credential identifiers merely for convenience.

## Known release blockers

The following items are explicitly unresolved and prevent Stable classification:

- the inherited HTTP authentication path in `internal/server/server.go` still logs remote address and submitted username on failed login and may log remote address plus username on successful login when request logging is enabled;
- the inherited short-term authentication cookie requires explicit review and regression evidence for `HttpOnly`, `Secure`, and an intentional `SameSite` policy;
- GitHub Dependency Review cannot currently operate because Dependency Graph is disabled for this fork;
- unique GoreeCloud Backup icon/favicon/launcher/PWA/packaging artwork remains required;
- representative packaged desktop validation and Glaze runtime/visual/accessibility acceptance remain required;
- planned GoreeCloud Manager, Monitor, Notify, API/CLI, and future Identity integrations remain governed implementation targets rather than completed claims;
- target-environment repository, scheduling, retention, monitoring, notification, rollback, and representative restore acceptance remain required.

These blockers are intentionally visible in source governance rather than being waived by branding, UI polish, or a partially green CI matrix.

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
- Wardveil security/status presentation backed by actual control evidence;
- protection-state evaluation;
- repository and snapshot health presentation;
- integrity and retention evidence;
- guided restore and restore-testing workflows;
- recovery-evidence recording;
- privacy-conscious observability and diagnostics;
- controlled integration APIs and events;
- packaging, release engineering, and long-term maintainability.

## Planned GoreeCloud integrations

Integrations are intentionally service-specific rather than a shared-database design.

- **GoreeCloud Manager** — protection summary, repository health, restore-verification state, and administrative deep links.
- **GoreeCloud Monitor** — independent availability and backup/verification health signals; Monitor must not become the backup authority.
- **GoreeCloud Notify** — structured backup failure, degradation, verification, and recovery-related events with no secret-bearing payloads.
- **GoreeCloud Identity** — future authentication/authorization integration only when its security model is appropriate for highly sensitive backup administration and a recovery-access path remains documented.
- **GoreeCloud API/CLI** — explicit, least-privilege, documented contracts rather than direct coupling to recovery-critical internal storage.

These integrations remain implementation targets unless explicitly identified elsewhere as completed.

## Recovery safety boundary

Development in this repository must not be treated as authorization to modify, remove, migrate, or replace an existing production Kopia repository.

A future production cutover requires separate target-environment acceptance, including representative restoration and recovery-evidence collection. Existing recovery points remain authoritative until a replacement has been proven.

No Glaze UI, Wardveil Security presentation, vulnerability scan, branding, build, or source-conformance success can substitute for restore validation.

## Validation

The repository retains inherited Kopia validation and adds GoreeCloud-owned UI and security gates.

The GoreeCloud UI workflow validates canonical Glaze UI/source behavior and server presentation tests. The GoreeCloud Security workflow verifies module integrity, GoreeCloud/Wardveil security contracts, reachable Go vulnerabilities, production Electron dependencies, and deterministic security evidence.

The broader inherited build, lint, test, coverage, compatibility, race-detector, HTML UI, licensing, and platform workflows remain important acceptance evidence and are not weakened merely to make a fork-specific change pass.

Dependency Review is also intentionally retained. If GitHub repository configuration prevents it from running, that configuration issue remains a readiness item rather than a reason to remove the security gate.

See [docs/goreecloud/PRODUCTION_READINESS.md](docs/goreecloud/PRODUCTION_READINESS.md) for the complete evidence model.

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
