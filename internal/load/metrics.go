package load

import (
	"fmt"
	"sort"
	"time"
)

// Sample records the outcome of a single scale operation.
type Sample struct {
	Worker    int
	Cycle     int
	Direction string // "up" | "down"
	From      int
	To        int
	Latency   time.Duration
	Err       error
}

// Report prints an aggregated latency table for all collected samples.
func Report(samples []Sample) {
	type bucket struct {
		latencies []time.Duration
		errors    int
	}

	buckets := map[string]*bucket{}
	for _, s := range samples {
		if _, ok := buckets[s.Direction]; !ok {
			buckets[s.Direction] = &bucket{}
		}
		if s.Err != nil {
			buckets[s.Direction].errors++
			continue
		}
		buckets[s.Direction].latencies = append(buckets[s.Direction].latencies, s.Latency)
	}

	fmt.Println()
	fmt.Printf("%-10s  %-6s  %-8s  %-8s  %-8s  %-8s  %-8s  %-8s  %s\n",
		"Direction", "Count", "Min", "Mean", "P50", "P95", "P99", "Max", "Errors")
	fmt.Println("--------------------------------------------------------------------------")

	for _, dir := range []string{"up", "down"} {
		b, ok := buckets[dir]
		if !ok {
			continue
		}
		if len(b.latencies) == 0 {
			fmt.Printf("%-10s  %-6d  %-8s  %-8s  %-8s  %-8s  %-8s  %-8s  %d\n",
				dir, 0, "-", "-", "-", "-", "-", "-", b.errors)
			continue
		}
		sort.Slice(b.latencies, func(i, j int) bool {
			return b.latencies[i] < b.latencies[j]
		})
		fmt.Printf("%-10s  %-6d  %-8s  %-8s  %-8s  %-8s  %-8s  %-8s  %d\n",
			dir,
			len(b.latencies),
			fmtDur(b.latencies[0]),
			fmtDur(meanDur(b.latencies)),
			fmtDur(percentile(b.latencies, 50)),
			fmtDur(percentile(b.latencies, 95)),
			fmtDur(percentile(b.latencies, 99)),
			fmtDur(b.latencies[len(b.latencies)-1]),
			b.errors,
		)
	}
}

func meanDur(d []time.Duration) time.Duration {
	var total time.Duration
	for _, v := range d {
		total += v
	}
	return total / time.Duration(len(d))
}

// percentile returns the p-th percentile (0–100) of a pre-sorted slice.
func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(p) / 100.0 * float64(len(sorted)))
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func fmtDur(d time.Duration) string {
	return d.Round(time.Millisecond).String()
}
