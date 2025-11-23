// ~/cs744-kv/loadgen/main.go
// Closed-loop multi-threaded load generator for CS744 KV server (Phase-2)
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Workload string

const (
	WorkPutAll    Workload = "putall"
	WorkGetAll    Workload = "getall"
	WorkGetPopular Workload = "getpopular"
	WorkGetPut    Workload = "getput"
)

type sample struct {
	latNs int64
	ok    bool
}

func main() {
	// Flags
	var server string
	var threads int
	var durationSec int
	var workload string
	var keyspace int
	var popular int
	var putPct int
	var out string

	flag.StringVar(&server, "url", "http://localhost:8080", "Base URL of KV server")
	flag.IntVar(&threads, "threads", 10, "Number of concurrent client threads")
	flag.IntVar(&durationSec, "duration", 300, "Duration of the test in seconds (min 300 recommended)")
	flag.StringVar(&workload, "workload", "getpopular", "Workload: putall|getall|getpopular|getput")
	flag.IntVar(&keyspace, "keyspace", 100000, "Number of unique keys used for getall/putall/getput")
	flag.IntVar(&popular, "popular", 100, "Number of hot keys for getpopular")
	flag.IntVar(&putPct, "putpct", 10, "Percentage of puts in getput (0-100)")
	flag.StringVar(&out, "out", "results.csv", "CSV output filename")
	flag.Parse()

	w := Workload(workload)
	if w != WorkPutAll && w != WorkGetAll && w != WorkGetPopular && w != WorkGetPut {
		fmt.Printf("Unknown workload: %s\n", workload)
		return
	}

	fmt.Printf("Starting loadgen: server=%s threads=%d duration=%ds workload=%s keyspace=%d popular=%d putpct=%d\n",
		server, threads, durationSec, workload, keyspace, popular, putPct)

	// Prepare keys
	keys := make([]string, keyspace)
	for i := 0; i < keyspace; i++ {
		keys[i] = fmt.Sprintf("k-%d", i)
	}
	hotKeys := keys
	if popular < keyspace {
		hotKeys = keys[:popular]
	}

	// Pre-seed DB for read workloads (only a small seed to ensure keys exist when the workload does GET)
	if w == WorkGetAll || w == WorkGetPopular || w == WorkGetPut {
		fmt.Println("Seeding DB with some keys (1000 keys) ...")
		seedCount := 1000
		if seedCount > keyspace {
			seedCount = keyspace
		}
		seedWG := sync.WaitGroup{}
		sem := make(chan struct{}, 50)
		for i := 0; i < seedCount; i++ {
			seedWG.Add(1)
			sem <- struct{}{}
			go func(k string) {
				defer seedWG.Done()
				put(server, k, "seed-val")
				<-sem
			}(keys[i])
		}
		seedWG.Wait()
		fmt.Println("Seed complete")
	}

	resultsCh := make(chan sample, 10_000_00) // buffered channel
	var totalRequests int64
	var totalErrors int64

	stopAt := time.Now().Add(time.Duration(durationSec) * time.Second)
	var wg sync.WaitGroup

	client := &http.Client{Timeout: 10 * time.Second}

	// Worker function (closed-loop): send -> wait -> send next
	worker := func(id int) {
		defer wg.Done()
		rnd := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))
		for time.Now().Before(stopAt) {
			start := time.Now()
			var ok bool
			switch w {
			case WorkPutAll:
				k := keys[rnd.Intn(len(keys))]
				ok = putWithClient(client, server, k, randomValue(rnd))
			case WorkGetAll:
				k := keys[rnd.Intn(len(keys))]
				ok = getWithClient(client, server, k)
			case WorkGetPopular:
				if rnd.Intn(100) < 90 && len(hotKeys) > 0 {
					k := hotKeys[rnd.Intn(len(hotKeys))]
					ok = getWithClient(client, server, k)
				} else {
					k := keys[rnd.Intn(len(keys))]
					ok = getWithClient(client, server, k)
				}
			case WorkGetPut:
				if rnd.Intn(100) < putPct {
					k := keys[rnd.Intn(len(keys))]
					ok = putWithClient(client, server, k, randomValue(rnd))
				} else {
					k := keys[rnd.Intn(len(keys))]
					ok = getWithClient(client, server, k)
				}
			}
			lat := time.Since(start)
			atomic.AddInt64(&totalRequests, 1)
			if !ok {
				atomic.AddInt64(&totalErrors, 1)
			}
			resultsCh <- sample{latNs: lat.Nanoseconds(), ok: ok}
		}
	}

	// Launch workers
	wg.Add(threads)
	for i := 0; i < threads; i++ {
		go worker(i)
	}

	// Collector goroutine to avoid blocking workers for long
	var latencies []int64
	var successCount int64
	collectDone := make(chan struct{})
	go func() {
		for s := range resultsCh {
			latencies = append(latencies, s.latNs)
			if s.ok {
				successCount++
			}
		}
		close(collectDone)
	}()

	// Wait for workers to finish
	wg.Wait()
	// close results channel and wait for collector
	close(resultsCh)
	<-collectDone

	totalReq := atomic.LoadInt64(&totalRequests)
	totalErr := atomic.LoadInt64(&totalErrors)
	durationF := float64(durationSec)

	// Compute metrics
	var sum int64
	for _, v := range latencies {
		sum += v
	}
	avgNs := int64(0)
	if totalReq > 0 {
		avgNs = sum / int64(len(latencies))
	}
	avgMs := float64(avgNs) / 1e6
	throughput := float64(totalReq) / durationF

	// percentiles
	p50, p90, p99 := percentileNs(latencies, 50), percentileNs(latencies, 90), percentileNs(latencies, 99)
	// convert to ms
	p50ms := float64(p50) / 1e6
	p90ms := float64(p90) / 1e6
	p99ms := float64(p99) / 1e6

	// Print summary
	fmt.Println("===== SUMMARY =====")
	fmt.Printf("Workload: %s\n", workload)
	fmt.Printf("Threads: %d, Duration(s): %d\n", threads, durationSec)
	fmt.Printf("Total requests: %d\n", totalReq)
	fmt.Printf("Successful requests: %d\n", successCount)
	fmt.Printf("Errors: %d\n", totalErr)
	fmt.Printf("Throughput (req/s): %.3f\n", throughput)
	fmt.Printf("Avg latency (ms): %.6f\n", avgMs)
	fmt.Printf("p50 (ms): %.6f p90 (ms): %.6f p99 (ms): %.6f\n", p50ms, p90ms, p99ms)

	// Write CSV
	f, err := os.Create(out)
	if err != nil {
    		fmt.Println("failed to create csv:", err)
    		return
	}
	defer f.Close()

	csvWriter := csv.NewWriter(f)
	defer csvWriter.Flush()

	headers := []string{
    	"timestamp", "workload", "threads", "duration_s", "total_requests",
    	"successful", "errors", "throughput_rps", "avg_ms", "p50_ms",
    	"p90_ms", "p99_ms",
	}	
	_ = csvWriter.Write(headers)

	row := []string{
    	time.Now().Format(time.RFC3339),
    	workload,
    	fmt.Sprintf("%d", threads),
    	fmt.Sprintf("%d", durationSec),
    	fmt.Sprintf("%d", totalReq),
    	fmt.Sprintf("%d", successCount),
    	fmt.Sprintf("%d", totalErr),
    	fmt.Sprintf("%.6f", throughput),
    	fmt.Sprintf("%.6f", avgMs),
    	fmt.Sprintf("%.6f", p50ms),
    	fmt.Sprintf("%.6f", p90ms),
    	fmt.Sprintf("%.6f", p99ms),
	}
	_ = csvWriter.Write(row)

	fmt.Printf("Results written to %s\n", out)

}

// helper functions

func randomValue(r *rand.Rand) string {
	return fmt.Sprintf("v-%d", r.Int63())
}

func put(base, key, value string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	return putWithClient(client, base, key, value)
}

func putWithClient(client *http.Client, base, key, value string) bool {
	body, _ := json.Marshal(map[string]string{"value": value})
	req, err := http.NewRequest("PUT", base+"/kv/"+key, bytes.NewBuffer(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	return false
}

func getWithClient(client *http.Client, base, key string) bool {
	resp, err := client.Get(base + "/kv/" + key)
	if err != nil {
		return false
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	return false
}

func percentileNs(data []int64, p float64) int64 {
	if len(data) == 0 {
		return 0
	}
	sort.Slice(data, func(i, j int) bool { return data[i] < data[j] })
	r := (p / 100.0) * float64(len(data)-1)
	lo := int(math.Floor(r))
	hi := int(math.Ceil(r))
	if lo == hi {
		return data[lo]
	}
	frac := r - float64(lo)
	return int64(float64(data[lo])*(1-frac) + float64(data[hi])*frac)
}

