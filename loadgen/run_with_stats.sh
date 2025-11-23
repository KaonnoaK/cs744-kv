#!/bin/bash

THREADS=$1
WORKLOAD=$2
OUTDIR="results_${WORKLOAD}_${THREADS}"
mkdir -p "$OUTDIR"

echo "[run] THREADS=$THREADS WORKLOAD=$WORKLOAD OUTDIR=$OUTDIR"

# Start stats collection
./collect_stats.sh "$OUTDIR" &

# Run load generator pinned to CPU2-4 for 300 seconds
taskset -c 2-4 ./loadgen \
   -threads $THREADS \
   -duration 300 \
   -workload $WORKLOAD \
   -keyspace 100000 \
   -popular 100 \
   -out "$OUTDIR/summary.csv"

# After experiment, stop the collectors
./stop_stats.sh "$OUTDIR"

echo "[done] results in $OUTDIR"

