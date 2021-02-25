#!/bin/sh
GIT_COMMIT=`git rev-parse HEAD`
GOOS=linux
GOARCH=arm
GOARM=6
CGO_ENABLED=0
go build -v -a -o bin/arm/go-raspberry-pi-timelapse -ldflags="-s -w -X main.GitCommit=${GIT_COMMIT}"
file bin/arm/go-raspberry-pi-timelapse
du bin/arm/go-raspberry-pi-timelapse
