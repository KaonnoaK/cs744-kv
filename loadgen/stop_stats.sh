#!/bin/bash

OUTDIR=$1
PIDS=$(cat "$OUTDIR/pids.txt")

echo "[stop] stopping collectors: $PIDS"

kill $PIDS

