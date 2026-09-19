#!/bin/sh
set -eu

backup_bin="/usr/local/bin/goreecloud-backup"
repository_password_path="${GOREECLOUD_BACKUP_REPOSITORY_PASSWORD_FILE:-/run/secrets/repository_password}"

if [ -r "$repository_password_path" ]; then
  KOPIA_PASSWORD="$(cat "$repository_password_path")"
  if [ -z "$KOPIA_PASSWORD" ]; then
    echo "repository password secret is empty" >&2
    exit 1
  fi
  export KOPIA_PASSWORD
fi

connect_existing_sftp_repository() {
  : "${KOPIA_PASSWORD:?repository password secret is required to connect the repository}"
  : "${GOREECLOUD_BACKUP_SFTP_HOST:?Set the verified SFTP host}"
  : "${GOREECLOUD_BACKUP_SFTP_PORT:?Set the verified SFTP port}"
  : "${GOREECLOUD_BACKUP_SFTP_USERNAME:?Set the verified SFTP username}"
  : "${GOREECLOUD_BACKUP_SFTP_REPOSITORY_PATH:?Set the verified SFTP repository path}"

  config_path="${KOPIA_CONFIG_PATH:-/app/config/repository.config}"

  if [ -e "$config_path" ]; then
    echo "repository config already exists; refusing to overwrite: $config_path" >&2
    exit 1
  fi

  exec "$backup_bin" repository connect sftp \
    --host "$GOREECLOUD_BACKUP_SFTP_HOST" \
    --port "$GOREECLOUD_BACKUP_SFTP_PORT" \
    --username "$GOREECLOUD_BACKUP_SFTP_USERNAME" \
    --path "$GOREECLOUD_BACKUP_SFTP_REPOSITORY_PATH" \
    --keyfile /run/secrets/kopia-sftp-ed25519 \
    --known-hosts /run/secrets/kopia-sftp-known_hosts
}

if [ "${1:-}" = "__goreecloud_connect_existing_sftp" ]; then
  shift
  if [ "$#" -ne 0 ]; then
    echo "internal repository-connect operation does not accept extra arguments" >&2
    exit 2
  fi
  connect_existing_sftp_repository
fi

exec "$backup_bin" "$@"
