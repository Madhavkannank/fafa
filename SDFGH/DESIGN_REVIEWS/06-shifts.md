# Design Review 06 — Shifts (Left Shift & Right Shift)

- **Cluster**: 6 — Shifts (`LeftShift`, `SignedRightShift`, `UnsignedRightShift`)
- **Status**: Draft / Pending User Review [GATE]
- **Author**: Lead Engineer (Go Porting Team)
- **Reference**: `GoogleChromeLabs/jsbi` (`jsbi/lib/jsbi.ts` lines 299–315, 1719–1824)

---

## 1. Overview & Operational Scope [GO IMPLEMENTATION SPECIFICATION]

Cluster 6 covers arbitrary-precision bitwise shift operations for `BigInt` values:
- `LeftShift(x, y *BigInt) (*BigInt, error)`
- `SignedRightShift(x, y *BigInt) (*BigInt, error)`
- `UnsignedRightShift(x, y *BigInt) (*BigInt, error)`

### Scope & Dependencies [GO IMPLEMENTATION SPECIFICATION]
- **Primitives Used**: 30-bit limb accessor methods (`Digit`, `SetDigit`, `Length`, `Sign`, `SetSign`, `Trim`, `Copy`, `OneDigit`, `Zero`), `absoluteAddOne`.
- **Dependencies**: None outside `jsbi` package primitives.

---

## 2. ECMAScript Specification Mapping & Negative Shift Dispatch [ECMAScript SPECIFICATION]

### 2.1 BigInt::leftShift ( x, y )
Ref: ECMAScript Language Specification (ECMA-262), Section *BigInt::leftShift (x, y)*.
- If $y < 0$, returns $\text{BigInt::signedRightShift}(x, -y)$.
- Else, returns $x \times 2^y$.

### 2.2 BigInt::signedRightShift ( x, y )
Ref: ECMAScript Language Specification (ECMA-262), Section *BigInt::signedRightShift (x, y)*.
- If $y < 0$, returns $\text{BigInt::leftShift}(x, -y)$.
- Else, returns $\lfloor x / 2^y \rfloor$ (floor division toward $-\infty$).

### 2.3 BigInt::unsignedRightShift ( x, y )
Ref: ECMAScript Language Specification (ECMA-262), Section *BigInt::unsignedRightShift (x, y)*.
- Throws a `TypeError` exception: *"BigInts have no unsigned right shift; use >> instead"*.

### 2.4 Negative Shift Dispatch Rules [ECMAScript SPECIFICATION]
When the shift count $y$ is negative ($y < 0$), the direction of shifting is inverted according to ECMAScript semantics:
- `LeftShift(x, negativeY)` $\rightarrow$ delegates to `SignedRightShift(x, abs(y))`
- `SignedRightShift(x, negativeY)` $\rightarrow$ delegates to `LeftShift(x, abs(y))`

---

## 3. JSBI vs Go Error Mapping [GO IMPLEMENTATION SPECIFICATION]

JSBI and Go handle operational errors differently due to runtime type system differences:

| Exceptional Condition | JSBI TS Exception [JSBI SOURCE] | Go Error Return [GO IMPLEMENTATION SPECIFICATION] | Classification |
| :--- | :--- | :--- | :--- |
| Unsigned Right Shift (`>>>`) | Throws `TypeError('BigInts have no unsigned right shift; use >> instead')` | Returns `nil, ErrType` | Go Implementation Decision |
| Shift Amount Exceeds Max (`shift > 1 << 30` or `length > 1`) | Throws `RangeError('BigInt too big')` | Returns `nil, ErrRange` | Go Implementation Decision |

---

## 4. Public API & Signature Specification [GO IMPLEMENTATION SPECIFICATION]

```go
// LeftShift shifts x left by y bits (or right if y is negative).
func LeftShift(x, y *BigInt) (*BigInt, error)

// SignedRightShift shifts x right by y bits with sign preservation (floor division toward -inf).
func SignedRightShift(x, y *BigInt) (*BigInt, error)

// UnsignedRightShift returns ErrType because BigInts do not support unsigned right shift.
func UnsignedRightShift(x, y *BigInt) (*BigInt, error)
```

---

## 5. Shift Amount Sentinel (`toShiftAmount`) [JSBI SOURCE]

### Source Reference (JSBI lines 1819–1824)
```ts
static __toShiftAmount(x: JSBI): number {
  if (x.length > 1) return -1;
  const value = x.__unsignedDigit(0);
  if (value > JSBI.__kMaxLengthBits) return -1;
  return value;
}
```

### Internal Sentinel Behavior [JSBI SOURCE]
- The return value `-1` is an **internal sentinel** indicating an invalid or unrepresentable shift count.
- **Unobservability**: The sentinel `-1` is never exposed directly to callers.
- **Translation**:
  - In `leftShiftByAbsolute`: `-1` is translated to JSBI `RangeError('BigInt too big')` / Go `ErrRange`.
  - In `rightShiftByAbsolute`: `-1` triggers `rightShiftByMaximum(sign)`, yielding `-1n` for negative numbers or `0n` for non-negative numbers.

---

## 6. Helper Prerequisites & Right-Shift Maximum Helper

### 6.1 Prerequisites for `rightShiftByAbsolute` [GO IMPLEMENTATION SPECIFICATION]
- `absoluteAddOne`: Increments magnitude by 1 for negative floor rounding. Required only when a negative arithmetic right shift discards non-zero bits and floor rounding must increase the magnitude by one.
- `Digit`: Reads 30-bit limb value at index $i$.
- `SetDigit`: Writes 30-bit limb value at index $i$.
- `Copy`: Creates independent duplicate for value independence.
- `Trim`: Normalizes leading zero limbs.
- `Length`: Returns total limb count.

### 6.2 `rightShiftByMaximum` Helper [JSBI SOURCE]
JSBI lines 1812–1817:
```ts
static __rightShiftByMaximum(sign: boolean): JSBI {
  if (sign) {
    return JSBI.__oneDigit(1, true);
  }
  return JSBI.__zero();
}
```
- **When Called**: Invoked when the requested shift removes all significant bits or when the shift amount exceeds the representable limit (`toShiftAmount` returns `-1` or `resultLength = length - digitShift <= 0`).
- **Why JSBI Uses It**: Under right shift, when the requested shift removes all significant bits or when the shift amount exceeds the representable limit, positive numbers reduce to zero ($0$) and negative numbers reduce to negative one ($-1$, due to floor division toward $-\infty$).
- **Observable Results**:
  - Non-negative input (`sign == false`): returns the canonical zero value (`0n`).
  - Negative input (`sign == true`): returns the canonical -1 value (`-1n`).
- **Go Behavior Mapping**: `rightShiftByMaximum(sign bool) *BigInt` returns the canonical -1 value (`OneDigit(1, true)`) if `sign` is true, else returns the canonical zero value (`Zero()`).

---

## 7. `leftShiftByAbsolute` & Limb Growth Proof [JSBI SOURCE]

### Source Reference (JSBI lines 1719–1751)
```ts
  static __leftShiftByAbsolute(x: JSBI, y: JSBI): JSBI {
    const shift = JSBI.__toShiftAmount(y);
    if (shift < 0) throw new RangeError('BigInt too big');
    const digitShift = (shift / 30) | 0;
    const bitsShift = shift % 30;
    const length = x.length;
    const grow = bitsShift !== 0 &&
                 (x.__digit(length - 1) >>> (30 - bitsShift)) !== 0;
    const resultLength = length + digitShift + (grow ? 1 : 0);
    const result = new JSBI(resultLength, x.sign);
    if (bitsShift === 0) {
      let i = 0;
      for (; i < digitShift; i++) result.__setDigit(i, 0);
      for (; i < resultLength; i++) {
        result.__setDigit(i, x.__digit(i - digitShift));
      }
    } else {
      let carry = 0;
      for (let i = 0; i < digitShift; i++) result.__setDigit(i, 0);
      for (let i = 0; i < length; i++) {
        const d = x.__digit(i);
        result.__setDigit(
            i + digitShift, ((d << bitsShift) & 0x3FFFFFFF) | carry);
        carry = d >>> (30 - bitsShift);
      }
      if (grow) {
        result.__setDigit(length + digitShift, carry);
      } else {
        if (carry !== 0) throw new Error('implementation bug');
      }
    }
    return result.__trim();
  }
```

### Mathematical Proof of Limb Growth Condition [GO IMPLEMENTATION SPECIFICATION]

**Theorem**: In `leftShiftByAbsolute`, the final carry after shifting the most significant digit $d = x.\text{Digit}(n-1)$ is non-zero ($\text{carry} \neq 0$) if and only if $\text{grow} = \text{true}$.

**Proof**:
1. In a 30-bit radix system, each limb stores integer values in $[0, 2^{30} - 1]$.
2. When shifting a digit $d$ left by $b = \text{bitsShift} \in [0, 29]$ bits, the upper bits shifted out of the 30-bit window are given by logical right shift:
   $$\text{carry}_{\text{final}} = \lfloor d / 2^{30 - b} \rfloor = d \gg (30 - b)$$
3. If $b = 0$, $\text{bitsShift} = 0 \implies \text{carry}_{\text{final}} = 0$, and `grow = false`.
4. If $b > 0$, the most significant digit $d = x.\text{Digit}(n-1)$ has upper bits $d \gg (30 - b)$.
   - $\text{grow}$ is defined as `bitsShift != 0 && (x.Digit(length - 1) >> (30 - bitsShift)) != 0`.
   - Therefore, $\text{grow} = \text{true} \iff d \gg (30 - b) \neq 0 \iff \text{carry}_{\text{final}} \neq 0$.
5. Thus, allocating `resultLength = length + digitShift + (grow ? 1 : 0)` guarantees exact limb capacity without dynamic reallocation. $\blacksquare$

---

## 8. Mathematical Explanation of `mustRoundDown` [ECMAScript SPECIFICATION]

### Floor Division Semantics vs. Truncation
Arithmetic right shift on signed integers corresponds to floor division toward $-\infty$:
$$\text{SignedRightShift}(x, k) = \left\lfloor \frac{x}{2^k} \right\rfloor$$

#### Examples
1. $-5 \gg 1$:
   $$\left\lfloor \frac{-5}{2} \right\rfloor = \lfloor -2.5 \rfloor = -3$$
   - Magnitude representation: $|-5| = 5$. Truncating right shift on magnitude $5 \gg 1 = 2$, yielding magnitude 2 (which as a negative number would be $-2$).
   - Because bits were lost ($5 \pmod 2 = 1 \neq 0$), floor division requires rounding down to $-3$.
   - JSBI achieves $-3$ by performing `absoluteAddOne` on magnitude 2: $2 + 1 = 3 \implies -3$.

2. $-7 \gg 2$:
   $$\left\lfloor \frac{-7}{4} \right\rfloor = \lfloor -1.75 \rfloor = -2$$
   - Magnitude $|-7| = 7$. Truncating right shift $7 \gg 2 = 1$.
   - Lost bits: $7 \pmod 4 = 3 \neq 0$, so `mustRoundDown = true`.
   - Magnitude increment: $1 + 1 = 2 \implies -2$.

---

## 9. Worked Execution Examples

Let $\text{digitShift} = \lfloor \text{shift} / 30 \rfloor$ and $\text{bitsShift} = \text{shift} \pmod{30}$.

### Example 1: $13 \ll 3$ ($x = 13$, shift $= 3$)
- **Operand $x$**: magnitude `[13]`, `Length() = 1`, `Sign() = false`.
- **Shift Parameters**: `shift = 3`, `digitShift = 0`, `bitsShift = 3`.
- **Growth Check**: `grow = (13 >> 27) != 0` $\implies$ `false`.
- **Allocation**: `resultLength = 1 + 0 + 0 = 1`.
- **Shift Loop**:
  - `i = 0`: $d = 13$. `(13 << 3) & 0x3FFFFFFF = 104`. `carry = 13 >> 27 = 0`.
- **Final Result**: `Digits: [104]` ($13 \times 8 = 104$).

### Example 2: $1 \ll 35$ ($x = 1$, shift $= 35$)
- **Operand $x$**: magnitude `[1]`, `Length() = 1`, `Sign() = false`.
- **Shift Parameters**: `shift = 35`, `digitShift = 1`, `bitsShift = 5`.
- **Growth Check**: `grow = (1 >> 25) != 0` $\implies$ `false`.
- **Allocation**: `resultLength = 1 + 1 + 0 = 2`.
- **Shift Loop**:
  - `i = 0`: `result.SetDigit(0, 0)`.
  - `i = 0` (limb shift): $d = 1$. `result.SetDigit(1, (1 << 5) & 0x3FFFFFFF = 32)`. `carry = 1 >> 25 = 0`.
- **Final Result**: `Digits: [0, 32]` ($1 \times 2^{35} = 34,359,738,368$).

### Example 3: $13 \gg 2$ ($x = 13$, shift $= 2$)
- **Operand $x$**: magnitude `[13]`, `Length() = 1`, `Sign() = false`.
- **Shift Parameters**: `shift = 2`, `digitShift = 0`, `bitsShift = 2`.
- **Allocation**: `resultLength = 1 - 0 = 1`.
- **Rounding Check**: `Sign() == false` $\implies$ `mustRoundDown = false`.
- **Shift Loop**:
  - `last = 0`: `carry = 13 >> 2 = 3`. `result.SetDigit(0, 3)`.
- **Final Result**: `Digits: [3]` ($\lfloor 13 / 4 \rfloor = 3$).

### Example 4: $-5 \gg 1$ ($x = -5$, shift $= 1$)
- **Operand $x$**: magnitude `[5]`, `Length() = 1`, `Sign() = true`.
- **Shift Parameters**: `shift = 1`, `digitShift = 0`, `bitsShift = 1`.
- **Rounding Check**: `Sign() == true`. Mask `(1 << 1) - 1 = 1`. `x.Digit(0) & 1 = 5 & 1 = 1 != 0` $\implies$ `mustRoundDown = true`.
- **Allocation**: `resultLength = 1 - 0 = 1`.
- **Shift Loop**:
  - `carry = 5 >> 1 = 2`. `result.SetDigit(0, 2)`.
- **Add-One Rounding**: `mustRoundDown == true` $\implies$ `absoluteAddOne([2]) = [3]`.
- **Final Result**: `Digits: [3], Sign: true` ($\lfloor -5 / 2 \rfloor = -3$).

---

## 10. Complexity Analysis [ENGINEERING GOAL]

Let $n = x.\text{Length()}$ be the limb count of operand $x$, and $\text{resultLength}$ be the resulting limb count.

| Function | Time Complexity | Space Complexity | Notes |
| :--- | :--- | :--- | :--- |
| `toShiftAmount` | $O(1)$ | $O(1)$ | Checks single 30-bit digit. |
| `rightShiftByMaximum` | $O(1)$ | $O(1)$ | Returns the canonical zero or canonical -1 value. |
| `leftShiftByAbsolute` | $O(\text{resultLength})$ | $O(\text{resultLength})$ | Allocates `resultLength = n + digitShift + (grow ? 1 : 0)`. |
| `rightShiftByAbsolute` | $O(n)$ | $O(\text{resultLength})$ | Scans input limbs $n$; result occupies $\text{resultLength}$ limbs. |
| `LeftShift` | $O(\text{resultLength})$ | $O(\text{resultLength})$ | $y \ge 0 \implies O(\text{resultLength})$ via left shift; $y < 0 \implies O(n)$ via `SignedRightShift` dispatch. |
| `SignedRightShift` | $O(n)$ | $O(\text{resultLength})$ | $y \ge 0 \implies O(n)$ via right shift; $y < 0 \implies O(\text{resultLength})$ via `LeftShift` dispatch. |
| `UnsignedRightShift` | $O(1)$ | $O(1)$ | Immediately returns `ErrType`. |

---

## 11. Benchmark Protocol Specification [GO IMPLEMENTATION SPECIFICATION]

Performance will be measured after implementation using `BenchmarkLeftShift`, `BenchmarkSignedRightShift`, and `BenchmarkUnsignedRightShift`:

```go
func BenchmarkLeftShift(b *testing.B)
func BenchmarkSignedRightShift(b *testing.B)
func BenchmarkUnsignedRightShift(b *testing.B)
```

---

## 12. Recommended Implementation Order [ENGINEERING GOAL]

1. `toShiftAmount` helper
2. `rightShiftByMaximum` helper
3. `leftShiftByAbsolute` implementation
4. `rightShiftByAbsolute` implementation
5. `LeftShift`, `SignedRightShift`, `UnsignedRightShift` public API wiring
6. Unit test suite in `tests/port/shift_test.go`
7. Differential fuzz driver in `fuzz/harness/fuzz_cluster6.go` & `oracle_cluster6.mjs`
