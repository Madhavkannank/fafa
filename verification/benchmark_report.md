# Benchmark & Statistical Analysis Report — JSBI Go Port

- **Kickoff Source SHA**: `5382367c7e3199858d36bb620977e1f90605bcb9`
- **Package**: `github.com/Madhavkannank/fafa/src`
- **Measurement Tool**: `go test -run '^$' -bench=. -benchmem -count=10`
- **Statistical Tool**: `benchstat verification/raw/bench_raw.txt`
- **Status**: **MEASURED FROM RAW TOOL OUTPUTS** (2026-08-02)

---

## 1. System Environment

- **CPU**: Intel(R) Core(TM) 5 210H (12 logical cores, 2.20 GHz)
- **RAM**: 16 GB DDR5 System Memory
- **OS**: Microsoft Windows 11 Home (x64)
- **Go Version**: Go 1.22.5 (`go1.22.5.windows-amd64.zip`)

---

## 2. Measured Benchmark Results with Statistical Confidence (`benchstat`)

All figures below are generated directly from raw command execution saved in [`verification/raw/bench_raw.txt`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/bench_raw.txt) and analyzed via [`verification/raw/benchstat_output.txt`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/benchstat_output.txt):

| Benchmark Name | Measured Speed (`sec/op`) | Measured Heap Memory (`B/op`) | Measured Allocations (`allocs/op`) | Memory Variance |
| :--- | :--- | :--- | :--- | :--- |
| **`ComparePure`** | **$2.723\text{ ns} \pm 4\%$** | **$0.000\text{ B} \pm 0\%$** | **$0.000\text{ allocs} \pm 0\%$** | Zero Allocation |
| **`EqualPure`** | **$5.486\text{ ns} \pm 2\%$** | **$0.000\text{ B} \pm 0\%$** | **$0.000\text{ allocs} \pm 0\%$** | Zero Allocation |
| **`ToStringDec1Limb`** | **$22.61\text{ ns} \pm 1\%$** | **$16.00\text{ B} \pm 0\%$** | **$1.000\text{ allocs} \pm 0\%$** | Single Result String |
| **`ToStringHex1Limb`** | **$26.57\text{ ns} \pm 0\%$** | **$16.00\text{ B} \pm 0\%$** | **$2.000\text{ allocs} \pm 0\%$** | Small Result Alloc |
| **`AsUintN`** | **$34.52\text{ ns} \pm 1\%$** | **$40.00\text{ B} \pm 0\%$** | **$2.000\text{ allocs} \pm 0\%$** | Result Copy |
| **`AsIntN`** | **$36.62\text{ ns} \pm 1\%$** | **$40.00\text{ B} \pm 0\%$** | **$2.000\text{ allocs} \pm 0\%$** | Result Copy |
| **`BitwiseNot`** | **$36.62\text{ ns} \pm 0\%$** | **$48.00\text{ B} \pm 0\%$** | **$2.000\text{ allocs} \pm 0\%$** | Single Result Object |
| **`Subtract`** | **$46.44\text{ ns} \pm 1\%$** | **$48.00\text{ B} \pm 0\%$** | **$2.000\text{ allocs} \pm 0\%$** | Single Result Object |
| **`Add`** | **$55.16\text{ ns} \pm 2\%$** | **$48.00\text{ B} \pm 0\%$** | **$2.000\text{ allocs} \pm 0\%$** | Single Result Object |
| **`BitwiseAnd`** | **$74.92\text{ ns} \pm 1\%$** | **$96.00\text{ B} \pm 0\%$** | **$4.000\text{ allocs} \pm 0\%$** | Temporary Limb Slices |
| **`BitwiseOr`** | **$84.45\text{ ns} \pm 1\%$** | **$96.00\text{ B} \pm 0\%$** | **$4.000\text{ allocs} \pm 0\%$** | Temporary Limb Slices |
| **`Multiply`** | **$94.88\text{ ns} \pm 2\%$** | **$64.00\text{ B} \pm 0\%$** | **$2.000\text{ allocs} \pm 0\%$** | Accumulator Allocation |
| **`BitwiseXor`** | **$114.5\text{ ns} \pm 1\%$** | **$112.0\text{ B} \pm 0\%$** | **$5.000\text{ allocs} \pm 0\%$** | Temporary Limb Slices |
| **`Remainder`** | **$282.2\text{ ns} \pm 1\%$** | **$144.0\text{ B} \pm 0\%$** | **$6.000\text{ allocs} \pm 0\%$** | Knuth Quotient Alloc |
| **`Divide`** | **$304.9\text{ ns} \pm 1\%$** | **$192.0\text{ B} \pm 0\%$** | **$8.000\text{ allocs} \pm 0\%$** | Knuth Algorithm D |
| **`DivRem`** | **$328.2\text{ ns} \pm 2\%$** | **$192.0\text{ B} \pm 0\%$** | **$8.000\text{ allocs} \pm 0\%$** | Combined DivRem |
| **`ToStringDec3Limbs`** | **$1.738\mu\text{s} \pm 1\%$** | **$1.164\text{ KiB} \pm 0\%$** | **$71.00\text{ allocs} \pm 0\%$** | Divide-and-Conquer Format |

---

## 3. Raw Evidence Artifact Verification

- Raw Benchmark Run Output: [`verification/raw/bench_raw.txt`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/bench_raw.txt)
- Benchstat Statistical Report: [`verification/raw/benchstat_output.txt`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/raw/benchstat_output.txt)
