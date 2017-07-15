#!/bin/sh
GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build -v -a -o bin/arm/go-raspberry-pi-timelapse
file bin/arm/go-raspberry-pi-timelapse
du bin/arm/go-raspberry-pi-timelapse
