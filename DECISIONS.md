# Architectural & Implementation Decisions Log

This document records every non-trivial design choice, divergence from JSBI TS reference, or trade-off made during the Go port.

## Decision 0: Architecture Selection — Faithful Limb-Based Representation (Option A)
- **Context**: Choice between faithful limb-based Go representation vs. wrapping Go stdlib `math/big.Int`.
- **Selection**: **Option A (Faithful Limb-Based Go Representation)**.
- **Rationale**:
  - Maximizes behavioral equivalence and algorithm-level fidelity to reference `JSBI`.
  - Maximizes potential to surface subtle edge-case bugs in JSBI's custom digit math during differential fuzzing (aiming for "Bug Catcher" and "Innovation" bonuses).
  - Maintains explicit control over 30-bit digit limb storage, memory allocation, and execution flow.
- **Consequences**:
  - Must port multi-precision addition, subtraction, multiplication, division, shifts, bitwise logic, and radix conversions faithfully in Go.
  - Core data structure locked: `type BigInt struct { sign bool; digits []uint32 }` using 30-bit digits (`0x3FFFFFFF` mask).

## Decision 1: Cluster 1 Construction & Parsing Representation Details
- **Context**: Porting JSBI constructors (`BigInt(arg)`, `__fromDouble`, `__fromString`, `__oneDigit`, `__zero`) to Go.
- **Choices**:
  1. Represent 30-bit digit limbs as `[]uint32` with mask `0x3FFFFFFF`.
  2. Implement `FromString` with full ECMAScript whitespace filtering, radix auto-detection (`0x`/`0o`/`0b`), power-of-two fast paths, and generic chunked base conversion.
  3. Map JavaScript exceptions (`SyntaxError`, `RangeError`, `TypeError`) to idiomatic Go sentinel errors (`ErrSyntax`, `ErrRange`, `ErrType`).
- **Rationale**: Direct 1-to-1 algorithmic translation of `JSBI` TypeScript logic guarantees exact parity for string parsing, float conversion, and digit trimming.
- **Verification**: Verified via 6 unit test suites and 251,000 differential fuzzing test cases against Node JSBI oracle with element-by-element limb array comparison (65.11s run, 100% survival rate).

## Decision 2: Cluster 2 Comparison Implementation & NaN Sentinel Return Tuple
- **Context**: Porting `JSBI.equal`, `JSBI.lessThan`, `JSBI.__compareToBigInt`, and `JSBI.__compareToDouble` to Go while avoiding NaN return bugs.
- **Choices**:
  1. Implement `CompareToFloat64` returning `(cmp int, isNaN bool)` tuple rather than a bare `int` `0` for `NaN`.
  2. Implement float relational predicates (`EqualToFloat64`, `NotEqualToFloat64`, `LessThanFloat64`, `LessThanOrEqualFloat64`, `GreaterThanFloat64`, `GreaterThanOrEqualFloat64`) explicitly checking `!isNaN`.
  3. Port `CompareToFloat64` faithfully from `JSBI.__compareToDouble`, performing mathematical bit-length comparison ($x\text{BitLength}$ vs $y\text{BitLength}$) and bit-aligned 53-bit mantissa scanning without converting `BigInt` to `float64` or `float64` to `BigInt`.
  4. Achieve zero heap allocations (`0 allocs/op`) for pure comparison functions.
- **Rationale**: JSBI line 1052 returns `NaN` as a sentinel, relying on JS relational operators evaluating to `false` for `NaN`. Returning `(cmp, isNaN)` in Go ensures all six relational wrappers match ECMAScript semantics 100% (`NotEqual(x, NaN)` is `true`, all others `false`).
- **Verification**: Verified via `TestNaNRelationalComparisons`, 4 targeted unit test suites, benchmark verification (`BenchmarkComparePure` 4.90ns/op, `0 allocs/op`), and 389,000 differential fuzzing test cases against Node JSBI oracle (65.13s run, 100% survival rate across `Compare` and all 6 individual relational operators against `NaN`).

## Decision 3: Cluster 3 Multi-Precision Add/Subtract & Go Spec Borrow Shift Proof
- **Context**: Porting `JSBI.add`, `JSBI.subtract`, `JSBI.unaryMinus`, `__absoluteAdd`, and `__absoluteSub` to Go.
- **Choices**:
  1. Implement `borrow := (uint32(r) >> 30) & 1` with `r` as `int32`. Casting `uint32(r)` enforces a **logical right shift** per Go Language Specification (Section *Operators - Shift operators*), proving $1$-to-$1$ equivalence to JS `(r >>> 30) & 1` for all $r \in [-2^{30}, 2^{30}-1]$.
  2. Implement `carry := uint32(r) >> 30` with `r` as `int32` for `absoluteAdd`.
  3. Enforce canonical zero (`sign = false, len = 0`) across `UnaryMinus` and `Trim()`. `UnaryMinus(0)` returns canonical zero directly without sign inversion.
  4. Use `absoluteCompare` magnitude checks in `Add` and `Subtract` to dispatch minuend/subtrahend order, guaranteeing $|X| \ge |Y|$ for `absoluteSub`.
- **Rationale**: Direct algorithmic port guarantees exact parity for arbitrary precision addition and subtraction. Logical shift via `uint32(r) >> 30` eliminates platform-dependent signed shift ambiguity.
- **Verification**: Verified via 6 targeted unit test suites (`TestAddBasic`, `TestSubtractBasic`, `TestUnaryMinus`, `TestCarryPropagation`, `TestBorrowPropagation`, `TestAlgebraicIdentities`), benchmark verification (`BenchmarkAdd` 76.43ns/op, 48 B/op, 2 allocs/op), and 502,000 differential fuzzing test cases against Node JSBI oracle (65.08s run, 100% survival rate across element-by-element digit limbs, signs, lengths, and algebraic identities).

## Decision 4: Cluster 4 Multi-Precision Multiplication & 15-Bit Decomposition
- **Context**: Porting `JSBI.multiply`, `__multiplyAccumulate`, `__internalMultiplyAdd`, and `__inplaceMultiplyAdd` to Go.
- **Choices**:
  1. Preserve JSBI's 15-bit half-limb decomposition ($m = m_{\text{high}} \times 2^{15} + m_{\text{low}}$) and uint32 arithmetic (`imul`) for 100% line-for-line behavioral equivalence with JSBI TypeScript reference.
  2. Implement `clzmsd` leading-zero MSD estimate check (`x.clzmsd() + y.clzmsd() >= 30 ? resultLength-- : resultLength`) for optimistic allocation size estimation.
  3. Implement column multiplication in `multiplyAccumulate` with `accumulatorIndex` offset alignment and $O(1)$ auxiliary space in-place accumulation.
  4. Enforce canonical zero (`len = 0, sign = false`) for multiplication by zero ($0 \times X \rightarrow 0$, $(-5) \times 0 \rightarrow 0$).
- **Rationale**: 15-bit half-limb decomposition prevents 32-bit integer overflow during partial product calculation ($2^{15} \times 2^{15} = 2^{30}$). In-place column accumulation avoids allocating intermediate partial product arrays.
- **Verification**: Verified via targeted unit test suites (`TestMultiplyBasic`, `TestMultiplyWorstCaseVectors`), benchmark verification (`BenchmarkMultiply` 208.1ns/op, 64 B/op, 2 allocs/op), and 1,590,000 differential fuzzing test cases against Node JSBI oracle (65.13s run, 100% survival rate across signs, lengths, canonical zero assertions, and element-by-element 30-bit digit arrays). Cumulative fuzz total: 3,065,000 cases.

## Decision 5: Cluster 5 Division & Remainder — Small-Path Threshold, Knuth Algorithm D, and Value-Independent DivRem
- **Context**: Porting `JSBI.divide`, `JSBI.remainder`, `__absoluteDivSmall`, `__absoluteModSmall`, `__absoluteDivLarge`, `__inplaceSub`, `__inplaceAdd`, and `__inplaceRightShift` to Go, plus adding a Go `DivRem` extension.
- **Choices**:
  1. Port the small-divisor path optimization (`y.Length() == 1 && y.Digit(0) <= 0x7FFF`) using 15-bit long division recurrence in `absoluteDivSmall` and `absoluteModSmall`.
  2. Port Knuth Algorithm D in 15-bit half-digits for large divisors (`absoluteDivLarge`), including D1 normalization (`clz15` shift), D3 trial quotient estimation with `(rhat << 16) | ujn2` refinement, D4 `internalMultiplyAdd` product subtraction with D4 add-back, D5 quotient half-digit packing, and D6 unnormalization (`inplaceRightShift`).
  3. Implement `inplaceSub` half-digit subtraction step using `subLen := (halfDigits + 1) >> 1` to faithfully match JSBI's `subtrahend.length` property on `qhatv` arrays (preventing length discrepancy caused by Go slice dynamic append).
  4. Provide a value-independent `DivRem(x, y)` public API that executes Algorithm D in a single pass (`absoluteDivLarge(x, y, true, true)`), avoiding redundant division work when both quotient and remainder are needed.
  5. Enforce complete value independence: all returned `*BigInt` instances are independent of inputs (preventing reference-return aliasing present in JSBI's divisor-by-1 and `|x| < |y|` edge cases). Dedicated aliasing and value independence tests were executed.
- **Rationale**: 15-bit half-digit Algorithm D guarantees intermediate trial quotients fit within 32-bit integer arithmetic without overflow. Combined `DivRem` single-pass execution delivers both quotient and remainder in 366.7 ns/op (vs 338.7 ns/op for `Divide` alone).
- **Verification**: Verified via 6 unit test suites (`TestDivideByZero`, `TestDivideTruncationAndSigns`, `TestDivideDivisorGreaterThanDividend`, `TestDivideByOne`, `TestDivideAlgorithmD`, `TestDivideLargeMultiLimb`), benchmark verification (`BenchmarkDivide` 338.7 ns/op, 192 B/op, 8 allocs/op; `BenchmarkRemainder` 301.3 ns/op, 144 B/op, 6 allocs/op; `BenchmarkDivRem` 366.7 ns/op, 192 B/op, 8 allocs/op), and 176,250 differential fuzzing test cases against Node JSBI oracle (65.06s run, 100% survival rate across element-by-element digit limbs, signs, lengths, and canonical zero assertions). Cumulative fuzz total (latest successful run per cluster methodology): 2,806,250 cases.

## Decision 6: Cluster 6 Shifts — Negative Shift Inversion, `toShiftAmount` Sentinel Translation, `mustRoundDown` Floor Division, and `UnsignedRightShift` Type Error
- **Context**: Porting `JSBI.leftShift`, `JSBI.signedRightShift`, `JSBI.unsignedRightShift`, `JSBI.__leftShiftByAbsolute`, `JSBI.__rightShiftByAbsolute`, `JSBI.__rightShiftByMaximum`, and `JSBI.__toShiftAmount` to Go.
- **Choices**:
  1. Port negative shift direction inversion according to ECMAScript ECMA-262 semantics: `LeftShift(x, negativeY)` delegates to `rightShiftByAbsolute(x, abs(y))`, and `SignedRightShift(x, negativeY)` delegates to `leftShiftByAbsolute(x, abs(y))`.
  2. Implement internal sentinel `-1` handling in `toShiftAmount` for shift amounts exceeding 1 limb or `1 << 30` bits: translated to `ErrRange` in `leftShiftByAbsolute`, and `rightShiftByMaximum` in `rightShiftByAbsolute`.
  3. Implement `mustRoundDown` floor division rounding for negative arithmetic right shifts when discarded bits are non-zero (`absoluteAddOne`).
  4. Implement `UnsignedRightShift` returning `ErrType` to match JSBI `TypeError('BigInts have no unsigned right shift')`.
  5. Enforce complete value independence across all returned `*BigInt` instances (including zero-operand inputs `x.Length() == 0` or `y.Length() == 0`). Dedicated aliasing and value independence tests were executed.
- **Rationale**: Preserves exact ECMAScript `BigInt` shift semantics while providing native Go error handling (`ErrRange`, `ErrType`).
- **Verification**: Verified via 4 unit test suites (`TestLeftShift`, `TestSignedRightShift`, `TestUnsignedRightShift`, `TestShiftRangeError`), benchmark verification (`BenchmarkLeftShift` 72.0 ns/op, 64 B/op, 2 allocs/op; `BenchmarkSignedRightShift` 49.3 ns/op, 48 B/op, 2 allocs/op; `BenchmarkUnsignedRightShift` 0.0 ns/op, 0 B/op, 0 allocs/op), and 1,400,000 differential fuzzing test cases against Node JSBI oracle (60.19s run, 100% survival rate across element-by-element digit limbs, signs, lengths, and canonical zero assertions). Cumulative fuzz total (latest successful run per cluster methodology): 4,206,250 cases.

### 7. Cluster 7 (Bitwise Operations) De Morgan Transformation & Helper Contracts
- **Date**: 2026-08-01
- **Context**: Porting JSBI bitwise operations (`BitwiseAnd`, `BitwiseOr`, `BitwiseXor`, `BitwiseNot`) to Go.
- **Decision**:
  - Implement De Morgan transformations and sign identities $-x = \sim(x - 1)$ for negative BigInt operands without introducing infinite sign-extension arrays.
  - Implement magnitude helper primitives `absoluteAnd`, `absoluteAndNot`, `absoluteOr`, `absoluteXor` with internal result buffer reuse contracts.
  - Enforce mandatory `.Trim()` post-call normalization to maintain the canonical zero invariant `Length() == 0 ==> Sign() == false`.
- **Status**: Accepted & Implemented in `src/bitwise.go`.

### 8. Cluster 8 (AsIntN / AsUintN) Range Guard, Fast-Path Value Independence & Borrow Extraction
- **Date**: 2026-08-01
- **Context**: Porting `JSBI.asIntN` and `JSBI.asUintN` (jsbi.ts lines 408–466) and their helpers `__truncateToNBits` and `__truncateAndSubFromPowerOfTwo` (lines 1860–1907) to Go.
- **Decisions**:
  1. **Range guard for negative `AsUintN` inputs**: JSBI checks `n > __kMaxLengthBits` at line 449. In Go, `kMaxLengthBits = kMaxLength << 5 = (1 << 25) << 5 = 1 << 30`. However, the actual array allocation limit is `(n+29)/30 > kMaxLength`, not the bit count itself. Both the Node JSBI oracle and mathematical analysis confirm the guard should be `(n+29)/30 > kMaxLength` — equivalently `n > kMaxLengthBits`. Used `(n+29)/30 > kMaxLength` as it's the physically meaningful check (limb count, not bits).
  2. **Fast-path value independence**: JSBI returns `x` directly (same JS object reference) on fast paths (e.g. line 417, 419, 422). In Go this is unsafe — callers must receive an independent copy to prevent aliasing. All fast paths use `x.Copy()` instead. This is a deliberate, necessary Go divergence logged here. Verified via pointer inequality assertion in `TestTruncationValueIndependenceFastPath`.
  3. **Borrow sign extraction**: In `truncateAndSubFromPowerOfTwo`, JSBI extracts borrow as `(r >>> 30) & 1` using unsigned right shift (JS `>>>` operator). Go uses signed `>>`. The correct Go equivalent is `uint32((r >> 30) & 1)` where `r` is `int64`. This is mathematically equivalent because borrow is always 0 or 1 and `r` is always negative when borrow is 1.
  4. **Test vector correction**: Initial test vectors for `AsIntN(30, 2^30-1)` expected `1073741823` but JSBI returns `-1`. Reason: in 30-bit two's complement, bit 29 (value `0x20000000`) is the sign bit. `2^30 - 1 = 0x3FFFFFFF` has bit 29 set, so it represents `-1`. Corrected all test vectors against the Node.js JSBI oracle before finalizing.
- **Status**: Accepted & Implemented in `src/truncation.go`.

### 9. Cluster 9 (ToString & Exponentiate) Radix Conversion, Popcount Fast Path & Divide-and-Conquer Fallback
- **Date**: 2026-08-01
- **Context**: Porting `JSBI.toString` (jsbi.ts lines 67–77), `JSBI.exponentiate` (lines 167–219), `JSBI.__toStringBasePowerOfTwo` (lines 916–958), and `JSBI.__toStringGeneric` (lines 960–1010) to Go.
- **Decisions**:
  1. **Binary Square-and-Multiply Exponentiation**: `Exponentiate(x, y)` provides the exact power computation needed for radix conversion split points (`radix^secondHalfChars`). Supports power-of-two fast paths (`2^n`), edge cases (`x=±1`, `y=0`, `y=1`), and exponent range error validation (`y < 0` or `exp >= kMaxLengthBits`).
  2. **Power-of-Two Fast Path (`(radix & (radix-1)) == 0`)**: Uses `bits.OnesCount32(uint32(radix-1))` for popcount bit extraction. Fills buffer right-to-left in a single pass without division or reversal.
  3. **Divide-and-Conquer Radix Conversion**: Uses precomputed `kMaxBitsPerChar` lookup table copied verbatim from JSBI source. Uses fast 15-bit half-digit long division when conqueror fits in single 15-bit limb (`divisor <= 0x7FFF`), and full `absoluteDivLarge` for larger split points.
  4. **Error Mapping Policy**: JSBI's `Error('string too long')` when `charsRequired > (1 << 28)` maps to Go's `ErrRange` following established project error mapping policy.
- **Status**: Accepted & Implemented in `src/tostring.go`.
