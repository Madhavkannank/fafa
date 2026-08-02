# Operational Throughput Report — JSBI Go Port

- **Kickoff Source SHA**: `5382367c7e3199858d36bb620977e1f90605bcb9`
- **Measurement Tool**: `bench/throughput/throughput.go`
- **Status**: Audit COMPLETE — Throughput (`ops/sec`) Measured & Documented

---

## 1. Environment & Setup

- **Hardware**: Intel(R) Core(TM) 5 210H (12 logical threads, 2.20 GHz base clock)
- **RAM**: 16 GB System Memory
- **Operating System**: Microsoft Windows 11 Home (x64)
- **Go Version**: Go 1.22.5 (`go1.22.5.windows-amd64.zip`)

---

## 2. Methodology & Instrumentation

1. **Iteration Loops**: Operational throughput is measured by executing single-threaded iteration loops of 500,000 to 10,000,000 operations per target.
2. **Warm-Up Phase**: 1,000 iterations are executed prior to timing to warm instruction caches and CPU frequency scaling governor state.
3. **High-Precision Timing**: Duration is measured using Go's `time.Now()` monotonic clock interface (`time.Since(start)`).
4. **Throughput Formula**: $\text{Throughput (ops/sec)} = \frac{\text{Total Iterations}}{\text{Elapsed Seconds}}$.

---

## 3. Measured Throughput Results

All figures trace directly to actual `go run bench/throughput/throughput.go` output executed on 2026-08-02:

| Operation | Iterations | Benchmark Duration | Measured Throughput (ops/sec) | MegaOps / sec (Mops/sec) |
| :--- | :--- | :--- | :--- | :--- |
| **Compare** | 10,000,000 | 81.85 ms | **122,167,694.71 ops/sec** | **122.17 Mops/sec** |
| **Add** | 5,000,000 | 228.01 ms | **21,928,401.14 ops/sec** | **21.93 Mops/sec** |
| **Multiply** | 5,000,000 | 233.38 ms | **21,424,075.43 ops/sec** | **21.42 Mops/sec** |
| **Divide** | 1,000,000 | 315.92 ms | **3,165,356.31 ops/sec** | **3.17 Mops/sec** |
| **ToString (Base 10, 3 Limbs)** | 500,000 | 906.09 ms | **551,824.12 ops/sec** | **0.55 Mops/sec** |

---

## 4. Observations

1. **Zero-Allocation Throughput**: Comparison operations (`Compare`) achieve **122.17 Million ops/sec** on a single CPU core.
2. **Arithmetic Throughput**: Addition (`Add`) and Multiplication (`Multiply`) achieve over **21.4 Million ops/sec** per core.
3. **Multi-Limb Radix Conversion**: Divide-and-conquer base-10 string formatting for multi-limb numbers achieves over **551,000 ops/sec** per core.
