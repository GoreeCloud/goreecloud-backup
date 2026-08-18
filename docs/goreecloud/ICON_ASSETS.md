# GoreeCloud Backup — Canonical Icon and Asset Contract

## Status

**Artwork integration contract established; canonical artwork is not yet supplied.**

This document defines how GoreeCloud Backup carries one unique product identity across web, PWA, desktop packaging, Linux/AppImage, and future Android clients. It deliberately does not create or approve artwork.

The application must not be classified visually complete or Stable while the canonical GoreeCloud Backup artwork remains absent or inherited Kopia/default-framework icon assets remain production-visible.

## Governing identity

GoreeCloud Backup must have its own recognizable primary symbol. The GoreeCloud platform logo must not be used as the application icon, and Wardveil Security identity must not become the Backup application icon.

The icon should communicate the application's Role and Purpose: preservation, backup, recovery, verification, or durable restoration. It should remain understandable without text and recognizable at favicon size.

The preferred design direction is a distinctive recovery/preservation symbol with Glaze UI family traits such as polished geometry, restrained depth, intentional negative space, and strong silhouette. A generic shield alone is not appropriate because Wardveil Security by GoreeCloud is the platform security identity rather than the Backup product identity.

## Canonical source

One approved source artwork file will be authoritative:

`branding/goreecloud-backup-icon.svg`

Requirements:

- original GoreeCloud-controlled artwork;
- square view box;
- no embedded remote resources, fonts, scripts, tracking, or external dependencies;
- no upstream Kopia logo or proprietary third-party mark;
- no text that becomes unreadable at launcher or favicon sizes;
- strong silhouette at 16 × 16 and 32 × 32 representations;
- suitable for light and dark surroundings;
- compatible with platform masking and adaptive-icon requirements;
- source-controlled and reviewed like application code.

The canonical SVG is the identity source. Raster, ICO, ICNS, maskable, monochrome, launcher, and notification forms are generated or adapted from it; they are not independently redesigned symbols.

## Web and PWA outputs

The browser-facing asset set must include at least:

- `app/public/favicon.ico` containing appropriate small favicon representations;
- `app/public/icon-16.png` — 16 × 16;
- `app/public/icon-32.png` — 32 × 32;
- `app/public/icon-48.png` — 48 × 48;
- `app/public/logo192.png` — 192 × 192;
- `app/public/logo512.png` — 512 × 512;
- `app/public/icon-maskable-512.png` — 512 × 512 maskable PWA form when the artwork supports safe masking.

`app/public/manifest.json`, browser metadata, install prompts, shortcuts, and any future social/install metadata must reference this GoreeCloud Backup identity rather than inherited Create React App or Kopia identity.

## Desktop and AppImage outputs

Desktop packaging must derive from the same canonical identity:

- `app/assets/icon.png` — canonical high-resolution Linux/Electron input, minimum 512 × 512;
- `app/assets/icon.icns` — macOS packaging derivative;
- `app/assets/icon.ico` — Windows packaging derivative;
- Linux desktop/AppImage launcher metadata must resolve to the same identity.

Electron Builder configuration must use these GoreeCloud-controlled assets explicitly. Falling back to Electron, Kopia, framework, or operating-system default artwork is a release failure.

## Android APK/AAB outputs

There is no Android application source tree in the current GoreeCloud Backup repository checkpoint. An APK icon therefore cannot honestly be marked implemented yet.

If/when an Android client is added, it must derive its launcher identity from the same canonical artwork and provide platform-appropriate resources, including:

- adaptive launcher foreground/background resources;
- legacy launcher PNGs for required density buckets;
- round launcher form when required by the target launcher ecosystem;
- monochrome/themed icon resource when supported;
- notification icon adapted from the same identity where Android requires a one-color notification glyph.

A future Android client may adjust safe-zone padding, masking, monochrome rendering, and platform corner treatment, but it must not introduce a different primary symbol.

## Wardveil relationship

Wardveil Security by GoreeCloud may appear on security status, protection controls, alerts, security findings, and related security surfaces. Wardveil does not replace the GoreeCloud Backup application identity.

The Backup icon may visually coexist with Wardveil badges or security indicators, but the two identities must remain distinguishable. A user should be able to tell whether an icon represents the Backup application itself or a security/protection status produced under Wardveil Security.

## Asset provenance and generation

Generated derivatives must be reproducible from the canonical source with a documented local/open-source toolchain. The build process must not require a hosted image-generation service or remote conversion API.

The repository should record:

- canonical source path;
- asset-generation command/tool versions;
- expected output paths and dimensions;
- checksums in release evidence when practical;
- any platform-specific adaptation and the reason for it.

Generated assets should not contain metadata that unnecessarily exposes local usernames, filesystem paths, editing history, or private environment details.

## Validation requirements

Before visual completion or Stable classification:

1. canonical artwork is present and approved;
2. inherited Kopia/default-framework product icons are removed from production-facing surfaces where legally permitted;
3. web favicon representations are visually reviewed at 16, 32, and 48 pixels;
4. 192 and 512 pixel PWA/install representations are verified;
5. AppImage/Linux launcher identity is verified in a representative desktop environment;
6. Windows/macOS package assets are verified when those packages are supported;
7. Android launcher/adaptive/notification assets are verified on representative devices if an Android client exists;
8. manifest/package metadata uses GoreeCloud Backup naming and asset paths;
9. light/dark surroundings and accessibility/high-contrast behavior do not make the symbol unreadable;
10. the icon remains recognizably the same product across every supported surface.

## Current inherited asset warning

The current repository checkpoint still contains inherited icon binaries in `app/assets` and `app/public`. Those files are temporary compatibility assets and must not be treated as the approved GoreeCloud Backup identity.

They may remain only until approved canonical artwork is supplied and the complete derivative set is generated, wired, reviewed, and validated. This exception is development-only and blocks visual-completion/Stable claims.
