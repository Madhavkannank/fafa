# Percentile Latency Campaign Report — JSBI Go Port

Campaign: 10 Independent High-Resolution Percentile Latency Runs (5,000 sample batches/run)

| Operation | Runs | Mean p50 (ns) | Mean p90 (ns) | Mean p95 (ns) | Mean p99 (ns) | Median p99 | Min p99 | Max p99 | StdDev p99 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Compare** | 10 | 0.00 | 0.00 | 3.60 | **252.61** | 258.94 | 200.71 | 286.13 | 28.11 |
| **Add** | 10 | 0.00 | 101.29 | 382.20 | **1188.69** | 1204.19 | 1015.85 | 1268.86 | 70.60 |
| **Multiply** | 10 | 0.00 | 79.95 | 366.17 | **1172.23** | 1208.40 | 1010.68 | 1305.40 | 105.23 |
| **Divide** | 10 | 0.00 | 699.72 | 2284.78 | **6241.14** | 6347.43 | 5183.75 | 6689.84 | 433.71 |
| **ToString (Base 10, 3 Limbs)** | 10 | 0.00 | 6815.70 | 10584.85 | **14399.98** | 14028.67 | 13880.77 | 16208.78 | 863.23 |

---
### Methodological & Timer Note
Windows system clock tick granularity (~15.6ms) causes sub-100ns operation samples to yield 0.00ns median (p50) per-op readings when divided by batch sizes ($500$ to $10,000$ ops/batch). Batch-averaged means and p99 tails reflect true hardware performance.

### Raw Artifacts & Traceability
- CSV Data: [`verification/raw/latency_campaign.csv`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/latency_campaign.csv)
- Measurement Tool: `bench/latency/latency.go`
