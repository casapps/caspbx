#!/usr/bin/env bash
set -euo pipefail

if ! command -v incus &>/dev/null; then
    echo "ERROR: incus not found. Install incus or use tests/docker.sh"
    exit 1
fi

PROJECTNAME=$(basename "$PWD")
PROJECTORG=$(basename "$(dirname "$PWD")")
CONTAINER_NAME="test-$PROJECTNAME-$$"
INCUS_IMAGE="images:debian/trixie"

mkdir -p "${TMPDIR:-/tmp}/$PROJECTORG"
BUILD_DIR=$(mktemp -d "${TMPDIR:-/tmp}/$PROJECTORG/$PROJECTNAME-XXXXXX")
trap 'rm -rf "$BUILD_DIR"; incus delete "$CONTAINER_NAME" --force 2>/dev/null || true' EXIT

GODIR="${HOME}/.local/share/go"
GOCACHE="${HOME}/.local/share/go/build"
mkdir -p "$GODIR" "$GOCACHE"

GO_DOCKER="docker run --rm -v $(pwd):/build -v ${BUILD_DIR}:/out -v ${GOCACHE}:/root/.cache/go-build -v ${GODIR}:/go -w /build -e CGO_ENABLED=0 golang:alpine"

echo "Building server binary in Docker..."
$GO_DOCKER go build -o "/out/$PROJECTNAME" ./src

echo "Launching Incus container..."
incus launch "$INCUS_IMAGE" "$CONTAINER_NAME"
sleep 2

incus file push "$BUILD_DIR/$PROJECTNAME" "$CONTAINER_NAME/usr/local/bin/"
incus exec "$CONTAINER_NAME" -- chmod +x "/usr/local/bin/$PROJECTNAME"
incus exec "$CONTAINER_NAME" -- bash -lc "apt-get update >/dev/null && apt-get install -y file >/dev/null"

echo "Running bootstrap tests in Incus..."
incus exec "$CONTAINER_NAME" -- bash -lc "
  set -e
  $PROJECTNAME --version
  $PROJECTNAME --help
  file /usr/local/bin/$PROJECTNAME
  $PROJECTNAME | grep -q 'bootstrap scaffold is active'
  $PROJECTNAME --status | grep -q 'bootstrap status: unavailable'
  cp /usr/local/bin/$PROJECTNAME /tmp/renamed-server
  chmod +x /tmp/renamed-server
  /tmp/renamed-server --help | grep -q 'renamed-server'
"

echo "Incus bootstrap tests completed successfully"
