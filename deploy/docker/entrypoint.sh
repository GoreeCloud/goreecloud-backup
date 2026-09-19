#!/bin/sh
set -eu

read_required_secret() {
  secret_name="$1"
  secret_path="$2"

  if [ ! -r "$secret_path" ]; then
    echo "required secret '$secret_name' is not readable at $secret_path" >&2
    exit 1
  fi

  secret_value="$(cat "$secret_path")"
  if [ -z "$secret_value" ]; then
    echo "required secret '$secret_name' is empty" >&2
    exit 1
  fi

  case "$secret_name" in
    repository_password)
      KOPIA_PASSWORD="$secret_value"
      export KOPIA_PASSWORD
      ;;
    server_password)
      KOPIA_SERVER_PASSWORD="$secret_value"
      export KOPIA_SERVER_PASSWORD
      ;;
    server_control_password)
      KOPIA_SERVER_CONTROL_PASSWORD="$secret_value"
      export KOPIA_SERVER_CONTROL_PASSWORD
      ;;
    *)
      echo "unsupported secret name: $secret_name" >&2
      exit 1
      ;;
  esac

  unset secret_value
}

read_required_secret   repository_password   "${GOREECLOUD_BACKUP_REPOSITORY_PASSWORD_FILE:-/run/secrets/repository_password}"

read_required_secret   server_password   "${GOREECLOUD_BACKUP_SERVER_PASSWORD_FILE:-/run/secrets/server_password}"

read_required_secret   server_control_password   "${GOREECLOUD_BACKUP_SERVER_CONTROL_PASSWORD_FILE:-/run/secrets/server_control_password}"

: "${KOPIA_SERVER_USERNAME:?KOPIA_SERVER_USERNAME is required}"
: "${KOPIA_SERVER_CONTROL_USER:?KOPIA_SERVER_CONTROL_USER is required}"

exec /usr/local/bin/goreecloud-backup "$@"
