#!/bin/sh
# Abort on error
set -e
# Print commands while running this
set -x

# Run static analysis and tests.
go vet ./...
go test -v -cover ./...

# This creates a build for local development, quick iteration.
GIT_COMMIT=`git rev-parse --short HEAD`
BUILT_AT=`date`
go build -v -ldflags="-X 'main.GitCommit=${GIT_COMMIT}' -X 'main.BuiltAt=${BUILT_AT}'"
