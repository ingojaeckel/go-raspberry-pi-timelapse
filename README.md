[![Build Status](https://img.shields.io/travis/ingojaeckel/go-raspberry-pi-timelapse.svg)](https://travis-ci.org/ingojaeckel/go-raspberry-pi-timelapse)

# Go Raspberry Pi Timelapse

This repository contains documentation and source code to help running a [Raspberry Pi Zero](https://www.raspberrypi.org/products/raspberry-pi-zero-w/) based timelapse camera.

## How does it work?

![The build](https://raw.githubusercontent.com/ingojaeckel/go-raspberry-pi-timelapse/master/docs/go-raspberry-pi-timelapse.jpg "How does it work?")

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

## Build overview

![The build](https://raw.githubusercontent.com/ingojaeckel/go-raspberry-pi-timelapse/master/docs/build.JPG "Build overview")

### Build steps

#### Step 1: Initial Build

1. Optional: If you want to use the battery backed clocked start by soldering the 2x20-pin strip on Pi Zero W board ([instructions](https://learn.adafruit.com/adding-a-real-time-clock-to-raspberry-pi/wiring-the-rtc)). After the soldering, attach the RTC to the 2x20-pin strip. Don't forget to insert the CR1220 battery into the RTC. 
2. Insert the micro SD card into the Pi Zero W.
3. Connect the Pi Zero W and the Pi Camera Board v2 via the Pi Zero Camera cable.
4. Insert the Pi Zero W into the C4Labs Zerbra Zero Case.

#### Step 2: Install and configure the software

1. Connect the Pi Zero W to a keyboard and screen using the USB & mini HDMI adapter.
2. Turn on the Pi Zero W plugging in a micro USB cable into the PWR IN connector (bottom right corner).
3. Install Raspbian Jessie Lite as described [here](https://www.raspberrypi.org/downloads/raspbian/).
4. To connect your Pi Zero W to your Wifi during development, enter the Wifi credentials in the `etc/os/interfaces` configuration file.
5. Run the `etc/os/1_setup_wifi_dev_mode.sh` script. Confirm that your Pi Zero W is able to connect to your Wifi.
6. Wrap up the installation by running the following scrips:
```
    ./etc/os/2_install_timelapse_dependencies.sh
    ./etc/os/3_install_timelapse_service.sh
```
7. Reboot the Pi Zero W. After the reboot it should start taking pictures every 30 minutes starting at 0:15h. E.g. if you turn it on at 2:04pm, the the system will take pictures at 2:15pm, 2:45pm, 3:15pm, etc.
8. Consider editing the timing inside the configuration file `/lib/systemd/system/timelapse.service`. Reboot to apply the changes. To take photos every 5 minutes starting at 0:00h, change the initial configuration to:
```
    [Unit]
    Description = Start Timelapse 
    After = network.target
    [Service]
    ExecStart = /home/pi/go-raspberry-pi-timelapse 300 0
    [Install]
    WantedBy = multi-user.target
```
9. Determine the IP address of the Pi Zero W and download the pictures it has taken via `http://<local IP address of Pi Zero W>:8080/`. The IP address will be local to your network. Depending on the configuration of your network it might look similar to: `192.168.1.110` resulting in the URL: `http://192.168.1.110:8080`.

#### Step 3: Deployment

1. The Pi Zero W will act as a Wifi access point when it is deployed. This allows you to check its health (temperature, disk usage, CPU usage, etc) and download the pictures via Wifi. To configure the Wifi for the deployment adjust the credentials in `etc/os/hostapd.conf`.
2. Install software required to run the Pi Zero W as a Wifi access point by running `./etc/os/4_setup_wifi_deploy_mode.sh`.
3. Install software required to allow other devices to connect to access point by running `./etc/os/5_setup_dhcpd.sh`.
4. Reboot the Pi Zero W. After the reboot it should be available as an access point. Look for the Wifi SSID you configured in `hostapd.conf` (default name: `timelapse-raspberry-pi`, default passphrase: `InsertTheRealPassword`). Connect to the Pi's Wifi.
5. On a device that is connected to access point open `http://192.168.0.1:8080`. Just like during development mode this will allow you to download the timelapse pictures and check the Pi's status.
