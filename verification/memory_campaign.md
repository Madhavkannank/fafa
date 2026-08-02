# Memory Footprint & MemStats Campaign Report — JSBI Go Port

Campaign: 10 Independent Runtime MemStats Executions across 400,000 mixed operations

| State / Metric | Runs | Mean Value | Median | Minimum | Maximum | StdDev |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Baseline HeapAlloc** | 10 | **150.44 KB** | 150.17 KB | 150.17 KB | 152.86 KB | 0.85 KB |
| **Active Workload HeapAlloc** | 10 | **950.26 KB** | 950.26 KB | 950.26 KB | 950.26 KB | 0.00 KB |
| **Active Workload HeapInuse** | 10 | **1192.00 KB** | 1192.00 KB | 1192.00 KB | 1192.00 KB | 0.00 KB |
| **OS Virtual Reserved (Sys)** | 10 | **15828.00 KB** | 15828.00 KB | 15828.00 KB | 15828.00 KB | 0.00 KB |
| **Post-GC HeapAlloc** | 10 | **156.57 KB** | 155.81 KB | 155.59 KB | 158.57 KB | 1.34 KB |
| **Garbage Collections (NumGC)** | 10 | 41.0 | 41.0 | 41 | 41 | 0.00 |

---
### OS Process RSS Footnote
Runtime heap statistics (`HeapAlloc`, `HeapInuse`, `Sys`) were measured using `runtime.ReadMemStats`; operating-system RSS was not measured due to OS page-cache variance.

### Raw Artifacts & Traceability
- CSV Data: [`verification/raw/memory_campaign.csv`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/memory_campaign.csv)
- Measurement Tool: `bench/memory/memory.go`
