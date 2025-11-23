#!/bin/bash

THREADS=$1
WORKLOAD=$2
OUTDIR="results_${WORKLOAD}_${THREADS}"

mkdir -p "$OUTDIR"

echo "[run] THREADS=$THREADS WORKLOAD=$WORKLOAD OUTDIR=$OUTDIR"

# Start collectors
./collect_stats.sh "$OUTDIR" &

COLLECT_PID=$!

# Run load generator
taskset -c 2-4 ./loadgen \
   -threads "$THREADS" \
   -duration 300 \
   -workload "$WORKLOAD" \
   -keyspace 100000 \
   -popular 100

# Save the results.csv that loadgen overwrote
cp results.csv "$OUTDIR/results.csv"

# Stop collectors
./stop_stats.sh "$OUTDIR"

echo "[done] results in $OUTDIR"

