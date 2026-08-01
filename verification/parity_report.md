# Source Parity Audit Report — JSBI Go Port

- **Kickoff Source SHA**: `5382367c7e3199858d36bb620977e1f90605bcb9`
- **Go Package**: `github.com/Madhavkannank/fafa/src`
- **Auditor**: Lead Engineer (Antigravity AI Agent)
- **Status**: Audit COMPLETE — 100% Exported API Parity Confirmed

---

## 1. Exported API Parity Table

| JSBI Function | Go Exported API | Implementation File | Control Flow Match | Edge Case Match | Status | Notes / Intentional Divergences |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `JSBI.BigInt(arg)` | `FromString`, `FromFloat64`, `BigIntVal` | `src/constructors.go`, `src/fromString.go` | EXACT | EXACT | VERIFIED | Go native typed constructors |
| `JSBI.toNumber(x)` | `ToFloat64(x)` | `src/constructors.go` | EXACT | EXACT | VERIFIED | Converts BigInt to float64 IEEE 754 |
| `JSBI.unaryMinus(x)` | `UnaryMinus(x)` | `src/add_sub.go` | EXACT | EXACT | VERIFIED | Negates sign, copy on non-zero |
| `JSBI.bitwiseNot(x)` | `BitwiseNot(x)` | `src/bitwise.go` | EXACT | EXACT | VERIFIED | `-(x + 1)` representation identity |
| `JSBI.exponentiate(x, y)`| `Exponentiate(x, y)` | `src/tostring.go` | EXACT | EXACT | VERIFIED | Binary square-and-multiply, 2^n fast path |
| `JSBI.multiply(x, y)` | `Multiply(x, y)` | `src/multiply.go` | EXACT | EXACT | VERIFIED | 15-bit limb product accumulation |
| `JSBI.divide(x, y)` | `Divide(x, y)` | `src/divide.go` | EXACT | EXACT | VERIFIED | Knuth Algorithm D & half-digit div |
| `JSBI.remainder(x, y)` | `Remainder(x, y)` | `src/divide.go` | EXACT | EXACT | VERIFIED | Sign matches dividend `x` |
| `JSBI.add(x, y)` | `Add(x, y)` | `src/add_sub.go` | EXACT | EXACT | VERIFIED | Absolute add/sub sign dispatch |
| `JSBI.subtract(x, y)` | `Subtract(x, y)` | `src/add_sub.go` | EXACT | EXACT | VERIFIED | Absolute add/sub sign dispatch |
| `JSBI.leftShift(x, y)` | `LeftShift(x, y)` | `src/shift.go` | EXACT | EXACT | VERIFIED | Positive & negative shift dispatch |
| `JSBI.signedRightShift` | `SignedRightShift(x, y)`| `src/shift.go` | EXACT | EXACT | VERIFIED | Floor division rounding for neg |
| `JSBI.unsignedRightShift`| `UnsignedRightShift(x, y)`| `src/shift.go` | EXACT | EXACT | VERIFIED | Always returns `ErrTypeError` |
| `JSBI.lessThan(x, y)` | `LessThan(x, y)` | `src/comparison.go` | EXACT | EXACT | VERIFIED | Compare(x,y) < 0 |
| `JSBI.lessThanOrEqual` | `LessThanOrEqual(x, y)`| `src/comparison.go` | EXACT | EXACT | VERIFIED | Compare(x,y) <= 0 |
| `JSBI.greaterThan(x, y)`| `GreaterThan(x, y)` | `src/comparison.go` | EXACT | EXACT | VERIFIED | Compare(x,y) > 0 |
| `JSBI.greaterThanOrEqual`| `GreaterThanOrEqual(x, y)`| `src/comparison.go` | EXACT | EXACT | VERIFIED | Compare(x,y) >= 0 |
| `JSBI.equal(x, y)` | `Equal(x, y)` | `src/comparison.go` | EXACT | EXACT | VERIFIED | Compare(x,y) == 0 |
| `JSBI.notEqual(x, y)` | `NotEqual(x, y)` | `src/comparison.go` | EXACT | EXACT | VERIFIED | Compare(x,y) != 0 |
| `JSBI.bitwiseAnd(x, y)`| `BitwiseAnd(x, y)` | `src/bitwise.go` | EXACT | EXACT | VERIFIED | De Morgan sign transformation |
| `JSBI.bitwiseOr(x, y)` | `BitwiseOr(x, y)` | `src/bitwise.go` | EXACT | EXACT | VERIFIED | De Morgan sign transformation |
| `JSBI.bitwiseXor(x, y)` | `BitwiseXor(x, y)` | `src/bitwise.go` | EXACT | EXACT | VERIFIED | De Morgan sign transformation |
| `JSBI.asIntN(n, x)` | `AsIntN(n, x)` | `src/truncation.go` | EXACT | EXACT | VERIFIED | Two's complement bit wrap, `x.Copy()` fast path |
| `JSBI.asUintN(n, x)` | `AsUintN(n, x)` | `src/truncation.go` | EXACT | EXACT | VERIFIED | Unsigned modular bit wrap, `x.Copy()` fast path |
| `JSBI.prototype.toString`| `ToString(x, radix)` | `src/tostring.go` | EXACT | EXACT | VERIFIED | Popcount power-of-2 fast path & divide-and-conquer |

---

## 2. Internal Helper Function Parity Audit

| JSBI Helper Primitive | Go Internal Primitive | Status | Notes |
| :--- | :--- | :--- | :--- |
| `__absoluteAdd` | `absoluteAdd` | VERIFIED | `src/add_sub.go` |
| `__absoluteSub` | `absoluteSub` | VERIFIED | `src/add_sub.go` |
| `__absoluteAddOne` | `absoluteAddOne` | VERIFIED | `src/add_sub.go` |
| `__absoluteSubOne` | `absoluteSubOne` | VERIFIED | `src/add_sub.go` |
| `__absoluteAnd` | `absoluteAnd` | VERIFIED | `src/bitwise.go` |
| `__absoluteAndNot` | `absoluteAndNot` | VERIFIED | `src/bitwise.go` |
| `__absoluteOr` | `absoluteOr` | VERIFIED | `src/bitwise.go` |
| `__absoluteXor` | `absoluteXor` | VERIFIED | `src/bitwise.go` |
| `__absoluteCompare` | `absoluteCompare` | VERIFIED | `src/comparison.go` |
| `__absoluteDivSmall` | `absoluteDivSmall` | VERIFIED | `src/divide.go` |
| `__absoluteModSmall` | `absoluteModSmall` | VERIFIED | `src/divide.go` |
| `__absoluteDivLarge` | `absoluteDivLarge` | VERIFIED | `src/divide.go` (Knuth Algorithm D) |
| `__leftShiftByAbsolute` | `leftShiftByAbsolute` | VERIFIED | `src/shift.go` |
| `__rightShiftByAbsolute`| `rightShiftByAbsolute`| VERIFIED | `src/shift.go` |
| `__rightShiftByMaximum` | `rightShiftByMaximum` | VERIFIED | `src/shift.go` |
| `__toShiftAmount` | `toShiftAmount` | VERIFIED | `src/shift.go` |
| `__truncateToNBits` | `truncateToNBits` | VERIFIED | `src/truncation.go` |
| `__truncateAndSubFromPowerOfTwo` | `truncateAndSubFromPowerOfTwo` | VERIFIED | `src/truncation.go` |
| `__toStringBasePowerOfTwo` | `toStringBasePowerOfTwo` | VERIFIED | `src/tostring.go` |
| `__toStringGeneric` | `toStringGeneric` | VERIFIED | `src/tostring.go` |

---

## 3. Logged Intentional Divergences

1. **Fast-Path Memory Independence**: JSBI returns `x` directly (same JS reference) when operations perform no modification (e.g., `AsIntN`, `AsUintN`, `LeftShift(x, 0)`). Go returns `x.Copy()` to prevent caller state mutation aliasing. (Logged in DECISIONS.md #8).
2. **Native Errors**: Go methods return native Go error values (`ErrRange`, `ErrTypeError`) instead of throwing JavaScript exception objects.
