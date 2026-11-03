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
  # -C -: resume an interrupted download of the .partial file.
  curl -fL --retry 5 --retry-all-errors -C - -o "$file.partial" "$url"
  if [ "$(sha256 "$file.partial")" != "$sum" ]; then
    rm -f "$file.partial"
    die "$url does not match the SHA-256 in config"
  fi
  mv "$file.partial" "$file"
}

# CustomPiOS starts only the chroot's first bash through qemu-arm-static; every
# program it runs needs the host kernel to run 32-bit ARM code (binfmt_misc).
# Docker Desktop registers that; plain Linux arm64/amd64 hosts may not.
check_arm32() {
  docker run --rm -q --platform linux/arm/v6 busybox:1.36.1 true >/dev/null 2>&1 ||
    die "this host cannot run 32-bit ARM programs. Register qemu with:
  docker run --privileged --rm tonistiigi/binfmt --install arm"
}

[ -f "$APP" ] || die "$APP is missing. Run 'make image', which builds it first."
docker info >/dev/null 2>&1 || die "Docker is not running. Start Docker Desktop (macOS) or the docker service."
check_arm32
fetch_base_image

# --privileged: CustomPiOS loop-mounts the image partitions and chroots into
# them. --init with TINI_KILL_PROCESS_GROUP: Ctrl-C reaches the whole build, so
# CustomPiOS unmounts the image and the container stops (a bare bash as PID 1
# would ignore the signal and keep building in the background).
# Afterwards, whether the build failed or not:
# - detach the loop devices the build left attached to images in workspace/
#   (they belong to the host, or Docker Desktop's VM, and would pile up),
#   leaving any that existed before alone;
# - hand workspace/ to the calling user (the container runs as root).
docker run --rm --privileged --init -e TINI_KILL_PROCESS_GROUP=1 \
  -e DIST_VERSION="${DIST_VERSION:-dev}" \
  -e HOST_UID="$(id -u)" -e HOST_GID="$(id -g)" \
  -v "$PWD":/distro \
  "$CUSTOMPIOS_IMAGE" bash -c '
    trap "echo Interrupted" INT TERM
    before=$(losetup -a | cut -d: -f1)
    build; rc=$?
    for dev in $(losetup -a | grep "/workspace/[^/]*\.img" | cut -d: -f1); do
      grep -qx "$dev" <<<"$before" || losetup -d "$dev"
    done
    chown -R "$HOST_UID:$HOST_GID" /distro/workspace || true
    exit $rc'

ls -l workspace/*.zip
