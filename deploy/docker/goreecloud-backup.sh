#!/bin/sh
set -eu

STACK_DIR="${GOREECLOUD_BACKUP_STACK_DIR:-/srv/docker/stacks/goreecloud-backup}"
COMPOSE_FILE="${STACK_DIR}/compose.yaml"
SOURCES_FILE="${STACK_DIR}/compose.sources.yaml"
ENV_FILE="${STACK_DIR}/.env"

for required_file in "$COMPOSE_FILE" "$SOURCES_FILE" "$ENV_FILE"; do
  if [ ! -r "$required_file" ]; then
    echo "required GoreeCloud Backup file is not readable: $required_file" >&2
    exit 1
  fi
done

compose() {
  docker compose     --env-file "$ENV_FILE"     --file "$COMPOSE_FILE"     --file "$SOURCES_FILE"     "$@"
}

# Render without resolving environment values so validation does not print
# reusable secret content if a future environment variable becomes sensitive.
compose config --no-env-resolution --quiet

# Repository reachability and authentication must succeed before a snapshot is
# attempted. Failure is explicit; scheduling/monitoring may decide how to alert.
compose run --rm backup repository status >/dev/null

compose run --rm backup snapshot create /source
