# Design Review 05: Divide & Remainder

- **Status**: UNDER REVISION — Awaiting User Approval (GATE)
- **Cluster**: 5 — Divide / Remainder
- **References**: `jsbi/lib/jsbi.ts` lines 236–267, 1509–1533, 1540–1605, 1607–1609, 1612–1696, 1698–1717, 1922–1936.

---

## 1. Objectives

Cluster 5 implements arbitrary-precision integer division and remainder operations for `BigInt` instances. The implementation must faithfully port JSBI's two-path strategy: a fast half-digit long-division algorithm for small divisors (≤ 0x7FFF), and Knuth Algorithm D over 15-bit half-digits for large divisors.

---

## 2. ECMAScript BigInt Division Semantics [ECMAScript SPECIFICATION]

BigInt division in ECMAScript truncates toward zero. This is not floor division.

**Source**: ECMAScript 2023 specification §6.1.6.2 BigInt::divide — let quotient be the mathematical value x/y. Return BigInt(truncate(quotient)), where `truncate` removes the fractional part toward zero.

### Truncation Examples [ECMAScript SPECIFICATION]

| Expression | Result | Sign Rule |
| :--- | :--- | :--- |
| `7n / 3n` | `2n` | positive ÷ positive → positive |
| `-7n / 3n` | `-2n` | negative ÷ positive → negative |
| `7n / -3n` | `-2n` | positive ÷ negative → negative |
| `-7n / -3n` | `2n` | negative ÷ negative → positive |

Floor division would give `-7n / 3n = -3n`. The implementation must truncate, not floor.

### Sign Rules [ECMAScript SPECIFICATION]

- `sign(quotient) = sign(dividend) XOR sign(divisor)`, with the exception that if the mathematical quotient is exactly zero, the result is `0n` (canonical positive zero).
- `sign(remainder) = sign(dividend)` always. The divisor sign has no effect on the remainder sign.

Verified in JSBI source:
- `jsbi.ts` line 239: `const resultSign = x.sign !== y.sign;` (XOR for quotient) [JSBI SOURCE]
- `jsbi.ts` line 265: `remainder.sign = x.sign;` (dividend sign for remainder) [JSBI SOURCE]

---

## 3. Division Correctness Invariants [ECMAScript SPECIFICATION + JSBI SOURCE]

For every successful integer division `Q, R = divmod(X, Y)`:

1. **Division Identity**: `X = Q × Y + R`
2. **Remainder Bound**: `|R| < |Y|`
3. **Quotient Sign**: `sign(Q) = sign(X) XOR sign(Y)` (if Q == 0, sign = false)
4. **Remainder Sign**: `sign(R) = sign(X)` always
5. **Canonical Zero Quotient**: If `Q == 0`, then `len(Q) == 0` and `sign(Q) == false`
6. **Canonical Zero Remainder**: If `R == 0`, then `len(R) == 0` and `sign(R) == false`

All six invariants will be checked element-by-element during differential fuzzing. Correctness claims not yet verified by fuzzing are labelled [INFERENCE] or [ENGINEERING GOAL].

---

## 4. Division-by-Zero Mapping [JSBI SOURCE → GO IMPLEMENTATION SPECIFICATION]

### JSBI Behaviour [JSBI SOURCE]

- `jsbi.ts` line 237: `if (y.length === 0) throw new RangeError('Division by zero');`
- `jsbi.ts` line 255: `if (y.length === 0) throw new RangeError('Division by zero');`

JSBI throws a `RangeError` exception when the divisor `BigInt` has zero length.

### Go Implementation Mapping [GO IMPLEMENTATION SPECIFICATION]

- `Divide(x, y)` and `Remainder(x, y)` return `(nil, ErrRange)` when `y.Length() == 0`.
- `DivRem(x, y)` returns `(nil, nil, ErrRange)` when `y.Length() == 0`.
- No partial result is returned. No panic occurs.
- The sentinel error is the existing exported `ErrRange` defined in `src/errors.go`.

### Fuzzing Requirement [GO IMPLEMENTATION SPECIFICATION]

For every `Y = 0` case, the fuzzer verifies that both JSBI and Go indicate error. Observable behaviour must be identical except for language error representation (JSBI throws; Go returns error value).

---

## 5. JSBI Source — Public Entrypoints [JSBI SOURCE]

**File**: `jsbi/lib/jsbi.ts` | **Lines 236–267**

```ts
236:   static divide(x: JSBI, y: JSBI): JSBI {
237:     if (y.length === 0) throw new RangeError('Division by zero');
238:     if (JSBI.__absoluteCompare(x, y) < 0) return JSBI.__zero();
239:     const resultSign = x.sign !== y.sign;
240:     const divisor = y.__unsignedDigit(0);
241:     let quotient;
242:     if (y.length === 1 && divisor <= 0x7FFF) {
243:       if (divisor === 1) {
244:         return resultSign === x.sign ? x : JSBI.unaryMinus(x);
245:       }
246:       quotient = JSBI.__absoluteDivSmall(x, divisor, null);
247:     } else {
248:       quotient = JSBI.__absoluteDivLarge(x, y, true, false);
249:     }
250:     quotient.sign = resultSign;
251:     return quotient.__trim();
252:   }
253:
254:   static remainder(x: JSBI, y: JSBI): JSBI {
255:     if (y.length === 0) throw new RangeError('Division by zero');
256:     if (JSBI.__absoluteCompare(x, y) < 0) return x;
257:     const divisor = y.__unsignedDigit(0);
258:     if (y.length === 1 && divisor <= 0x7FFF) {
259:       if (divisor === 1) return JSBI.__zero();
260:       const remainderDigit = JSBI.__absoluteModSmall(x, divisor);
261:       if (remainderDigit === 0) return JSBI.__zero();
262:       return JSBI.__oneDigit(remainderDigit, x.sign);
263:     }
264:     const remainder = JSBI.__absoluteDivLarge(x, y, false, true);
265:     remainder.sign = x.sign;
266:     return remainder.__trim();
267:   }
```

### Special Cases Observed Directly From Source [JSBI SOURCE]

**`divide` — line 238**: If `|X| < |Y|`, returns `__zero()` immediately without entering any division algorithm.

**`divide` — line 242–244**: If divisor is 1 (small path), JSBI returns `x` unchanged if the sign already matches `resultSign`, or `unaryMinus(x)` if the sign is wrong. When the sign matches, this is a reference return to the input — no copy is made. The Go implementation must copy to preserve value semantics.

**`remainder` — line 256**: If `|X| < |Y|`, `remainder` returns `x` directly. This is a reference to the input. The Go implementation must return a copy.

**`remainder` — line 259**: If divisor is 1, returns `__zero()`. Remainder of any integer divided by 1 is 0.

---

## 6. Small-Divisor Path Activation Condition [JSBI SOURCE]

**File**: `jsbi/lib/jsbi.ts` | **Line 242**:
```ts
if (y.length === 1 && divisor <= 0x7FFF)
```

**What JSBI guarantees** [JSBI SOURCE]: The small path activates when the divisor has exactly one 30-bit limb **and** the value of that limb is at most 0x7FFF (decimal 32767, 15 bits). Divisors with one limb but a value above 0x7FFF — up to 0x3FFFFFFF (the maximum 30-bit value) — use Algorithm D instead. The JSBI source does not document the reason this specific threshold was chosen.

**Possible implementation rationale** [INFERENCE — not documented in JSBI source]:
The small-divisor algorithm forms each division input as `(remainder << 15) | halfDigit(i)`, a 30-bit quantity. For each quotient half-digit `input / divisor` to fit within 15 bits, the divisor must satisfy `floor((2^30 - 1) / divisor) < 2^15`. This is satisfied exactly when `divisor <= 0x7FFF`. This is a plausible but unverified reason; the actual rationale may differ.

---

## 7. Half-Digit Representation [JSBI SOURCE]

JSBI splits each 30-bit digit limb into two 15-bit half-digits to enable Algorithm D and the small-divisor path.

### Accessor Definitions [JSBI SOURCE]

**File**: `jsbi/lib/jsbi.ts` | **Lines 1922–1936**

```ts
1922:   __halfDigitLength(): number {
1923:     const len = this.length;
1924:     if (this.__unsignedDigit(len - 1) <= 0x7FFF) return len * 2 - 1;
1925:     return len*2;
1926:   }
1927:   __halfDigit(i: number): number {
1928:     return (this[i >>> 1] >>> ((i & 1) * 15)) & 0x7FFF;
1929:   }
1930:   __setHalfDigit(i: number, value: number): void {
1931:     const digitIndex = i >>> 1;
1932:     const previous = this.__digit(digitIndex);
1933:     const updated = (i & 1) ? (previous & 0x7FFF) | (value << 15) :
1934:                             (previous & 0x3FFF8000) | (value & 0x7FFF);
1935:     this.__setDigit(digitIndex, updated);
1936:   }
```

### Derived Formulas [JSBI SOURCE — directly derived from lines 1927–1929]

For half-digit index `i` (0 = least significant):
- **Digit (limb) index**: `i >> 1`
- **Half position**: `i & 1` → 0 = low 15 bits; 1 = high 15 bits
- **Extraction**: `digit[i>>1] >>> ((i & 1) * 15) & 0x7FFF`
- **Setting low half** (`i & 1 == 0`): `(digit & 0x3FFF8000) | (value & 0x7FFF)`
- **Setting high half** (`i & 1 == 1`): `(digit & 0x7FFF) | (value << 15)`

### `__halfDigitLength` Invariant [JSBI SOURCE — lines 1922–1925]

From line 1924: if the most-significant digit ≤ 0x7FFF, the top half is zero, so effective half-digit count is `len*2 - 1`. Otherwise it is `len*2`. This definition excludes leading zero half-digits at the most-significant position.

---

## 8. Complexity Notation

Throughout this document:
- **H_X** = half-digit count of operand X = `X.halfDigitLength()`
- **L_X** = limb count of operand X = `X.Length()`
- **H_Y** = half-digit count of operand Y
- Relationship: `L_X * 2 - 1 <= H_X <= L_X * 2`

---

## 9. Small-Divisor Path: `absoluteDivSmall` [JSBI SOURCE]

### Source [JSBI SOURCE]

**File**: `jsbi/lib/jsbi.ts` | **Lines 1509–1523**

```ts
1509:   static __absoluteDivSmall(x: JSBI, divisor: number,
1510:       quotient: JSBI|null = null): JSBI {
1511:     if (quotient === null) quotient = new JSBI(x.length, false);
1512:     let remainder = 0;
1513:     for (let i = x.length * 2 - 1; i >= 0; i -= 2) {
1514:       let input = ((remainder << 15) | x.__halfDigit(i)) >>> 0;
1515:       const upperHalf = (input / divisor) | 0;
1516:       remainder = (input % divisor) | 0;
1517:       input = ((remainder << 15) | x.__halfDigit(i - 1)) >>> 0;
1518:       const lowerHalf = (input / divisor) | 0;
1519:       remainder = (input % divisor) | 0;
1520:       quotient.__setDigit(i >>> 1, (upperHalf << 15) | lowerHalf);
1521:     }
1522:     return quotient;
1523:   }
```

**Observation** [JSBI SOURCE]: The loop at line 1513 starts from `x.length * 2 - 1` (not `halfDigitLength() - 1`). This means it always iterates over `x.length` full digit pairs, including any leading zero half-digits. The quotient buffer has the same limb count as `x`. The caller (`divide`, line 250–251) applies the sign and trims.

### Algorithm Description [JSBI SOURCE]

Processes pairs of half-digits from most significant to least significant (`i -= 2`). Each pair processes the high half (index `i`) then the low half (index `i-1`) of one full digit:

1. **High half** (index `i`):
   - `input = (remainder << 15) | halfDigit(i)`
   - `upperHalf = floor(input / divisor)`
   - `remainder = input mod divisor`

2. **Low half** (index `i-1`):
   - `input = (remainder << 15) | halfDigit(i-1)`
   - `lowerHalf = floor(input / divisor)`
   - `remainder = input mod divisor`

3. **Pack digit**: `setDigit(i >> 1, (upperHalf << 15) | lowerHalf)`

### Mathematical Recurrence [DERIVED FROM JSBI SOURCE — lines 1514–1519]

At each step: `new_input = (current_remainder << 15) | next_half_digit`

This is long division in base 2^15. The remainder from each half-digit step serves as the "carry" into the next, scaled by 2^15 before the next half-digit is appended. No approximation or normalization is required because each quotient digit is computed exactly via integer division.

### Why Normalization is Unnecessary [INFERENCE]

Algorithm D requires normalization to guarantee the trial quotient estimate errs by at most 2. The small-divisor algorithm computes the exact quotient digit via `(input / divisor) | 0` with no estimation. There is no approximation step and therefore no normalization requirement.

### Complexity [GO IMPLEMENTATION SPECIFICATION]

Time: O(L_X). Auxiliary space: O(L_X) for the quotient buffer (when `quotient == nil`); O(1) otherwise.

---

## 10. Small-Divisor Path: `absoluteModSmall` [JSBI SOURCE]

### Source [JSBI SOURCE]

**File**: `jsbi/lib/jsbi.ts` | **Lines 1525–1532**

```ts
1525:   static __absoluteModSmall(x: JSBI, divisor: number): number {
1526:     let remainder = 0;
1527:     for (let i = x.length * 2 - 1; i >= 0; i--) {
1528:       const input = ((remainder << 15) | x.__halfDigit(i)) >>> 0;
1529:       remainder = (input % divisor) | 0;
1530:     }
1531:     return remainder;
1532:   }
```

### Algorithm Description [JSBI SOURCE]

Processes every half-digit from most significant to least significant (step 1, not 2). At each step: `remainder = ((remainder << 15) | halfDigit(i)) mod divisor`. Returns the final `remainder` as a plain number.

Same mathematical recurrence as `absoluteDivSmall` — long division base 2^15 — but the quotient digits are discarded.

### Output [JSBI SOURCE]

Returns a JavaScript number in range `[0, divisor-1]`. Since divisor ≤ 0x7FFF, the return value fits in 15 bits. The Go return type `uint32` can hold all values in this range.

### Complexity [GO IMPLEMENTATION SPECIFICATION]

Time: O(L_X). Auxiliary space: O(1).

---

## 11. Helper Function Contracts [GO IMPLEMENTATION SPECIFICATION]

### 11.1 `absoluteDivSmall`

**Prerequisites**: `halfDigit`, `setDigit` accessors.

| Property | Specification |
| :--- | :--- |
| **Purpose** | Divide `*BigInt x` by a small divisor ≤ 0x7FFF, returning the absolute-value quotient |
| **Input assumptions** | `x.Length() >= 1`; `2 <= divisor <= 0x7FFF` (divisor == 1 is handled by the caller) |
| **Output guarantee** | Returns an independent quotient value with same limb count as `x`, before trimming; caller must call `Trim()` and set sign |
| **Operand modification** | `x` is not modified. If `quotient` parameter is non-nil, it is mutated in-place |
| **Independence** | If `quotient == nil`, returns a newly created `*BigInt` that does not alias `x` or any other input |
| **Allocation expectation** | Expected: zero allocations when `quotient != nil`; expected: a new value creation when `quotient == nil`. Actual behaviour depends on compiler escape analysis [ENGINEERING GOAL] |
| **Worst-case complexity** | O(L_X) |
| **Applicable invariants** | Result is unsigned (sign = false); `Trim()` may remove leading zero limbs |

### 11.2 `absoluteModSmall`

**Prerequisites**: `halfDigit` accessor.

| Property | Specification |
| :--- | :--- |
| **Purpose** | Compute `|x| mod divisor` for small divisor ≤ 0x7FFF |
| **Input assumptions** | `x.Length() >= 1`; `2 <= divisor <= 0x7FFF` |
| **Output guarantee** | Returns exact remainder as `uint32` in range `[0, divisor-1]` |
| **Operand modification** | `x` is not modified |
| **Independence** | Returns a plain `uint32` scalar; no aliasing concern |
| **Allocation expectation** | Expected: zero allocations [ENGINEERING GOAL] |
| **Worst-case complexity** | O(L_X) |
| **Applicable invariants** | Result < divisor always |

### 11.3 `absoluteDivLarge`

**Prerequisites**: `halfDigit`, `setHalfDigit`, `halfDigitLength` accessors; `specialLeftShift`; `internalMultiplyAdd`; `inplaceSub`; `inplaceAdd`; `inplaceRightShift`.

| Property | Specification |
| :--- | :--- |
| **Purpose** | Knuth Algorithm D division for divisors > 0x7FFF or with multiple limbs |
| **Input assumptions** | `dividend.Length() >= 1`; `divisor.Length() >= 1`; `|dividend| >= |divisor|` (guaranteed by caller); at least one of `wantQuotient`, `wantRemainder` is true |
| **Output guarantee** | Returns `(quotient, remainder)` where either may be nil if not requested; neither is trimmed; caller must call `Trim()` on each and set sign |
| **Operand modification** | Original operands are not modified; internal normalized copies are created independently |
| **Independence** | Returned values do not alias `dividend`, `divisor`, or any internal temporary |
| **Allocation expectation** | Expected: one new value per requested result plus internal temporaries (`u`, `qhatv`). Actual behaviour depends on escape analysis [ENGINEERING GOAL] |
| **Worst-case complexity** | O(H_X × H_Y) |
| **Applicable invariants** | Quotient is unsigned; remainder is unnormalized inside algorithm, then right-shifted before return |

### 11.4 `specialLeftShift` [JSBI SOURCE]

**File**: `jsbi/lib/jsbi.ts` | **Lines 1698–1717**

**Prerequisites**: `digit`, `setDigit` accessors.

| Property | Specification |
| :--- | :--- |
| **Purpose** | Shift `x` left by `shift` bits within 30-bit limbs; optionally append a zero limb at MSD |
| **Input assumptions** | `0 <= shift <= 29`; `addDigit` is 0 or 1 |
| **Output guarantee** | Returns a newly created independent `*BigInt` of length `x.Length() + addDigit` |
| **Operand modification** | `x` is not modified |
| **Independence** | Returned value does not alias `x` |
| **Complexity** | O(L_X) |

### 11.5 `internalMultiplyAdd` [JSBI SOURCE]

Used in Algorithm D step D4 (line 1579: `JSBI.__internalMultiplyAdd(divisor, qhat, 0, n2, qhatv)`).

**Prerequisites**: `digit`, `setDigit` accessors.

| Property | Specification |
| :--- | :--- |
| **Purpose** | Compute `qhatv = divisor * qhat + addend`, storing result in `qhatv` |
| **Operand modification** | `qhatv` is overwritten; `divisor` is not modified |
| **Independence** | `qhatv` must not alias `divisor` |
| **Complexity** | O(L_divisor) |

### 11.6 `inplaceSub` and `inplaceAdd` [JSBI SOURCE]

Used in Algorithm D step D4 (lines 1580–1583).

**Prerequisites**: `halfDigit`, `setHalfDigit` accessors; `digit`, `setDigit` accessors (used internally by the unrolled full-digit steps in `inplaceSub`).

| Property | Specification |
| :--- | :--- |
| **Purpose** | `inplaceSub`: subtract `subtrahend` from `this` starting at half-digit offset `startIndex` over `halfDigits` half-digits; return borrow. `inplaceAdd`: add `summand` to `this` at half-digit offset; return carry |
| **Operand modification** | `this` is mutated in-place; `subtrahend`/`summand` are not modified |
| **Independence** | `this` must not alias `subtrahend`/`summand` |
| **Return value** | Borrow or carry (0 or 1) from the final half-digit step |
| **Complexity** | O(halfDigits) |

### 11.7 `inplaceRightShift` [JSBI SOURCE]

Used in Algorithm D unnormalization step (line 1595).

**Prerequisites**: `digit`, `setDigit` accessors.

| Property | Specification |
| :--- | :--- |
| **Purpose** | Shift `this` right by `shift` bits within 30-bit limbs, in-place |
| **Operand modification** | `this` is mutated in-place |
| **Complexity** | O(L_this) |

---

## 12. Helper Dependency Diagram [GO IMPLEMENTATION SPECIFICATION]

```
Divide(x, y)
├── absoluteCompare(x, y)          — magnitude check; returns zero immediately if |x| < |y|
├── absoluteDivSmall(x, divisor)   — small path: y.Length()==1 && divisor <= 0x7FFF
│   └── (no sub-helpers; uses halfDigit, setDigit accessors)
└── absoluteDivLarge(x, y, q, r)  — large path: all other cases
    ├── specialLeftShift(divisor, shift, 0)   — D1: normalize divisor copy
    ├── specialLeftShift(dividend, shift, 1)  — D1: normalize dividend copy
    ├── internalMultiplyAdd(divisor, qhat, …) — D4: compute qhatv = qhat * V
    ├── inplaceSub(qhatv, j, n+1)             — D4: U -= qhatv at offset j
    ├── inplaceAdd(divisor, j, n)             — D4: add-back if borrow
    └── inplaceRightShift(shift)              — D6: unnormalize remainder

Remainder(x, y)
├── absoluteCompare(x, y)
├── absoluteModSmall(x, divisor)             — small path
└── absoluteDivLarge(x, y, false, true)      — large path

DivRem(x, y)
├── absoluteCompare(x, y)
├── [small path]: absoluteDivSmall + absoluteModSmall (separate calls)
└── [large path]: absoluteDivLarge(x, y, true, true)  (single call)
```

---

## 13. Value Independence [GO IMPLEMENTATION SPECIFICATION]

1. **Returned values are independent of inputs**: Every `*BigInt` returned by `Divide`, `Remainder`, `DivRem`, `absoluteDivSmall`, and `absoluteDivLarge` does not alias `x`, `y`, or any internal temporary. The caller may read and modify returned values without affecting the operands.

2. **Operands are never modified**: `x` and `y` are read-only within any public function. Internal normalization creates independent copies via `specialLeftShift`; it does not mutate the caller's values.

3. **Internal temporaries are not exposed**: `u` (normalized dividend), `qhatv` (trial product), and the normalized divisor copy are internal to `absoluteDivLarge`. No reference to them is returned to the caller.

4. **No aliasing between results**: When `DivRem` returns both quotient and remainder, neither aliases the other.

5. **JSBI reference-return exceptions** [JSBI SOURCE — lines 244 and 256]: JSBI returns direct references to input objects in two edge cases (`divide` with sign-matched divisor == 1; `remainder` when `|x| < |y|`). The Go port must never do this — it must return independent values in all paths.

---

## 14. Large-Divisor Path: Knuth Algorithm D [JSBI SOURCE + EXTERNAL THEOREM]

### 14.1 Source [JSBI SOURCE]

**File**: `jsbi/lib/jsbi.ts` | **Lines 1540–1601**

```ts
1543:     const n = divisor.__halfDigitLength();
1544:     const n2 = divisor.length;
1545:     const m = dividend.__halfDigitLength() - n;
1546:     let q = null;
1547:     if (wantQuotient) {
1548:       q = new JSBI((m + 2) >>> 1, false);
1549:       q.__initializeDigits();
1550:     }
1551:     const qhatv = new JSBI((n + 2) >>> 1, false);
1552:     qhatv.__initializeDigits();
1553:     // D1.
1554:     const shift = JSBI.__clz15(divisor.__halfDigit(n - 1));
1555:     if (shift > 0) {
1556:       divisor = JSBI.__specialLeftShift(divisor, shift, 0);
1557:     }
1558:     const u = JSBI.__specialLeftShift(dividend, shift, 1);
1559:     // D2.
1560:     const vn1 = divisor.__halfDigit(n - 1);
1561:     let halfDigitBuffer = 0;
1562:     for (let j = m; j >= 0; j--) {
1563:       // D3.
1564:       let qhat = 0x7FFF;
1565:       const ujn = u.__halfDigit(j + n);
1566:       if (ujn !== vn1) {
1567:         const input = ((ujn << 15) | u.__halfDigit(j + n - 1)) >>> 0;
1568:         qhat = (input / vn1) | 0;
1569:         let rhat = (input % vn1) | 0;
1570:         const vn2 = divisor.__halfDigit(n - 2);
1571:         const ujn2 = u.__halfDigit(j + n - 2);
1572:         while ((JSBI.__imul(qhat, vn2) >>> 0) > (((rhat << 16) | ujn2) >>> 0)) {
1573:           qhat--;
1574:           rhat += vn1;
1575:           if (rhat > 0x7FFF) break;
1576:         }
1577:       }
1578:       // D4.
1579:       JSBI.__internalMultiplyAdd(divisor, qhat, 0, n2, qhatv);
1580:       let c = u.__inplaceSub(qhatv, j, n + 1);
1581:       if (c !== 0) {
1582:         c = u.__inplaceAdd(divisor, j, n);
1583:         u.__setHalfDigit(j + n, (u.__halfDigit(j + n) + c) & 0x7FFF);
1584:         qhat--;
1585:       }
1586:       if (wantQuotient) {
1587:         if (j & 1) {
1588:           halfDigitBuffer = qhat << 15;
1589:         } else {
1590:           (q as JSBI).__setDigit(j >>> 1, halfDigitBuffer | qhat);
1591:         }
1592:       }
1593:     }
1594:     if (wantRemainder) {
1595:       u.__inplaceRightShift(shift);
1596:       if (wantQuotient) {
1597:         return {quotient: (q as JSBI), remainder: u};
1598:       }
1599:       return u;
1600:     }
1601:     if (wantQuotient) return (q as JSBI);
```

### 14.2 Normalization Invariant [JSBI SOURCE]

**`__clz15`** [JSBI SOURCE] — Lines 1607–1609:
```ts
static __clz15(value: number): number {
    return JSBI.__clz30(value) - 15;
}
```
Counts leading zeros in a 15-bit field. Result is in range `[0, 14]`.

**Normalization procedure** [JSBI SOURCE — lines 1554–1558]:
```
shift = clz15(divisor.halfDigit(n - 1))   // 0..14
if shift > 0: divisor = specialLeftShift(divisor, shift, 0)
u = specialLeftShift(dividend, shift, 1)   // always creates u, even if shift == 0
```

**Post-normalization invariant** [DERIVED FROM JSBI SOURCE — clz15 definition]:
After normalization, `divisor.halfDigit(n-1) >= 0x4000`.

Proof: `clz15(v) = 0` means bit 14 of `v` is set, i.e., `v >= 0x4000`. The shift moves the top set bit of `halfDigit(n-1)` into position 14. Therefore after the shift, `halfDigit(n-1) >= 0x4000`.

**Why This Invariant Is Required** [EXTERNAL THEOREM — see Section 14.3]: Knuth's theorem requires the leading half-digit of the normalized divisor to be ≥ b/2 (where b = 2^15 = 0x8000), i.e., ≥ 0x4000. This is exactly the post-normalization invariant above.

**Normalization preserves quotient correctness** [INFERENCE — from elementary divisibility]:
Shifting both dividend and divisor by the same factor 2^shift multiplies both by 2^shift. Since division is: `floor(X / Y) = floor((X * 2^shift) / (Y * 2^shift))`, the integer quotient is unchanged. The remainder must be right-shifted by `shift` to recover the original remainder.

### 14.3 Trial Quotient Bound [EXTERNAL THEOREM — Knuth TAOCP Vol. 2 §4.3.1]

**This bound is not proven within this document. It is an imported correctness result.**

**Statement of External Theorem** [EXTERNAL THEOREM]: Given a base `b = 2^15` and a normalized divisor with leading half-digit `v_{n-1} >= b/2 = 0x4000`, the trial quotient estimate:

```
qhat = min(b - 1,  floor((u_{j+n} * b + u_{j+n-1}) / v_{n-1}))
```

satisfies: `q_j <= qhat <= q_j + 2` where `q_j` is the exact j-th quotient half-digit.

**Source**: Knuth, Donald E. *The Art of Computer Programming*, Volume 2: Seminumerical Algorithms, 3rd Edition, Section 4.3.1, Theorem B.

**Dependency chain for the Go implementation** [DERIVED FROM EXTERNAL THEOREM]:
1. Normalization step (Section 14.2) establishes `v_{n-1} >= 0x4000`, satisfying the theorem's precondition.
2. The D3 refinement loop (line 1572–1576) can reduce `qhat` by at most 2 because the theorem bounds the overestimate.
3. After D3, `qhat <= q_j + 1` (Knuth shows D3 eliminates the `+2` case for normalized divisors).
4. If D4 `inplaceSub` returns borrow (`c != 0`), `qhat` overestimated by exactly 1; the add-back corrects it.
5. After D4, `qhat == q_j` exactly.

**What is not proven here**: Knuth's theorem itself. Any implementation bug in normalization violates the theorem's precondition and may cause an unbounded error in `qhat`.

### 14.4 Trial Quotient Refinement: Why `rhat << 16` [JSBI SOURCE — line 1572]

```ts
while ((JSBI.__imul(qhat, vn2) >>> 0) > (((rhat << 16) | ujn2) >>> 0))
```

`ujn2 = halfDigit(j + n - 2)` is in range `[0, 0x7FFF]` (15 bits). The test `(rhat << 16) | ujn2` represents `rhat * 2^15 + ujn2` but shifted one extra bit to avoid overlap: shifting `rhat` by 16 places it entirely above `ujn2`'s 15-bit range. The `>>> 0` coerces to unsigned 32-bit for correct comparison. After `rhat += vn1` in the loop, `rhat` can exceed 0x7FFF and the loop breaks (line 1575) to prevent overflow.

### 14.5 D4 Add-Back Step [JSBI SOURCE — lines 1580–1584]

```ts
let c = u.__inplaceSub(qhatv, j, n + 1);
if (c !== 0) {
    c = u.__inplaceAdd(divisor, j, n);
    u.__setHalfDigit(j + n, (u.__halfDigit(j + n) + c) & 0x7FFF);
    qhat--;
}
```

If `inplaceSub` returns borrow `c != 0`, then `qhat * V > U[j..j+n]`, meaning `qhat` overestimated by 1. The divisor is added back at the same offset. The carry from `inplaceAdd` propagates into `halfDigit(j + n)`. Then `qhat` is decremented. After this, `qhat` equals the exact quotient half-digit `q_j`.

### 14.6 Half-Digit Quotient Packing [JSBI SOURCE — lines 1586–1592]

```ts
if (wantQuotient) {
    if (j & 1) {
        halfDigitBuffer = qhat << 15;
    } else {
        (q as JSBI).__setDigit(j >>> 1, halfDigitBuffer | qhat);
    }
}
```

The loop produces one 15-bit quotient half-digit per iteration. Two consecutive half-digits are packed into one 30-bit limb:

- `j` **odd**: `qhat` is the **high** half of limb `j >> 1`. Stored in `halfDigitBuffer = qhat << 15`.
- `j` **even**: `qhat` is the **low** half of limb `j >> 1`. Written as `halfDigitBuffer | qhat`.

Packing is delayed because both halves must be computed before the full 30-bit limb can be stored.

**Worked packing example** (m = 3, j runs 3 → 2 → 1 → 0, 4 iterations producing 2 limbs):

| j | `j & 1` | `qhat` (example) | Action |
| :--- | :--- | :--- | :--- |
| 3 | odd | 0x1A3F | `halfDigitBuffer = 0x1A3F << 15 = 0x0D1F8000` |
| 2 | even | 0x0B22 | `setDigit(1, 0x0D1F8000 | 0x0B22) = setDigit(1, 0x0D1F8B22)` |
| 1 | odd | 0x3FF1 | `halfDigitBuffer = 0x3FF1 << 15 = 0x1FF88000` |
| 0 | even | 0x2C00 | `setDigit(0, 0x1FF88000 | 0x2C00) = setDigit(0, 0x1FF8AC00)` |

**Boundary**: When `m` is even, `j` starts at an even value. The first iteration writes to `halfDigitBuffer = 0` — the high half for limb `j >> 1` is stored as 0 combined with the first `qhat`. This is correct because for an even-j starting loop, the high half of the highest quotient limb is zero.

---

## 15. Algorithm D — Worked Example

A complete step-by-step trace of Algorithm D through `internalMultiplyAdd` and `inplaceSub` internals requires the full source of those helpers (JSBI lines 1471–1507 and 1624–1684), which are not reproduced in this design review. Because a partial trace would leave intermediate arithmetic unverified, the worked example is omitted from this document.
The algorithm structure is fully specified in Sections 14.1–14.6 (source quotes, normalization invariant, trial quotient bound, D3 refinement, D4 add-back, and D5 packing). Correctness is verified by differential fuzzing against the Node.js JSBI oracle (Section 19) rather than by a hand-traced example.

**What the fuzzer verifies that a worked example would not**: the fuzzer exercises 65+ seconds of random operand pairs, including deliberate add-back trigger vectors and large borrow/carry chains, providing stronger correctness evidence than any single hand-traced example.

---


## 16. `DivRem` Specification [GO IMPLEMENTATION SPECIFICATION — Extension to JSBI]

JSBI does not expose a combined `divrem` function. `DivRem` is a Go implementation extension.

### Behaviour [GO IMPLEMENTATION SPECIFICATION]

```go
func DivRem(x, y *BigInt) (quotient *BigInt, remainder *BigInt, err error)
```

- If `y.Length() == 0`: returns `(nil, nil, ErrRange)`.
- If `|x| < |y|`: returns `(zero, copy(x), nil)` without entering any division algorithm.
- If divisor is large (not the small path): calls `absoluteDivLarge(dividend, divisor, true, true)` once, obtaining both quotient and remainder from a **single Algorithm D execution**, directly porting JSBI lines 1596–1598. This avoids duplicated Algorithm D work compared with calling `Divide()` then `Remainder()` separately.
- If divisor is small: **the implementation may compute quotient and remainder via one pass or two separate passes** — this is an open Go implementation choice. `absoluteDivSmall` does not return a remainder, and `absoluteModSmall` does not return a quotient, so one natural approach uses two passes (`absoluteDivSmall` then `absoluteModSmall`). An alternative single-pass implementation that computes both simultaneously is equally valid if implemented correctly. The choice must be documented in `DECISIONS.md` when made.

### Classification [GO IMPLEMENTATION SPECIFICATION]

- The large-divisor single-execution path is a direct port of JSBI lines 1596–1598.
- The small-divisor path strategy is a Go implementation decision, not mandated by JSBI.

---

## 17. Algorithmic Complexity

### Table 1: Algorithmic Complexity [GO IMPLEMENTATION SPECIFICATION]

Notation: H_X = half-digit count of X; H_Y = half-digit count of Y; L_X = limb count of X.

| Function | Time | Auxiliary Space |
| :--- | :--- | :--- |
| `absoluteDivSmall` | O(L_X) | O(L_X) when allocating quotient; O(1) otherwise |
| `absoluteModSmall` | O(L_X) | O(1) |
| `absoluteDivLarge` | O(H_X × H_Y) | O(H_X + H_Y) |
| `Divide` (small path) | O(L_X) | O(L_X) |
| `Divide` (large path) | O(H_X × H_Y) | O(H_X + H_Y) |
| `Remainder` (small path) | O(L_X) | O(1) |
| `Remainder` (large path) | O(H_X × H_Y) | O(H_X + H_Y) |
| `DivRem` (large path) | O(H_X × H_Y) | O(H_X + H_Y) |

### Table 2: Go Engineering Targets [ENGINEERING GOAL — non-normative]

These are targets to measure against after implementation. They are not algorithmic guarantees. Actual heap allocation counts depend on compiler escape analysis, Go version, and implementation details.

| Function | Expected Heap Allocations | Benchmark |
| :--- | :--- | :--- |
| `Divide` | Low (target: ~2) | `BenchmarkDivide -test.benchmem` |
| `Remainder` | Low (target: ~2) | `BenchmarkRemainder -test.benchmem` |
| `DivRem` (large) | Low (target: ~4) | `BenchmarkDivRem -test.benchmem` |

---

## 18. Recommended Implementation Order [GO IMPLEMENTATION SPECIFICATION]

Implement in this sequence to minimize debugging surface at each step:

1. **`absoluteModSmall`** — simplest helper, O(1) state, no allocation, easy to unit-test
2. **`absoluteDivSmall`** — builds on the same recurrence, easy to verify with small examples
3. **`specialLeftShift`** — normalization primitive needed by Algorithm D; test independently
4. **`Divide` and `Remainder` — small path only** — wire up path selection, error handling, sign logic, canonical zero; verify with differential fuzzer before touching Algorithm D
5. **`inplaceRightShift`** — unnormalization; test independently
6. **`absoluteDivLarge` — skeleton with D1 normalization** — implement and test normalization only, verify `v[n-1] >= 0x4000` post-condition
7. **`absoluteDivLarge` — D3 trial quotient** — implement estimation and refinement, verify against oracle for small cases
8. **`absoluteDivLarge` — D4 multiply/subtract and add-back** — implement `internalMultiplyAdd`, `inplaceSub`, `inplaceAdd`; add-back is the highest-risk step
9. **`absoluteDivLarge` — D5 quotient packing** — implement half-digit packing; verify with even/odd m boundary cases
10. **`absoluteDivLarge` — D6 unnormalization** — wire in `inplaceRightShift`
11. **`Divide` and `Remainder` — large path** — wire up `absoluteDivLarge` into public API
12. **`DivRem`** — implement using `absoluteDivLarge(true, true)` for large path
13. **Full differential fuzzing** — run ≥ 65s against oracle, including all required vectors

---

## 19. Differential Fuzzing Plan [GO IMPLEMENTATION SPECIFICATION]

- **Harness File**: `fuzz/harness/fuzz_cluster5.go`
- **Oracle Helper**: `fuzz/harness/oracle_cluster5.mjs`
- **Oracle Engine**: Node.js executing the original JSBI library (`jsbi/dist/jsbi-cjs.js`). The oracle is a test-time dependency only; it is not shipped with the Go port.
- **Target Duration**: ≥ 65 continuous seconds.

### Oracle Comparison Protocol [GO IMPLEMENTATION SPECIFICATION]

For each test case, the harness provides identical operands to both the Go port and to Node.js running JSBI. The Node.js oracle executes `JSBI.divide(x, y)` and `JSBI.remainder(x, y)` and exports `divSign`, `divLen`, `divDigits: Array.from(...)`, `remSign`, `remLen`, `remDigits: Array.from(...)`. The Go harness then checks:
- Boolean sign match for quotient and remainder independently
- Limb count match for quotient and remainder independently
- **Element-by-element 30-bit digit array match**: `go.Digit(i) == oracle.Digits[i]` for all `i`
- **Canonical zero assertion**: If `len == 0`, assert `sign == false`
- **Division identity check** [ENGINEERING GOAL — secondary verification, may be added]: `Q * Y + R == X`

### Required Test Vectors [GO IMPLEMENTATION SPECIFICATION]

| Vector Category | Specific Cases |
| :--- | :--- |
| Division by zero | `Y = 0` |
| Divisor > Dividend | `|X| < |Y|` → Q = 0, R = X |
| Division by ±1 | `Y = 1n`, `Y = -1n` |
| Equal operands | `X == Y` → Q = 1 or -1, R = 0 |
| Small path: arbitrary divisors | Random `Y.Length()==1 && Y.Digit(0) <= 0x7FFF` |
| Small path: divisor boundary | `Y.Digit(0) = 0x7FFE`, `0x7FFF` |
| Large single-limb divisor | `Y.Digit(0) = 0x8000`, `0x8001` → forces Algorithm D |
| One limb ÷ many limbs | 1-limb dividend, 3+ limb divisor → Q = 0 |
| Many limbs ÷ one limb | 100+ limb dividend, 1-limb small divisor |
| Power-of-two divisors | `Y = 2^k` for k = 1…60 |
| Normalization shift = 0 | Leading Y half-digit = `0x7FFF` |
| Normalization shift = 14 | Leading Y half-digit = `0x4000` |
| All normalization shifts 0..14 | Divisors designed for each specific shift value |
| Divisors with leading half = 0x4000 | Exact boundary of normalization invariant |
| Divisors with leading half = 0x4001 | One above the boundary |
| Alternating half-digit patterns | `0x55555555…` / `0xAAAAAAAA…` operands |
| All-0x7FFF half-digits divisor | Maximum value in each half-digit position |
| Negative dividend | `X < 0, Y > 0` |
| Negative divisor | `X > 0, Y < 0` |
| Both negative | `X < 0, Y < 0` |
| Operands differing by exactly 1 limb | `X.Length() = Y.Length() + 1` |
| Add-back trigger vectors | Crafted inputs where D4 add-back fires |
| Multiple consecutive add-backs | Multiple j iterations each triggering add-back |
| Large borrow chains | Multi-limb borrow propagation in `inplaceSub` |
| Large carry chains | Multi-limb carry propagation in `inplaceAdd` |
| Quotient with leading zero limbs | Cases where `Trim()` removes leading zeros |
| Canonical zero remainder | R = 0 exactly |
| Random 1–100 limb operands | Broad random coverage |
| Random 100–2000 limb operands | Large operand coverage |
| 500+ limb operands | Stress test of Algorithm D scalability |
| 1000+ limb operands | Maximum scale test |
| Max 30-bit limb values | All digits = 0x3FFFFFFF |
| Repeated borrow chains | Alternating subtraction patterns causing cascading borrows |
| Repeated carry chains | Patterns causing cascading carries in add-back |

---

## 20. Final Acceptance Checklist [GO IMPLEMENTATION SPECIFICATION]

The following criteria must all be satisfied before Cluster 5 is considered complete and committed.

### Code
- [ ] `absoluteModSmall` implemented and unit-tested independently
- [ ] `absoluteDivSmall` implemented and unit-tested independently
- [ ] `specialLeftShift` implemented and unit-tested independently
- [ ] `absoluteDivLarge` implemented with all 6 Algorithm D steps
- [ ] `Divide` implemented covering both paths + all edge cases
- [ ] `Remainder` implemented covering both paths + all edge cases
- [ ] `DivRem` implemented with documented small-path strategy in `DECISIONS.md`

### Tests
- [ ] Dedicated unit test suite `TestDivideBasic` passes
- [ ] Dedicated unit test suite `TestRemainderBasic` passes
- [ ] Dedicated unit test suite `TestDivRemBasic` passes
- [ ] Standing regression suite (Clusters 1–5) passes with 0 failures
- [ ] All six correctness invariants (Section 3) verified by targeted test cases

### Fuzzing
- [ ] `fuzz_cluster5.go` harness implemented with oracle_cluster5.mjs
- [ ] Fuzzing run ≥ 65 continuous seconds with 0 mismatches
- [ ] All required vector categories from Section 19 covered by the harness
- [ ] Results appended to `fuzz/log.txt` from live run output

### Benchmarks
- [ ] `BenchmarkDivide` executed and reported with `-test.benchmem`
- [ ] `BenchmarkRemainder` executed and reported with `-test.benchmem`
- [ ] `BenchmarkDivRem` executed and reported with `-test.benchmem`
- [ ] Results recorded in `bench/results.json`

### Documentation
- [ ] `DECISIONS.md` entry for Cluster 5 written (including DivRem small-path strategy)
- [ ] `CHANGELOG.md` entry written
- [ ] `README.md` status table updated
- [ ] `SDFGH/PROJECT_STATUS.md` updated
- [ ] `SDFGH/ppp.md` timeline entry appended

---

## 21. Self Review [ENGINEERING INFERENCE]

1. **Which external correctness dependency is highest risk?**
   The Knuth trial quotient bound theorem (Section 14.3). If normalization is implemented incorrectly, the precondition `v_{n-1} >= 0x4000` fails silently and `qhat` may be unboundedly wrong.

2. **Which implementation is highest risk?**
   `absoluteDivLarge` → `inplaceSub` and `inplaceAdd` half-digit index arithmetic (JSBI lines 1624–1684). The odd/even boundary unrolling is the most complex indexing in the codebase.

3. **Which bug would survive unit tests but be caught by fuzzing?**
   An add-back correction that fires for odd-j iterations and corrupts the half-digit buffer, producing wrong quotient packing only in certain operand size combinations.

4. **Which invariant is easiest to violate?**
   Remainder sign invariant: `sign(R) = sign(X)`. If `remainder.sign = x.sign` (line 265) is accidentally omitted, all negative dividends produce wrong sign.

5. **Which assumption is not verified by this document?**
   The D3 refinement loop always terminates. JSBI line 1575 breaks when `rhat > 0x7FFF`. This termination condition is visible in the source but not formally proven here.

6. **Which assumption is most difficult to verify by inspection alone?**
   The correctness of `inplaceSub` and `inplaceAdd` half-digit index arithmetic (JSBI lines 1624–1684). The odd/even unrolling is intricate and must be verified both by unit tests and by differential fuzzing against JSBI running under Node.js.
