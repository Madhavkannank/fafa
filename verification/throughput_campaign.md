# Operational Throughput Campaign Report  JSBI Go Port

Campaign: 10 Independent Single-Threaded Throughput Benchmark Runs

| Operation | Runs | Mean (ops/sec) | Mean (Mops/sec) | Median (Mops/sec) | Min (Mops/sec) | Max (Mops/sec) | StdDev (Mops/sec) |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Compare** | 10 | **115,453,917.47** | **115.45 Mops** | 113.78 Mops | 103.83 Mops | 132.90 Mops | 9.59 Mops |
| **Add** | 10 | **18,315,238.65** | **18.32 Mops** | 18.67 Mops | 16.32 Mops | 20.89 Mops | 1.41 Mops |
| **Multiply** | 10 | **16,923,066.69** | **16.92 Mops** | 17.19 Mops | 15.37 Mops | 18.30 Mops | 0.92 Mops |
| **Divide** | 10 | **2,904,770.65** | **2.90 Mops** | 2.98 Mops | 2.44 Mops | 3.10 Mops | 0.20 Mops |
| **ToString (Base 10, 3 Limbs)** | 10 | **535,100.44** | **0.54 Mops** | 0.54 Mops | 0.51 Mops | 0.55 Mops | 0.01 Mops |

---
### Raw Artifacts & Traceability
- CSV Data: [`verification/raw/throughput_campaign.csv`](verification/raw/throughput_campaign.csv)
- Measurement Tool: `bench/throughput/throughput.go`
