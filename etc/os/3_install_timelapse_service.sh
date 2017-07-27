#!/bin/sh

cp timelapse_service.sh /etc/init.d/
chmod 755 /etc/init.d/timelapse_service.sh

ln -fs /etc/rc5.d/S100timelapse /etc/init.d/timelapse_service.sh
