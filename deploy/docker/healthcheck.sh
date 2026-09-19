#!/bin/sh
set -eu

secret_path="${GOREECLOUD_BACKUP_SERVER_CONTROL_PASSWORD_FILE:-/run/secrets/server_control_password}"

if [ ! -r "$secret_path" ]; then
  exit 1
fi

control_password="$(cat "$secret_path")"
if [ -z "$control_password" ]; then
  exit 1
fi

KOPIA_SERVER_USERNAME="${KOPIA_SERVER_CONTROL_USER:-server-control}" \
KOPIA_SERVER_PASSWORD="$control_password" \
  /usr/local/bin/goreecloud-backup server status \
    --address=http://127.0.0.1:51515 >/dev/null 2>&1
