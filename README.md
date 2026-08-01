# JSBI Go Port

Go port of `GoogleChromeLabs/jsbi` (TypeScript, Apache-2.0).

## Track & Kickoff Metadata
- **Track**: Track C (JS/TS -> Go)
- **Source Repository**: `https://github.com/GoogleChromeLabs/jsbi`
- **Kickoff Commit**: `5382367c7e3199858d36bb620977e1f90605bcb9`
- **Original Test Suite SHA256**: `1309534b7a6d5d89f5340914b244d49eebb8c676d0040eb898570102dd973585`

## Architecture Policy
- **Selected Strategy**: Option A (Faithful Limb-Based Go Representation).
- **Limb Data Model**: 30-bit digit slice (`[]uint32` with mask `0x3FFFFFFF`) + boolean `sign`.

## One-Command Build & Test
```bash
# Run all unit tests
export GOTMPDIR='c:/Users/madha/OneDrive/Desktop/port TS-GO/tmp' && ./go_sdk/go/bin/go.exe test -c -o tmp/test.exe ./tests/port && ./tmp/test.exe -test.v

# Run allocation benchmarks
./tmp/test.exe -test.run '^$' -test.bench=. -test.benchmem
```

## Status & Progress

| Cluster | Status | Unit Tests | Differential Fuzzing | Fuzz Duration | Survival Rate |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **1. Construction & Parsing** | COMPLETE | 6/6 PASS | 251,000 cases (Element-by-Element Limb Match) | 65.11s | 100% |
| **2. Comparison** | COMPLETE | 4/4 PASS | 389,000 cases (Compare + 6 Relational Operators + NaN) | 65.13s | 100% |
| **3. Add / Subtract** | COMPLETE | 6/6 PASS (16/16 Total) | 525,000 cases (Add + Sub + Limb Match + Canonical Zero + Identities) | 65.11s | 100% |
| **4. Multiply** | PENDING | - | - | - | - |
| **5. Divide / Remainder** | PENDING | - | - | - | - |
| **6. Shifts** | PENDING | - | - | - | - |
| **7. Bitwise** | PENDING | - | - | - | - |
| **8. asIntN / asUintN** | PENDING | - | - | - | - |
| **9. toString / Radix** | PENDING | - | - | - | - |

- **Allocation Performance**:
  - `Compare` and `Equal`: `0 B/op, 0 allocs/op` (4.90 ns/op and 11.29 ns/op).
  - `Add` and `Subtract`: `48 B/op, 2 allocs/op` (76.43 ns/op and 73.37 ns/op).
- **Original JSBI Test Suite**: 5 files verified passing unmodified on clean checkout.
