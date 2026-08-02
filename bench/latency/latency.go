package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	jsbi "github.com/Madhavkannank/fafa/src"
)

type LatencyStats struct {
	Name     string
	Samples  int
	Min      float64
	Mean     float64
	Median   float64 // p50
	P90      float64
	P95      float64
	P99      float64
	Max      float64
}

func main() {
	fmt.Println("=== JSBI Go Port — Latency Percentile Measurement Tool ===")
	fmt.Println("Hardware: Intel(R) Core(TM) 5 210H, Windows 11 x64")
	fmt.Println("Toolchain: Go 1.22.5 windows/amd64")
	fmt.Println("Methodology: Monotonic clock (time.Now()), 1,000 warm-up batches (discarded), 10,000 measured sample batches (1,000 ops/batch).")
	fmt.Println()

	// Operands setup
	a1, _ := jsbi.FromString("1073741823", 10)
	b1, _ := jsbi.FromString("536870911", 10)
	aLarge, _ := jsbi.FromString("1237940039285380274899124224", 10)
	bLarge, _ := jsbi.FromString("9876543210123456789", 10)

	targets := []struct {
		name      string
		batchSize int
		fn        func()
	}{
		{
			name:      "Compare",
			batchSize: 10000,
			fn:        func() { _ = jsbi.Compare(a1, b1) },
		},
		{
			name:      "Add",
			batchSize: 5000,
			fn:        func() { _ = jsbi.Add(a1, b1) },
		},
		{
			name:      "Multiply",
			batchSize: 5000,
			fn:        func() { _ = jsbi.Multiply(a1, b1) },
		},
		{
			name:      "Divide",
			batchSize: 1000,
			fn:        func() { _, _ = jsbi.Divide(aLarge, bLarge) },
		},
		{
			name:      "ToString (Base 10, 3 Limbs)",
			batchSize: 500,
			fn:        func() { _, _ = jsbi.ToString(aLarge, 10) },
		},
	}

	const warmupBatches = 100
	const sampleBatches = 5000

	statsList := make([]LatencyStats, 0, len(targets))

	for _, t := range targets {
		// Warm-up phase (excluded from stats)
		for i := 0; i < warmupBatches; i++ {
			for b := 0; b < t.batchSize; b++ {
				t.fn()
			}
		}

		// Measurement phase
		durations := make([]float64, sampleBatches)
		var totalDur float64

		for i := 0; i < sampleBatches; i++ {
			start := time.Now()
			for b := 0; b < t.batchSize; b++ {
				t.fn()
			}
			elapsed := float64(time.Since(start).Nanoseconds()) / float64(t.batchSize)
			durations[i] = elapsed
			totalDur += elapsed
		}

		sort.Float64s(durations)

		stats := LatencyStats{
			Name:    t.name,
			Samples: sampleBatches,
			Min:     durations[0],
			Mean:    totalDur / float64(sampleBatches),
			Median:  percentile(durations, 50.0),
			P90:     percentile(durations, 90.0),
			P95:     percentile(durations, 95.0),
			P99:     percentile(durations, 99.0),
			Max:     durations[sampleBatches-1],
		}
		statsList = append(statsList, stats)

		fmt.Printf("[%s] Samples=%d batches (%d ops/batch) | Min=%.2fns Mean=%.2fns Median(p50)=%.2fns P90=%.2fns P95=%.2fns P99=%.2fns Max=%.2fns\n",
			stats.Name, stats.Samples, t.batchSize, stats.Min, stats.Mean, stats.Median, stats.P90, stats.P95, stats.P99, stats.Max)
	}

	// Output Markdown Table to stdout
	fmt.Println("\n### Latency Percentiles Summary Table (ns)")
	fmt.Println("| Operation | Sample Batches | Ops/Batch | Min (ns) | Mean (ns) | Median / p50 (ns) | p90 (ns) | p95 (ns) | p99 (ns) | Max (ns) |")
	fmt.Println("| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |")
	for i, s := range statsList {
		t := targets[i]
		fmt.Printf("| **%s** | %d | %d | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f | %.2f |\n",
			s.Name, s.Samples, t.batchSize, s.Min, s.Mean, s.Median, s.P90, s.P95, s.P99, s.Max)
	}

	_ = os.Stdout.Sync()
}

func percentile(sorted []float64, p float64) float64 {
	index := (p / 100.0) * float64(len(sorted)-1)
	i := int(index)
	frac := index - float64(i)
	if i >= len(sorted)-1 {
		return sorted[len(sorted)-1]
	}
	return sorted[i] + frac*(sorted[i+1]-sorted[i])
}
