#!/bin/bash

OUTDIR=$1
mkdir -p "$OUTDIR"

# Save PID list inside OUTDIR
PIDFILE="$OUTDIR/pids.txt"

# Start sar collectors
sar -u 1 > "$OUTDIR/cpu.log" &
echo $! >> "$PIDFILE"

sar -d 1 > "$OUTDIR/disk.log" &
echo $! >> "$PIDFILE"

vmstat 1 > "$OUTDIR/vmstat.log" &
echo $! >> "$PIDFILE"

echo "[collector] writing stats to $OUTDIR"

