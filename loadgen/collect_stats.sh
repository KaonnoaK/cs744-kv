#!/bin/bash

OUTDIR=$1
mkdir -p "$OUTDIR"

echo "[collector] writing stats to $OUTDIR"

# CPU stats
mpstat 1 > "$OUTDIR/cpu.log" &
CPU_PID=$!

# Disk I/O
iostat -x 1 > "$OUTDIR/disk.log" &
DISK_PID=$!

# Memory / swap / system
vmstat 1 > "$OUTDIR/vmstat.log" &
VM_PID=$!

# Network (optional)
sar -n DEV 1 > "$OUTDIR/net.log" &
NET_PID=$!

# Save PIDs
echo $CPU_PID $DISK_PID $VM_PID $NET_PID > "$OUTDIR/pids.txt"

