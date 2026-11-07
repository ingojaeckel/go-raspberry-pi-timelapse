#!/bin/sh
# Builds the Go server (with the frontend build embedded: run `npm run build`
# first) and runs it in a fresh folder, for the Playwright smoke test. PATH
# leaves out ffmpeg and rpicam-still, so no real camera is ever used.
set -eu
dir=$(mktemp -d)
trap 'rm -rf "$dir"' EXIT
(cd "$(dirname "$0")/../../backend" && go build -tags nethttpomithttp2 -o "$dir/timelapse" .)
cd "$dir"
PATH=/usr/bin:/bin ./timelapse -port "127.0.0.1:${E2E_PORT:-18080}" -storageFolder photos
