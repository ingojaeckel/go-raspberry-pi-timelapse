#!/bin/sh
# This is taken from https://www.raspberrypi.org/documentation/configuration/wireless/access-point.md

# Commented out by default. Enable if necessary.
# sudo apt-get update
sudo apt-get install -y dnsmasq hostapd
sudo systemctl stop dnsmasq
sudo systemctl stop hostapd

echo "sudo nano /etc/dhcpcd.conf"
echo
echo "Add \"denyinterfaces wlan0\" to the end of the file (but above any other added interface lines) and save the file."
