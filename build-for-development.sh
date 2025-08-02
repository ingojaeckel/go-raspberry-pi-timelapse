#!/bin/sh
# Abort on error
set -e
# Print commands while running this
set -x

# Run static analysis and tests.
go vet ./...
go test -v -cover ./...

# This creates a build for local development, quick iteration.
# OpenCV support will be enabled if OpenCV libraries are available
GIT_COMMIT=`git rev-parse --short HEAD`
BUILT_AT=`date`

# Try to build with OpenCV support, fall back to basic build if not available
if pkg-config --exists opencv4 2>/dev/null; then
    echo "Building with native Go OpenCV support (OpenCV detected)"
    CGO_ENABLED=1 go build -v -tags opencv -ldflags="-X 'main.gitCommit=${GIT_COMMIT}' -X 'main.builtAt=${BUILT_AT}'"
else
    echo "Building without OpenCV support (OpenCV not detected)"
    go build -v -ldflags="-X 'main.gitCommit=${GIT_COMMIT}' -X 'main.builtAt=${BUILT_AT}'"
fi
