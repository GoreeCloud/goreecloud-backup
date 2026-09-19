#!/bin/sh
set -eu

STACK_DIR="${GOREECLOUD_BACKUP_STACK_DIR:-/srv/docker/stacks/goreecloud-backup}"
COMPOSE_FILE="${STACK_DIR}/compose.yaml"
SOURCES_FILE="${STACK_DIR}/compose.sources.yaml"
ENV_FILE="${STACK_DIR}/.env"
ACTION="${1:-backup}"

if [ "$#" -gt 1 ]; then
  echo "usage: $0 [backup|connect-repository|repository-status|validate]" >&2
  exit 2
fi

for required_file in "$COMPOSE_FILE" "$SOURCES_FILE" "$ENV_FILE"; do
  if [ ! -r "$required_file" ]; then
    echo "required GoreeCloud Backup file is not readable: $required_file" >&2
    exit 1
  fi
done

compose() {
  docker compose \
    --env-file "$ENV_FILE" \
    --file "$COMPOSE_FILE" \
    --file "$SOURCES_FILE" \
    "$@"
}

validate_compose() {
  # Do not render resolved environment values to stdout. Secret material belongs
  # in files, but this also avoids exposing future sensitive environment values.
  compose config --no-env-resolution --quiet
}

repository_status() {
  compose run --rm backup repository status
}

validate_compose

case "$ACTION" in
  validate)
    exit 0
    ;;

  repository-status)
    repository_status
    ;;

  connect-repository)
    if repository_status >/dev/null 2>&1; then
      echo "repository is already connected; refusing to replace the existing local repository configuration" >&2
      exit 1
    fi

    compose run --rm backup __goreecloud_connect_existing_sftp
    repository_status
    ;;

  backup)
    # Repository reachability and authentication must succeed before a snapshot
    # is attempted. Failure is explicit; scheduling/monitoring handles alerting.
    repository_status >/dev/null
    compose run --rm backup snapshot create /source
    ;;

  *)
    echo "unsupported action: $ACTION" >&2
    echo "usage: $0 [backup|connect-repository|repository-status|validate]" >&2
    exit 2
    ;;
esac
