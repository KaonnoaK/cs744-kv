package main

import (
	"sync"
	"time"
)

var (
	mu     sync.Mutex
	count  int64
	latSum time.Duration
)

func record(duration time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	count++
	latSum += duration
}

func getMetrics() map[string]interface{} {
	mu.Lock()
	defer mu.Unlock()

	avg := float64(latSum.Milliseconds())
	if count > 0 {
		avg /= float64(count)
	}

	return map[string]interface{}{
		"total_requests": count,
		"avg_latency_ms": avg,
	}
}

