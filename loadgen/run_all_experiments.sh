#!/bin/bash

WORKLOADS=("getpopular" "getall" "putall" "getput")
THREADS=(10 50 100 200 400)

echo "[master] Starting full Phase-2 experiment suite..."

for W in "${WORKLOADS[@]}"; do
  for T in "${THREADS[@]}"; do
    echo "====================================="
    echo "[master] Running workload=$W threads=$T"
    echo "====================================="
    ./run_with_stats.sh $T $W
    echo "[master] Sleeping 10 seconds before next run..."
    sleep 10
  done
done

echo "[master] All experiments finished!"

