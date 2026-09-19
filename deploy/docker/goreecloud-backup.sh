#!/bin/sh
set -eu
umask 027

STACK_DIR="${GOREECLOUD_BACKUP_STACK_DIR:-/srv/docker/stacks/goreecloud-backup}"
COMPOSE_FILE="${STACK_DIR}/compose.yaml"
SOURCES_FILE="${STACK_DIR}/compose.sources.yaml"
ENV_FILE="${STACK_DIR}/.env"
SOURCE_SCOPE_VALIDATOR="${STACK_DIR}/validate-source-scope.py"
SNAPSHOT_EVIDENCE_VALIDATOR="${STACK_DIR}/validate-snapshot-evidence.py"
STATUS_RECORDER="${STACK_DIR}/record-status.py"
STATUS_DIR="${GOREECLOUD_BACKUP_STATUS_DIR:-${STACK_DIR}/status}"
STATUS_FILE="${STATUS_DIR}/status.json"
EVENTS_FILE="${STATUS_DIR}/events.jsonl"
LOCK_FILE="${STACK_DIR}/.operation.lock"
ACTION="${1:-backup}"
OPERATION_ID=""
OPERATION_FINALIZED=0
FAILURE_CATEGORY="configuration_failure"

if [ "$#" -gt 1 ]; then
  echo "usage: $0 [backup|acceptance-recovery|connect-repository|repository-status|validate]" >&2
  exit 2
fi

for required_file in \
  "$COMPOSE_FILE" \
  "$SOURCES_FILE" \
  "$ENV_FILE" \
  "$SOURCE_SCOPE_VALIDATOR" \
  "$SNAPSHOT_EVIDENCE_VALIDATOR" \
  "$STATUS_RECORDER"; do
  if [ ! -r "$required_file" ]; then
    echo "required GoreeCloud Backup file is not readable: $required_file" >&2
    exit 1
  fi
done

for required_command in docker python3 flock; do
  if ! command -v "$required_command" >/dev/null 2>&1; then
    echo "required GoreeCloud Backup command is unavailable: $required_command" >&2
    exit 1
  fi
done

exec 9>"$LOCK_FILE"
if ! flock -n 9; then
  echo "another GoreeCloud Backup operation is already active" >&2
  exit 1
fi

record_started() {
  python3 "$STATUS_RECORDER" \
    --status-file "$STATUS_FILE" \
    --events-file "$EVENTS_FILE" \
    --operation "$ACTION" \
    started
}

record_terminal() {
  state="$1"
  shift
  python3 "$STATUS_RECORDER" \
    --status-file "$STATUS_FILE" \
    --events-file "$EVENTS_FILE" \
    --operation "$ACTION" \
    --operation-id "$OPERATION_ID" \
    "$@" \
    "$state"
  OPERATION_FINALIZED=1
}

on_exit() {
  rc=$?
  trap - EXIT

  if [ "$rc" -ne 0 ] && [ "$OPERATION_FINALIZED" -ne 1 ] && [ -n "$OPERATION_ID" ]; then
    python3 "$STATUS_RECORDER" \
      --status-file "$STATUS_FILE" \
      --events-file "$EVENTS_FILE" \
      --operation "$ACTION" \
      --operation-id "$OPERATION_ID" \
      --failure-category "$FAILURE_CATEGORY" \
      failed >/dev/null 2>&1 || true
  fi

  exit "$rc"
}

trap on_exit EXIT
trap 'FAILURE_CATEGORY=interrupted; exit 130' INT
trap 'FAILURE_CATEGORY=interrupted; exit 143' TERM

OPERATION_ID="$(record_started)"

compose() {
  docker compose \
    --env-file "$ENV_FILE" \
    --file "$COMPOSE_FILE" \
    --file "$SOURCES_FILE" \
    "$@"
}

validate_compose() {
  compose config --no-env-resolution --quiet
}

validate_source_scope() {
  FAILURE_CATEGORY="source_scope_failure"
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
  FAILURE_CATEGORY="repository_failure"
  compose run --rm backup repository status
}

validate_repository_identity() {
  FAILURE_CATEGORY="repository_identity_failure"
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
  FAILURE_CATEGORY="restore_target_failure"
  compose run --rm --entrypoint /bin/sh backup -c '
    set -eu
    test -d /restore
    test -w /restore
    test -x /restore
  ' >/dev/null
}

parse_snapshot_evidence() {
  evidence="$1"
  old_ifs="$IFS"
  IFS="$(printf '\t')"
  set -- $evidence
  IFS="$old_ifs"

  if [ "$#" -ne 2 ]; then
    echo "snapshot evidence parser did not return exactly two identifiers" >&2
    exit 1
  fi

  SNAPSHOT_ID="$1"
  ROOT_OBJECT_ID="$2"
}

create_validated_snapshot() {
  tag="$1"
  FAILURE_CATEGORY="snapshot_failure"
  expected_source="$(expected_snapshot_source)"

  snapshot_json="$(
    compose run --rm backup snapshot create \
      --fail-fast \
      --json \
      --tags "$tag" \
      /source
  )"

  evidence="$(
    printf '%s\n' "$snapshot_json" \
      | python3 "$SNAPSHOT_EVIDENCE_VALIDATOR" \
          --expected-source "$expected_source"
  )"

  parse_snapshot_evidence "$evidence"
}

run_acceptance_recovery() {
  validate_source_scope
  validate_repository_identity
  validate_restore_target
  create_validated_snapshot "goreecloud:acceptance"

  restore_target="/restore/acceptance-${SNAPSHOT_ID}"

  FAILURE_CATEGORY="integrity_verification_failure"
  compose run --rm backup snapshot verify \
    --verify-files-percent=100 \
    "$SNAPSHOT_ID"

  FAILURE_CATEGORY="restore_failure"
  compose run --rm --entrypoint /bin/sh backup -c '
    set -eu
    test ! -e "$1"
  ' sh "$restore_target" >/dev/null

  compose run --rm backup snapshot restore \
    --skip-owners \
    --write-files-atomically \
    "$ROOT_OBJECT_ID" \
    "$restore_target"

  compose run --rm --entrypoint /bin/sh backup -c '
    set -eu
    target="$1"
    test -d "$target"
    test -n "$(ls -A "$target")"
  ' sh "$restore_target" >/dev/null

  FAILURE_CATEGORY="status_recording_failure"
  record_terminal technical-restore --snapshot-id "$SNAPSHOT_ID"

  printf 'acceptance_snapshot_id=%s\n' "$SNAPSHOT_ID"
  printf 'acceptance_root_object_id=%s\n' "$ROOT_OBJECT_ID"
  printf 'acceptance_restore_target=%s\n' "$restore_target"
  echo "acceptance restore retained for application-specific validation"
}

FAILURE_CATEGORY="configuration_failure"
validate_compose

case "$ACTION" in
  acceptance-recovery)
    run_acceptance_recovery
    ;;

  validate)
    validate_source_scope
    FAILURE_CATEGORY="status_recording_failure"
    record_terminal succeeded
    ;;

  repository-status)
    validate_repository_identity
    repository_status
    FAILURE_CATEGORY="status_recording_failure"
    record_terminal succeeded
    ;;

  connect-repository)
    FAILURE_CATEGORY="repository_connection_failure"
    if repository_status >/dev/null 2>&1; then
      echo "repository is already connected; refusing to replace the existing local repository configuration" >&2
      exit 1
    fi

    compose run --rm backup __goreecloud_connect_existing_sftp
    validate_repository_identity
    repository_status
    FAILURE_CATEGORY="status_recording_failure"
    record_terminal succeeded
    ;;

  backup)
    validate_source_scope
    validate_repository_identity
    create_validated_snapshot "goreecloud:scheduled"
    FAILURE_CATEGORY="status_recording_failure"
    record_terminal succeeded --snapshot-id "$SNAPSHOT_ID"
    printf 'scheduled_snapshot_id=%s\n' "$SNAPSHOT_ID"
    ;;

  *)
    FAILURE_CATEGORY="unsupported_action"
    echo "unsupported action: $ACTION" >&2
    echo "usage: $0 [backup|acceptance-recovery|connect-repository|repository-status|validate]" >&2
    exit 2
    ;;
esac
