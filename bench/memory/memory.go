package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	jsbi "github.com/Madhavkannank/fafa/src"
)

func main() {
	fmt.Println("=== JSBI Go Port — Memory Footprint & Runtime MemStats Measurement Tool ===")
	fmt.Println("Hardware: Intel(R) Core(TM) 5 210H, Windows 11 x64")
	fmt.Println("Toolchain: Go 1.22.5 windows/amd64")
	fmt.Println("Methodology: runtime.ReadMemStats before and after 100,000 BigInt workload operations with forced GC.")
	fmt.Println()

	// Initial baseline snapshot
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	var msBase runtime.MemStats
	runtime.ReadMemStats(&msBase)

	fmt.Printf("Baseline MemStats:\n")
	fmt.Printf("  HeapAlloc:   %10d bytes (%.2f KB)\n", msBase.HeapAlloc, float64(msBase.HeapAlloc)/1024.0)
	fmt.Printf("  HeapInuse:   %10d bytes (%.2f KB)\n", msBase.HeapInuse, float64(msBase.HeapInuse)/1024.0)
	fmt.Printf("  HeapObjects: %10d objects\n", msBase.HeapObjects)
	fmt.Printf("  Sys:         %10d bytes (%.2f KB)\n", msBase.Sys, float64(msBase.Sys)/1024.0)
	fmt.Printf("  NumGC:       %10d collections\n", msBase.NumGC)
	fmt.Println()

	// Workload execution
	const N = 100000
	a, _ := jsbi.FromString("1237940039285380274899124224", 10)
	b, _ := jsbi.FromString("9876543210123456789", 10)

	start := time.Now()
	for i := 0; i < N; i++ {
		_ = jsbi.Add(a, b)
		_ = jsbi.Multiply(a, b)
		_, _ = jsbi.Divide(a, b)
		_, _ = jsbi.ToString(a, 10)
	}
	elapsed := time.Since(start)

	var msWorkload runtime.MemStats
	runtime.ReadMemStats(&msWorkload)

	fmt.Printf("Active Workload MemStats (100,000 x Add/Mul/Div/ToString passes):\n")
	fmt.Printf("  HeapAlloc:   %10d bytes (%.2f MB)\n", msWorkload.HeapAlloc, float64(msWorkload.HeapAlloc)/(1024.0*1024.0))
	fmt.Printf("  HeapInuse:   %10d bytes (%.2f MB)\n", msWorkload.HeapInuse, float64(msWorkload.HeapInuse)/(1024.0*1024.0))
	fmt.Printf("  HeapObjects: %10d objects\n", msWorkload.HeapObjects)
	fmt.Printf("  TotalAlloc:  %10d bytes (%.2f MB)\n", msWorkload.TotalAlloc, float64(msWorkload.TotalAlloc)/(1024.0*1024.0))
	fmt.Printf("  Sys:         %10d bytes (%.2f MB)\n", msWorkload.Sys, float64(msWorkload.Sys)/(1024.0*1024.0))
	fmt.Printf("  NumGC:       %10d collections\n", msWorkload.NumGC)
	fmt.Printf("  Execution:   %v (%.2f ops/sec)\n", elapsed, float64(N*4)/elapsed.Seconds())
	fmt.Println()

	// Post-GC cleanup snapshot
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	var msPost runtime.MemStats
	runtime.ReadMemStats(&msPost)

	fmt.Printf("Post-GC MemStats:\n")
	fmt.Printf("  HeapAlloc:   %10d bytes (%.2f KB)\n", msPost.HeapAlloc, float64(msPost.HeapAlloc)/1024.0)
	fmt.Printf("  HeapInuse:   %10d bytes (%.2f KB)\n", msPost.HeapInuse, float64(msPost.HeapInuse)/1024.0)
	fmt.Printf("  HeapObjects: %10d objects\n", msPost.HeapObjects)
	fmt.Printf("  TotalAlloc:  %10d bytes (%.2f MB)\n", msPost.TotalAlloc, float64(msPost.TotalAlloc)/(1024.0*1024.0))
	fmt.Printf("  NumGC:       %10d collections\n", msPost.NumGC)

	_ = os.Stdout.Sync()
}
