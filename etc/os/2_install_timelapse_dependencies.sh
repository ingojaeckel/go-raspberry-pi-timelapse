#!/bin/sh
# Echo this is deprecated. Remove this in favor of setup.sh.

TARGET=/home/pi/go-raspberry-pi-timelapse

aptitude update
aptitude safe-upgrade -y
apt-get install -y htop screen git python-smbus i2c-tools
wget https://github.com/ingojaeckel/go-raspberry-pi-timelapse/raw/master/bin/arm/go-raspberry-pi-timelapse -o ${TARGET}
chmod 755 ${TARGET}

echo "Whenever you are ready run: ${TARGET}"
