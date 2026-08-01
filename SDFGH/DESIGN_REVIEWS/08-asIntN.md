# Design Review 08 — Fixed-Width Truncation (`AsIntN`, `AsUintN`)

- **Cluster**: 8 — Fixed-Width Truncation (`AsIntN`, `AsUintN`)
- **Status**: Draft / Pending User Review [GATE]
- **Author**: Lead Engineer (Go Porting Team)
- **Reference**: `GoogleChromeLabs/jsbi` (`jsbi/lib/jsbi.ts` lines 408–466, 1860–1907)

---

## 1. Overview & Operational Scope [GO IMPLEMENTATION SPECIFICATION]

Cluster 8 implements fixed-width integer bit truncation for arbitrary-precision `BigInt` values:
- `AsIntN(bits int, x *BigInt) (*BigInt, error)`
- `AsUintN(bits int, x *BigInt) (*BigInt, error)`

### Scope & Dependencies [GO IMPLEMENTATION SPECIFICATION]
- **Primitives Used**: 30-bit limb accessors (`Digit`, `SetDigit`, `Length`, `Sign`, `SetSign`, `Trim`, `Copy`, `Zero`), Go native error types (`ErrRange`).
- **Dependencies**: None outside `jsbi` package primitives.

---

## 2. ECMAScript Specification Mapping [ECMAScript SPECIFICATION]

Ref: ECMAScript Language Specification (ECMA-262), Sections *BigInt.asIntN* and *BigInt.asUintN*.
- `BigInt.asIntN(bits, BigInt)` wraps a BigInt to a signed $N$-bit integer ($[-2^{N-1}, 2^{N-1}-1]$).
- `BigInt.asUintN(bits, BigInt)` wraps a BigInt to an unsigned $N$-bit integer ($[0, 2^N-1]$).
- JSBI simulates two's complement bit wrapping on magnitude-and-sign digit arrays using `__truncateToNBits` and `__truncateAndSubFromPowerOfTwo`.

---

## 3. JSBI Source Reference & Line Mapping [JSBI SOURCE]

| JSBI TS Function | JSBI TS Lines | Target Go Identifier |
| :--- | :--- | :--- |
| `JSBI.asIntN(n, x)` | 408–437 | `AsIntN(bits int, x *BigInt) (*BigInt, error)` |
| `JSBI.asUintN(n, x)` | 439–466 | `AsUintN(bits int, x *BigInt) (*BigInt, error)` |
| `JSBI.__truncateToNBits(n, x)` | 1860–1874 | `truncateToNBits(bits int, x *BigInt) *BigInt` |
| `JSBI.__truncateAndSubFromPowerOfTwo(n, x, resultSign)` | 1876–1907 | `truncateAndSubFromPowerOfTwo(bits int, x *BigInt, resultSign bool) *BigInt` |

---

## 4. JSBI Error & Sentinel Mapping [GO IMPLEMENTATION SPECIFICATION]

### Error Mapping Decision [GO IMPLEMENTATION SPECIFICATION]
- **JSBI Source**: Throws `RangeError('Invalid value: not (convertible to) a safe integer')` when `bits < 0` (lines 411–414, 442–445).
- **JSBI Source**: Throws `RangeError('BigInt too big')` when `bits > kMaxLengthBits` on negative `asUintN` inputs (lines 449–451).
- **Go Target**: Map both error conditions to Go sentinel `ErrRange`.

---

## 5. Control Flow & Detailed Dispatch Tables [JSBI SOURCE]

### 5.1 Distinguishing `bits == kMaxLengthBits` vs `bits > kMaxLengthBits` [JSBI SOURCE]
In JSBI, `kMaxLengthBits = 1 << 30` ($1,073,741,824$ bits):
- **`bits == kMaxLengthBits`**:
  - Any representable JSBI `BigInt` has fewer than $2^{30}$ bits. Therefore, if $x \ge 0$, no truncation is needed and JSBI returns $x$ directly.
  - If $x < 0$ in `asUintN`, JSBI executes `__truncateAndSubFromPowerOfTwo(kMaxLengthBits, x, false)` to compute $2^{\text{kMaxLengthBits}} - |x|$.
- **`bits > kMaxLengthBits`**:
  - For $x \ge 0$, JSBI returns $x$ directly since $x$ cannot exceed `kMaxLengthBits` bits.
  - For $x < 0$ in `asUintN`, allocating an array of size $2^{\text{kMaxLengthBits}}$ would exceed maximum memory limits. JSBI throws `RangeError('BigInt too big')` (mapped to `ErrRange` in Go).

### 5.2 `AsIntN` Dispatch Table [JSBI SOURCE]

| Bit Width Condition | Operand State | JSBI Control Flow / Action | Result |
| :--- | :--- | :--- | :--- |
| `x.Length() == 0` | Any | `return x` | `Zero()` |
| `bits < 0` | Any | `throw RangeError` | `nil, ErrRange` |
| `bits == 0` | Any | `return JSBI.__zero()` | `Zero()` |
| `bits == kMaxLengthBits` | Any | `return x` | Independent Copy of $x$ |
| `bits > kMaxLengthBits` | Any | `return x` | Independent Copy of $x$ |
| `x.Length() < neededLength` | Any | `return x` | Independent Copy of $x$ |
| Sign-bit `(topDigit & compareDigit) == 0` | Any | `return __truncateToNBits(bits, x)` | Positive truncated result |
| Sign-bit set ($1$) | Positive ($x \ge 0$) | `return __truncateAndSubFromPowerOfTwo(bits, x, true)` | Negative result |
| Sign-bit set ($1$) | Negative ($x < 0$) | Complex check $\implies$ `__truncateAndSubFromPowerOfTwo(bits, x, false)` or `__truncateToNBits(bits, x)` | Truncated result |

### 5.3 `AsUintN` Dispatch Table [JSBI SOURCE]

| Bit Width Condition | Operand State | JSBI Control Flow / Action | Result |
| :--- | :--- | :--- | :--- |
| `x.Length() == 0` | Any | `return x` | `Zero()` |
| `bits < 0` | Any | `throw RangeError` | `nil, ErrRange` |
| `bits == 0` | Any | `return JSBI.__zero()` | `Zero()` |
| `bits == kMaxLengthBits` | Positive ($x \ge 0$) | `return x` | Independent Copy of $x$ |
| `bits == kMaxLengthBits` | Negative ($x < 0$) | `return __truncateAndSubFromPowerOfTwo(bits, x, false)` | Unsigned wrapped result |
| `bits > kMaxLengthBits` | Positive ($x \ge 0$) | `return x` | Independent Copy of $x$ |
| `bits > kMaxLengthBits` | Negative ($x < 0$) | `throw RangeError('BigInt too big')` | `nil, ErrRange` |
| `x.Sign() == true` (Negative) | Any valid `bits` | `return __truncateAndSubFromPowerOfTwo(bits, x, false)` | Unsigned wrapped result |
| `x.Sign() == false` & fits in `bits` | Positive | `return x` | Independent Copy of $x$ |
| `x.Sign() == false` & exceeds `bits` | Positive | `return __truncateToNBits(bits, x)` | Unsigned truncated result |

---

## 6. Helper Prerequisites & Formal Contracts [GO IMPLEMENTATION SPECIFICATION]

### 6.1 Helper Prerequisites [GO IMPLEMENTATION SPECIFICATION]
For truncation helpers `truncateToNBits` and `truncateAndSubFromPowerOfTwo`, the required primitives are:
- `Digit(i int) uint32`: Reads 30-bit limb at index $i$.
- `SetDigit(i int, val uint32)`: Writes 30-bit limb at index $i$.
- `Length() int`: Returns current limb count.
- `Sign() bool`: Reads sign flag (`true` for negative).
- `SetSign(sign bool)`: Sets sign flag.
- `Trim() *BigInt`: Removes leading zero limbs and normalizes zero.
- **Borrow / Carry Mechanics**:
  - `truncateToNBits`: Bitwise mask operation on top limb: `msd = (msd << (32 - drop)) >>> (32 - drop)`.
  - `truncateAndSubFromPowerOfTwo`: Subtraction borrow propagation `borrow = (r >>> 30) & 1` across limbs $0 \dots \text{neededDigits}-1$.

### 6.2 Contract: `truncateToNBits(bits int, x *BigInt) *BigInt`
- **Purpose**: Copies magnitude digits $0 \dots \text{neededDigits}-2$ directly from $x$, masks top digit to `bits % 30` bits, sets sign to `x.Sign()`, and calls `.Trim()`.
- **Inputs**: `bits`: target bit width ($> 0$). `x`: input `*BigInt`.
- **Outputs**: `*BigInt` containing truncated magnitude.
- **Input Immutability**: $x$ is **never modified**.
- **Aliasing Guarantee**: Returns a newly allocated internal result object.

### 6.3 Contract: `truncateAndSubFromPowerOfTwo(bits int, x *BigInt, resultSign bool) *BigInt`
- **Purpose**: Simulates two's complement subtraction $2^{\text{bits}} - |x|$ across digits $0 \dots \text{neededDigits}-1$, sets sign to `resultSign`, and calls `.Trim()`.
- **Inputs**: `bits`: target bit width. `x`: input `*BigInt`. `resultSign`: sign flag for output.
- **Outputs**: `*BigInt` containing two's complement wrapped magnitude.
- **Input Immutability**: $x$ is **never modified**.
- **Aliasing Guarantee**: Returns a newly allocated internal result object.

---

## 7. Helper Coverage Map [GO IMPLEMENTATION SPECIFICATION]

| Helper Function | Exercised By Public APIs | Conditions |
| :--- | :--- | :--- |
| `truncateToNBits` | `AsIntN(bits, x)` | Positive $x$ where sign bit is 0, or negative $x$ after two's complement reduction. |
| | `AsUintN(bits, x)` | Positive $x$ whose bit length exceeds `bits`. |
| `truncateAndSubFromPowerOfTwo` | `AsIntN(bits, x)` | Positive $x$ where sign bit is 1 (wraps to negative), or negative $x$ needing subtraction from $2^N$. |
| | `AsUintN(bits, x)` | Any negative $x$ ($x < 0$) for valid bit width $N \le \text{kMaxLengthBits}$. |

---

## 8. Canonical Zero Section [GO IMPLEMENTATION SPECIFICATION]

- Truncating any BigInt to 0 bits (`AsIntN(0, x)`, `AsUintN(0, x)`) returns canonical zero `0n`.
- Truncating a multiple of $2^N$ to $N$ bits returns canonical zero.
- **Strict Invariant**: `Length() == 0 ==> Sign() == false`.

---

## 9. Worked Execution Examples

### Example 1: `AsIntN(4, 7)` ($x = 7 = 0\text{b}0111$)
- **Bits**: `4`, $2^3 = 8$ (sign-bit mask `0b1000`).
- **Sign-bit**: `7 & 8 == 0` (Sign bit NOT set).
- **Result**: `7` (`0b0111`).

### Example 2: `AsIntN(4, 9)` ($x = 9 = 0\text{b}1001$)
- **Bits**: `4`, sign-bit mask `0b1000`.
- **Sign-bit**: `9 & 8 == 8` (Sign bit IS set).
- **Negative Path**: `truncateAndSubFromPowerOfTwo(4, 9, true)` $\implies -(2^4 - 9) = -(16 - 9) = -7$.
- **Result**: `-7`.

### Example 3: Multi-Limb `AsUintN(65, -0x4000000000000001)`
- **Operand $x$**: $x = -4,611,686,018,427,387,905_{10} = -0\text{x}4000000000000001$.
  - Magnitude 30-bit limbs: `x.Digit(0) = 1`, `x.Digit(1) = 0`, `x.Digit(2) = 1` ($1 \times 2^{60} + 1$). `Sign() = true`.
- **Bits $N$**: `65`.
- **`neededDigits`**: $\lfloor(65 + 29)/30\rfloor = 3$ limbs.
- **Execution Path**: Negative $x \implies \text{truncateAndSubFromPowerOfTwo}(65, x, \text{false})$.
  - `i = 0`: $0 - 1 - 0 = -1 \implies$ digit $0 = 0\text{x}3FFFFFFF$, borrow $= 1$.
  - `i = 1`: $0 - 0 - 1 = -1 \implies$ digit $1 = 0\text{x}3FFFFFFF$, borrow $= 1$.
  - `i = 2` (top digit, `msdBitsConsumed = 65 % 30 = 5`):
    - `msd = 1`. `drop = 32 - 5 = 27`.
    - `minuendMsd = 1 << 5 = 32`.
    - `resultMsd = (32 - 1 - 1) & 31 = 30`.
- **Trimming**: Result digits `[0x3FFFFFFF, 0x3FFFFFFF, 30]`, `Sign() = false`.
- **Final Result**: $30 \times 2^{60} + (2^{60}-1) = 2^{65} - (2^{60}+1) = 36,893,488,147,419,103,231_{10}$.

---

## 10. Value Independence Guarantee for Fast Paths [GO IMPLEMENTATION SPECIFICATION]

- `AsIntN` and `AsUintN` **never modify** input $x$.
- **Fast Path Verification**: On fast paths (e.g. `bits >= kMaxLengthBits` or `x.Length() < neededLength`), the implementation **never returns the input pointer directly**.
- **Requirement**: Always allocate a new `BigInt` copy: `return Copy(x)`.
- **Assertion**:
  $$\text{returnedPointer} \ne \text{inputPointer}$$
  Mutating the returned object (`SetDigit(0, 999)`) must leave the original $x$ 100% byte-for-byte unmodified.

---

## 11. Differential Fuzzing Protocol [GO IMPLEMENTATION SPECIFICATION]

- **Oracle Engine**: Node.js v24 executing `JSBI.asIntN` and `JSBI.asUintN`.
- **Mandatory Bit Widths ($N$)**:
  $$N \in \{0, 1, 29, 30, 31, 59, 60, 61, 2^{30}-1, 2^{30}, 2^{30}+1\}$$
- **Mandatory Operands**:
  - Positive operands ($x \ge 0$)
  - Negative operands ($x < 0$)
  - Zero ($x = 0$)
  - Powers of two ($2^k$)
  - Alternating masks (`0x15555555`, `0x2AAAAAAA`)
  - Multi-limb operands (1 to 32 limbs)
- **Comparison Protocol**: Sign match, length match, digit-by-digit match, canonical zero verification.
- **Immediate Termination**: On any mismatch, print context and execute `os.Exit(1)`.
- **Target Duration**: $60.0\text{s}+$ continuous run.

---

## 12. Complexity Analysis [ENGINEERING GOAL]

Let $K = \text{neededDigits} = \lfloor(N + 29) / 30\rfloor$.

| Function | Time Complexity | Space Complexity | Notes |
| :--- | :--- | :--- | :--- |
| `truncateToNBits` | $O(K)$ | $O(K)$ | Copies and masks $K$ digits. |
| `truncateAndSubFromPowerOfTwo` | $O(K)$ | $O(K)$ | Subtracts $K$ digits from power of 2. |
| `AsIntN` | $O(K)$ | $O(K)$ | Bounds checks + truncation. |
| `AsUintN` | $O(K)$ | $O(K)$ | Bounds checks + truncation. |

---

## 13. Benchmark Expectations & Protocol [GO IMPLEMENTATION SPECIFICATION]

Planned Benchmark Protocol in `tests/port/truncation_test.go`:
```go
func BenchmarkAsIntN(b *testing.B)
func BenchmarkAsUintN(b *testing.B)
```

### Allocation Targets [ENGINEERING GOAL]
- `AsIntN`, `AsUintN`: $\le 64$ B/op, $\le 2$ allocs/op for 1–2 limb BigInt inputs.

---

## 14. Final Self Review [ENGINEERING GOAL]

1. **Weakest Proof**: The mathematical proof that `truncateAndSubFromPowerOfTwo` produces exact two's complement bit wrapping across arbitrary limb boundaries when $N \pmod{30} \ne 0$.
2. **Highest-Risk Implementation Area**: Correct minuend digit calculation in `truncateAndSubFromPowerOfTwo` when `msdBitsConsumed != 0` (lines 1898–1904).
3. **Bug Most Likely to Survive Unit Tests**: Off-by-one bit mask shift in `compareDigit = 1 << ((n - 1) % 30)` when $N$ is an exact multiple of 30.
4. **Easiest Invariant to Violate**: Returning the original input object directly without allocating a copy on fast paths (`n >= kMaxLengthBits`), violating value independence.
5. **Remaining Assumptions**: Assuming JSBI's `kMaxLengthBits = 1 << 30` bit limit fits cleanly within Go `int` capacity.
