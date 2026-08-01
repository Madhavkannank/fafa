# Design Review 09 — String Formatting (`ToString`)

- **Cluster**: 9 — String Formatting (`ToString`)
- **Status**: Draft / Pending User Review [GATE]
- **Author**: Lead Engineer (Go Porting Team)
- **Reference**: `GoogleChromeLabs/jsbi` (`jsbi/lib/jsbi.ts` lines 67–77, 916–1010)

---

## 1. Overview & Operational Scope [GO IMPLEMENTATION SPECIFICATION]

Cluster 9 implements radix conversion from BigInt to string:

- `ToString(x *BigInt, radix int) (string, error)`

This is the final output cluster. It converts the internal 30-bit limb representation into a human-readable ASCII string for any radix in `[2, 36]`.

### Scope & Dependencies [GO IMPLEMENTATION SPECIFICATION]
- **Primitives Used**: `Digit`, `Length`, `Sign`, `clz30`, half-digit accessors (`halfDigit`, `setHalfDigit`), `absoluteDivLarge` (Cluster 5), `Exponentiate` (Cluster 4 / external).
- **Dependencies**: Cluster 4 (Multiply/Exponentiate), Cluster 5 (Division).

---

## 2. ECMAScript Specification Mapping [ECMAScript SPECIFICATION]

Ref: ECMAScript Language Specification (ECMA-262), Section *BigInt::toString*.
- `BigInt.toString(radix)` returns the shortest string representation in the given radix.
- Radix must be an integer in `[2, 36]`; anything outside throws a `RangeError`.
- A negative BigInt is represented with a leading `-` character; zero returns `"0"`.

---

## 3. Algorithm — JSBI Source Control Flow [JSBI SOURCE]

### 3.1 Entry Point (`toString`, lines 67–77)

```
toString(radix = 10):
  if radix < 2 || radix > 36 → throw RangeError
  if length == 0              → return "0"
  if (radix & (radix-1)) == 0 → __toStringBasePowerOfTwo(this, radix)  [fast path]
  else                         → __toStringGeneric(this, radix, false)  [general path]
```

**Power-of-two detection**: `(radix & (radix - 1)) == 0` detects radices 2, 4, 8, 16, 32 in O(1). These are the "easy" radices — each output character maps to exactly `log2(radix)` bits of the input, allowing a simple bit-extraction loop without any division.

### 3.2 Power-of-Two Fast Path (`__toStringBasePowerOfTwo`, lines 916–958)

**Algorithm**:
1. Compute `bitsPerChar = popcount(radix - 1)` — the number of bits per output character (using a 3-step bit-count trick).
2. Compute `charMask = radix - 1` — bitmask to extract one character worth of bits.
3. Compute total bit length: `bitLength = length * 30 - clz30(msd)`.
4. Compute `charsRequired = ceil(bitLength / bitsPerChar)`. Add 1 if negative.
5. Allocate `result[charsRequired]`, fill right-to-left.
6. Walk limbs from index 0 upward, maintaining a `digit` accumulator and `availableBits`:
   - For each limb, extract as many full characters as possible from the combined bits.
   - After all full limbs, drain the MSD.
7. If negative, prepend `'-'`.
8. Assert `pos == -1` as a consistency check.

**Key invariants**:
- `availableBits` counts bits already extracted from the current limb but not yet consumed.
- Characters are filled from index `charsRequired-1` down to `0` (right-to-left, avoiding later reversal).
- MSD is handled separately after the inner loop.

**Error guard**: JSBI throws `Error('string too long')` if `charsRequired > (1 << 28)`. In Go: return `ErrRange` for this case. This mapping follows the project's established error-mapping policy used in previous clusters: non-range-check JS errors that indicate an out-of-range condition are uniformly mapped to `ErrRange`. No new error type is introduced.

### 3.3 General (Divide-and-Conquer) Path (`__toStringGeneric`, lines 960–1010)

**Algorithm**:
1. **Base case — zero limbs**: return `""` (recursive call returns `""` — top-level call never reaches here due to early exit in entry point).
2. **Base case — one limb**: `digit.toString(radix)` — in Go: `strconv.FormatUint(uint64(x.Digit(0)), radix)`. Prepend `"-"` if top-level and negative.
3. **Recursive divide-and-conquer**:
   - Estimate `charsRequired` using the `__kMaxBitsPerChar` lookup table (ceiling of `bitLength / log2(radix)`, scaled by `kBitsPerCharTableMultiplier = 32`):
     ```
     charsRequired = ceil(bitLength * 32 / (kMaxBitsPerChar[radix] - 1))
     secondHalfChars = (charsRequired + 1) >> 1
     ```
   - Compute `conqueror = radix ^ secondHalfChars` (via `Exponentiate`).
   - **Fast division path** (if `conqueror` fits in one limb, `divisor <= 0x7FFF`):
     - Manual half-digit long division: iterate `i = length*2-1 .. 0`, compute `quotient` and `remainder` using 15-bit arithmetic.
     - `secondHalf = remainder.toString(radix)` (native JS, Go: `strconv.FormatUint`).
   - **Full division path** (multi-limb conqueror):
     - Call `absoluteDivLarge(x, conqueror, true, true)` to get `{quotient, remainder}`.
     - `secondHalf = __toStringGeneric(remainder, radix, true)` (recursive).
   - Recurse: `firstHalf = __toStringGeneric(quotient, radix, true)`.
   - Pad `secondHalf` with leading zeros to length `secondHalfChars`.
   - Concatenate `firstHalf + secondHalf`. Prepend `"-"` if top-level negative.

**Complexity**: O(n^1.585 · log n) — divide-and-conquer reduces the recursion depth to O(log n) instead of O(n^2) for naive repeated division.

---

## 4. Dispatch Table [JSBI SOURCE]

| Condition | Path Taken |
|---|---|
| `radix < 2 || radix > 36` | Error: `ErrRange` |
| `x.Length() == 0` | Return `"0"` immediately |
| `(radix & (radix-1)) == 0` | `toStringBasePowerOfTwo` (bit extraction, no division) |
| All other radix values (non-power-of-two) | `toStringGeneric` (divide-and-conquer recursion) |
| `toStringGeneric` base case: `length == 1` | `strconv.FormatUint(uint64(digit), radix)` |
| `toStringGeneric` conqueror fits in 1 limb & `<= 0x7FFF` | Fast half-digit division loop |
| `toStringGeneric` conqueror multi-limb | `absoluteDivLarge` |

---

## 5. Helper Contracts [GO IMPLEMENTATION SPECIFICATION]

### `toStringBasePowerOfTwo(x *BigInt, radix int) string`
- **Purpose**: Convert BigInt to string for power-of-two radices (2, 4, 8, 16, 32).
- **Inputs**: `x` — the value to convert; `radix` — a power of two in `[2, 32]`.
- **Output**: Decimal string with optional leading `"-"`.
- **Precondition**: `x.Length() > 0`; caller ensures `(radix & (radix-1)) == 0`.
- **Modifies inputs**: No.
- **Allocation**: Allocates one `[]byte` of length `charsRequired`.
- **Trim required**: No (reads only, does not mutate `x`).

### `toStringGeneric(x *BigInt, radix int, isRecursiveCall bool) string`
- **Purpose**: Convert BigInt to string for arbitrary radix using divide-and-conquer.
- **Inputs**: `x` — the value to convert (magnitude only; sign handled by top-level call); `radix` — integer in `[2, 36]`; `isRecursiveCall` — suppresses sign prefix and `"0"` return in recursive sub-calls.
- **Output**: String fragment (may be empty on recursive zero-length sub-call).
- **Precondition**: `x.Length() > 0` for meaningful output; `isRecursiveCall == false` at top level.
- **Modifies inputs**: No (reads only; `quotient.__trim()` and `remainder.__trim()` are called on freshly allocated results from division, not on `x`).
- **Allocation**: Allocates intermediate `quotient`, `conqueror`, and `remainder` BigInts; allocates string output.
- **Trim required**: Caller trims `quotient` and `remainder` after division before recursing.

---

## 6. `__kMaxBitsPerChar` Lookup Table [JSBI SOURCE]

JSBI uses a static precomputed table `__kMaxBitsPerChar` indexed by radix (0–36), containing `ceil(log2(radix)) * 32` — that is, the ceiling of `log2(radix)`, scaled by `kBitsPerCharTableMultiplier = 32`:

```
kMaxBitsPerChar = [0, 0, 32, 51, 64, 75, 83, 90, 96, 102, 107, 111, 115,
                   119, 122, 126, 128, 131, 134, 136, 139, 141, 143, 145,
                   147, 149, 151, 153, 154, 156, 158, 159, 160, 162, 163, 165, 166]
```

> **Note**: This lookup table must be copied verbatim from the JSBI source. Individual values must not be manually derived or modified.

Usage in `toStringGeneric`:
```
bitLength = length * 30 - clz30(msd)
maxBitsPerChar = kMaxBitsPerChar[radix]         // ceil(log2(radix)) * 32
minBitsPerChar = maxBitsPerChar - 1             // floor(log2(radix)) * 32 approximately
charsRequired = ceil(bitLength * 32 / minBitsPerChar)
```

This overestimates `charsRequired` slightly (padding with leading zeros later corrects the second-half string), ensuring no buffer underrun.

---

## 7. Edge Cases [JSBI SOURCE]

| Case | JSBI Behavior | Go Mapping |
|---|---|---|
| `x == 0` | `"0"` (early exit in `toString`) | Return `"0"` |
| `radix < 2 or > 36` | `RangeError('toString() radix argument must be between 2 and 36')` | Return `("", ErrRange)` |
| Negative `x` | Leading `"-"` prepended once at top level only | Same |
| `x.Length() == 1` (recursive) | `digit.toString(radix)` without sign | `strconv.FormatUint(uint64(d), radix)` |
| `charsRequired > (1 << 28)` | `Error('string too long')` | Return `("", ErrRange)` |
| Radix 10 | `toStringGeneric` (non-power-of-two path) | Same |
| Radix 16 | `toStringBasePowerOfTwo` (4 bits per char) | Same |
| `secondHalf` too short | Left-padded with `'0'` to exactly `secondHalfChars` width | Same, use `strings.Repeat("0", ...)` |

---

## 8. `popcount`-based `bitsPerChar` Computation [JSBI SOURCE]

JSBI computes `bitsPerChar = popcount(radix - 1)` using a 3-pass bit-parallel trick at lines 918–922:

```js
let bits = radix - 1;
bits = ((bits >>> 1) & 0x55) + (bits & 0x55);    // 2-bit sums
bits = ((bits >>> 2) & 0x33) + (bits & 0x33);    // 4-bit sums
bits = ((bits >>> 4) & 0x0F) + (bits & 0x0F);    // 8-bit sum
const bitsPerChar = bits;
```

For valid power-of-two radices: radix-1 has exactly `k` bits set where `2^k == radix`.
- radix 2 → `bits = 1`
- radix 4 → `bits = 2`
- radix 8 → `bits = 3`
- radix 16 → `bits = 4`
- radix 32 → `bits = 5`

**Go equivalent**: Use `math/bits.OnesCount32(uint32(radix - 1))` — identical result, more readable.

---

## 9. Half-Digit Fast Division [JSBI SOURCE]

When `conqueror.Length() == 1 && conqueror.Digit(0) <= 0x7FFF`, JSBI uses a fast 15-bit half-digit division loop (lines 986–993). This parallels the same half-digit machinery used in `absoluteDivSmall` (Cluster 5).

```
quotient = new JSBI(x.length, false), __initializeDigits()
remainder = 0
for i = x.length * 2 - 1 .. 0:
    input = (remainder << 15) | x.__halfDigit(i)
    quotient.__setHalfDigit(i, floor(input / divisor))
    remainder = input % divisor
secondHalf = remainder.toString(radix)
```

**Go implementation**: Requires `HalfDigit(i int) uint32` and `SetHalfDigit(i int, v uint32)` accessors on `*BigInt` — these split each 30-bit digit into two 15-bit halves, indexed as `halfDigit(2*limbIndex)` = low 15 bits, `halfDigit(2*limbIndex + 1)` = high 15 bits. These were already needed for Cluster 5 division; verify they already exist in `src/divide.go` or `src/bigint.go`.

---

## 10. Value Independence [GO IMPLEMENTATION SPECIFICATION]

`ToString` is a pure read — it never modifies `x`. Internal `quotient` and `remainder` objects are freshly allocated by the division helpers. No aliasing concerns arise from the caller's perspective.

---

## 11. Worked Examples [JSBI SOURCE]

### Example 1: `ToString(0, 10)` = `"0"`
- `x.Length() == 0` → return `"0"` immediately (entry point early exit).

### Example 2: `ToString(255, 16)` = `"ff"` — power-of-two path
- `radix = 16`, `(16 & 15) == 0` → `toStringBasePowerOfTwo`.
- `bitsPerChar = popcount(15) = 4`, `charMask = 15`.
- `255 = 0xFF`. Digits: `[0xFF]` → single 30-bit limb.
- `bitLength = 1*30 - clz30(0xFF) = 30 - 22 = 8`.
- `charsRequired = ceil(8/4) = 2`.
- Loop i=0 (last limb is also index 0, handled in MSD section):
  - `current = (0 | 0xFF << 0) & 0xF = 0xF = 15` → `'f'`, `pos=1`.
  - `digit = 0xFF >>> 4 = 0xF`.
  - `digit & 0xF = 15` → `'f'`, `pos=0`.
  - `digit >>> 4 = 0` → stop.
- Result (built right-to-left): `"ff"`.

### Example 3: `ToString(-1000, 10)` = `"-1000"` — generic path, single limb
- `x.Length() == 1`, `x.Sign() == true`.
- `(10 & 9) != 0` → `toStringGeneric(x, 10, false)`.
- `length == 1` base case: `result = FormatUint(1000, 10) = "1000"`. `isRecursiveCall == false && sign` → `result = "-1000"`.
- Return `"-1000"`.

### Example 4: Multi-limb, radix 10 (conceptual)
- `x` = large number, e.g. `2^60 + 1`, two 30-bit limbs.
- `bitLength = 2*30 - clz30(msd)`.
- `secondHalfChars = ceil(charsRequired / 2)`.
- `conqueror = 10^secondHalfChars` computed via `Exponentiate`.
- If `conqueror` fits in one half-digit-safe limb: fast loop; otherwise `absoluteDivLarge`.
- Recurse on quotient; pad second half; concatenate.

---

## 12. Canonical Zero [GO IMPLEMENTATION SPECIFICATION]

A canonical zero `BigInt` has `Length() == 0` and `Sign() == false`. The entry point handles this as the first check: `if x.Length() == 0 → return "0"`. No downstream helper is ever called on a zero-length input.

---

## 13. Helper Coverage Map [GO IMPLEMENTATION SPECIFICATION]

| Public API Call | Entry Point | Helpers Called |
|---|---|---|
| `ToString(x, 2/4/8/16/32)` | entry point guard | `toStringBasePowerOfTwo` → `clz30`, `bits.OnesCount32` |
| `ToString(x, non-power-of-two)` | entry point guard | `toStringGeneric` → `Exponentiate`, `absoluteDivLarge` (or fast loop), `FormatUint`, recursive `toStringGeneric` |
| `ToString(x=0, any)` | entry point guard | none |

---

## 14. Differential Fuzzing Protocol [GO IMPLEMENTATION SPECIFICATION]

**Oracle**: Node.js `JSBI.toString(x, radix)` vs. Go `ToString(x, radix)`.

**Mandatory test radices**:
- Every radix from `2` through `36` inclusive must be exercised in the differential fuzzing run. This is an exhaustive requirement, not a representative sample. All 5 power-of-two radices (`2, 4, 8, 16, 32`) must hit the `toStringBasePowerOfTwo` path; all 30 remaining radices (`3, 5, 6, 7, 9, 10, 11, 12, 13, 14, 15, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 33, 34, 35, 36`) must hit the `toStringGeneric` path.

**Mandatory operand distribution**:
- Zero
- One-limb operands (positive and negative)
- Two-limb operands (positive and negative)
- Eight-limb operands
- Sixteen-limb operands
- Powers of two (`2^29`, `2^30`, `2^59`, `2^60`, `2^89`, `2^90`)
- Very large random numbers (20–30 random decimal digits)
- Boundary: `-1`, `1`, `2^30 - 1`, `2^30`

**Dedicated unit tests for the fast division path**: The fast half-digit division path (`conqueror.Length() == 1 && divisor <= 0x7FFF`) must be covered by dedicated unit tests with hand-crafted inputs that are known to force this branch. Random fuzzing may not reliably reach this optimisation path because the split-point `secondHalfChars` may be large enough to make `conqueror` multi-limb for typical random inputs. Construct inputs of exactly 2–4 limbs at carefully chosen radices (e.g. radix 10 with small `secondHalfChars`) to guarantee the fast path executes.

**Verification per case**:
1. `gotString == oracleString` (byte-for-byte identical)
2. Error agreement: if oracle throws, Go returns non-nil error, and vice versa.

**Minimum run**: 60 continuous seconds. Record total case count, radix distribution, and operand size distribution in `fuzz/log.txt`.

---

## 15. Benchmark Targets [ENGINEERING GOAL]

| Operation | Target |
|---|---|
| `ToString(x, 16)` 1 limb | ≤ 80 ns/op, ≤ 48 B/op, ≤ 2 allocs/op |
| `ToString(x, 10)` 1 limb | ≤ 80 ns/op, ≤ 64 B/op, ≤ 2 allocs/op |
| `ToString(x, 10)` 3 limbs | ≤ 300 ns/op |

These are engineering goals, not JSBI properties. Benchmark targets are engineering goals only and are not correctness or pass/fail requirements.
