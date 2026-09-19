# GoreeCloud Backup VPS container packaging

This directory defines the controlled GoreeCloud-owned container packaging and Compose surface intended for the future replacement of the current Kopia runtime on `goreecloud-vps-01`.

It is a **stabilization candidate**, not evidence that a production cutover has occurred.

## Safety and governance properties

The deployment definition intentionally:

- requires an exact GoreeCloud Backup image reference instead of `latest`;
- requires exact tag-and-digest build and runtime base images when building the image;
- does not use `privileged: true`;
- drops Linux capabilities and enables `no-new-privileges`;
- uses a read-only container root filesystem;
- requires an explicitly validated runtime UID/GID instead of silently defaulting to root;
- mounts the protected source tree read-only;
- provides a separate writable restore-validation directory rather than restoring directly into production source data;
- does not disable CSRF protection;
- publishes no host port;
- attaches only to a verified private Gateway Docker network;
- reads reusable passwords from protected Compose secret files;
- keeps repository credential persistence disabled;
- includes an authenticated server-status health check;
- bounds container log retention.

## Production target layout

The authoritative production stack belongs under:

`/srv/docker/stacks/goreecloud-backup/`

Persistent state should use the approved GoreeCloud Docker hierarchy, such as:

- `/srv/docker/appdata/goreecloud-backup/config/`
- `/srv/docker/appdata/goreecloud-backup/cache/`
- `/srv/docker/logs/goreecloud-backup/`
- `/srv/docker/appdata/goreecloud-backup/restore-tests/`
- `/srv/docker/secrets/goreecloud-backup/`

The live VPS paths, ownership, source scope, repository backend, private Gateway network, and secret delivery must be verified on the server before production use.

## Image build

The image build deliberately has no default base-image references. A release build must provide exact tag-and-digest values:

`GO_BUILD_IMAGE=<exact-tag>@sha256:<digest>`

`RUNTIME_IMAGE=<exact-tag>@sha256:<digest>`

It must also provide the exact GoreeCloud Backup version, exact source commit SHA, and a unique output image tag before running:

`deploy/docker/build-image.sh`

The resulting local image ID is only build evidence. Production pinning still requires the published image's approved tag and digest.

## Compose configuration

Copy `.env.example` to the protected authoritative environment file only in the approved VPS stack directory, then replace every placeholder from verified live state.

Do not place active passwords in the environment file. The Compose definition reads the repository, UI-server, and server-control passwords from protected host files through Compose secrets.

A local-filesystem backup repository may require an additional backend-specific read/write mount. Do not invent that mount in source control: preserve the existing repository architecture until the live repository type and path are verified, then document the minimum required mount in the authoritative VPS deployment.

## Cutover boundary

This packaging must not replace the current Kopia container until all of the following are proven for one exact Release Candidate:

1. mandatory exact-head source and build checks pass;
2. the image is built and pinned by exact release tag and digest;
3. the current Kopia repository type, configuration, credentials path, schedules, source paths, and recovery points are inventoried from the live VPS;
4. a pre-cutover recovery point is created and independently verified;
5. GoreeCloud Backup can open the intended repository without destructive migration;
6. representative backup, integrity verification, isolated restore, and restored-data validation succeed;
7. Gateway/private-network access, authentication, monitoring, logging, and notifications are verified;
8. rollback to the retained Kopia runtime is tested or otherwise credibly demonstrated;
9. only then is the old runtime stopped and GoreeCloud Backup made authoritative;
10. Kopia is not retired or deleted until post-cutover observation and Stable qualification permit it.

The existing production recovery points remain authoritative until those gates are satisfied.
