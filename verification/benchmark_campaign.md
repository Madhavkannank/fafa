# Benchmark Campaign Report  JSBI Go Port

Campaign: 8 Independent Benchmark Suite Executions

| Benchmark | Runs | Mean (ns/op) | Median (ns/op) | Min (ns/op) | Max (ns/op) | StdDev | 95% CI | Mean B/op | Mean Allocs/op |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **BenchmarkAdd** | 8 | 55.31 | 54.03 | 52.20 | 64.91 | 4.24 | 2.94 | 48.0 B | 2.0 |
| **BenchmarkSubtract** | 8 | 48.89 | 47.97 | 46.59 | 52.70 | 2.38 | 1.65 | 48.0 B | 2.0 |
| **BenchmarkBitwiseNot** | 8 | 38.35 | 37.92 | 37.48 | 41.57 | 1.35 | 0.93 | 48.0 B | 2.0 |
| **BenchmarkBitwiseAnd** | 8 | 77.38 | 76.38 | 73.93 | 84.93 | 3.80 | 2.63 | 96.0 B | 4.0 |
| **BenchmarkBitwiseOr** | 8 | 85.71 | 84.49 | 83.31 | 93.91 | 3.53 | 2.44 | 96.0 B | 4.0 |
| **BenchmarkBitwiseXor** | 8 | 117.04 | 114.75 | 112.90 | 132.80 | 6.61 | 4.58 | 112.0 B | 5.0 |
| **BenchmarkComparePure** | 8 | 2.74 | 2.72 | 2.67 | 2.92 | 0.08 | 0.05 | 0.0 B | 0.0 |
| **BenchmarkEqualPure** | 8 | 5.59 | 5.55 | 5.25 | 6.14 | 0.28 | 0.20 | 0.0 B | 0.0 |
| **BenchmarkDivide** | 8 | 309.06 | 304.25 | 302.80 | 334.40 | 10.90 | 7.55 | 192.0 B | 8.0 |
| **BenchmarkRemainder** | 8 | 290.93 | 285.90 | 282.00 | 310.40 | 11.63 | 8.06 | 144.0 B | 6.0 |
| **BenchmarkDivRem** | 8 | 333.80 | 332.70 | 320.80 | 357.60 | 10.80 | 7.48 | 192.0 B | 8.0 |
| **BenchmarkMultiply** | 8 | 107.73 | 96.62 | 93.75 | 177.20 | 28.42 | 19.70 | 64.0 B | 2.0 |
| **BenchmarkToStringHex1Limb** | 8 | 28.50 | 27.02 | 26.36 | 36.78 | 3.58 | 2.48 | 16.0 B | 2.0 |
| **BenchmarkToStringDec1Limb** | 8 | 22.67 | 22.59 | 22.32 | 23.18 | 0.32 | 0.22 | 16.0 B | 1.0 |
| **BenchmarkToStringDec3Limbs** | 8 | 1771.62 | 1744.00 | 1723.00 | 1962.00 | 79.23 | 54.90 | 1192.0 B | 71.0 |
| **BenchmarkAsIntN** | 8 | 37.07 | 36.98 | 36.41 | 38.53 | 0.66 | 0.46 | 40.0 B | 2.0 |
| **BenchmarkAsUintN** | 8 | 34.70 | 34.72 | 33.51 | 36.17 | 0.81 | 0.56 | 40.0 B | 2.0 |

---
### Raw Artifacts & Traceability
- CSV Data: [`verification/raw/benchmark_campaign.csv`](verification/raw/benchmark_campaign.csv)
- JSON Data: [`verification/raw/benchmark_campaign.json`](verification/raw/benchmark_campaign.json)
- Execution Command: `go test -run '^$' -bench=. -benchmem ./tests/port/...` (8 runs)
- Date Generated: 2026-08-02
- Git Commit SHA: `5382367c7e3199858d36bb620977e1f90605bcb9`
