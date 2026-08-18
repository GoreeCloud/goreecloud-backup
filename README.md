# GoreeCloud Backup

GoreeCloud Backup is the GoreeCloud-maintained backup and recovery platform built on the mature [Kopia](https://github.com/kopia/kopia) codebase.

> **Development status:** early maintained-fork foundation. This repository is not yet approved to replace any existing production Kopia deployment or repository.

## Project direction

GoreeCloud Backup preserves Kopia's proven backup engine while GoreeCloud develops its own product layer, operational model, recovery verification, integrations, Glaze UI experience, and long-term fork-to-native transition path.

The project is recovery-first: creating a snapshot is not by itself evidence that data is recoverable. GoreeCloud Backup is being designed around repository integrity, meaningful retention, monitoring, restore testing, recovery evidence, and clear protection states.

Initial priorities are:

- preserve upstream repository compatibility and recovery behavior where practical;
- maintain explicit upstream provenance and Apache-2.0 licensing obligations;
- establish controlled GoreeCloud build, security, and release workflows;
- introduce the GoreeCloud Backup product identity without hiding the Kopia foundation;
- integrate the Glaze UI design language progressively rather than through a risky one-shot rewrite;
- add GoreeCloud-specific backup policy, verification, restore, monitoring, and integration layers;
- keep existing production backup systems untouched until replacement capability is proven through restoration.

## Upstream foundation

This repository is forked from [`kopia/kopia`](https://github.com/kopia/kopia). Kopia provides the underlying encrypted snapshot, deduplication, compression, repository, storage-backend, CLI, server, and UI foundations that GoreeCloud Backup initially inherits.

The exact fork baseline and upstream-maintenance rules are recorded in [UPSTREAM.md](UPSTREAM.md).

## Architecture

The current architecture remains substantially Kopia-derived. GoreeCloud-specific changes will be introduced in controlled layers so that recovery-critical behavior can be validated independently.

Planned product areas include:

- backup and snapshot management;
- repositories and storage targets;
- retention policies;
- integrity verification;
- guided restores and restore testing;
- protection-state evaluation;
- recovery evidence;
- GoreeCloud Manager integration;
- GoreeCloud Monitor integration;
- GoreeCloud Notify integration;
- future GoreeCloud Identity integration where appropriate.

See [docs/goreecloud/DEVELOPMENT.md](docs/goreecloud/DEVELOPMENT.md) for the maintained-fork development rules.

## Recovery safety boundary

Development in this repository must not be treated as authorization to modify, remove, migrate, or replace an existing production Kopia repository.

A future production cutover requires separate validation, including representative restoration and recovery-evidence collection. Existing recovery points remain authoritative until a replacement has been proven.

## Building

The inherited Kopia build infrastructure remains in place during the initial fork phase. See [BUILD.md](BUILD.md) for the current upstream-derived build process.

GoreeCloud-specific build and release changes will be added incrementally and validated in pull requests before they are relied upon for production artifacts.

## Licensing and attribution

The inherited Kopia code is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).

GoreeCloud Backup preserves required upstream copyright, license, notice, provenance, and attribution information. Product branding may change where permitted, but upstream authorship and licensing must not be obscured.

## Upstream Kopia project

For upstream Kopia documentation, releases, support resources, and contribution guidance, use the official Kopia project resources:

- [Kopia repository](https://github.com/kopia/kopia)
- [Kopia documentation](https://kopia.io/docs/)
- [Kopia website](https://kopia.io/)

GoreeCloud Backup is independently maintained by GoreeCloud and is not the official Kopia distribution.
