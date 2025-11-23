import os
import pandas as pd
import matplotlib.pyplot as plt

def load_summary(folder):
    path = os.path.join(folder, "summary.csv")
    if not os.path.exists(path):
        return None
    try:
        df = pd.read_csv(path)
        return df
    except:
        return None

def plot_metric(metric, ylabel, outname, results):
    plt.figure(figsize=(10, 6))
    for workload, data in results.items():
        threads = sorted(data.keys())
        values = [data[t][metric] for t in threads]
        plt.plot(threads, values, marker='o', label=workload)

    plt.xlabel("Threads")
    plt.ylabel(ylabel)
    plt.title(outname.replace(".png", ""))
    plt.legend()
    plt.grid(True)
    plt.savefig(outname, dpi=300)
    plt.close()


def main():
    folders = [d for d in os.listdir(".") if d.startswith("results_")]

    workloads = {}
    for folder in folders:
        parts = folder.split("_")
        workload = parts[1]
        threads = int(parts[2])

        df = load_summary(folder)
        if df is None:
            continue

        if workload not in workloads:
            workloads[workload] = {}

        workloads[workload][threads] = df.iloc[0]

    # PLOT THROUGHPUT
    plot_metric("throughput_rps", "Throughput (requests/sec)", "throughput.png", workloads)

    # PLOT LATENCIES
    plot_metric("avg_ms", "Average Latency (ms)", "avg_latency.png", workloads)
    plot_metric("p50_ms", "p50 Latency (ms)", "p50_latency.png", workloads)
    plot_metric("p90_ms", "p90 Latency (ms)", "p90_latency.png", workloads)
    plot_metric("p99_ms", "p99 Latency (ms)", "p99_latency.png", workloads)

    # Also generate per-workload latency graphs
    for wl in workloads.keys():
        wl_data = {wl: workloads[wl]}
        plot_metric("avg_ms", "Average Latency (ms)", f"{wl}_avg_latency.png", wl_data)
        plot_metric("p50_ms", "p50 Latency (ms)", f"{wl}_p50_latency.png", wl_data)
        plot_metric("p90_ms", "p90 Latency (ms)", f"{wl}_p90_latency.png", wl_data)
        plot_metric("p99_ms", "p99 Latency (ms)", f"{wl}_p99_latency.png", wl_data)

    print("✔ All throughput + latency graphs generated!")

if __name__ == "__main__":
    main()

