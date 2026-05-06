#!/usr/bin/env bash
set -euo pipefail

PROJECTNAME=$(basename "$PWD")
PROJECTORG=$(basename "$(dirname "$PWD")")

mkdir -p "${TMPDIR:-/tmp}/$PROJECTORG"
BUILD_DIR=$(mktemp -d "${TMPDIR:-/tmp}/$PROJECTORG/$PROJECTNAME-XXXXXX")
trap 'rm -rf "$BUILD_DIR"' EXIT

GODIR="${HOME}/.local/share/go"
GOCACHE="${HOME}/.local/share/go/build"
mkdir -p "$GODIR" "$GOCACHE"

GO_DOCKER="docker run --rm -v $(pwd):/build -v ${BUILD_DIR}:/out -v ${GOCACHE}:/root/.cache/go-build -v ${GODIR}:/go -w /build -e CGO_ENABLED=0 golang:alpine"

echo "Building server binary in Docker..."
$GO_DOCKER go build -o "/out/$PROJECTNAME" ./src

echo "Testing bootstrap binary in Docker (Alpine)..."
docker run --rm -v "$BUILD_DIR:/app" alpine:latest sh -c "
  set -e
  apk add --no-cache bash file >/dev/null
  chmod +x /app/$PROJECTNAME

  echo '=== Version Check ==='
  /app/$PROJECTNAME --version

  echo '=== Help Check ==='
  /app/$PROJECTNAME --help

  echo '=== Binary Info ==='
  file /app/$PROJECTNAME

  echo '=== Bootstrap Start Check ==='
  /app/$PROJECTNAME | grep -q 'bootstrap scaffold is active'

  echo '=== Status Check ==='
  /app/$PROJECTNAME --status | grep -q 'bootstrap status: unavailable'

  echo '=== Binary Rename Check ==='
  cp /app/$PROJECTNAME /app/renamed-server
  chmod +x /app/renamed-server
  /app/renamed-server --help | grep -q 'renamed-server'
"

echo "Docker bootstrap tests completed successfully"
