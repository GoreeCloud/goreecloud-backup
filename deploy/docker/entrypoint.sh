#!/bin/sh
set -eu

repository_password_path="${GOREECLOUD_BACKUP_REPOSITORY_PASSWORD_FILE:-/run/secrets/repository_password}"

if [ -r "$repository_password_path" ]; then
  KOPIA_PASSWORD="$(cat "$repository_password_path")"
  if [ -z "$KOPIA_PASSWORD" ]; then
    echo "repository password secret is empty" >&2
    exit 1
  fi
  export KOPIA_PASSWORD
fi

exec /usr/local/bin/goreecloud-backup "$@"
