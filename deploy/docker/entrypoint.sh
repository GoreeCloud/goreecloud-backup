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
  : "${GOREECLOUD_BACKUP_CLIENT_HOSTNAME:?Set the repository client hostname}"
  : "${GOREECLOUD_BACKUP_CLIENT_USERNAME:?Set the repository client username}"
  : "${GOREECLOUD_BACKUP_REPOSITORY_ID:?Set the expected preserved repository Unique ID}"
  : "${GOREECLOUD_BACKUP_SFTP_HOST:?Set the verified SFTP host}"
  : "${GOREECLOUD_BACKUP_SFTP_PORT:?Set the verified SFTP port}"
  : "${GOREECLOUD_BACKUP_SFTP_USERNAME:?Set the verified SFTP username}"
  : "${GOREECLOUD_BACKUP_SFTP_REPOSITORY_PATH:?Set the verified SFTP repository path}"

  config_path="${KOPIA_CONFIG_PATH:-/app/config/repository.config}"

  if [ -e "$config_path" ]; then
    echo "repository config already exists; refusing to overwrite: $config_path" >&2
    exit 1
  fi

  exec "$backup_bin" repository connect \
    --override-hostname "$GOREECLOUD_BACKUP_CLIENT_HOSTNAME" \
    --override-username "$GOREECLOUD_BACKUP_CLIENT_USERNAME" \
    sftp \
    --host "$GOREECLOUD_BACKUP_SFTP_HOST" \
    --port "$GOREECLOUD_BACKUP_SFTP_PORT" \
    --username "$GOREECLOUD_BACKUP_SFTP_USERNAME" \
    --path "$GOREECLOUD_BACKUP_SFTP_REPOSITORY_PATH" \
    --keyfile /run/secrets/kopia-sftp-ed25519 \
    --known-hosts /run/secrets/kopia-sftp-known_hosts
}

validate_connected_repository() {
  : "${GOREECLOUD_BACKUP_CLIENT_HOSTNAME:?Set the repository client hostname}"
  : "${GOREECLOUD_BACKUP_CLIENT_USERNAME:?Set the repository client username}"
  : "${GOREECLOUD_BACKUP_REPOSITORY_ID:?Set the expected preserved repository Unique ID}"

  status_output="$("$backup_bin" repository status)"

  printf '%s\n' "$status_output" \
    | grep -Eq "^Hostname:[[:space:]]+${GOREECLOUD_BACKUP_CLIENT_HOSTNAME}$" \
    || {
      echo "repository client hostname does not match the governed deployment identity" >&2
      exit 1
    }

  printf '%s\n' "$status_output" \
    | grep -Eq "^Username:[[:space:]]+${GOREECLOUD_BACKUP_CLIENT_USERNAME}$" \
    || {
      echo "repository client username does not match the governed deployment identity" >&2
      exit 1
    }

  printf '%s\n' "$status_output" \
    | grep -Eq '^Read-only:[[:space:]]+false \
    || {
      echo "repository connection is unexpectedly read-only" >&2
      exit 1
    }

  printf '%s\n' "$status_output" \
    | grep -Eq '^Storage type:[[:space:]]+sftp \
    || {
      echo "repository storage type is not the governed SFTP boundary" >&2
      exit 1
    }

  printf '%s\n' "$status_output" \
    | grep -Eq "^Unique ID:[[:space:]]+${GOREECLOUD_BACKUP_REPOSITORY_ID}$" \
    || {
      echo "connected repository Unique ID does not match the preserved recovery repository" >&2
      exit 1
    }

  echo "repository identity validated"
}

case "${1:-}" in
  __goreecloud_connect_existing_sftp)
    shift
    if [ "$#" -ne 0 ]; then
      echo "internal repository-connect operation does not accept extra arguments" >&2
      exit 2
    fi
    connect_existing_sftp_repository
    ;;

  __goreecloud_validate_repository)
    shift
    if [ "$#" -ne 0 ]; then
      echo "internal repository-validation operation does not accept extra arguments" >&2
      exit 2
    fi
    validate_connected_repository
    ;;

  *)
    exec "$backup_bin" "$@"
    ;;
esac
