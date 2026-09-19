# GoreeCloud Backup VPS CLI/container packaging

This directory defines the controlled GoreeCloud-owned **headless CLI/container** packaging intended for the first replacement of the current Kopia workload on `goreecloud-vps-01`.

It is a stabilization candidate, not evidence that a production cutover has occurred.

## Why the first replacement is CLI-only

The authoritative GoreeCloud Kopia change record documents the known-good production workload as an on-demand CLI Compose service driven by a systemd wrapper/timer, with no published service port. The first GoreeCloud Backup replacement deliberately preserves that smaller runtime boundary instead of introducing an always-running web or API server during a recovery-system migration.

A future GoreeCloud Backup web/admin service can be qualified separately. It is not required to replace the current VPS backup job.

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
- mounts the SFTP private key and known_hosts read-only at compatibility paths;
- keeps repository credential persistence disabled;
- keeps backup source mounts out of the base Compose file so stale documentation cannot silently expand or shrink protection scope;
- provides a separate writable restore-validation location;
- retains bounded local container logs.

## Production target

The intended stack path is:

`/srv/docker/stacks/goreecloud-backup/`

The first controlled replacement may deliberately continue using verified existing Kopia application-data or secret paths when doing so gives the safest rollback and repository-compatibility boundary. Cosmetic path renaming is not a release requirement.

Before deployment, create a live `compose.sources.yaml` from fresh readback of the authoritative current Kopia Compose file. Do not reconstruct the source list from old documentation.

## Known-good historical baseline to reverify

The current GoreeCloud change record documents the prior verified baseline as:

- Kopia 0.23.1;
- `/srv/docker/stacks/kopia/compose.yaml`;
- an on-demand CLI container with `restart: "no"`;
- no published Kopia service port;
- an encrypted SFTP repository stored off-VPS and reached over the NetBird private path;
- persistent config, cache, logs, and temporary data;
- source mounts read-only under `/source`;
- SFTP key and known_hosts mounted read-only under `/run/secrets`;
- a systemd service/timer attempting backup around 00:00, 06:00, 12:00, and 18:00 with up to five minutes randomized delay;
- historically successful 100-percent snapshot verification and real isolated restore validation.

That record is planning evidence only. Run `vps-preflight.sh` against the live VPS before using any of it for cutover.

## Image build

The image build has no default base-image references. A release build must provide exact tag-and-digest values:

`GO_BUILD_IMAGE=<exact-tag>@sha256:<digest>`

`RUNTIME_IMAGE=<exact-tag>@sha256:<digest>`

It must also provide the exact GoreeCloud Backup version, exact source commit SHA, and a unique output image tag before running:

`deploy/docker/build-image.sh`

A local image ID is only build evidence. Production pinning requires the published image's approved tag and immutable digest.

## Controlled source scope

The base `compose.yaml` contains no backup-source bind mounts. Before Release Candidate acceptance:

1. run the read-only VPS preflight;
2. read the live Kopia Compose bind mounts;
3. compare them with current application/recovery requirements;
4. create `compose.sources.yaml` with only the verified required source mounts;
5. keep every source mount read-only;
6. preserve intentional exclusions such as secrets and live database files unless a separately validated application-consistent method replaces them.

## Scheduling

The included systemd unit/timer files model the historically documented four-times-daily cadence. They are templates, not proof of the live current schedule. Verify the existing timer before installing or enabling the GoreeCloud units.

The wrapper validates Compose without resolving environment values, proves repository access, and only then creates a snapshot of `/source`.

## Cutover boundary

Do not replace or retire the current Kopia runtime until one exact GoreeCloud Backup Release Candidate has all applicable source/release gates and the live VPS completes all of these:

1. capture fresh current Kopia Compose, image/digest, source mounts, application-data paths, secret-file paths, systemd units, repository status, and snapshot state;
2. create a fresh pre-cutover Kopia recovery point;
3. run required integrity verification, including 100-percent verification where the governed recovery procedure requires it;
4. complete an isolated representative restore and validate restored data;
5. stage the exact GoreeCloud Backup image by tag and digest without destructive repository migration;
6. prove the GoreeCloud binary can read the existing repository and existing recovery points;
7. create a new GoreeCloud Backup snapshot, verify it, and restore representative data;
8. verify the scheduled job, failure behavior, logging, monitoring/notification path, and recovery evidence;
9. demonstrate a credible rollback to the retained Kopia image/configuration;
10. only then disable the Kopia timer and enable the GoreeCloud Backup timer;
11. observe post-cutover operation before retiring old Kopia deployment material.

Existing production recovery points remain authoritative until these gates are satisfied.
