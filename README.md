[![Build Status](https://img.shields.io/travis/ingojaeckel/go-raspberry-pi-timelapse.svg)](https://travis-ci.org/ingojaeckel/go-raspberry-pi-timelapse)

# Go Raspberry Pi Timelapse

This repository contains documentation and source code to help running a [Raspberry Pi Zero](https://www.raspberrypi.org/products/raspberry-pi-zero-w/) based timelapse camera.

## Parts List

This project was created for a timelapse system consisting of the following core components: a Pi Zero W with a camera board, a case, and an SD card. The following sections describe the components in more detail.

### Core parts for regular deployment (about $65)

* [Raspberry Pi Zero W](https://www.adafruit.com/product/3400) ($10)
* [Raspberry Pi Camera Board v2](https://www.adafruit.com/product/3099) ($30)
* [Raspberry Pi Zero Camera Cable](https://www.adafruit.com/product/3157) ($6)
* [C4Labs Zebra Zero Case](https://www.adafruit.com/product/3003) ($7)
* [microSDHC Card](https://www.adafruit.com/product/2767) ($12)

### Additional parts to include a battery backed clock (about $8)

* [Break-away 0.1" 2x20-pin Strip Dual Male Header](https://www.adafruit.com/product/2822) ($1)
* [CR1220 12mm Diameter - 3V Lithium Coin Cell Battery](https://www.adafruit.com/product/380) ($1)
* [Adafruit PiRTC - PCF8523 Real Time Clock for Raspberry Pi](https://www.adafruit.com/product/3386) ($6)

### Additional Parts for Development (about $6)

* [USB OTG Host Cable - MicroB OTG male to A female](https://www.adafruit.com/product/1099) ($3)
* [Mini HDMI Plug to Standard HDMI Jack Adapter](https://www.adafruit.com/product/2819) ($3)

## Setting up Pi for Timelapse

* Install Raspbian Jessie Lite, see [docs](https://www.raspberrypi.org/downloads/raspbian/)
* Install required software and run the main service:
```
    aptitude update
    aptitude safe-upgrade -y
    apt-get install -y htop screen
    wget https://github.com/ingojaeckel/go-raspberry-pi-timelapse/raw/master/bin/arm/go-raspberry-pi-timelapse
    ./go-raspberry-pi-timelapse
```
