# GoreeCloud Backup — Glaze UI Conformance

## Status

Target design system: **Glaze UI 1.0**  
Canonical source revision: `d6e446fd8ef251259d16368d50aad90d9287a774`  
Security identity: **Wardveil Security by GoreeCloud**  
Current implementation state: **Source conformance in active development; visual acceptance pending**

This record defines the current GoreeCloud Backup presentation boundary and its stable-release gates. It does not authorize production replacement of an existing Kopia deployment or repository.

## Design-System Boundary

GoreeCloud Backup currently consumes a locally vendored snapshot of the canonical Glaze UI web and accessibility foundation and applies a product-specific presentation adapter around the inherited Kopia HTML UI.

The intended surface hierarchy is **Canvas, Solid, Raised, Glaze, and Overlay**. The application canvas uses restrained GoreeCloud gradients. Readability-first content uses solid surfaces. Cards and data regions use raised surfaces. Navigation may use selective Glaze translucency. Dialogs, menus, popovers, and toasts use the strongest Overlay separation.

The compatibility overlay is temporary. It is not a permanent substitute for owning and building the frontend source directly. The exit criteria and exact upstream frontend provenance remain recorded in `FRONTEND.md`.

## Wardveil Security Presentation Contract

Wardveil Security by GoreeCloud is the official platform-wide security and protection identity used when GoreeCloud Backup presents security posture, security controls, or protection assurance. Wardveil is a presentation and identity layer; it does not replace the underlying backup engine, repository security model, GoreeCloud policies, or technical authorities.

GoreeCloud Backup may use the approved phrases **Wardveil Security**, **Wardveil**, and **Protected by Wardveil** where the interface is actually communicating security or protection posture and the scope has authoritative evidence. It must not invent unapproved Wardveil module names, and it must not imply that a successful snapshot, a green security scan, process health, or Wardveil branding proves recoverability.

Wardveil-facing surfaces must remain Glaze UI surfaces. They should communicate calm protection, trust, privacy, clarity, and control rather than stereotypical cybersecurity imagery. Security status must always preserve the identity of the underlying control or evidence source so the brand never obscures technical authority.

For GoreeCloud Backup specifically, Wardveil presentation is appropriate for areas such as repository credential protection, vulnerability/update posture, desktop-shell hardening status, secret-handling safeguards, integrity/security findings, authorization denial, and security-related recovery readiness. Recovery state remains governed by GoreeCloud Backup's recovery evidence and restore-verification model.

## Component Coverage

The application adapter defines Glaze presentation for the inherited Bootstrap component vocabulary, including:

- navigation bars and off-canvas navigation;
- cards and accordions;
- dialogs, menus, popovers, and toasts;
- buttons, destructive actions, pagination, and disabled states;
- text inputs, selects, input groups, checkboxes, and switches;
- tabs, pills, list groups, and selected navigation states;
- tables and responsive data surfaces;
- success, warning, degraded, denied-access, and error alerts;
- progress and loading indicators;
- breadcrumbs, badges, code, and preformatted technical content.

The adapter preserves inherited application behavior and backup semantics. It changes presentation, hierarchy, interaction feedback, and product identity only.

## Application State Contract

GoreeCloud Backup must present operational state deliberately instead of collapsing all non-success conditions into one generic error.

Required Glaze UI state families are:

- **Loading** — work has started or authoritative state is being retrieved. Long-running backup/restore operations must identify what is happening and must not leave an indefinite ambiguous spinner when progress/status information is available.
- **Success** — the specific requested operation completed. Success copy must identify the operation and must not promote a successful snapshot into a `Restore Verified` or blanket security claim.
- **Empty** — the queried collection or state is legitimately empty. Empty states should explain the next safe action rather than resemble a failure.
- **Warning** — attention is required but the operation/service may remain usable.
- **Degraded** — a required protection, repository, monitoring, verification, or recovery-readiness signal is incomplete, stale, or failing while some service capability remains available.
- **Denied access** — authentication/authorization/CSRF or other access-control logic rejected the operation. Denied states must not reveal whether hidden protected objects exist or leak authorization internals.
- **Error** — the requested operation failed. Error surfaces must be actionable, preserve safe diagnostic context, and avoid exposing credentials, repository secrets, protected file contents, or untrusted raw backend output.
- **Destructive confirmation** — deletion, retention changes, repository maintenance, overwrite restore, credential revocation, or other destructive/recovery-sensitive actions require explicit scope, consequence, and confirmation appropriate to the risk.

Security-related warning, degraded, denied, and error states may carry Wardveil identity where that improves clarity. The originating technical control must remain visible.

## Adaptive Layout

GoreeCloud Backup targets the canonical Glaze adaptive ranges:

- **Compact** — through 599 px: touch-first targets, stacked controls, viewport-safe dialogs, compact gutters, and horizontally resilient data regions.
- **Medium** — 600 through 1023 px: comfortable touch targets, wider dialogs, and increased composition spacing.
- **Expanded** — 1024 through 1439 px: higher information density with larger content gutters and data-oriented modal sizing.
- **Wide** — 1440 px and above: bounded reading width while permitting data-heavy surfaces and large dialogs to use the available workspace.

The current compatibility layer cannot yet transform every upstream navigation composition into a product-specific rail, split view, or bottom-navigation model. That remains part of the controlled frontend-source transition rather than an undocumented production exception.

## Accessibility and Resilience

The current source boundary includes:

- practical 44 px minimum targets and 48 px coarse-pointer targets;
- visible keyboard focus treatment;
- readable light and dark semantic colors;
- **Reduced motion** handling that removes nonessential transform and transition movement;
- **Reduced transparency** handling with solid-surface fallbacks;
- unsupported-backdrop-filter solid fallbacks;
- **Increased contrast** treatment with stronger boundaries and text recovery;
- **Forced colors** support for operating-system high-contrast modes;
- print-safe removal of navigation and transient overlays;
- local/system fonts and no required remote UI assets.

Accessibility treatments are designed to preserve the Glaze UI visual hierarchy rather than replacing it with a separate visual language.

## Privacy and Dependency Contract

The GoreeCloud-owned Glaze layer introduces **No remote UI dependencies**. The vendored design-system CSS, accessibility CSS, and product adapter are served locally from the application. The Glaze layer does not add analytics, tracking, advertising, third-party fonts, remote icon delivery, or a new telemetry path.

Inherited Kopia frontend dependencies and external service/provider behavior are separate compatibility and upstream-maintenance concerns. They are not reclassified as Glaze UI dependencies.

Error, diagnostic, and Wardveil security surfaces must not render reusable credentials, raw authentication material, protected backup contents, or other sensitive values merely because a backend error included them. Production-facing error mapping and log/display redaction remain part of runtime acceptance.

## Product Identity

The application name and browser metadata are GoreeCloud Backup. Required Kopia licensing and attribution remain preserved separately from user-facing product identity.

A unique canonical GoreeCloud Backup application icon and derived favicon/launcher set are still required before visual completion under the GoreeCloud Application Branding and User Interface Design standard. The current inherited favicon and remaining upstream secondary branding therefore keep this candidate below visual-completion status.

Wardveil Security must receive and retain its own distinguishable security identity rather than becoming the GoreeCloud Backup application icon. Backup identity, Glaze UI design identity, and Wardveil security identity are complementary but separate.

## Automated Conformance

`scripts/validate_goreecloud_ui.py` is the fail-closed source validator for the current presentation boundary. It verifies:

- canonical Glaze semantic-token markers and source revision provenance;
- canonical motion values and minimum target sizing;
- accessibility and resilience markers;
- application component-state coverage;
- Compact, Medium, Expanded, and Wide source rules;
- local-only UI dependency behavior in the GoreeCloud-owned CSS boundary;
- server integration and presentation tests;
- this conformance record and its stable-release caveats.

The GoreeCloud UI GitHub Actions workflow runs this validator in addition to module verification and server-package tests. The GoreeCloud Security workflow separately enforces the Wardveil-facing security/source contract and security maintenance controls.

## Stable-Release Gate

Source conformance does not equal stable release. **Visual acceptance pending** remains the authoritative state until representative runtime evaluation confirms:

- light and dark appearance quality;
- Compact, Medium, Expanded, and Wide behavior;
- keyboard navigation and focus visibility;
- reduced-motion behavior;
- reduced-transparency and no-blur fallbacks;
- increased-contrast and forced-colors behavior;
- loading, success, empty, warning, degraded, denied-access, error, and destructive states;
- long-running backup and restore feedback without silent/ambiguous failure;
- safe error redaction and actionable diagnostic presentation;
- consistent product identity across primary and secondary screens;
- correct Wardveil Security presentation wherever security posture is surfaced;
- a unique GoreeCloud Backup icon, favicon, and supported launcher surfaces;
- no material upstream-default presentation remaining on production-facing surfaces unless explicitly documented as an approved exception.

Production backup migration, repository replacement, destructive repository operations, or retirement of an existing backup path require separate restore and recovery acceptance evidence and are outside this visual conformance record.
