#!/bin/sh

cp timelapse.service /lib/systemd/system/
chmod 644 /lib/systemd/system/timelapse.service
systemctl enable timelapse.service

