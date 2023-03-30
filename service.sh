#!/bin/sh

exitOnErr() {
  if [ $? -ne 0 ]; then
    echo $1
    exit 1
  fi
}

start() {
  nohup ./loghub >>out.log 2>&1 &
  exitOnErr "failed to start service"
}


stop() {
  if pid=`pgrep loghub`; then
    echo "Stopping ($pid)..."
    kill $pid
    while `kill -0 $pid 2>/dev/null`; do
      echo 'Waiting for terminated'
      sleep 1
    done
  fi
}

restart() {
  stop
  start
}

update() {
  stop
  rm -f loghub && rz && chmod o+x loghub
  exitOnErr "failed to update binary"
  start
}

case "$1" in
  'start')
    start
    ;;
  'stop')
    stop
    ;;
  'restart')
    restart
    ;;
  'update')
    update
    ;;
  *)
    echo "Usage $0 {start|stop|restart|update}"
    exit 1
esac

