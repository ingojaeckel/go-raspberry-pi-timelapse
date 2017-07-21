#!/bin/sh
TARGET=/tmp/go-raspberry-pi-timelapse

aptitude update
aptitude safe-upgrade -y
apt-get install -y htop screen
wget https://github.com/ingojaeckel/go-raspberry-pi-timelapse/raw/master/bin/arm/go-raspberry-pi-timelapse -o ${TARGET}

echo "Whenever you are ready run: ${TARGET}"
