import csv, glob, os
import matplotlib.pyplot as plt

def read_results(pattern):
    files = glob.glob(pattern)
    results = []
    for f in files:
        with open(f) as fd:
            r = list(csv.DictReader(fd))[0]
            results.append({
                "threads": int(r["threads"]),
                "throughput": float(r["throughput_rps"]),
                "avg_ms": float(r["avg_ms"]),
                "p99_ms": float(r["p99_ms"]),
                "file": f
            })
    return sorted(results, key=lambda x: x["threads"])

def plot(workload):
    results = read_results(f"results_{workload}_*/*summary.csv")

    threads = [r["threads"] for r in results]
    tput = [r["throughput"] for r in results]
    avg = [r["avg_ms"] for r in results]

    plt.figure()
    plt.plot(threads, tput, marker='o')
    plt.title(f"Throughput vs Threads ({workload})")
    plt.xlabel("Threads")
    plt.ylabel("Throughput (req/s)")
    plt.savefig(f"{workload}_throughput.png")

    plt.figure()
    plt.plot(threads, avg, marker='o')
    plt.title(f"Avg Latency vs Threads ({workload})")
    plt.xlabel("Threads")
    plt.ylabel("Avg latency (ms)")
    plt.savefig(f"{workload}_latency.png")

for w in ["getpopular", "getall", "putall", "getput"]:
    plot(w)

