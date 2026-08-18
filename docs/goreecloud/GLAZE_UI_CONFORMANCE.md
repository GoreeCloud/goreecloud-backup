# GoreeCloud Backup — Glaze UI Conformance

## Status

Target design system: **Glaze UI 1.0**  
Canonical source revision: `d6e446fd8ef251259d16368d50aad90d9287a774`  
Current implementation state: **Source conformance in active development; visual acceptance pending**

This record defines the current GoreeCloud Backup presentation boundary and its stable-release gates. It does not authorize production replacement of an existing Kopia deployment or repository.

## Design-System Boundary

GoreeCloud Backup currently consumes a locally vendored snapshot of the canonical Glaze UI web and accessibility foundation and applies a product-specific presentation adapter around the inherited Kopia HTML UI.

The intended surface hierarchy is **Canvas, Solid, Raised, Glaze, and Overlay**. The application canvas uses restrained GoreeCloud gradients. Readability-first content uses solid surfaces. Cards and data regions use raised surfaces. Navigation may use selective Glaze translucency. Dialogs, menus, popovers, and toasts use the strongest Overlay separation.

The compatibility overlay is temporary. It is not a permanent substitute for owning and building the frontend source directly. The exit criteria and exact upstream frontend provenance remain recorded in `FRONTEND.md`.

## Component Coverage

The application adapter defines Glaze presentation for the inherited Bootstrap component vocabulary, including:

- navigation bars and off-canvas navigation;
- cards and accordions;
- dialogs, menus, popovers, and toasts;
- buttons, destructive actions, pagination, and disabled states;
- text inputs, selects, input groups, checkboxes, and switches;
- tabs, pills, list groups, and selected navigation states;
- tables and responsive data surfaces;
- success, warning, and error alerts;
- progress and loading indicators;
- breadcrumbs, badges, code, and preformatted technical content.

The adapter preserves inherited application behavior and backup semantics. It changes presentation, hierarchy, interaction feedback, and product identity only.

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

## Product Identity

The application name and browser metadata are GoreeCloud Backup. Required Kopia licensing and attribution remain preserved separately from user-facing product identity.

A unique canonical GoreeCloud Backup application icon and derived favicon/launcher set are still required before visual completion under the GoreeCloud Application Branding and User Interface Design standard. The current inherited favicon and remaining upstream secondary branding therefore keep this candidate below visual-completion status.

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

The GoreeCloud UI GitHub Actions workflow runs this validator in addition to module verification and server-package tests.

## Stable-Release Gate

Source conformance does not equal stable release. **Visual acceptance pending** remains the authoritative state until representative runtime evaluation confirms:

- light and dark appearance quality;
- Compact, Medium, Expanded, and Wide behavior;
- keyboard navigation and focus visibility;
- reduced-motion behavior;
- reduced-transparency and no-blur fallbacks;
- increased-contrast and forced-colors behavior;
- loading, empty, warning, error, and destructive states;
- consistent product identity across primary and secondary screens;
- a unique GoreeCloud Backup icon, favicon, and supported launcher surfaces;
- no material upstream-default presentation remaining on production-facing surfaces unless explicitly documented as an approved exception.

Production backup migration, repository replacement, destructive repository operations, or retirement of an existing backup path require separate restore and recovery acceptance evidence and are outside this visual conformance record.
