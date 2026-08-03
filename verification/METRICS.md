# Central Metrics Registry — JSBI Go Port

This registry serves as the authoritative metadata and raw artifact traceability index for the Go port of \GoogleChromeLabs/jsbi\.

---

## 1. Measurement Campaigns & Execution Metadata

| Campaign / Measurement Category | Execution Count | Workflow Description | Raw Artifact Location | Report Location |
| :--- | :--- | :--- | :--- | :--- |
| **Benchmark Campaign** | 8 Executions | 8 independent benchmark suite passes | [\erification/raw/benchmark_campaign.csv\](verification/raw/benchmark_campaign.csv) | [\erification/benchmark_campaign.md\](verification/benchmark_campaign.md) |
| **Coverage Campaign** | 8 Executions | 8 independent statement coverage passes | [\erification/raw/coverage_campaign.csv\](verification/raw/coverage_campaign.csv) | [\erification/coverage_campaign.md\](verification/coverage_campaign.md) |
| **Benchstat Statistical Analysis** | 10 Iterations | Toolchain \go test -bench=. -count=10\ run | [\erification/raw/benchstat_output.txt\](verification/raw/benchstat_output.txt) | [\erification/benchmark_report.md\](verification/benchmark_report.md) |
| **Percentile Latency Campaign** | 10 Executions | 10 runs (5,000 sample batches/run) | [\erification/raw/latency_campaign.csv\](verification/raw/latency_campaign.csv) | [\erification/latency_campaign.md\](verification/latency_campaign.md) |
| **Operational Throughput Campaign** | 10 Executions | 10 single-threaded throughput runs | [\erification/raw/throughput_campaign.csv\](verification/raw/throughput_campaign.csv) | [\erification/throughput_campaign.md\](verification/throughput_campaign.md) |
| **Runtime Memory MemStats Campaign** | 10 Executions | 10 runs across 400,000 mixed operations | [\erification/raw/memory_campaign.csv\](verification/raw/memory_campaign.csv) | [\erification/memory_campaign.md\](verification/memory_campaign.md) |
| **Full Regression Suite Campaign** | 10 Executions | 10 runs of 44 unit test suites | [\erification/raw/coverage.out\](verification/raw/coverage.out) | [\erification/test_campaign.md\](verification/test_campaign.md) |
| **Differential Fuzzing Campaign** | 9,696,250 Cases | Fuzzing against Node.js JSBI reference oracle | [\uzz/log.txt\](fuzz/log.txt) | [\erification/fuzz_campaign.md\](verification/fuzz_campaign.md) |
| **Static Code & Security Analysis** | 5 Tools | \go vet\, \staticcheck\, \golangci-lint\, \govulncheck\, \gosec\ | [\erification/raw/static/\](verification/raw/static/) | [\erification/static_campaign.md\](verification/static_campaign.md) |

---

## 2. Key Statement & Performance Figures

- **Statement Coverage**: **88.7%** (Raw: [\erification/raw/coverage_summary.txt\](verification/raw/coverage_summary.txt))
- **\ComparePure\ Execution Speed**: **2.723 ns/op ±4%** (Raw: [\erification/raw/benchstat_output.txt\](verification/raw/benchstat_output.txt))
- **\EqualPure\ Execution Speed**: **5.486 ns/op ±2%** (Raw: [\erification/raw/benchstat_output.txt\](verification/raw/benchstat_output.txt))
- **\Add\ Execution Speed**: **55.16 ns/op ±2%** (Raw: [\erification/raw/benchstat_output.txt\](verification/raw/benchstat_output.txt))
- **\Subtract\ Execution Speed**: **46.44 ns/op ±1%** (Raw: [\erification/raw/benchstat_output.txt\](verification/raw/benchstat_output.txt))
- **\Multiply\ Execution Speed**: **94.88 ns/op ±2%** (Raw: [\erification/raw/benchstat_output.txt\](verification/raw/benchstat_output.txt))
- **\Divide\ Execution Speed**: **304.9 ns/op ±1%** (Raw: [\erification/raw/benchstat_output.txt\](verification/raw/benchstat_output.txt))
- **\golangci-lint\ Issue Count**: **0 issues** (Raw: [\erification/raw/static/golangci_lint.txt\](verification/raw/static/golangci_lint.txt))
- **\govulncheck\ Vulnerabilities**: **0 vulnerabilities** (Raw: [\erification/raw/static/govulncheck.txt\](verification/raw/static/govulncheck.txt))
- **Unmodified Original Upstream Test Integrity**: **5 / 5 files unmodified** (Raw: [\erification/original_test_integrity.md\](verification/original_test_integrity.md))

---

## 3. Workflow Execution Count Explanation

Different measurement workflows intentionally use different execution counts based on their design purpose:

1. **Campaign Automation Suite** (\enchmark_campaign.csv\ and \coverage_campaign.csv\):
   - Configured for **8 complete, end-to-end execution passes** across the package test and benchmark suite.
2. **Go Toolchain Benchstat Suite** (\enchstat_output.txt\):
   - Executed via \go test -bench=. -benchmem -count=10\ yielding **10 statistical iterations** for \enchstat\ statistical analysis.
3. **Micro-Benchmarking Tools** (\latency_campaign.csv\, \	hroughput_campaign.csv\, \memory_campaign.csv\):
   - Executed across **10 independent runs** to sample memory stats, ops/sec throughput, and tail latency percentiles.

These different execution counts represent distinct, purpose-built measurement workflows rather than data inconsistencies.
