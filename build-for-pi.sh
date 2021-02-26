#!/bin/sh
set -e -x
go vet ./...
go test -v -cover ./...

GIT_COMMIT=`git rev-parse --short HEAD`
BUILT_AT=`date`
GOOS=linux \
GOARCH=arm \
GOARM=6 \
CGO_ENABLED=0 \
go build -v -a -o bin/arm/go-raspberry-pi-timelapse -ldflags="-s -w -X 'main.gitCommit=${GIT_COMMIT}' -X 'main.builtAt=${BUILT_AT}'"
file bin/arm/go-raspberry-pi-timelapse
du bin/arm/go-raspberry-pi-timelapse
