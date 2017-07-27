#!/bin/sh

RETVAL=0
prog="go-raspberry-pi-timelapse"
EXECUTABLE="/home/pi/${prog}"

start() {
	echo -n "Starting $prog"
	nohup ${EXECUTABLE} &
	RETVAL=$?
	[ "$RETVAL" = 0 ] && touch /var/lock/subsys/$prog
	echo
}

stop() {
	echo -n "Stopping $prog"
	killall -9 $prog
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
		echo "Usage: $0 {start|stop|status}"
		RETVAL=1
esac
exit $RETVAL
