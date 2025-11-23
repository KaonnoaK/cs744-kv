import os
import csv
import re
import matplotlib.pyplot as plt
from collections import defaultdict

# -------------------------------------
#  Directory scanning
# -------------------------------------

def find_result_folders(base="."):
    dirs = []
    for d in os.listdir(base):
        if os.path.isdir(d) and re.match(r"results_.+_\d+", d):
            dirs.append(d)
    return dirs

# -------------------------------------
#  Parse results.csv inside each folder
# -------------------------------------

def parse_results(path):
    resultfile = os.path.join(path, "results.csv")
    if not os.path.exists(resultfile):
        return None

    with open(resultfile) as f:
        r = list(csv.DictReader(f))

    if len(r) == 0:
        return None

    row = r[0]  # Only one row

    return {
        "workload": row["workload"],
        "threads": int(float(row["threads"])),
        "throughput": float(row["throughput_rps"]),
        "p50": float(row["p50_ms"]),
        "p90": float(row["p90_ms"]),
        "p99": float(row["p99_ms"]),
    }

# -------------------------------------
#  Plotting helpers
# -------------------------------------

def plot_metric(workload, data, metric, ylabel):
    plt.figure(figsize=(8,5))
    data = sorted(data, key=lambda x: x["threads"])

    x = [d["threads"] for d in data]
    y = [d[metric] for d in data]

    plt.plot(x, y, marker="o")
    plt.title(f"{workload.upper()} — {ylabel}")
    plt.xlabel("Threads")
    plt.ylabel(ylabel)
    plt.grid(True)

    os.makedirs("plots", exist_ok=True)
    outfile = f"plots/{workload}_{metric}.png"
    plt.savefig(outfile)
    print(f"[+] Saved {outfile}")
    plt.close()

# -------------------------------------
#  Main: Load → Aggregate → Plot
# -------------------------------------

def main():
    folders = find_result_folders(".")
    print("[*] Found result folders:", folders)

    workloads = defaultdict(list)

    for folder in folders:
        res = parse_results(folder)
        if res:
            workloads[res["workload"]].append(res)

    print("[*] Workloads found:", workloads.keys())

    # generate all plots per workload
    for w, data in workloads.items():
        plot_metric(w, data, "throughput", "Throughput (req/s)")
        plot_metric(w, data, "p50", "p50 Latency (ms)")
        plot_metric(w, data, "p90", "p90 Latency (ms)")
        plot_metric(w, data, "p99", "p99 Latency (ms)")

    print("\n✔ All graphs generated in ./plots/\n")

if __name__ == "__main__":
    main()

