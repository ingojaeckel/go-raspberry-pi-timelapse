#!/bin/sh
GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -a -o bin/arm/go-raspberry-pi-timelapse
