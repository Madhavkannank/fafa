# Design Review 07 — Bitwise Operations (`BitwiseAnd`, `BitwiseOr`, `BitwiseXor`, `BitwiseNot`)

- **Cluster**: 7 — Bitwise Operations (`BitwiseAnd`, `BitwiseOr`, `BitwiseXor`, `BitwiseNot`)
- **Status**: Draft / Pending User Review [GATE]
- **Author**: Lead Engineer (Go Porting Team)
- **Reference**: `GoogleChromeLabs/jsbi` (`jsbi/lib/jsbi.ts` lines 158–165, 345–406, 1297–1410)

---

## 1. Overview & Operational Scope [GO IMPLEMENTATION SPECIFICATION]

Cluster 7 covers arbitrary-precision bitwise logic operations for `BigInt` values:
- `BitwiseAnd(x, y *BigInt) *BigInt`
- `BitwiseOr(x, y *BigInt) *BigInt`
- `BitwiseXor(x, y *BigInt) *BigInt`
- `BitwiseNot(x *BigInt) *BigInt`

### Scope & Dependencies [GO IMPLEMENTATION SPECIFICATION]
- **Primitives Used**: 30-bit limb accessor methods (`Digit`, `SetDigit`, `Length`, `Sign`, `SetSign`, `Trim`, `Copy`, `Zero`), `absoluteAddOne`, `absoluteSubOne`.
- **Dependencies**: None outside `jsbi` package primitives.

---

## 2. ECMAScript Specification Mapping [ECMAScript SPECIFICATION]

Ref: ECMAScript Language Specification (ECMA-262), Sections *BigInt::bitwiseAND*, *BigInt::bitwiseOR*, *BigInt::bitwiseXOR*, *BigInt::bitwiseNOT*.
- ECMAScript defines BigInt bitwise operations as operating on two's complement binary representations with infinite sign extension.
- JSBI maps two's complement representations to magnitude-and-sign arithmetic using De Morgan's laws and the identity $-x = \sim(x - 1)$.

---

## 3. JSBI Source Reference & Line Mapping [JSBI SOURCE]

| JSBI TS Function | JSBI TS Lines | Target Go Identifier |
| :--- | :--- | :--- |
| `JSBI.bitwiseNot(x)` | 158–165 | `BitwiseNot(x *BigInt) *BigInt` |
| `JSBI.bitwiseAnd(x, y)` | 345–363 | `BitwiseAnd(x, y *BigInt) *BigInt` |
| `JSBI.bitwiseOr(x, y)` | 386–406 | `BitwiseOr(x, y *BigInt) *BigInt` |
| `JSBI.bitwiseXor(x, y)` | 365–384 | `BitwiseXor(x, y *BigInt) *BigInt` |
| `JSBI.__absoluteAnd(x, y, result)` | 1297–1324 | `absoluteAnd(x, y, result *BigInt) *BigInt` |
| `JSBI.__absoluteAndNot(x, y, result)` | 1326–1350 | `absoluteAndNot(x, y, result *BigInt) *BigInt` |
| `JSBI.__absoluteOr(x, y, result)` | 1352–1382 | `absoluteOr(x, y, result *BigInt) *BigInt` |
| `JSBI.__absoluteXor(x, y, result)` | 1384–1410 | `absoluteXor(x, y, result *BigInt) *BigInt` |

---

## 4. Control Flow & Explicit Dispatch Tables [JSBI SOURCE]

### 4.1 `BitwiseNot` Dispatch Table [JSBI SOURCE]

| Operand Sign | JSBI Derivation Identity | JSBI Control Flow / Helper Sequence | Result Sign |
| :--- | :--- | :--- | :--- |
| $x \ge 0$ | $\sim x = -(x + 1)$ | `return JSBI.__absoluteAddOne(x, true)` | `true` (Negative) |
| $x < 0$ | $\sim(-x) = x - 1$ | `return JSBI.__absoluteSubOne(x).__trim()` | `false` (Non-negative) |

### 4.2 `BitwiseAnd` Dispatch Table [JSBI SOURCE]

| Sign $x$ | Sign $y$ | Swap Condition | JSBI Control Flow / Helper Sequence | Result Sign |
| :--- | :--- | :--- | :--- | :--- |
| $+x$ | $+y$ | None | `return JSBI.__absoluteAnd(x, y).__trim()` | `false` |
| $-x$ | $-y$ | None | `resLength = max(x.len, y.len) + 1`<br>`res = __absoluteSubOne(x, resLength)`<br>`y1 = __absoluteSubOne(y)`<br>`res = __absoluteOr(res, y1, res)`<br>`return __absoluteAddOne(res, true, res).__trim()` | `true` |
| $+x$ | $-y$ | None | `return __absoluteAndNot(x, __absoluteSubOne(y)).__trim()` | `false` |
| $-x$ | $+y$ | `if (x.sign) [x, y] = [y, x]` | Swaps to $+y \ \& \ (-x)$, then calls `__absoluteAndNot(y, __absoluteSubOne(x)).__trim()` | `false` |

### 4.3 `BitwiseOr` Dispatch Table [JSBI SOURCE]

| Sign $x$ | Sign $y$ | Swap Condition | JSBI Control Flow / Helper Sequence | Result Sign |
| :--- | :--- | :--- | :--- | :--- |
| $+x$ | $+y$ | None | `resLength = max(x.len, y.len)`<br>`return __absoluteOr(x, y).__trim()` | `false` |
| $-x$ | $-y$ | None | `resLength = max(x.len, y.len)`<br>`res = __absoluteSubOne(x, resLength)`<br>`y1 = __absoluteSubOne(y)`<br>`res = __absoluteAnd(res, y1, res)`<br>`return __absoluteAddOne(res, true, res).__trim()` | `true` |
| $+x$ | $-y$ | None | `resLength = max(x.len, y.len)`<br>`res = __absoluteSubOne(y, resLength)`<br>`res = __absoluteAndNot(res, x, res)`<br>`return __absoluteAddOne(res, true, res).__trim()` | `true` |
| $-x$ | $+y$ | `if (x.sign) [x, y] = [y, x]` | Swaps to $+y \ \| \ (-x)$, then runs mixed-sign path above. | `true` |

### 4.4 `BitwiseXor` Dispatch Table [JSBI SOURCE]

| Sign $x$ | Sign $y$ | Swap Condition | JSBI Control Flow / Helper Sequence | Result Sign |
| :--- | :--- | :--- | :--- | :--- |
| $+x$ | $+y$ | None | `return __absoluteXor(x, y).__trim()` | `false` |
| $-x$ | $-y$ | None | `resLength = max(x.len, y.len)`<br>`res = __absoluteSubOne(x, resLength)`<br>`y1 = __absoluteSubOne(y)`<br>`return __absoluteXor(result, y1, res).__trim()` | `false` |
| $+x$ | $-y$ | None | `resLength = max(x.len, y.len) + 1`<br>`res = __absoluteSubOne(y, resLength)`<br>`res = __absoluteXor(res, x, res)`<br>`return __absoluteAddOne(res, true, res).__trim()` | `true` |
| $-x$ | $+y$ | `if (x.sign) [x, y] = [y, x]` | Swaps to $+y \ \hat{} \ (-x)$, then runs mixed-sign path above. | `true` |

---

## 5. Formal Helper Contracts [GO IMPLEMENTATION SPECIFICATION]

### 5.1 Contract: `absoluteAnd(x, y, result *BigInt) *BigInt`
- **Purpose**: Computes magnitude bitwise AND ($x \ \& \ y$) limb by limb up to $\min(\text{len}(x), \text{len}(y))$.
- **Inputs**: $x, y \in \text{BigInt}$ (non-nil). `result`: optional reusable output buffer (or `nil`).
- **Outputs**: `*BigInt` containing raw un-trimmed limb result.
- **Input Modification**: $x$ and $y$ are **never modified**.
- **Buffer Reuse**: If `result` is non-nil, its internal digit slice is mutated and returned. Otherwise, the implementation allocates a new internal result object of length $\min(\text{len}(x), \text{len}(y))$.
- **Post-Call Trimming**: Callers **must call `.Trim()`** on the return value to normalize leading zeros.
- **Output Length**: $N = \min(\text{len}(x), \text{len}(y))$ (or `result.Length()`).
- **Aliasing Guarantee**: The returned instance does **not alias** $x$ or $y$.

### 5.2 Contract: `absoluteAndNot(x, y, result *BigInt) *BigInt`
- **Purpose**: Computes magnitude bitwise AND NOT ($x \ \& \sim y$) limb by limb up to $\text{len}(x)$.
- **Inputs**: $x, y \in \text{BigInt}$ (non-nil). `result`: optional reusable output buffer (or `nil`).
- **Outputs**: `*BigInt` containing raw un-trimmed limb result.
- **Input Modification**: $x$ and $y$ are **never modified**.
- **Buffer Reuse**: If `result` is non-nil, its internal digit slice is mutated and returned. Otherwise, the implementation allocates a new internal result object of length $\text{len}(x)$.
- **Post-Call Trimming**: Callers **must call `.Trim()`** on the return value to normalize leading zeros.
- **Output Length**: $N = \text{len}(x)$ (or `result.Length()`).
- **Aliasing Guarantee**: The returned instance does **not alias** $x$ or $y$.

### 5.3 Contract: `absoluteOr(x, y, result *BigInt) *BigInt`
- **Purpose**: Computes magnitude bitwise OR ($x \ | \ y$) limb by limb up to $\max(\text{len}(x), \text{len}(y))$.
- **Inputs**: $x, y \in \text{BigInt}$ (non-nil). `result`: optional reusable output buffer (or `nil`).
- **Outputs**: `*BigInt` containing raw un-trimmed limb result.
- **Input Modification**: $x$ and $y$ are **never modified**.
- **Buffer Reuse**: If `result` is non-nil, its internal digit slice is mutated and returned. Otherwise, the implementation allocates a new internal result object of length $\max(\text{len}(x), \text{len}(y))$.
- **Post-Call Trimming**: Callers **must call `.Trim()`** on the return value to normalize leading zeros.
- **Output Length**: $N = \max(\text{len}(x), \text{len}(y))$ (or `result.Length()`).
- **Aliasing Guarantee**: The returned instance does **not alias** $x$ or $y$.

### 5.4 Contract: `absoluteXor(x, y, result *BigInt) *BigInt`
- **Purpose**: Computes magnitude bitwise XOR ($x \ \hat{} \ y$) limb by limb up to $\max(\text{len}(x), \text{len}(y))$.
- **Inputs**: $x, y \in \text{BigInt}$ (non-nil). `result`: optional reusable output buffer (or `nil`).
- **Outputs**: `*BigInt` containing raw un-trimmed limb result.
- **Input Modification**: $x$ and $y$ are **never modified**.
- **Buffer Reuse**: If `result` is non-nil, its internal digit slice is mutated and returned. Otherwise, the implementation allocates a new internal result object of length $\max(\text{len}(x), \text{len}(y))$.
- **Post-Call Trimming**: Callers **must call `.Trim()`** on the return value to normalize leading zeros.
- **Output Length**: $N = \max(\text{len}(x), \text{len}(y))$ (or `result.Length()`).
- **Aliasing Guarantee**: The returned instance does **not alias** $x$ or $y$.

---

## 6. Canonical Zero Section [GO IMPLEMENTATION SPECIFICATION]

### Normalization of Zero Results
Bitwise operations between non-zero operands can yield zero (e.g. `BitwiseAnd(12, 3)` $\implies 0$, or `BitwiseXor(x, x)` $\implies 0$).

- **Trimming**: Every public bitwise operation executes `.__trim()` / `.Trim()` on the final `*BigInt` result before returning to the caller.
- **Canonical Zero Representation**: A zero value must have:
  $$\text{Length}() == 0 \implies \text{Sign}() == \text{false}$$
- **Strict Invariant**: Under no circumstances may a bitwise operation return a `*BigInt` with `Length() == 0` and `Sign() == true` (negative zero).

---

## 7. Worked Execution Examples

### Example 1: Multi-Limb `BitwiseAnd(0x100000005, 0x100000003)`
- **Operand $x$**: $x = 0\text{x}100000005 = 4,294,967,301_{10}$.
  - Radixt 30-bit limbs: `x.Digit(0) = 5`, `x.Digit(1) = 4` (since $4 \times 2^{30} + 5 = 4,294,967,301$).
  - `Length() = 2`, `Sign() = false`.
- **Operand $y$**: $y = 0\text{x}100000003 = 4,294,967,299_{10}$.
  - Limbs: `y.Digit(0) = 3`, `y.Digit(1) = 4`.
  - `Length() = 2`, `Sign() = false`.
- **Both Positive Path**: `absoluteAnd(x, y)`
  - `i = 0`: $5 \ \& \ 3 = 1$.
  - `i = 1`: $4 \ \& \ 4 = 4$.
- **Result**: `Digits: [1, 4]`, `Sign: false` ($0\text{x}100000001$).

### Example 2: Multi-Limb `BitwiseXor(0x100000005, 0x100000003)`
- **Operands**: $x = [5, 4]$, $y = [3, 4]$.
- **Both Positive Path**: `absoluteXor(x, y)`
  - `i = 0`: $5 \ \hat{} \ 3 = 6$.
  - `i = 1`: $4 \ \hat{} \ 4 = 0$.
- **Trimming**: Top digit `0` trimmed $\implies$ `Length() = 1`.
- **Result**: `Digits: [6]`, `Sign: false`.

### Example 3: Negative Multi-Limb `BitwiseXor(-0x40000001, -0x40000001)`
- **Operands**: $x = -1,073,741,825_{10}$, $y = -1,073,741,825_{10}$.
  - Magnitude limbs: $x = [1, 1]$, `Sign() = true`.
- **Both Negative Path**: $(-x) \ \hat{} \ (-y) = (x - 1) \ \hat{} \ (y - 1)$.
  - `absoluteSubOne([1, 1])` $\implies [0, 1]$.
  - `absoluteXor([0, 1], [0, 1])` $\implies [0, 0]$.
- **Trimming**: Both zeros trimmed $\implies$ `Length() = 0`.
- **Result**: `Digits: []`, `Sign: false` (Canonical Zero `0n`).

### Example 4: Carry Propagation Across Limbs (`BitwiseNot(0x3FFFFFFF)`)
- **Operand $x$**: $x = 0\text{x}3FFFFFFF = 1,073,741,823_{10}$ (`Digit(0) = 0x3FFFFFFF`).
- **Positive Path**: `absoluteAddOne([0x3FFFFFFF], true)`.
  - `i = 0`: $0\text{x}3FFFFFFF + 1 = 0\text{x}40000000 \implies$ digit $0$, carry $1$.
  - `i = 1`: carry stored in the newly allocated most-significant limb.
- **Result**: `Digits: [0, 1]`, `Sign: true` ($-0\text{x}40000000$).

---

## 8. Value Independence Guarantee [GO IMPLEMENTATION SPECIFICATION]

1. **Input Immutability**: Operands $x$ and $y$ are **never modified** under any bitwise operation.
2. **Result Independence**: Returned `*BigInt` instances **never alias** $x$ or $y$.
3. **Internal Buffer Isolation**: Reusable intermediate `result` buffers passed inside `absoluteAnd`, `absoluteOr`, `absoluteXor`, and `absoluteAndNot` are internal to temporary calculations and never leaked to callers.
4. **Observable Objects**: All user-visible return objects are independent heap-allocated instances.

---

## 9. Differential Fuzzing Protocol [GO IMPLEMENTATION SPECIFICATION]

- **Oracle Engine**: Node.js v24 executing `GoogleChromeLabs/jsbi` ESM distribution (`jsbi/dist/jsbi.mjs`), invoking `JSBI.bitwiseAnd`, `JSBI.bitwiseOr`, `JSBI.bitwiseXor`, and `JSBI.bitwiseNot`.
- **Comparison Protocol**:
  1. **Sign Match**: Assert `goResult.Sign() == oracleRes.Sign`.
  2. **Length Match**: Assert `goResult.Length() == oracleRes.Length`.
  3. **Canonical Zero Assertion**: Assert `goResult.Length() == 0 ==> goResult.Sign() == false`.
  4. **Element-by-Element Limb Match**: Assert `goResult.Digit(i) == oracleRes.Digits[i]` for every 30-bit limb.
  5. **Immediate Termination**: On any mismatch, print context and execute `os.Exit(1)`.
- **Target Duration**: $60.0\text{s}+$ continuous run.
- **Representative Worst-Case Vectors**:
  - Boundary values ($2^{30}-1, 2^{30}, 2^{60}-1$).
  - Negative numbers requiring `absoluteSubOne` borrow propagation.
  - Multi-limb operands where `absoluteSubOne` propagates a borrow across multiple consecutive zero limbs.
  - Alternating bit pattern masks (`0x2AAAAAAA`, `0x15555555`).
  - Bitwise operations yielding zero outputs (`x ^ x`, `x & 0`).

---

## 10. Complexity Analysis [ENGINEERING GOAL]

Let $n = \max(x.\text{Length()}, y.\text{Length()})$ be the maximum limb count of the operands.

| Function | Time Complexity | Space Complexity | Notes |
| :--- | :--- | :--- | :--- |
| `absoluteAnd` | $O(\min(x.\text{Length()}, y.\text{Length()}))$ | $O(\min(x.\text{Length()}, y.\text{Length()}))$ | Bitwise AND of overlapping limbs. |
| `absoluteAndNot` | $O(x.\text{Length()})$ | $O(x.\text{Length()})$ | Bitwise AND NOT of limbs. |
| `absoluteOr` | $O(n)$ | $O(n)$ | Bitwise OR of all limbs. |
| `absoluteXor` | $O(n)$ | $O(n)$ | Bitwise XOR of all limbs. |
| `BitwiseNot` | $O(x.\text{Length()})$ | $O(x.\text{Length()})$ | Calls `absoluteAddOne` or `absoluteSubOne`. |
| `BitwiseAnd` | $O(n)$ | $O(n)$ | Executes magnitude helpers by sign combination. |
| `BitwiseOr` | $O(n)$ | $O(n)$ | Executes magnitude helpers by sign combination. |
| `BitwiseXor` | $O(n)$ | $O(n)$ | Executes magnitude helpers by sign combination. |

---

## 11. Benchmark Expectations & Protocol [GO IMPLEMENTATION SPECIFICATION]

Planned Benchmark Protocol in `tests/port/bitwise_test.go`:
```go
func BenchmarkBitwiseAnd(b *testing.B)
func BenchmarkBitwiseOr(b *testing.B)
func BenchmarkBitwiseXor(b *testing.B)
func BenchmarkBitwiseNot(b *testing.B)
```

### Allocation Targets [ENGINEERING GOAL]
- `BitwiseNot`: $\le 64$ B/op, $\le 2$ allocs/op.
- `BitwiseAnd`, `BitwiseOr`, `BitwiseXor`: $\le 128$ B/op, $\le 4$ allocs/op (accounting for intermediate `absoluteSubOne` buffers on negative paths).

---

## 12. Final Self Review [ENGINEERING GOAL]

1. **Weakest Proof**: The mathematical proof that $(-x) \ | \ (-y) == -(((x-1) \ \& \ (y-1)) + 1)$ holds under finite 30-bit limb sign-extension boundary conditions without limb overflow.
2. **Highest-Risk Implementation Area**: Intermediate `result` buffer reuse in `absoluteOr` and `absoluteAndNot` during multi-step negative sign combinations (where `result` array size must equal $\max(\text{len}(x), \text{len}(y)) + 1$ to prevent index out of bounds).
3. **Bug Most Likely to Survive Unit Tests**: Incorrect borrow propagation in `absoluteSubOne` when subtracting 1 from a multi-limb number with trailing `0` limbs (e.g. `[0, 1] - 1 = [0x3FFFFFFF, 0]`).
4. **Easiest Invariant to Violate**: Forgetting to call `.Trim()` on intermediate helper return values, resulting in leading zero limbs remaining in negative bitwise results.
5. **Remaining Assumptions**: Assuming JSBI's De Morgan transformation for negative BigInt bitwise operations matches ECMAScript two's complement specification across all edge cases.
