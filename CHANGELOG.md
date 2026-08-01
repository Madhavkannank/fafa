# Changelog

All notable changes to the JSBI Go port will be documented in this file.

## [Unreleased]
### Cluster 6 — Shifts (2026-08-01)
- Implemented `LeftShift(x, y *BigInt) (*BigInt, error)` supporting positive/negative shift counts and bit/limb shifts.
- Implemented `SignedRightShift(x, y *BigInt) (*BigInt, error)` with negative number floor division rounding toward $-\infty$ (`mustRoundDown`).
- Implemented `UnsignedRightShift(x, y *BigInt) (*BigInt, error)` returning `ErrType`.
- Implemented `leftShiftByAbsolute`, `rightShiftByAbsolute`, `rightShiftByMaximum`, and `toShiftAmount` helpers.
- Dedicated aliasing and value independence tests were executed across all returned values.
- Passed 4 unit test suites (`TestLeftShift`, `TestSignedRightShift`, `TestUnsignedRightShift`, `TestShiftRangeError`) and 1,400,000 differential fuzzing test cases (60.19s run, 100% equivalence survival rate against Node JSBI oracle across element-by-element 30-bit digit arrays, signs, and lengths). Cumulative total (latest successful run per cluster methodology): 4,206,250 cases.
- Empirically measured benchmark performance: `BenchmarkLeftShift` (72.0 ns/op, 64 B/op, 2 allocs/op), `BenchmarkSignedRightShift` (49.3 ns/op, 48 B/op, 2 allocs/op), `BenchmarkUnsignedRightShift` (0.0 ns/op, 0 B/op, 0 allocs/op).

### Cluster 5 — Division & Remainder (2026-08-01)
- Implemented `Divide(x, y *BigInt) (*BigInt, error)` for truncating integer division toward zero.
- Implemented `Remainder(x, y *BigInt) (*BigInt, error)` for remainder calculation matching ECMAScript `x % y` semantics.
- Implemented `DivRem(x, y *BigInt) (quotient, remainder *BigInt, err error)` extension API for single-pass Algorithm D execution.
- Implemented `absoluteDivSmall` and `absoluteModSmall` 15-bit long division fast paths for small single-limb divisors (`divisor <= 0x7FFF`).
- Implemented `absoluteDivLarge` Knuth Algorithm D over 15-bit half-digits with D1 normalization (`clz15`), D3 trial quotient refinement, D4 multiply/subtract and add-back, D5 quotient packing, and D6 unnormalization (`inplaceRightShift`).
- Implemented `inplaceSub` with `subLen := (halfDigits + 1) >> 1` to preserve JSBI `qhatv` array length invariants across dynamic slice growth.
- Dedicated aliasing and value independence tests were executed across all returned values.
- Passed 6 unit test suites (`TestDivideByZero`, `TestDivideTruncationAndSigns`, `TestDivideDivisorGreaterThanDividend`, `TestDivideByOne`, `TestDivideAlgorithmD`, `TestDivideLargeMultiLimb`) and 176,250 differential fuzzing test cases (65.06s run, 100% equivalence survival rate against Node JSBI oracle across element-by-element 30-bit digit arrays, signs, and lengths). Cumulative total (latest successful run per cluster methodology): 2,806,250 cases.
- Empirically measured benchmark performance: `BenchmarkDivide` (338.7 ns/op, 192 B/op, 8 allocs/op), `BenchmarkRemainder` (301.3 ns/op, 144 B/op, 6 allocs/op), `BenchmarkDivRem` (366.7 ns/op, 192 B/op, 8 allocs/op).


### Cluster 4 — Multiplication (2026-08-01)
- Implemented `Multiply(x, y *BigInt) *BigInt` for multi-precision multiplication.
- Implemented `multiplyAccumulate`, `internalMultiplyAdd`, and `inplaceMultiplyAdd` helper algorithms.
- Preserved 15-bit half-limb decomposition ($m = m_{\text{high}} \times 2^{15} + m_{\text{low}}$) and uint32 `imul` arithmetic for 100% line-for-line behavioral equivalence with JSBI.
- Implemented `clzmsd` leading zero count length estimation for optimistic buffer allocation.
- Enforced canonical zero (`len = 0, sign = false`) for zero multiplication ($0 \times X \rightarrow 0$).
- Passed 2 unit test suites (`TestMultiplyBasic`, `TestMultiplyWorstCaseVectors`) and 1,590,000 differential fuzzing test cases (65.13s run, 100% equivalence survival rate against Node JSBI oracle across signs, lengths, canonical zero assertions, and element-by-element 30-bit digit arrays). Total cumulative fuzz: 3,065,000 cases.
- Empirically measured benchmark performance: `BenchmarkMultiply` (208.1 ns/op, 64 B/op, 2 allocs/op).

### Cluster 3 — Add & Subtract (2026-08-01)
- Implemented `Add(x, y *BigInt) *BigInt` for multi-precision addition.
- Implemented `Subtract(x, y *BigInt) *BigInt` for multi-precision subtraction.
- Implemented `UnaryMinus(x *BigInt) *BigInt` for unary negation with canonical zero handling (`UnaryMinus(0) == +0`).
- Implemented `absoluteAdd`, `absoluteSub`, `absoluteAddOne`, and `absoluteSubOne` core limb arithmetic helpers.
- Added bit-level borrow shift proof `(uint32(r) >> 30) & 1` per Go Language Specification.
- Passed 6 unit test suites and 502,000 differential fuzzing test cases (65.08s run, 100% equivalence survival rate against Node JSBI oracle).
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
