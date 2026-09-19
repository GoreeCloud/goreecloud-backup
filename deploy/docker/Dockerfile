# syntax=docker/dockerfile:1.7

# Both base images are intentionally required build arguments.
# Release builds must pass exact tag-and-digest references.
ARG GO_BUILD_IMAGE
ARG RUNTIME_IMAGE

FROM ${GO_BUILD_IMAGE} AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG GOREECLOUD_BACKUP_VERSION
ARG GOREECLOUD_BACKUP_COMMIT

RUN test -n "${GOREECLOUD_BACKUP_VERSION}" \
    && test -n "${GOREECLOUD_BACKUP_COMMIT}" \
    && CGO_ENABLED=0 go build \
      -trimpath \
      -ldflags "-s -w -X github.com/kopia/kopia/repo.BuildVersion=${GOREECLOUD_BACKUP_VERSION} -X github.com/kopia/kopia/repo.BuildInfo=${GOREECLOUD_BACKUP_COMMIT} -X github.com/kopia/kopia/repo.BuildGitHubRepo=GoreeCloud/goreecloud-backup" \
      -o /out/goreecloud-backup \
      github.com/kopia/kopia

FROM ${RUNTIME_IMAGE} AS runtime

ENV DEBIAN_FRONTEND=noninteractive \
    KOPIA_CONFIG_PATH=/app/config/repository.config \
    KOPIA_CACHE_DIRECTORY=/app/cache \
    KOPIA_LOG_DIR=/app/logs \
    KOPIA_PERSIST_CREDENTIALS_ON_CONNECT=false \
    KOPIA_CHECK_FOR_UPDATES=false

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/goreecloud-backup /usr/local/bin/goreecloud-backup
COPY deploy/docker/entrypoint.sh /usr/local/bin/goreecloud-backup-entrypoint
COPY deploy/docker/healthcheck.sh /usr/local/bin/goreecloud-backup-healthcheck

RUN chmod 0755 \
      /usr/local/bin/goreecloud-backup \
      /usr/local/bin/goreecloud-backup-entrypoint \
      /usr/local/bin/goreecloud-backup-healthcheck \
    && mkdir -p /app/config /app/cache /app/logs /source /restore

EXPOSE 51515

ENTRYPOINT ["/usr/local/bin/goreecloud-backup-entrypoint"]
CMD ["server", "start", "--address=http://0.0.0.0:51515", "--persistent-logs"]
