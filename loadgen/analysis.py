#!/usr/bin/env python3
import os
import re
import csv
import json
import argparse
import statistics

##############################################
# Helpers to parse CPU, DISK, VMSTAT, SUMMARY
##############################################

def parse_cpu_log(path):
    """
    Reads cpu.log and extracts %usr, %sys, %idle per line.
    """
    usr, sys, idle = [], [], []

    cpu_line = re.compile(r"all\s+([\d\.]+)\s+[\d\.]+\s+([\d\.]+)\s+[\d\.]+\s+[\d\.]+\s+[\d\.]+\s+[\d\.]+\s+[\d\.]+\s+[\d\.]+\s+([\d\.]+)")

    with open(path) as f:
        for line in f:
            m = cpu_line.search(line)
            if m:
                usr.append(float(m.group(1)))
                sys.append(float(m.group(2)))
                idle.append(float(m.group(3)))

    return {
        "cpu_usr_avg": statistics.mean(usr) if usr else 0,
        "cpu_sys_avg": statistics.mean(sys) if sys else 0,
        "cpu_idle_avg": statistics.mean(idle) if idle else 100,
        "cpu_busy_avg": 100 - statistics.mean(idle) if idle else 0
    }


def parse_disk_log(path):
    """
    Reads disk.log and extracts disk util % (last column).
    """
    util_vals = []

    util_re = re.compile(r"\s+\d+\.\d+$")  # last float in line

    with open(path) as f:
        for line in f:
            parts = line.split()
            if len(parts) > 0 and parts[-1].replace('.', '', 1).isdigit():
                try:
                    v = float(parts[-1])
                    util_vals.append(v)
                except:
                    pass

    return {
        "disk_util_avg": statistics.mean(util_vals) if util_vals else 0
    }


def parse_vmstat_log(path):
    """
    Reads vmstat log and returns avg run queue, user CPU, sys CPU, idle CPU.
    """
    r_vals, usr_vals, sys_vals, idle_vals = [], [], [], []

    with open(path) as f:
        for line in f:
            parts = line.split()
            if len(parts) == 17 and parts[0].isdigit():  # vmstat data lines
                r_vals.append(int(parts[0]))
                usr_vals.append(float(parts[12]))
                sys_vals.append(float(parts[13]))
                idle_vals.append(float(parts[14]))

    return {
        "vm_r_avg": statistics.mean(r_vals) if r_vals else 0,
        "vm_usr_avg": statistics.mean(usr_vals) if usr_vals else 0,
        "vm_sys_avg": statistics.mean(sys_vals) if sys_vals else 0,
        "vm_idle_avg": statistics.mean(idle_vals) if idle_vals else 0
    }


def parse_summary_csv(path):
    """
    Reads the summary.csv (1-line CSV).
    """
    with open(path) as f:
        reader = csv.DictReader(f)
        row = next(reader)

    # convert to proper types
    for k, v in row.items():
        try:
            row[k] = float(v)
        except:
            row[k] = v

    return row


#################################
# Bottleneck Classifier
#################################

def classify_bottleneck(cpu_busy, disk_util, vm_r):
    """
    Simple rule-based bottleneck detection like a student would write.
    """

    if cpu_busy > 85:
        return "CPU-bound"

    if disk_util > 80:
        return "Disk I/O bound"

    if vm_r > 2 * os.cpu_count():
        return "Run-queue saturated (CPU contention)"

    return "No clear bottleneck (likely cache or network stall)"


#################################
# Main analysis
#################################

def analyze_directory(d):
    print(f"\n=== Analyzing {d} ===")

    cpu_path = os.path.join(d, "cpu.log")
    disk_path = os.path.join(d, "disk.log")
    vm_path = os.path.join(d, "vmstat.log")
    summary_path = os.path.join(d, "summary.csv")

    result = {}

    if os.path.exists(cpu_path):
        result.update(parse_cpu_log(cpu_path))

    if os.path.exists(disk_path):
        result.update(parse_disk_log(disk_path))

    if os.path.exists(vm_path):
        result.update(parse_vmstat_log(vm_path))

    if os.path.exists(summary_path):
        result.update(parse_summary_csv(summary_path))

    # Add classification
    result["bottleneck"] = classify_bottleneck(
        result.get("cpu_busy_avg", 0),
        result.get("disk_util_avg", 0),
        result.get("vm_r_avg", 0)
    )

    # Print pretty JSON
    print(json.dumps(result, indent=4))

    # Save result.json
    with open(os.path.join(d, "analysis.json"), "w") as f:
        json.dump(result, f, indent=4)


#################################
# CLI
#################################

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("directory", nargs="?", help="Directory to analyze")
    parser.add_argument("--all", action="store_true", help="Analyze all workload dirs automatically")
    args = parser.parse_args()

    if args.all:
        for d in os.listdir("."):
            if d.startswith("results_") and os.path.isdir(d):
                analyze_directory(d)
    else:
        if not args.directory:
            print("Usage: python3 analysis.py results_getall_400")
            exit(1)
        analyze_directory(args.directory)

