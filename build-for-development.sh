#!/bin/sh
# This creates a build for local development, quick iteration.
GIT_COMMIT=`git rev-parse --short HEAD`
BUILT_AT=`date`
go build -v -ldflags="-X 'main.GitCommit=${GIT_COMMIT}' -X 'main.BuiltAt=${BUILT_AT}'"
