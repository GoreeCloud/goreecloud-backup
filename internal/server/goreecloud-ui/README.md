# GoreeCloud Backup UI Overlay

This directory contains the presentation-only GoreeCloud layer applied to the currently embedded Kopia HTML UI.

## Boundary

The backup engine, repository format, encryption, deduplication, storage providers, snapshot behavior, retention, maintenance, and restore implementation remain in the inherited Kopia codebase. The files here do not replace or modify those behaviors.

The current upstream HTML interface is supplied to the Go server through the separate `github.com/kopia/htmluibuild` module. At the foundation baseline, `go.mod` pins that module to revision `29e3e25473e0`. The server-side overlay delegates upstream UI assets unchanged and transforms only the top-level HTML presentation to load GoreeCloud-controlled local styles and product metadata.

## Glaze UI provenance

The vendored reusable files are copied from the canonical `GoreeCloud/glaze-ui` repository at exact revision:

`d6e446fd8ef251259d16368d50aad90d9287a774`

Canonical source files:

- `css/glaze.css`
- `css/glaze.accessibility.css`

The canonical design-system source is MIT licensed. A copy of that license is preserved in `GLAZE-UI-LICENSE`.

`backup.css` is the GoreeCloud Backup-specific adapter. It maps the inherited React/Bootstrap presentation to Glaze UI semantic surfaces, controls, focus behavior, light/dark appearance, accessibility fallbacks, and compact-layout behavior without changing application logic.

## Development status

This is an incremental maintained-fork boundary, not the final GoreeCloud Backup frontend. It intentionally avoids a large frontend transplant before source, build, and recovery behavior are proven stable. A later phase may move the full HTML UI source under GoreeCloud control or replace it with an independently maintained frontend while continuing to use the proven backup APIs where appropriate.

Upstream attribution and Apache-2.0 obligations remain preserved elsewhere in the repository and release materials.
