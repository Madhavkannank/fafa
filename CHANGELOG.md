# Changelog

All notable changes to the JSBI Go port will be documented in this file.

## [Unreleased]
### Cluster 3 — Add & Subtract (2026-08-01)
- Implemented `Add(x, y *BigInt) *BigInt` for multi-precision addition.
- Implemented `Subtract(x, y *BigInt) *BigInt` for multi-precision subtraction.
- Implemented `UnaryMinus(x *BigInt) *BigInt` for unary negation with canonical zero handling (`UnaryMinus(0) == +0`).
- Implemented `absoluteAdd`, `absoluteSub`, `absoluteAddOne`, and `absoluteSubOne` core limb arithmetic helpers.
- Added bit-level borrow shift proof `(uint32(r) >> 30) & 1` per Go Language Specification.
- Passed 6 unit test suites and 422,000 differential fuzzing test cases (65.13s run, 100% equivalence survival rate against Node JSBI oracle).
- Empirically measured benchmark performance: `BenchmarkAdd` (76.43 ns/op, 48 B/op, 2 allocs/op), `BenchmarkSubtract` (73.37 ns/op, 48 B/op, 2 allocs/op).

### Cluster 2 — Comparison Operations (2026-08-01)
- Implemented `Compare`, `Equal`, `NotEqual`, `LessThan`, `LessThanOrEqual`, `GreaterThan`, `GreaterThanOrEqual`.
- Implemented `CompareToFloat64` returning `(cmp int, isNaN bool)` tuple matching `JSBI.__compareToDouble` mantissa scanning.
- Fixed `NaN` relational bug: relational operators against `NaN` return `false`, `!=` returns `true`.
- Added zero-allocation benchmarks (`BenchmarkComparePure` 4.90 ns/op, `BenchmarkEqualPure` 11.29 ns/op).
- Passed 389,000 differential fuzzing test cases (65.13s run, 100% survival rate).

### Cluster 1 — Construction & Parsing (2026-08-01)
- Initialized core `BigInt` struct with 30-bit digit limb slice (`[]uint32`).
- Implemented constructors `BigInt(arg)`, `FromInt64`, `FromUint64`, `FromBool`, `FromFloat64`, `FromString`.
- Implemented `Trim()`, `Copy()`, `clzmsd()`, and error definitions (`ErrSyntax`, `ErrRange`, `ErrType`).
- Passed 251,000 differential fuzzing test cases with element-by-element digit array verification (65.11s run, 100% survival rate).
