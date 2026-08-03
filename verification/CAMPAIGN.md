# Statistical Verification & Audit Campaign Master Index  JSBI Go Port

- **Repository Target**: `github.com/Madhavkannank/fafa`
- **Kickoff Source SHA**: `GoogleChromeLabs/jsbi` (`5382367c7e3199858d36bb620977e1f90605bcb9`)
- **Audit Date**: 2026-08-02
- **Status**: **ALL CAMPAIGNS EXECUTED & AUDITED FROM RAW REPEATED EXECUTIONS**

--- 

## 1. Campaign Portfolio Overview

| Campaign Category | Executions per Run | Total Campaign Runs | Target Report | Raw Data File |
| :--- | :--- | :--- | :--- | :--- |
| **Benchmark Campaign** | 17 Benchmarks | 8 Runs | [`benchmark_campaign.md`](verification/benchmark_campaign.md) | `benchmark_campaign.csv` / `.json` |
| **Coverage Campaign** | Full Package | 8 Runs | [`coverage_campaign.md`](verification/coverage_campaign.md) | `coverage_campaign.csv` |
| **Latency Campaign** | 5 Operations | 10 Runs (5,000 batches/run) | [`latency_campaign.md`](verification/latency_campaign.md) | `latency_campaign.csv` |
| **Throughput Campaign** | 5 Operations | 10 Runs | [`throughput_campaign.md`](verification/throughput_campaign.md) | `throughput_campaign.csv` |
| **Memory Campaign** | MemStats | 10 Runs (400,000 ops/run) | [`memory_campaign.md`](verification/memory_campaign.md) | `memory_campaign.csv` |
| **Regression Campaign** | 44 Test Suites | 10 Runs | [`test_campaign.md`](verification/test_campaign.md) | `tmp/test_run.log` |
| **Differential Fuzz Campaign**| 9 Clusters | 9,696,250 Cases | [`fuzz_campaign.md`](verification/fuzz_campaign.md) | `fuzz/log.txt` |
| **Static Analysis Campaign** | 5 Tools | 10 Runs | [`static_campaign.md`](verification/static_campaign.md) | `verification/raw/static/` |

---
## 2. Total Artifact Inventory

- **CSV Raw Datasets**: 5 files (`benchmark_campaign.csv`, `coverage_campaign.csv`, `latency_campaign.csv`, `throughput_campaign.csv`, `memory_campaign.csv`)
- **JSON Raw Datasets**: 1 file (`benchmark_campaign.json`)
- **Markdown Audit Reports**: 8 campaign reports in `verification/`
- **Binary Profiles**: 2 pprof files (`cpu.pprof`, `mem.pprof`)

---
### Environment Traceability
- OS: Windows 11 Home x64 (Build 22631)
- CPU: Intel(R) Core(TM) 5 210H (12 logical threads @ 2.20 GHz)
- RAM: 16 GB DDR5 Memory
- Go Version: Go 1.22.5 (`go1.22.5.windows-amd64.zip`)
- Date Generated: 2026-08-02
