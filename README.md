# GoogleChromeLabs/jsbi — Pure Go Port

> [!NOTE]
> **Cross-Repository Bug Audit & Evidence Index**:
> Verified bug findings and edge-case behavior analyses are documented in [`SDFGH/BUGS.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/SDFGH/BUGS.md) and indexed in [`Kavinraj696/verified-bugs`](https://github.com/Kavinraj696/verified-bugs).

- **Upstream Reference**: [`GoogleChromeLabs/jsbi`](https://github.com/GoogleChromeLabs/jsbi) (`5382367c7e3199858d36bb620977e1f90605bcb9`)
- **Package Path**: `github.com/Madhavkannank/fafa/src`
- **Go Version**: Go 1.20+ static library (`package jsbi`)
- **Status**: Pure Go Port Complete — All 9 Functional Clusters & Audit Campaigns Complete

---

## Bonus Criteria Verification Evidence

| Bonus Category | Target Criteria | Measured Evidence Status | Verification Proof Artifact Location |
| :--- | :--- | :--- | :--- |
| **Differential Fuzz Survivor** | 60s+ execution without divergences | 9,696,250 cases (0 mismatches) | [`fuzz/log.txt`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/fuzz/log.txt) & [`verification/fuzz_campaign.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/fuzz_campaign.md) |
| **Zero Unsafe** | Zero unsafe pointer usage | 0 `unsafe` imports in package `src` | [`src/`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/src/) & [`verification/static_campaign.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/static_campaign.md) |
| **Bug Audit** | Document latent bugs across repos | Cross-repository audit documented | [`SDFGH/BUGS.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/SDFGH/BUGS.md) & [`Kavinraj696/verified-bugs`](https://github.com/Kavinraj696/verified-bugs) |
| **Decision Log** | Architectural trade-offs schema | 10 architectural decision entries | [`DECISIONS.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/DECISIONS.md) |

---

## Quick Start & Verification

### 1. Build
```bash
make build
# or: go build -v ./src/...
```

### 2. Run Test Suite
```bash
make test
# or: go test -v ./tests/port/...
```

### 3. Run Benchmarks & Statistical Analysis (`benchstat`)
```bash
make bench
# or: go test -run '^$' -bench=. -benchmem -count=10 ./tests/port/... > verification/raw/bench_raw.txt
#     benchstat verification/raw/bench_raw.txt > verification/raw/benchstat_output.txt
```

### 4. Differential Fuzzing Campaign
```bash
make fuzz
# or: go run fuzz/harness/fuzz_cluster9.go
```

### 5. Repository Audit & Evidence Pipeline
```bash
make verify
```

---


- **Verified Bugs & Edge Cases Audit**: Documented in [`SDFGH/BUGS.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/SDFGH/BUGS.md) and cross-referenced in **[`Kavinraj696/verified-bugs`](https://github.com/Kavinraj696/verified-bugs)** (Bug Hunter Bonus +3).

## Central Evidence & Artifact Index

For complete execution metadata and raw data traceability, see [`verification/METRICS.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/METRICS.md).

| Evidence Category | Raw Artifact Location | Report Location |
| :--- | :--- | :--- |
| **Central Metrics Registry** | [`verification/METRICS.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/METRICS.md) | Authoritative Metadata Index |
| **Statistical Benchmarks** | [`verification/raw/benchstat_output.txt`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/benchstat_output.txt) | [`verification/benchmark_report.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/benchmark_report.md) |
| **Benchmark Campaign** | [`verification/raw/benchmark_campaign.csv`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/benchmark_campaign.csv) | [`verification/benchmark_campaign.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/benchmark_campaign.md) |
| **Statement Coverage** | [`verification/raw/coverage_summary.txt`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/coverage_summary.txt) | [`verification/coverage_campaign.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/coverage_campaign.md) |
| **Percentile Latency** | [`verification/raw/latency_campaign.csv`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/latency_campaign.csv) | [`verification/latency_campaign.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/latency_campaign.md) |
| **Operational Throughput** | [`verification/raw/throughput_campaign.csv`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/throughput_campaign.csv) | [`verification/throughput_campaign.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/throughput_campaign.md) |
| **Runtime Memory Stats** | [`verification/raw/memory_campaign.csv`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/memory_campaign.csv) | [`verification/memory_campaign.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/memory_campaign.md) |
| **CPU Profile (`pprof`)** | [`verification/raw/cpu_top.txt`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/cpu_top.txt) | [`verification/final_audit.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/final_audit.md) |
| **Memory Profile (`pprof`)** | [`verification/raw/mem_top.txt`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/mem_top.txt) | [`verification/final_audit.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/final_audit.md) |
| **Static Code Analysis** | [`verification/raw/static/`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/static/) | [`verification/static_campaign.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/static_campaign.md) |
| **Differential Fuzzing** | [`fuzz/log.txt`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/fuzz/log.txt) | [`verification/fuzz_campaign.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/fuzz_campaign.md) |

---

## Measured Performance & Verification Summary

Every audited metric below was verified against the corresponding raw artifact saved in [`verification/raw/`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/):

| Operation | Measured Execution Speed | Measured Heap Memory | Measured Allocations | Evaluation |
| :--- | :--- | :--- | :--- | :--- |
| **`ComparePure`** | **$2.723\text{ ns} \pm 4\%$** | **$0\text{ B/op} \pm 0\%$** | **$0\text{ allocs/op} \pm 0\%$** | Zero Heap Allocation |
| **`EqualPure`** | **$5.486\text{ ns} \pm 2\%$** | **$0\text{ B/op} \pm 0\%$** | **$0\text{ allocs/op} \pm 0\%$** | Zero Heap Allocation |
| **`Add`** | **$55.16\text{ ns} \pm 2\%$** | **$48\text{ B/op} \pm 0\%$** | **$2\text{ allocs/op} \pm 0\%$** | Single Result Allocation |
| **`Subtract`** | **$46.44\text{ ns} \pm 1\%$** | **$48\text{ B/op} \pm 0\%$** | **$2\text{ allocs/op} \pm 0\%$** | Single Result Allocation |
| **`Multiply`** | **$94.88\text{ ns} \pm 2\%$** | **$64\text{ B/op} \pm 0\%$** | **$2\text{ allocs/op} \pm 0\%$** | 15-Bit Decomposition |
| **`Divide`** | **$304.9\text{ ns} \pm 1\%$** | **$192\text{ B/op} \pm 0\%$** | **$8\text{ allocs/op} \pm 0\%$** | Knuth Algorithm D |
| **`DivRem`** | **$328.2\text{ ns} \pm 2\%$** | **$192\text{ B/op} \pm 0\%$** | **$8\text{ allocs/op} \pm 0\%$** | Single-Pass Division |
| **`ToStringDec1Limb`** | **$22.61\text{ ns} \pm 1\%$** | **$16\text{ B/op} \pm 0\%$** | **$1\text{ allocs/op} \pm 0\%$** | Fast Integer Format |

### Key Summary Metrics
- **Statement Coverage**: **88.7% of statements** in package `src` (verified via `go tool cover`).
- **Differential Fuzz Total**: **9,696,250 test cases** against live Node.js JSBI reference oracle with 0 mismatches.
- **Original Test Integrity**: 5/5 original test files in `tests/original/` are unmodified and passing.
- **Static Analysis**: `golangci-lint` (0 issues), `govulncheck` (0 vulnerabilities), `gosec` (0 critical flaws).
- **Thread Safety**: Lock-free thread-safety for read operations is inferred from value-independence immutability design. Race detector execution was excluded on Windows host due to missing CGO GCC toolchain (`CGO_ENABLED=1` required).
- **Memory Footprint**: Runtime heap statistics (`HeapAlloc`, `HeapInuse`, `Sys`) were measured using `runtime.ReadMemStats`; operating-system RSS was not measured due to OS page-cache variance.

> [!NOTE]
> **Measurement Workflow Execution Counts**: Different measurement workflows intentionally use different execution counts based on their design purpose:
> - The Benchmark Campaign (`benchmark_campaign.csv`) and Coverage Campaign (`coverage_campaign.csv`) archive 8 complete, end-to-end execution passes.
> - Statistical benchstat analysis (`benchstat_output.txt`) independently executes `go test -bench=. -benchmem -count=10` containing 10 statistical iterations.
> - Percentile Latency, Throughput, Memory MemStats, and Regression test campaigns archive 10 runs.
> 
> These represent distinct, purpose-built measurement workflows rather than data inconsistencies. See [`verification/METRICS.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/METRICS.md) for full traceability.

---

## Repository Layout

```
your-port/
├── README.md               ← Migration rationale, build instructions & verification index
├── DECISIONS.md            ← 10 architectural divergence decisions in 7-part schema
├── CHANGELOG.md            ← One line per meaningful code change
├── Dockerfile              ← Runnable artifact build & test container manifest
├── Makefile                ← One-command build, test, bench, fuzz, and verify targets
├── .port-mortem.toml       ← Track letter, source URL, kickoff hash & submission metadata
│
├── src/                    ← Pure Go implementation of JSBI BigInt engine
├── tests/
│   ├── original/           ← Unmodified upstream JSBI TypeScript test files (5/5)
│   └── port/               ← Go unit, property, stress & verification test suites
│
├── fuzz/
│   ├── harness/            ← Go fuzzer drivers & Node ESM oracle driver
│   └── log.txt             ← Real fuzzing log (9,696,250 cases, 0 mismatches)
│
├── bench/
│   ├── methodology.md      ← Measurement methodology & machine specs
│   └── results.json        ← Machine-readable JSON metrics (p99, throughput, memory, fuzz)
│
└── verification/           ← Verification & Audit Campaign artifacts
    ├── METRICS.md          ← Central Metrics Registry & Metadata Index
    ├── CAMPAIGN.md         ← Master Verification Campaign Index
    └── raw/                ← Raw tool outputs (benchstat, coverage, pprof, CSVs, static logs)
```
