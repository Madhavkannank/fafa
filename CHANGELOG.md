# Changelog

All notable changes to the JSBI Go Port project will be documented in this file.

## [Unreleased] - Cluster 2 Complete

### Added
- **Cluster 1 Implementation (`src/`)**: `BigInt` struct (`sign bool`, `digits []uint32` using 30-bit limbs), `FromInt64`, `FromUint64`, `FromFloat64`, `FromBool`, `BigIntVal`, `FromString` (with radix auto-detection and JS whitespace predicate), sentinel errors `ErrSyntax`, `ErrRange`, `ErrType`.
- **Cluster 2 Implementation (`src/comparison.go`)**: `Compare`, `Equal`, `NotEqual`, `LessThan`, `LessThanOrEqual`, `GreaterThan`, `GreaterThanOrEqual`, `CompareToFloat64`, `EqualToFloat64`, `CompareToInt64`, dynamic operator helpers (`EQ`, `NE`, `LT`, `LE`, `GT`, `GE`).
- **Unit Test Suite (`tests/port/`)**:
  - `construction_test.go` — 6/6 test suites PASS.
  - `comparison_test.go` — 3/3 test suites PASS (including targeted tests for `NaN`, `+Inf`, `-Inf`, `±0`, subnormals, $2^{53}-1$, $2^{53}$, $2^{53}+1$, and multi-limb values).
  - Standing regression suite: 9/9 total test suites PASS cleanly.
- **Allocation Benchmarks (`tests/port/comparison_test.go`)**:
  - `BenchmarkComparePure`: 2.885 ns/op, `0 B/op`, `0 allocs/op`.
  - `BenchmarkEqualPure`: 6.876 ns/op, `0 B/op`, `0 allocs/op`.
  - Achieved Go Implementation Goal of zero allocations for comparison operations.
- **Differential Fuzzing Harnesses (`fuzz/harness/`)**:
  - Cluster 1: `fuzz_cluster1.go` & `oracle.mjs` — 251,000 cases in 65.11s (100% survival rate, element-by-element 30-bit limb array match).
  - Cluster 2: `fuzz_cluster2.go` & `oracle_cluster2.mjs` — 692,000 cases in 65.03s (100% survival rate across `Compare` and all 6 individual comparison operators).
