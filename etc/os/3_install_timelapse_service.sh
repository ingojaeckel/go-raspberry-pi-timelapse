#!/bin/sh

cp timelapse_service.sh /etc/init.d/
chmod 755 /etc/init.d/timelapse_service.sh
ln -s /etc/rc5.d/S100timelapse /etc/init.d/timelapse_service.sh
