# GoreeCloud Backup Frontend Architecture

## Purpose

This record defines the inspected frontend source, generated-build, and server-embedding boundaries for GoreeCloud Backup. It exists to keep user-interface work reproducible, attributable, independently maintainable, and isolated from repository-critical backup behavior.

The current server-side Glaze UI overlay is a compatibility bridge. It is not the intended permanent mechanism for owning the GoreeCloud Backup frontend.

## Inspected upstream chain

The GoreeCloud Backup server currently depends on the following exact upstream chain:

```text
kopia/htmlui source
  commit 553ca95fc08f7990e569a2ee11330c200de56f69
        |
        | Vite production build
        v
kopia/htmluibuild generated module
  commit 29e3e25473e03df418bc5bf5e9a01b5886fc6026
        |
        | go:embed build
        v
GoreeCloud/goreecloud-backup
  go.mod pseudo-version:
  github.com/kopia/htmluibuild
  v0.0.1-0.20260804002937-29e3e25473e0
```

The `htmluibuild` commit explicitly records that it was generated from `kopia/htmlui` commit `553ca95fc08f7990e569a2ee11330c200de56f69`. This gives GoreeCloud an exact source-to-generated-asset provenance point rather than treating the embedded frontend as an opaque binary asset.

## Upstream frontend baseline

The inspected `kopia/htmlui` source is an Apache-2.0 React/Vite application. At the recorded baseline it includes, among other dependencies and structures:

- React 19.
- Vite.
- React Bootstrap and Bootstrap 5.
- React Router.
- Font Awesome.
- TanStack Table.
- Vitest-based source tests.
- page, component, context, form, CSS, and utility modules.
- a development proxy for `/api` requests to a local Kopia server.

The frontend communicates with the server through established `/api/v1/...` contracts. Those API boundaries are more important to preserve than the upstream visual implementation.

## Current upstream-facing presentation surfaces

The inspected frontend contains removable upstream presentation identity including:

- Kopia logo artwork in the application navigation.
- Kopia favicon and launcher artwork.
- `KopiaUI` page-title identity.
- `Kopia UI` document metadata.
- a legacy web manifest that still contains Create React App sample naming.
- the `kopia` body identifier and related CSS selectors.
- user-facing navigation and presentation implemented directly in the upstream React source.

These surfaces should be changed through controlled frontend source maintenance rather than indefinitely rewriting generated HTML after build time.

## Architecture decision

GoreeCloud Backup will move toward direct control of the frontend source corresponding to the embedded build while preserving Kopia's mature server and repository engine.

The target ownership chain is:

```text
Verified Kopia frontend baseline
        |
        v
GoreeCloud-controlled frontend source
  - GoreeCloud Backup product identity
  - Glaze UI
  - accessibility and adaptive behavior
  - preserved API compatibility
        |
        v
Reproducible GoreeCloud frontend build
        |
        v
GoreeCloud-controlled embedded frontend artifact/module
        |
        v
GoreeCloud Backup server
```

A separate frontend repository may be used if necessary to preserve the upstream source/build separation cleanly. The final repository name and generated-artifact arrangement must be selected and recorded before that repository is created. This document does not claim that such a repository already exists.

## Compatibility bridge

The current `internal/server` GoreeCloud UI filesystem overlay remains approved as an interim compatibility layer because it:

- leaves inherited generated JavaScript and application behavior unchanged;
- changes only controlled presentation resources and top-level HTML metadata;
- allows Glaze UI work to begin without altering repository format, encryption, content addressing, deduplication, retention, garbage collection, provider writes, or restore semantics;
- provides an immediately testable boundary in the main GoreeCloud Backup repository.

The overlay should be removed or reduced after GoreeCloud controls the generated frontend itself. It must not become a permanent second rendering system that duplicates the source frontend.

## Frontend development safety

Frontend development must use a disposable or synthetic test repository. Production Kopia repositories must not be used as ordinary UI-development fixtures.

User-interface changes must not silently alter:

- repository formats;
- encryption behavior;
- content-addressing behavior;
- snapshot serialization;
- provider writes;
- retention deletion;
- garbage collection;
- restore semantics;
- authentication or authorization contracts.

A frontend change that requires one of those behaviors to change must be treated as a separate higher-risk backup-engine change and reviewed under the recovery-safety gates in `DEVELOPMENT.md`.

## Build and CI requirements

A future GoreeCloud-controlled frontend build must be reproducible from a recorded source revision. At minimum, the frontend pipeline should:

1. use the repository's locked dependency graph with `npm ci` or the equivalent locked install;
2. run frontend tests;
3. run formatting/lint validation;
4. run license and dependency-security checks appropriate to the repository;
5. build production assets from the exact reviewed source revision;
6. record the source revision used to produce generated assets;
7. avoid build workflows that require reusable personal tokens merely to move generated assets between GoreeCloud repositories;
8. use least-privilege GitHub Actions permissions;
9. pin third-party GitHub Actions by immutable commit SHA where practical;
10. preserve Apache-2.0 upstream licensing and attribution.

Generated frontend assets must never be accepted only because a build completed. The source revision, build method, tests, licensing, and resulting integration must be traceable.

## Glaze UI migration direction

Glaze UI integration should move progressively from the compatibility adapter into the frontend source itself.

The controlled migration order is:

1. semantic tokens and global surfaces;
2. navigation and application shell;
3. typography and spacing;
4. buttons, forms, dialogs, cards, tables, alerts, tabs, and status surfaces;
5. light and dark appearance;
6. compact, medium, expanded, and wide layout behavior;
7. keyboard focus and assistive-technology semantics;
8. reduced motion and reduced transparency;
9. increased contrast and forced-colors behavior;
10. final unique GoreeCloud Backup product artwork and launcher assets.

Glaze UI work must improve product clarity. Backup-health, restore, verification, repository, and destructive-action information must remain operationally obvious even when decorative transparency or animation is unavailable.

## Branding migration direction

The frontend source should ultimately carry the canonical GoreeCloud Backup identity across applicable surfaces including:

- page titles;
- application navigation;
- favicon;
- web manifest;
- PWA/launcher artwork;
- accessibility labels;
- application metadata;
- empty states and onboarding;
- settings/about surfaces;
- error and connection states.

Required Kopia copyright, license, provenance, and attribution information must remain available even when removable Kopia product branding is no longer the primary user-facing identity.

## Test reconciliation

The inherited end-to-end suite contains assertions tied to the old `KopiaUI` browser title. GoreeCloud intentionally changes that user-facing identity to `GoreeCloud Backup` while preserving title-prefix HTML escaping. Those assertions must be reconciled to the GoreeCloud product identity rather than weakening or bypassing the test.

A separate inherited checkpoint-cleanup test also failed in one Linux CI execution while passing or not reproducing on other platforms in the same run. It is not caused by an intentional frontend behavior change based on the inspected failure evidence. It must be rerun and investigated if persistent; it must not be silently ignored simply to make the branch green.

## Exit criteria for the compatibility overlay

The current server-side overlay may be removed or materially reduced only after all of the following are true:

- GoreeCloud controls the frontend source revision used by the product.
- GoreeCloud has a reproducible and reviewed frontend build.
- the generated build has recorded source provenance.
- GoreeCloud Backup branding is applied directly in source.
- Glaze UI behavior is implemented directly in source.
- frontend unit/integration tests pass.
- server integration tests pass.
- title-prefix and CSRF behavior remain correct.
- light/dark and accessibility fallbacks are visually validated.
- the resulting server still uses the established API contracts without repository-critical behavioral regression.

Until then, the overlay remains a limited, documented compatibility bridge rather than evidence of a completed frontend fork.