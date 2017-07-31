#!/bin/sh
# This is taken from https://www.raspberrypi.org/documentation/configuration/wireless/access-point.md

sudo service dhcpcd restart
sudo ifdown wlan0
sudo ifup wlan0
sudo cp dnsmasq.conf /etc/dnsmasq.conf
sudo mkdir -p /etc/hostapd
sudo cp hostapd.conf /etc/hostapd/hostapd.conf

echo "sudo nano /etc/default/hostapd"
echo "Find the line with #DAEMON_CONF, and replace it with this:"
echo "DAEMON_CONF=\"/etc/hostapd/hostapd.conf\""
echo
echo "Run the following once all configuration is in place:"
echo
echo "sudo service hostapd start"
echo "sudo service dnsmasq start"
