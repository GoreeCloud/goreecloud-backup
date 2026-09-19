# GoreeCloud Backup VPS CLI/container packaging

This directory defines the controlled GoreeCloud-owned **headless CLI/container** packaging intended to restore first-party scheduled backup protection on `goreecloud-vps-01`.

It is a stabilization candidate, not evidence that a production deployment or Stable qualification has occurred.

## Current replacement state

The canonical GoreeCloud Backup stabilization task records that the former VPS-side Kopia container, image, network, stack/configuration paths, Kopia systemd units, running process, and host executable were already retired on September 19, 2026 under the approved immediate-retirement exception.

The preserved off-VPS Kopia repository and its historical recovery points remain recovery evidence. This packaging must therefore **not** assume that a live Kopia Compose stack, timer, image, or local configuration still exists on the VPS.

## Why the first replacement is CLI-only

The last verified Kopia workload used an on-demand CLI pattern with scheduled execution and no published service port. The first GoreeCloud Backup VPS runtime deliberately preserves that smaller operational boundary instead of introducing an always-running web or API server into the recovery path.

A future GoreeCloud Backup web/admin surface is qualified separately. It is not required for the headless VPS backup component to protect current VPS data.

## Safety properties

The deployment definition intentionally:

- requires an exact GoreeCloud Backup image tag and digest;
- requires exact tag-and-digest build and runtime base images;
- uses `restart: "no"` for the on-demand CLI workload;
- publishes no host port and defines no web-server command;
- does not use `privileged: true`, host networking, or the Docker socket;
- drops Linux capabilities and enables `no-new-privileges`;
- uses a read-only container root filesystem;
- requires a runtime UID/GID selected from fresh live source-readability evidence;
- injects the repository password from a protected host file;
- mounts the SFTP private key and known_hosts read-only at compatibility paths where the preserved repository requires them;
- keeps repository credential persistence disabled;
- keeps backup source mounts out of the base Compose file so historical documentation cannot silently expand or shrink protection scope;
- provides a separate writable restore-validation location;
- retains bounded local container logs.

## Production target

The intended stack path is:

`/srv/docker/stacks/goreecloud-backup/`

Persistent application-data and secret paths must be selected from the current GoreeCloud Backup deployment design and fresh VPS permission/readability evidence. Do not recreate removed Kopia VPS paths merely for cosmetic compatibility.

The preserved off-VPS repository may remain Kopia-format compatible. Repository-format compatibility is separate from re-creating the retired Kopia runtime.

## Historical recovery baseline

The GoreeCloud change/task records preserve historical evidence including:

- Kopia 0.23.1 as the prior VPS backup implementation;
- an on-demand CLI workload with no published service port;
- an encrypted SFTP repository stored off-VPS over the private network path;
- read-only backup source mounts;
- an historical four-times-daily schedule;
- historical 100-percent snapshot verification and isolated restore validation.

Those records are historical recovery evidence. They are **not** a source of truth for current VPS source paths, current credentials, current application state, or a live Kopia deployment.

## Image build

The image build has no default base-image references. A release build must provide exact tag-and-digest values:

`GO_BUILD_IMAGE=<exact-tag>@sha256:<digest>`

`RUNTIME_IMAGE=<exact-tag>@sha256:<digest>`

It must also provide the exact GoreeCloud Backup version, exact source commit SHA, and a unique output image tag before running:

`deploy/docker/build-image.sh`

A local image ID is only build evidence. Production pinning requires the published image's approved tag and immutable digest.

## Controlled source scope

The base `compose.yaml` contains no backup-source bind mounts. Before Release Candidate acceptance:

1. run the read-only VPS preflight and confirm that the retired Kopia runtime has not unexpectedly reappeared;
2. identify the **current** VPS workloads and data whose loss would require recovery;
3. use each workload's current authoritative recovery requirements plus fresh live path/readability evidence to define backup scope;
4. use database-native or application-approved exports for live databases where file-level copying is not a valid recovery method;
5. create `compose.sources.yaml` with only the verified required file-level source mounts;
6. keep source mounts read-only;
7. document intentional exclusions and the separate mechanism protecting excluded application-consistent data.

Do not reconstruct the source list from the deleted Kopia Compose file or from historical documentation.

## Scheduling

The included systemd unit/timer files model the historically used four-times-daily cadence as a candidate schedule. Because the old Kopia timer has already been retired, the schedule must be accepted against current recovery objectives before the GoreeCloud Backup timer is enabled.

The wrapper validates Compose without resolving environment values, proves repository access, and only then creates a snapshot of `/source`.

## VPS component qualification boundary

The headless VPS runtime may be lifecycle-qualified independently from the desktop/UI component. A Stable VPS component does not make unfinished desktop or web surfaces Stable.

Before enabling production scheduled protection, one exact GoreeCloud Backup VPS Release Candidate must establish all applicable headless-runtime gates, including:

1. confirm the retired Kopia VPS runtime remains absent and record any unexpected legacy residuals;
2. define and review the current VPS protection scope from authoritative workload requirements and live path evidence;
3. preserve the off-VPS historical Kopia repository/recovery evidence without destructive repository migration;
4. build and publish an exact GoreeCloud Backup candidate image with source commit, version, immutable digest, and approved base-image identities;
5. stage the exact image, configuration, credentials, and source bindings without enabling the production timer;
6. prove repository authentication and required access paths without exposing reusable secrets;
7. create a new candidate backup from current required sources and verify repository/integrity state;
8. perform an isolated representative restore and validate the restored data;
9. validate application-consistent recovery for data that cannot be protected by file copying alone;
10. verify failure behavior, logs, monitoring/notification routing, missed-run behavior, and recovery evidence;
11. demonstrate a credible recovery/rollback path, including access to preserved historical Kopia recovery points if the new runtime is unavailable;
12. enable the GoreeCloud Backup timer only after the candidate has passed the applicable Release Candidate cutover gates;
13. observe scheduled operation and repeat representative recovery validation before Stable promotion.

The preserved historical recovery points remain authoritative evidence until GoreeCloud Backup has independently demonstrated current backup creation and validated restoration.
