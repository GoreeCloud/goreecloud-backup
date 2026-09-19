#!/bin/sh
set -eu

KOPIA_STACK_DIR="${KOPIA_STACK_DIR:-/srv/docker/stacks/kopia}"
KOPIA_COMPOSE_FILE="${KOPIA_COMPOSE_FILE:-${KOPIA_STACK_DIR}/compose.yaml}"
KOPIA_TIMER="${KOPIA_TIMER:-goreecloud-kopia-backup.timer}"
KOPIA_SERVICE_UNIT="${KOPIA_SERVICE_UNIT:-goreecloud-kopia-backup.service}"

TARGET_STACK_DIR="${GOREECLOUD_BACKUP_STACK_DIR:-/srv/docker/stacks/goreecloud-backup}"
TARGET_COMPOSE_FILE="${TARGET_STACK_DIR}/compose.yaml"
TARGET_SOURCES_FILE="${TARGET_STACK_DIR}/compose.sources.yaml"
TARGET_ENV_FILE="${TARGET_STACK_DIR}/.env"

echo "=== GoreeCloud Backup VPS replacement preflight ==="
echo "Read-only: this script does not stop services, change repositories, modify timers, or print secret contents."

echo
echo "=== Docker / Compose versions ==="
docker version --format 'Docker server: {{.Server.Version}}'
docker compose version

blocker=0

echo
echo "=== Retired Kopia VPS runtime residual check ==="

if [ -r "$KOPIA_COMPOSE_FILE" ]; then
  echo "unexpected_legacy_compose=present:$KOPIA_COMPOSE_FILE"
  echo "The canonical task records the VPS-side Kopia stack as retired; unexpected reappearance requires review." >&2
  blocker=1
else
  echo "legacy_kopia_compose=absent"
fi

legacy_containers="$(docker ps -a --format '{{.Names}} {{.Image}}' | grep -Ei '(^|[[:space:]/])kopia([[:space:]/:@]|$)|kopia/kopia' || true)"
if [ -n "$legacy_containers" ]; then
  echo "unexpected_legacy_containers:"
  printf '%s\n' "$legacy_containers"
  blocker=1
else
  echo "legacy_kopia_containers=absent"
fi

legacy_images="$(docker image ls --format '{{.Repository}}:{{.Tag}} {{.ID}}' | grep -E '^kopia/kopia:' || true)"
if [ -n "$legacy_images" ]; then
  echo "unexpected_legacy_images:"
  printf '%s\n' "$legacy_images"
  blocker=1
else
  echo "legacy_kopia_images=absent"
fi

for unit in "$KOPIA_TIMER" "$KOPIA_SERVICE_UNIT"; do
  if systemctl list-unit-files --no-legend "$unit" 2>/dev/null | grep -q .; then
    echo "unexpected_legacy_unit=$unit"
    systemctl is-enabled "$unit" || true
    systemctl is-active "$unit" || true
    blocker=1
  else
    echo "legacy_unit_absent=$unit"
  fi
done

echo
echo "=== GoreeCloud Backup target material ==="

for required_file in "$TARGET_COMPOSE_FILE" "$TARGET_SOURCES_FILE" "$TARGET_ENV_FILE"; do
  if [ -r "$required_file" ]; then
    echo "target_file_readable=$required_file"
  else
    echo "target_file_missing=$required_file"
    blocker=1
  fi
done

if [ -r "$TARGET_SOURCES_FILE" ]; then
  echo
  echo "=== Candidate source bindings declared in compose.sources.yaml ==="
  # The file is expected to contain paths only; do not print the .env or secret files.
  sed -n -e '/^[[:space:]]*source:/p' -e '/^[[:space:]]*target:/p' -e '/^[[:space:]]*read_only:/p' "$TARGET_SOURCES_FILE"
fi

echo
echo "=== Current scheduling target ==="
systemctl is-enabled goreecloud-backup.timer || true
systemctl is-active goreecloud-backup.timer || true
systemctl is-enabled goreecloud-backup.service || true
systemctl is-active goreecloud-backup.service || true

if [ "$blocker" -ne 0 ]; then
  echo
  echo "Preflight found unresolved replacement-readiness blockers." >&2
  exit 1
fi

echo
echo "Preflight passed the non-destructive host/material checks."
echo "Repository access, backup creation, integrity verification, restore validation, and schedule acceptance remain separate required evidence."
