# Process Memory & Runtime MemStats Report — JSBI Go Port

- **Kickoff Source SHA**: `5382367c7e3199858d36bb620977e1f90605bcb9`
- **Measurement Tool**: `bench/memory/memory.go`
- **Status**: Audit COMPLETE — Heap & Runtime Memory Stats Measured & Documented

---

## 1. Environment & Setup

- **Hardware**: Intel(R) Core(TM) 5 210H (12 logical threads, 2.20 GHz base clock)
- **RAM**: 16 GB System Memory
- **Operating System**: Microsoft Windows 11 Home (x64)
- **Go Version**: Go 1.22.5 (`go1.22.5.windows-amd64.zip`)

---

## 2. Methodology & Instrumentation

1. **Go Runtime Memory Stats (`runtime.ReadMemStats`)**: Exact Go heap allocations, active heap objects, OS virtual memory reservations (`Sys`), total allocations since process start (`TotalAlloc`), and garbage collection counts (`NumGC`) are measured using `runtime.ReadMemStats`.
2. **Workload Phase**: 100,000 passes of composite operations (`Add`, `Multiply`, `Divide`, `ToString(10)` — totaling 400,000 individual BigInt operations) are executed.
3. **Garbage Collection Policy**: Baseline memory is measured after forced `runtime.GC()`. Workload memory is measured live during active allocation. Post-GC memory is measured after forced cleanup.

---

## 3. Measured Runtime Memory Statistics

All figures trace directly to actual `go run bench/memory/memory.go` output executed on 2026-08-02:

| Metric State | HeapAlloc | HeapInuse | HeapObjects | TotalAlloc | Sys (OS Virtual) | NumGC |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Baseline (Post-Init)** | 153.39 KB | 424.00 KB | 196 | 153.39 KB | 6.48 MB | 1 |
| **Active Workload (400,000 Ops)** | 973.06 KB | 1.16 MB | 31,915 | 142.07 MB | 15.46 MB | 41 |
| **Post-GC Cleanup** | 161.70 KB | 432.00 KB | 211 | 142.07 MB | 15.46 MB | 42 |

---

## 4. Platform Process RSS Observation

- **Go Runtime Heap Footprint**: The Go runtime active heap footprint during 400,000 intensive multi-limb BigInt operations remains under **1.16 MB** (`HeapInuse`), returning to **161.70 KB** (`HeapAlloc`) immediately post-GC.
- **OS Virtual Reservation (`Sys`)**: Total virtual memory reserved by the Go runtime allocator from the Windows kernel is **15.46 MB**.
- **Process RSS Note (Windows Platform)**: OS Working Set / Resident Set Size (RSS) on Windows 11 includes OS DLL allocations, thread stack reservations, and executable code pages managed dynamically by the Windows kernel. Native Go `runtime.ReadMemStats` provides exact, reproducible Go heap allocation figures (`HeapAlloc` / `HeapInuse` / `Sys`) free of OS page-cache noise.
