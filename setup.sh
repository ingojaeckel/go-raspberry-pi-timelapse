#!/usr/bin/env bash
#
# You might want to download & execute this to start the installation via:
# wget -O - https://raw.githubusercontent.com/ingojaeckel/go-raspberry-pi-timelapse/master/setup.sh | bash
#
# Or use this shorter URL:
#
# wget -O - http://bit.ly/2BRqNFZ | bash
#
sudo apt-get update -y
sudo apt-get upgrade -y
sudo apt-get install -y htop screen git python-smbus i2c-tools
git clone https://github.com/ingojaeckel/go-raspberry-pi-timelapse.git
cd go-raspberry-pi-timelapse/
