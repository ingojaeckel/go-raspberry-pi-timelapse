[![Build Status](https://img.shields.io/travis/ingojaeckel/go-raspberry-pi-timelapse.svg)](https://travis-ci.org/ingojaeckel/go-raspberry-pi-timelapse)

# Go Raspberry Pi Timelapse

Golang application to take timelapse pictures on a ARM platform like Raspberry Pi.

## Setting up Pi for Timelapse

* Install Raspbian Jessie Lite, see [docs](https://www.raspberrypi.org/downloads/raspbian/)
* Install required software:

    aptitude update
    aptitude safe-upgrade -y
    apt-get install -y htop screen
    wget https://github.com/ingojaeckel/go-raspberry-pi-timelapse/raw/master/bin/arm/go-raspberry-pi-timelapse

* Run the main service: `./go-raspberry-pi-timelapse`