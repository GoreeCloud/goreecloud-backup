#!/bin/sh
set -eu

KOPIA_STACK_DIR="${KOPIA_STACK_DIR:-/srv/docker/stacks/kopia}"
KOPIA_COMPOSE_FILE="${KOPIA_COMPOSE_FILE:-${KOPIA_STACK_DIR}/compose.yaml}"
KOPIA_SERVICE="${KOPIA_SERVICE:-kopia}"
KOPIA_TIMER="${KOPIA_TIMER:-goreecloud-kopia-backup.timer}"
KOPIA_SERVICE_UNIT="${KOPIA_SERVICE_UNIT:-goreecloud-kopia-backup.service}"

echo "=== GoreeCloud Backup replacement preflight ==="
echo "This script is read-only. It does not stop Kopia, change a repository, or print secret file contents."

if [ ! -r "$KOPIA_COMPOSE_FILE" ]; then
  echo "Kopia Compose file not found/readable: $KOPIA_COMPOSE_FILE" >&2
  exit 1
fi

echo
echo "=== Docker / Compose versions ==="
docker version --format 'Docker server: {{.Server.Version}}'
docker compose version

echo
echo "=== Current Kopia Compose services ==="
docker compose --file "$KOPIA_COMPOSE_FILE" config --no-env-resolution --services

if ! docker compose --file "$KOPIA_COMPOSE_FILE" config --no-env-resolution --services | grep -Fxq "$KOPIA_SERVICE"; then
  echo "expected Kopia service '$KOPIA_SERVICE' is not present; set KOPIA_SERVICE to the verified service name" >&2
  exit 1
fi

echo
echo "=== Current Kopia image reference ==="
docker compose --file "$KOPIA_COMPOSE_FILE" config --no-env-resolution --images

echo
echo "=== Current bind-mount source/target/read-only metadata ==="
docker compose --file "$KOPIA_COMPOSE_FILE" config --format json   | python3 -c '
import json, sys
data = json.load(sys.stdin)
service = data.get("services", {}).get(sys.argv[1], {})
for v in service.get("volumes", []):
    if isinstance(v, dict) and v.get("type") == "bind":
        print(f"{v.get('source','?')} -> {v.get('target','?')} read_only={bool(v.get('read_only', False))}")
' "$KOPIA_SERVICE"

echo
echo "=== Current scheduling units ==="
systemctl is-enabled "$KOPIA_TIMER" || true
systemctl is-active "$KOPIA_TIMER" || true
systemctl is-enabled "$KOPIA_SERVICE_UNIT" || true
systemctl is-active "$KOPIA_SERVICE_UNIT" || true

echo
echo "=== Current repository status through the existing Kopia stack ==="
docker compose --file "$KOPIA_COMPOSE_FILE" run --rm "$KOPIA_SERVICE" repository status

echo
echo "=== Current snapshot summary ==="
docker compose --file "$KOPIA_COMPOSE_FILE" run --rm "$KOPIA_SERVICE" snapshot list --all --json   | python3 -c '
import json, sys
data = json.load(sys.stdin)
items = data if isinstance(data, list) else []
print(f"snapshot_count={len(items)}")
if items:
    latest = max(items, key=lambda x: x.get("startTime", ""))
    print("latest_snapshot_id=" + str(latest.get("id", "unknown")))
    print("latest_start_time=" + str(latest.get("startTime", "unknown")))
'

echo
echo "Preflight completed. Preserve this output with the cutover evidence."
