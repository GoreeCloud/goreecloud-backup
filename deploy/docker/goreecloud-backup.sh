#!/bin/sh
set -eu

STACK_DIR="${GOREECLOUD_BACKUP_STACK_DIR:-/srv/docker/stacks/goreecloud-backup}"
COMPOSE_FILE="${STACK_DIR}/compose.yaml"
SOURCES_FILE="${STACK_DIR}/compose.sources.yaml"
ENV_FILE="${STACK_DIR}/.env"
SOURCE_SCOPE_VALIDATOR="${STACK_DIR}/validate-source-scope.py"
ACTION="${1:-backup}"

if [ "$#" -gt 1 ]; then
  echo "usage: $0 [backup|connect-repository|repository-status|validate]" >&2
  exit 2
fi

for required_file in "$COMPOSE_FILE" "$SOURCES_FILE" "$ENV_FILE" "$SOURCE_SCOPE_VALIDATOR"; do
  if [ ! -r "$required_file" ]; then
    echo "required GoreeCloud Backup file is not readable: $required_file" >&2
    exit 1
  fi
done

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required for fail-closed backup source validation" >&2
  exit 1
fi

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

validate_source_scope() {
  targets="$(
    compose config --format json \
      | python3 "$SOURCE_SCOPE_VALIDATOR" --require-host-readable
  )"

  for target in $targets; do
    compose run --rm --entrypoint /bin/sh backup -c '
      set -eu
      source_target="$1"
      test -e "$source_target"
      test -r "$source_target"
      if [ -d "$source_target" ]; then
        test -x "$source_target"
      fi
    ' sh "$target" >/dev/null
  done
}

repository_status() {
  compose run --rm backup repository status
}

validate_compose

case "$ACTION" in
  validate)
    validate_source_scope
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
    # A backup is not permitted until the declared protection scope is explicit,
    # read-only, present on the host, and readable by the configured container
    # identity. This prevents an empty /source snapshot from looking successful.
    validate_source_scope

    # Repository reachability and authentication must succeed before a snapshot
    # is attempted. Failure is explicit; scheduling/monitoring handles alerting.
    repository_status >/dev/null
    compose run --rm backup snapshot create --fail-fast /source
    ;;

  *)
    echo "unsupported action: $ACTION" >&2
    echo "usage: $0 [backup|connect-repository|repository-status|validate]" >&2
    exit 2
    ;;
esac
