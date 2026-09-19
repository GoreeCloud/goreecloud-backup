#!/bin/sh
set -eu

: "${GO_BUILD_IMAGE:?Set GO_BUILD_IMAGE to an exact tag@sha256 digest}"
: "${RUNTIME_IMAGE:?Set RUNTIME_IMAGE to an exact tag@sha256 digest}"
: "${GOREECLOUD_BACKUP_VERSION:?Set the exact GoreeCloud Backup version}"
: "${GOREECLOUD_BACKUP_COMMIT:?Set the exact source commit SHA}"
: "${GOREECLOUD_BACKUP_IMAGE_TAG:?Set a unique destination image tag}"

require_digest_reference() {
  value="$1"
  label="$2"

  case "$value" in
    *@sha256:*)
      ;;
    *)
      echo "$label must be an exact tag-and-digest image reference" >&2
      exit 1
      ;;
  esac
}

require_digest_reference "$GO_BUILD_IMAGE" GO_BUILD_IMAGE
require_digest_reference "$RUNTIME_IMAGE" RUNTIME_IMAGE

case "$GOREECLOUD_BACKUP_COMMIT" in
  ""|*[!0-9a-f]*)
    echo "GOREECLOUD_BACKUP_COMMIT must be a lowercase hexadecimal commit SHA" >&2
    exit 1
    ;;
esac

if [ "${#GOREECLOUD_BACKUP_COMMIT}" -ne 40 ]; then
  echo "GOREECLOUD_BACKUP_COMMIT must be a full 40-character commit SHA" >&2
  exit 1
fi

case "$GOREECLOUD_BACKUP_IMAGE_TAG" in
  *:latest|latest)
    echo "GOREECLOUD_BACKUP_IMAGE_TAG must not use latest" >&2
    exit 1
    ;;
esac

if [ "${GOREECLOUD_BACKUP_PUSH:-false}" = "true" ]; then
  : "${GOREECLOUD_BACKUP_METADATA_FILE:?Set a buildx metadata output path for pushed builds}"
  : "${GOREECLOUD_BACKUP_SBOM_GENERATOR:?Set an exact tag@sha256 SBOM generator image}"

  require_digest_reference "$GOREECLOUD_BACKUP_SBOM_GENERATOR" GOREECLOUD_BACKUP_SBOM_GENERATOR

  docker buildx version >/dev/null

  docker buildx build --pull \
    --file deploy/docker/Dockerfile \
    --build-arg "GO_BUILD_IMAGE=$GO_BUILD_IMAGE" \
    --build-arg "RUNTIME_IMAGE=$RUNTIME_IMAGE" \
    --build-arg "GOREECLOUD_BACKUP_VERSION=$GOREECLOUD_BACKUP_VERSION" \
    --build-arg "GOREECLOUD_BACKUP_COMMIT=$GOREECLOUD_BACKUP_COMMIT" \
    --tag "$GOREECLOUD_BACKUP_IMAGE_TAG" \
    --provenance=mode=max \
    --attest "type=sbom,generator=$GOREECLOUD_BACKUP_SBOM_GENERATOR" \
    --metadata-file "$GOREECLOUD_BACKUP_METADATA_FILE" \
    --push \
    .

  exit 0
fi

docker build --pull \
  --file deploy/docker/Dockerfile \
  --build-arg "GO_BUILD_IMAGE=$GO_BUILD_IMAGE" \
  --build-arg "RUNTIME_IMAGE=$RUNTIME_IMAGE" \
  --build-arg "GOREECLOUD_BACKUP_VERSION=$GOREECLOUD_BACKUP_VERSION" \
  --build-arg "GOREECLOUD_BACKUP_COMMIT=$GOREECLOUD_BACKUP_COMMIT" \
  --tag "$GOREECLOUD_BACKUP_IMAGE_TAG" \
  .

docker image inspect \
  --format 'Built image ID: {{.Id}}' \
  "$GOREECLOUD_BACKUP_IMAGE_TAG"
