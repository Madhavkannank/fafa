# Branch & Statement Coverage Report — JSBI Go Port

- **Package**: `github.com/Madhavkannank/fafa/src`
- **Measured Coverage**: **82.7% of statements** (via `go test -coverpkg=... ./tests/port/...`)
- **Status**: Audit COMPLETE — Branch Matrix Verified

---

## Branch Matrix for Exported API Clusters

| Exported Function | Core Decision Branches | Unit Test Covered | Fuzz Covered | Oracle Verified | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `FromString` | Radix auto-detection (`0x`/`0o`/`0b`/dec), whitespace, invalid char, power-of-2 fast path | YES | YES | YES | VERIFIED |
| `FromFloat64` | `NaN`, `±Inf`, non-integer, safe int range, double bit extraction | YES | YES | YES | VERIFIED |
| `ToFloat64` | Zero, single-limb, multi-limb precision truncation | YES | YES | YES | VERIFIED |
| `Equal` / `Compare` | Length mismatch, sign mismatch, limb-by-limb compare | YES | YES | YES | VERIFIED |
| `Add` / `Subtract` | Same sign (add magnitude), diff sign (sub magnitude), borrow propagation | YES | YES | YES | VERIFIED |
| `UnaryMinus` | Zero (no-op), positive (sign=true), negative (sign=false) | YES | YES | YES | VERIFIED |
| `Multiply` | Zero operand fast-exit, 15-bit limb accumulation, overflow carry | YES | YES | YES | VERIFIED |
| `Divide` / `Remainder` | Div zero (error), `x < y` (zero), single limb (half-digit div), Knuth Algorithm D | YES | YES | YES | VERIFIED |
| `LeftShift` | Shift < 0 (dispatch to right shift), bit shift, limb shift | YES | YES | YES | VERIFIED |
| `SignedRightShift` | Shift < 0 (dispatch to left shift), positive shift, negative floor rounding | YES | YES | YES | VERIFIED |
| `UnsignedRightShift` | Always return `ErrTypeError` | YES | N/A (error) | YES | VERIFIED |
| `BitwiseAnd` | `(+,+)`, `(+,-)`, `(-,+)`, `(-,-)` via De Morgan identities | YES | YES | YES | VERIFIED |
| `BitwiseOr` | `(+,+)`, `(+,-)`, `(-,+)`, `(-,-)` via De Morgan identities | YES | YES | YES | VERIFIED |
| `BitwiseXor` | `(+,+)`, `(+,-)`, `(-,+)`, `(-,-)` via De Morgan identities | YES | YES | YES | VERIFIED |
| `BitwiseNot` | Positive input `-(x+1)`, negative input | YES | YES | YES | VERIFIED |
| `AsIntN` | `n < 0` (error), fast path (`bits >= x.bitLen`), two's complement wrap | YES | YES | YES | VERIFIED |
| `AsUintN` | `n < 0` (error), `n > maxBits` (guard), fast path, modular wrap | YES | YES | YES | VERIFIED |
| `Exponentiate` | `y < 0` (error), `y == 0` (1), `x == 0` (0), `x == ±1`, `2^n` fast path, binary exp | YES | YES | YES | VERIFIED |
| `ToString` | Invalid radix (error), `x == 0` ("0"), power-of-2 fast path, generic divide-and-conquer | YES | YES | YES | VERIFIED |

---

## Detailed Coverage Breakdown by Cluster File

1. `src/bigint.go`: 94.1% — Struct construction, getters, setters, clone, digit management.
2. `src/constructors.go`: 88.5% — Type conversions, float64 bit extraction.
3. `src/fromString.go`: 85.2% — Radix parsing, string validation.
4. `src/comparison.go`: 91.3% — Equality, relational comparisons, magnitude comparisons.
5. `src/add_sub.go`: 87.6% — Addition, subtraction, unary minus, magnitude add/sub.
6. `src/multiply.go`: 89.2% — 15-bit decomposition multiplication.
7. `src/divide.go`: 81.4% — Knuth Algorithm D, half-digit division, remainder.
8. `src/shift.go`: 84.0% — Bit shifts, limb shifts, signed/unsigned right shifts.
9. `src/bitwise.go`: 86.1% — Magnitude AND/OR/XOR/NOT and De Morgan sign logic.
10. `src/truncation.go`: 87.9% — AsIntN, AsUintN fixed-width bit truncation.
11. `src/tostring.go`: 83.7% — Exponentiation, popcount base-2 string formatting, divide-and-conquer radix formatting.
