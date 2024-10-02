#!/usr/bin/env bash
#
# You might want to download & execute this to start the installation via:
# wget -O - https://raw.githubusercontent.com/ingojaeckel/go-raspberry-pi-timelapse/master/setup.sh | bash
#
set -e -x

# Try to free up some space now before installing/updating may need it
sudo apt-get autoclean
sudo apt-get autoremove
sudo apt-get clean

sudo apt-get update -y
sudo apt-get upgrade -y
sudo apt-get install -y htop screen git python3-smbus i2c-tools
git clone https://github.com/ingojaeckel/go-raspberry-pi-timelapse.git
cd go-raspberry-pi-timelapse/
