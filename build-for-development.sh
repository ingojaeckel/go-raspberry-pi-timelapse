#!/bin/sh
# This creates a build for local development, quick iteration.
GIT_COMMIT=`git rev-parse HEAD`
go build -v -ldflags="-X main.GitCommit=${GIT_COMMIT}"
