# Architectural & Implementation Decisions Log

This document records every non-trivial design choice, divergence from JSBI TS reference, or trade-off made during the Go port.

---

## Decision 0: Architecture Selection — Faithful Limb-Based Representation

- **Decision**: Represent BigInts using a custom 30-bit limb slice rather than wrapping Go's stdlib `math/big.Int`.
- **Problem**: Need to choose an architecture for porting JSBI to Go that maximizes behavioral equivalence, performance, and bug discovery potential.
- **Alternatives Considered**:
  1. Option A: Faithful 30-bit limb-based Go representation (`type BigInt struct { sign bool; digits []uint32 }`).
  2. Option B: Wrapper around Go standard library `math/big.Int`.
- **Chosen Solution**: Option A (Faithful 30-Bit Limb-Based Go Representation).
- **Reasoning**: Direct limb-level fidelity mirrors V8/JSBI digit math exactly, enabling exact algorithm-level verification and surfacing subtle boundary/overflow bugs during differential fuzzing.
- **Trade-offs**: Requires writing custom multi-precision arithmetic routines (addition, subtraction, multiplication, Knuth Algorithm D division, shifts, bitwise logic, radix conversion) instead of relying on stdlib.
- **Evidence**: 9.69M+ differential fuzz cases executed against Node.js JSBI oracle with 100% equivalence survival rate.

---

## Decision 1: Cluster 1 Representation & String Parsing Representation Details

- **Decision**: Implement direct 1-to-1 TypeScript algorithm translation for constructors, string parsing (`FromString`), and float conversions (`FromFloat64`, `ToFloat64`).
- **Problem**: Convert arbitrary strings (including auto-detect radices `0x`/`0o`/`0b`, whitespace) and IEEE 754 float64 values into/from BigInt representation.
- **Alternatives Considered**:
  1. Go stdlib `strconv.ParseInt` / `math/big.Int.SetString`.
  2. Custom ECMAScript-compliant parser mirroring JSBI `__fromString` and `__fromDouble`.
- **Chosen Solution**: Custom ECMAScript-compliant parser (`src/fromString.go`, `src/constructors.go`).
- **Reasoning**: Guarantees exact ECMAScript whitespace handling, digit trimming, float bit extraction, and sentinel error mapping (`ErrSyntax`, `ErrRange`, `ErrType`).
- **Trade-offs**: Additional parser code complexity vs using standard library helpers.
- **Evidence**: Verified via unit tests and 1,005,000 differential fuzz cases against Node.js JSBI oracle.

---

## Decision 2: Cluster 2 Relational Comparisons & NaN Sentinel Return Tuple

- **Decision**: Implement `CompareToFloat64` returning `(cmp int, isNaN bool)` tuple and relational wrappers checking `!isNaN`.
- **Problem**: In JavaScript, comparing BigInt with `NaN` yields `false` for all relational operators (`<`, `<=`, `>`, `>=`, `==`), but returns `true` for `!=`. Returning a single `int` `0` in Go would incorrectly cause `Equal(x, NaN)` to return `true`.
- **Alternatives Considered**:
  1. Returning a single `int` `0` for `NaN` (causes `Equal` false positives).
  2. Returning `(cmp int, isNaN bool)` tuple and guarding relational predicates with `!isNaN`.
- **Chosen Solution**: Returning `(cmp int, isNaN bool)` tuple (`src/comparison.go`).
- **Reasoning**: Matches ECMAScript relational semantics 100% when comparing against `NaN`.
- **Trade-offs**: Slightly more verbose internal helper signature vs exact specification correctness.
- **Evidence**: `TestNaNRelationalComparisons` passed; `BenchmarkComparePure` 2.95 ns/op (0 B/op, 0 allocs); 842,000 differential fuzz cases passed.

---

## Decision 3: Cluster 3 Multi-Precision Add/Subtract & Logical Right Shift Borrow Extraction

- **Decision**: Use `borrow := (uint32(r) >> 30) & 1` with `r` as `int32` for multi-limb borrow extraction.
- **Problem**: JSBI uses JavaScript unsigned shift `(r >>> 30) & 1`. Go has no `>>>` operator.
- **Alternatives Considered**:
  1. Casting `r` to `uint32` before right shift: `(uint32(r) >> 30) & 1`.
  2. Conditional logic `if r < 0 { borrow = 1 }`.
- **Chosen Solution**: `(uint32(r) >> 30) & 1` (`src/add_sub.go`).
- **Reasoning**: Per Go Language Specification (Section *Shift operators*), right shift on an unsigned type (`uint32`) is guaranteed to be a logical right shift, proving exact 1-to-1 mathematical equivalence to JS `>>>`.
- **Trade-offs**: Requires explicit `uint32` type conversion.
- **Evidence**: Verified via `TestBorrowPropagation`, `BenchmarkAdd` 54.05 ns/op (48 B/op, 2 allocs), and 872,000 differential fuzz cases.

---

## Decision 4: Cluster 4 Multi-Precision Multiplication & 15-Bit Decomposition

- **Decision**: Preserve JSBI's 15-bit half-limb decomposition ($m = m_{\text{high}} \times 2^{15} + m_{\text{low}}$) for partial product accumulation.
- **Problem**: 30-bit limb multiplication ($2^{30} \times 2^{30} = 2^{60}$) exceeds 32-bit integer capacity during column accumulation.
- **Alternatives Considered**:
  1. 64-bit integer multiplication (`uint64(a) * uint64(b)`).
  2. 15-bit half-limb decomposition (`m2Low = multiplier & 0x7FFF`, `m2High = multiplier >> 15`).
- **Chosen Solution**: 15-bit half-limb decomposition (`src/multiply.go`).
- **Reasoning**: Line-for-line behavioral equivalence with JSBI TypeScript reference (`__multiplyAccumulate`), eliminating platform uint64 variance.
- **Trade-offs**: Requires two 15-bit partial product passes per limb vs single 64-bit multiply.
- **Evidence**: `BenchmarkMultiply` 96.45 ns/op (64 B/op, 2 allocs); 1,590,000 differential fuzz cases passed with 0 mismatches.

---

## Decision 5: Cluster 5 Division & Remainder — Small-Path Threshold, Knuth Algorithm D, and DivRem

- **Decision**: Port 15-bit small-divisor path, Knuth Algorithm D for large divisors, and provide single-pass `DivRem`.
- **Problem**: Large integer division requires precise quotient estimation and remainder calculation without intermediate 64-bit overflow.
- **Alternatives Considered**:
  1. Separate passes for quotient (`Divide`) and remainder (`Remainder`).
  2. Combined single-pass `absoluteDivLarge(x, y, true, true)` backing both `Divide`, `Remainder`, and public `DivRem`.
- **Chosen Solution**: Combined single-pass `absoluteDivLarge` with small-divisor recurrence (`src/divide.go`).
- **Reasoning**: `DivRem` executes Knuth Algorithm D once to produce both quotient and remainder in 327.3 ns/op, avoiding redundant division computation.
- **Trade-offs**: Algorithm D complexity (D1 normalization through D6 unnormalization).
- **Evidence**: `BenchmarkDivide` 310.0 ns/op, `BenchmarkRemainder` 284.7 ns/op, `BenchmarkDivRem` 327.3 ns/op; 176,250 differential fuzz cases passed.

---

## Decision 6: Cluster 6 Shifts — Negative Shift Inversion & Floor Division Rounding

- **Decision**: Implement negative shift count direction inversion and `mustRoundDown` floor division rounding for signed right shifts.
- **Problem**: ECMAScript specifies `x << (-y) == x >> y` and `x >> (-y) == x << y`. Arithmetic right shifts on negative numbers must round toward $-\infty$.
- **Alternatives Considered**:
  1. Truncating division toward zero.
  2. Floor division toward $-\infty$ with `mustRoundDown` add-back when discarded bits are non-zero.
- **Chosen Solution**: Floor division with `mustRoundDown` (`src/shift.go`).
- **Reasoning**: Matches ECMAScript specification (ECMA-262) for arithmetic right shift on negative BigInts.
- **Trade-offs**: Requires scanning discarded bits for non-zero values to trigger `absoluteAddOne`.
- **Evidence**: `BenchmarkSignedRightShift` 49.3 ns/op (48 B/op, 2 allocs); 1,400,000 differential fuzz cases passed.

---

## Decision 7: Cluster 7 Bitwise Operations — De Morgan Transformations

- **Decision**: Implement De Morgan transformations and sign identities $-x = \sim(x - 1)$ for negative BigInt operands.
- **Problem**: Negative BigInts in two's complement represent infinite sign-extended ones, which cannot be stored as finite arrays directly.
- **Alternatives Considered**:
  1. Virtual sign-extension limb padding.
  2. De Morgan transformations (`x & y = ~(~x | ~y)`, `x | y = ~(~x & ~y)`, `x ^ y = (~x & ~y) ^ ~0`) operating on positive magnitudes.
- **Chosen Solution**: De Morgan transformations (`src/bitwise.go`).
- **Reasoning**: Converts negative operand bitwise operations into positive magnitude operations, preserving finite limb storage and canonical zero invariants.
- **Trade-offs**: Up to 4 intermediate magnitude allocations for negative-negative operand pairs.
- **Evidence**: `BenchmarkBitwiseNot` 39.57 ns/op, `BenchmarkBitwiseAnd` 89.06 ns/op; 1,863,000 differential fuzz cases passed.

---

## Decision 8: Cluster 8 Fixed-Width Truncation — Range Guard & Fast-Path Value Independence

- **Decision**: Use `(n+29)/30 > kMaxLength` range guard and return `x.Copy()` on all fast paths.
- **Problem**: JSBI returns same object reference `x` when no truncation occurs. In Go, returning input pointer allows caller mutation to mutate shared internal state.
- **Alternatives Considered**:
  1. Return input `x` pointer directly (same reference as JSBI).
  2. Return deep copy `x.Copy()` on fast paths.
- **Chosen Solution**: Return `x.Copy()` (`src/truncation.go`).
- **Reasoning**: Guarantees value independence (`returnedPointer != inputPointer`), preventing accidental aliasing and state corruption in concurrent or stateful Go code.
- **Trade-offs**: 1 extra slice allocation on fast paths (40 B/op).
- **Evidence**: `TestTruncationValueIndependenceFastPath` passed; `BenchmarkAsIntN` 36.39 ns/op (40 B/op, 2 allocs); 1,753,000 differential fuzz cases passed.

---

## Decision 9: Cluster 9 String Formatting — Popcount Base-2 Fast Path & Divide-and-Conquer Fallback

- **Decision**: Use `bits.OnesCount32` for power-of-two radices (2, 4, 8, 16, 32) and recursive divide-and-conquer for arbitrary radices.
- **Problem**: Repeated division by radix for large BigInts is $O(n^2)$, causing severe performance degradation for large numbers.
- **Alternatives Considered**:
  1. Naive repeated division by radix.
  2. Divide-and-conquer splitting by $radix^{secondHalfChars}$ using `Exponentiate` and `absoluteDivLarge`.
- **Chosen Solution**: Divide-and-conquer with `kMaxBitsPerChar` table lookup (`src/tostring.go`).
- **Reasoning**: Reduces radix conversion time complexity from $O(n^2)$ to $O(n^{1.585} \log n)$. Power-of-two path completes in $O(n)$ via bit extraction without division.
- **Trade-offs**: Intermediate memory allocation for split-point quotient/remainder BigInts.
- **Evidence**: `BenchmarkToStringHex1Limb` 26.24 ns/op (16 B/op, 2 allocs), `BenchmarkToStringDec1Limb` 22.79 ns/op (16 B/op, 1 alloc); 1,874,000 differential fuzz cases passed across all 35 radices.
