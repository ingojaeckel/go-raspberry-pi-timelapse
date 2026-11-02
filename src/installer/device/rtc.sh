#!/usr/bin/env bash
# This is adapted from [Adafruit](https://learn.adafruit.com/adding-a-real-time-clock-to-raspberry-pi/set-rtc-time):

echo "Ensure you have done this first:"
echo
echo "sudo i2cdetect -y 1"
echo "... Should output \"68\""
echo
echo "sudo echo \"dtoverlay=i2c-rtc,pcf8523\" >> /boot/config.txt"
echo "sudo reboot"
echo
read -rsp $'Press any key to continue...\n' -n1 key

sudo i2cdetect -y 1
read -rsp $'This should have printed UU. Press any key to continue...\n' -n1 key
sudo apt-get -y remove fake-hwclock && sudo update-rc.d -f fake-hwclock remove

echo "Comment out these lines in /lib/udev/hwclock-set:"
echo
echo "#if [ -e /run/systemd/system ] ; then"
echo "# exit 0"
echo "#fi"

# Sync time from Pi to RTC, see https://learn.adafruit.com/adding-a-real-time-clock-to-raspberry-pi/set-rtc-time#sync-time-from-pi-to-rtc

echo "To display the current time run:"
echo "sudo hwclock -D -r"
echo
echo "To write the current network time to the RTC run:"
echo sudo hwclock -w
echo
read -rsp $'Press any key to reboot...\n' -n1 key
sudo reboot
