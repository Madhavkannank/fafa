package main

import (
	"fmt"
	"os"
	"time"

	jsbi "github.com/Madhavkannank/fafa/src"
)

type ThroughputResult struct {
	Name       string
	Iterations int64
	Duration   time.Duration
	OpsPerSec  float64
}

func main() {
	fmt.Println("=== JSBI Go Port — Operations Per Second Throughput Tool ===")
	fmt.Println("Hardware: Intel(R) Core(TM) 5 210H, Windows 11 x64")
	fmt.Println("Toolchain: Go 1.22.5 windows/amd64")
	fmt.Println("Methodology: Single-threaded iteration loops measuring exact elapsed time for 1,000,000 to 10,000,000 operations.")
	fmt.Println()

	// Operands setup
	a1, _ := jsbi.FromString("1073741823", 10)
	b1, _ := jsbi.FromString("536870911", 10)
	aLarge, _ := jsbi.FromString("1237940039285380274899124224", 10)
	bLarge, _ := jsbi.FromString("9876543210123456789", 10)

	targets := []struct {
		name       string
		iterations int64
		fn         func()
	}{
		{
			name:       "Compare",
			iterations: 10000000,
			fn:         func() { _ = jsbi.Compare(a1, b1) },
		},
		{
			name:       "Add",
			iterations: 5000000,
			fn:         func() { _ = jsbi.Add(a1, b1) },
		},
		{
			name:       "Multiply",
			iterations: 5000000,
			fn:         func() { _ = jsbi.Multiply(a1, b1) },
		},
		{
			name:       "Divide",
			iterations: 1000000,
			fn:         func() { _, _ = jsbi.Divide(aLarge, bLarge) },
		},
		{
			name:       "ToString (Base 10, 3 Limbs)",
			iterations: 500000,
			fn:         func() { _, _ = jsbi.ToString(aLarge, 10) },
		},
	}

	results := make([]ThroughputResult, 0, len(targets))

	for _, t := range targets {
		// Warmup
		for i := 0; i < 1000; i++ {
			t.fn()
		}

		start := time.Now()
		for i := int64(0); i < t.iterations; i++ {
			t.fn()
		}
		elapsed := time.Since(start)

		opsPerSec := float64(t.iterations) / elapsed.Seconds()
		res := ThroughputResult{
			Name:       t.name,
			Iterations: t.iterations,
			Duration:   elapsed,
			OpsPerSec:  opsPerSec,
		}
		results = append(results, res)

		fmt.Printf("[%s] Iterations=%d | Duration=%v | Throughput=%.2f ops/sec (%.2f Mops/sec)\n",
			res.Name, res.Iterations, res.Duration, res.OpsPerSec, res.OpsPerSec/1e6)
	}

	// Print Markdown Table
	fmt.Println("\n### Throughput Summary Table")
	fmt.Println("| Operation | Iterations | Benchmark Duration | Measured Throughput (ops/sec) | MegaOps / sec |")
	fmt.Println("| :--- | :--- | :--- | :--- | :--- |")
	for _, r := range results {
		fmt.Printf("| **%s** | %d | %v | **%.2f ops/sec** | **%.2f Mops/sec** |\n",
			r.Name, r.Iterations, r.Duration, r.OpsPerSec, r.OpsPerSec/1e6)
	}

	_ = os.Stdout.Sync()
}
