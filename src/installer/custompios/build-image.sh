#!/usr/bin/env bash
# Builds the Pi Zero W disk image with CustomPiOS inside Docker, so it runs the
# same on macOS (Docker Desktop, Apple Silicon or Intel) and Linux, including CI.
#
# Usually called by `make image`, which first builds the app into
# modules/timelapsepi/filesystem/root/opt/timelapse/. Output: workspace/*.zip
#
# Environment:
#   DIST_VERSION   version in the image name (make image: git describe)
set -euo pipefail
cd "$(dirname "$0")"

# CustomPiOS 2.x build container, pinned (multi-arch: amd64, arm64).
CUSTOMPIOS_IMAGE=ghcr.io/guysoft/custompios@sha256:28c77a0c32b95b984e1c9a517b0e71df9f632f0a99909a15c799b4094f8a8d22
APP=modules/timelapsepi/filesystem/root/opt/timelapse/timelapse

die() { echo "Error: $*" >&2; exit 1; }

sha256() {
  if command -v sha256sum >/dev/null; then sha256sum "$1"; else shasum -a 256 "$1"; fi | cut -d' ' -f1
}

# Downloads the base image pinned in ./config unless it is already there.
fetch_base_image() {
  local url sum file
  url=$(DIST_PATH=. bash -c 'source ./config && echo "$TIMELAPSEPI_BASE_IMAGE_URL"')
  sum=$(DIST_PATH=. bash -c 'source ./config && echo "$TIMELAPSEPI_BASE_IMAGE_SHA256"')
  file=image/${url##*/}
  mkdir -p image
  if [ -f "$file" ] && [ "$(sha256 "$file")" = "$sum" ]; then
    return
  fi
  echo "Downloading $url"
  curl -fL --retry 3 -o "$file.partial" "$url"
  [ "$(sha256 "$file.partial")" = "$sum" ] || die "$url does not match the SHA-256 in config"
  mv "$file.partial" "$file"
}

[ -x "$APP" ] || die "$APP is missing. Run 'make image', which builds it first."
docker info >/dev/null 2>&1 || die "Docker is not running. Start Docker Desktop (macOS) or the docker service."
fetch_base_image

# --privileged: CustomPiOS loop-mounts the image partitions and chroots into
# them; 32-bit ARM code runs through qemu-arm-static from the container.
# CustomPiOS leaves a loop device attached to the image; loop devices belong to
# the host (or Docker Desktop's VM), so detach it whether the build failed or not.
docker run --rm --privileged \
  -e DIST_VERSION="${DIST_VERSION:-dev}" \
  -v "$PWD":/distro \
  "$CUSTOMPIOS_IMAGE" bash -c '
    build; rc=$?
    losetup -a | grep "/workspace/.*timelapsepi" | cut -d: -f1 | xargs -r losetup -d
    exit $rc'

ls -l workspace/*.zip
