#!/bin/sh
set -eu

STACK_DIR="${GOREECLOUD_BACKUP_STACK_DIR:-/srv/docker/stacks/goreecloud-backup}"
COMPOSE_FILE="${STACK_DIR}/compose.yaml"
SOURCES_FILE="${STACK_DIR}/compose.sources.yaml"
ENV_FILE="${STACK_DIR}/.env"
SOURCE_SCOPE_VALIDATOR="${STACK_DIR}/validate-source-scope.py"
SNAPSHOT_EVIDENCE_VALIDATOR="${STACK_DIR}/validate-snapshot-evidence.py"
ACTION="${1:-backup}"

if [ "$#" -gt 1 ]; then
  echo "usage: $0 [backup|acceptance-recovery|connect-repository|repository-status|validate]" >&2
  exit 2
fi

for required_file in "$COMPOSE_FILE" "$SOURCES_FILE" "$ENV_FILE" "$SOURCE_SCOPE_VALIDATOR" "$SNAPSHOT_EVIDENCE_VALIDATOR"; do
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

validate_repository_identity() {
  compose run --rm backup __goreecloud_validate_repository >/dev/null
}

expected_snapshot_source() {
  compose config --format json | python3 -c '
import json
import sys

config = json.load(sys.stdin)
environment = config["services"]["backup"].get("environment", {})
hostname = environment.get("GOREECLOUD_BACKUP_CLIENT_HOSTNAME", "")
username = environment.get("GOREECLOUD_BACKUP_CLIENT_USERNAME", "")
if not hostname or not username:
    raise SystemExit("repository client identity is missing from Compose")
print(f"{username}@{hostname}:/source")
'
}

validate_restore_target() {
  compose run --rm --entrypoint /bin/sh backup -c '
    set -eu
    test -d /restore
    test -w /restore
    test -x /restore
  ' >/dev/null
}

run_acceptance_recovery() {
  validate_source_scope
  validate_repository_identity
  validate_restore_target

  expected_source="$(expected_snapshot_source)"

  snapshot_json="$(
    compose run --rm backup snapshot create \
      --fail-fast \
      --json \
      --tags goreecloud:acceptance \
      /source
  )"

  evidence="$(
    printf '%s\n' "$snapshot_json" \
      | python3 "$SNAPSHOT_EVIDENCE_VALIDATOR" \
          --expected-source "$expected_source"
  )"

  old_ifs="$IFS"
  IFS="$(printf '\t')"
  set -- $evidence
  IFS="$old_ifs"

  if [ "$#" -ne 2 ]; then
    echo "snapshot evidence parser did not return exactly two identifiers" >&2
    exit 1
  fi

  snapshot_id="$1"
  root_object_id="$2"
  restore_target="/restore/acceptance-${snapshot_id}"

  compose run --rm backup snapshot verify \
    --verify-files-percent=100 \
    "$snapshot_id"

  compose run --rm --entrypoint /bin/sh backup -c '
    set -eu
    test ! -e "$1"
  ' sh "$restore_target" >/dev/null

  compose run --rm backup snapshot restore \
    --skip-owners \
    --write-files-atomically \
    "$root_object_id" \
    "$restore_target"

  compose run --rm --entrypoint /bin/sh backup -c '
    set -eu
    target="$1"
    test -d "$target"
    test -n "$(ls -A "$target")"
  ' sh "$restore_target" >/dev/null

  printf 'acceptance_snapshot_id=%s\n' "$snapshot_id"
  printf 'acceptance_root_object_id=%s\n' "$root_object_id"
  printf 'acceptance_restore_target=%s\n' "$restore_target"
  echo "acceptance restore retained for application-specific validation"
}

validate_compose

case "$ACTION" in
  acceptance-recovery)
    run_acceptance_recovery
    ;;

  validate)
    validate_source_scope
    ;;

  repository-status)
    validate_repository_identity
    repository_status
    ;;

  connect-repository)
    if repository_status >/dev/null 2>&1; then
      echo "repository is already connected; refusing to replace the existing local repository configuration" >&2
      exit 1
    fi

    compose run --rm backup __goreecloud_connect_existing_sftp
    validate_repository_identity
    repository_status
    ;;

  backup)
    # A backup is not permitted until the declared protection scope is explicit,
    # read-only, present on the host, and readable by the configured container
    # identity. This prevents an empty /source snapshot from looking successful.
    validate_source_scope

    # The persistent repository config must resolve to the exact governed
    # client identity and preserved off-VPS repository before any write occurs.
    validate_repository_identity

    compose run --rm backup snapshot create --fail-fast /source
    ;;

  *)
    echo "unsupported action: $ACTION" >&2
    echo "usage: $0 [backup|acceptance-recovery|connect-repository|repository-status|validate]" >&2
    exit 2
    ;;
esac
