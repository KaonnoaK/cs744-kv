#!/bin/bash

OUTDIR=$1
PIDFILE="$OUTDIR/pids.txt"

if [ ! -f "$PIDFILE" ]; then
    echo "[stop] ERROR: PID file missing: $PIDFILE"
    exit 1
fi

echo "[stop] stopping collectors:"
for pid in $(cat "$PIDFILE"); do
    kill $pid 2>/dev/null
done

rm "$PIDFILE"

