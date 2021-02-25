#!/bin/sh
GIT_COMMIT=`git rev-parse --short HEAD`
BUILT_AT=`date`
GOOS=linux
GOARCH=arm
GOARM=6
CGO_ENABLED=0
go build -v -a -o bin/arm/go-raspberry-pi-timelapse -ldflags="-s -w -X 'main.GitCommit=${GIT_COMMIT}' -X 'main.BuiltAt=${BUILT_AT}'"
file bin/arm/go-raspberry-pi-timelapse
du bin/arm/go-raspberry-pi-timelapse
