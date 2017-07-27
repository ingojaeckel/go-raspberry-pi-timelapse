#!/bin/sh
# source function library
. /etc/rc.d/init.d/functions

RETVAL=0
prog="timelapse"
EXECUTABLE="/home/pi/go-raspberry-pi-timelapse"

start() {
	echo -n $"Starting $prog:"
	nohup ${EXECUTABLE} &
	RETVAL=$?
	[ "$RETVAL" = 0 ] && touch /var/lock/subsys/$prog
	echo
}

stop() {
	echo -n $"Stopping $prog:"
	killproc $prog -TERM
	RETVAL=$?
	[ "$RETVAL" = 0 ] && rm -f /var/lock/subsys/$prog
	echo
}

case "$1" in
	start)
		start
		;;
	stop)
		stop
		;;
	status)
		status $prog
		RETVAL=$?
		;;
	*)	(10)
		echo $"Usage: $0 {start|stop|status}"
		RETVAL=1
esac
exit $RETVAL
