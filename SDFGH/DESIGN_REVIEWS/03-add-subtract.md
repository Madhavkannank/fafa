# Design Review 03: Add & Subtract

- **Status**: REVIEW COMPLETE — Awaiting User Approval (GATE)
- **Cluster**: 3 — Add / Subtract
- **References**: `jsbi/lib/jsbi.ts` lines 151–156, 269–297, 611–620, 1215–1295.

---

## 1. Objectives & Target API Scope
Cluster 3 implements multi-precision addition, subtraction, and unary negation for `BigInt` instances, matching JSBI reference semantics and ECMAScript specifications.

### Target Go API Surface
```go
package jsbi

// Public Arithmetic Functions
func Add(x, y *BigInt) *BigInt
func Subtract(x, y *BigInt) *BigInt
func UnaryMinus(x *BigInt) *BigInt

// Internal Helper Functions
func absoluteAdd(x, y *BigInt, resultSign bool) *BigInt
func absoluteSub(x, y *BigInt, resultSign bool) *BigInt
func absoluteAddOne(x, y *BigInt, resultSign bool) *BigInt
func absoluteSubOne(x *BigInt) *BigInt
```

---

## 2. Mathematical & Language Spec Proofs for Carry and Borrow

### 2.1 Carry Proof (`carry = uint32(r) >> 30`)
- **Context**: In multi-precision addition across 30-bit limbs ($0 \le d < 2^{30}$):
  $$r = \text{int32}(x.\text{Digit}(i)) + \text{int32}(y.\text{Digit}(i)) + \text{int32}(\text{carry}_{\text{in}})$$
- **Exact Go Type**: `r` is `int32` (signed 32-bit integer).
- **Bounds of $r$**:
  - Minimum possible value: $0 + 0 + 0 = 0$.
  - Maximum possible value: $(2^{30} - 1) + (2^{30} - 1) + 1 = 2 \times 2^{30} - 1 = 0x7FFFFFFF = 2147483647$.
  - Therefore, $r \in [0, 2^{31} - 1]$.
- **Evaluation**:
  - `uint32(r) >> 30` performs a **logical right shift** per Go Language Specification ("Shift operators implement logical shifts if the left operand is an unsigned integer").
  - If $0 \le r \le 2^{30} - 1 = 1073741823$: `uint32(r) >> 30` = `0`.
  - If $2^{30} \le r \le 2^{31} - 1 = 2147483647$: `uint32(r) >> 30` = `1`.
- **Conclusion**: $\text{carry} \in \{0, 1\}$ for all valid limb additions. Overflow into bit 31 never occurs.

### 2.2 Borrow Proof (`borrow = (uint32(r) >> 30) & 1`)
- **Context**: In multi-precision subtraction across 30-bit limbs ($0 \le d < 2^{30}$):
  $$r = \text{int32}(x.\text{Digit}(i)) - \text{int32}(y.\text{Digit}(i)) - \text{int32}(\text{borrow}_{\text{in}})$$

#### Proof & Go Language Specification Verification:
1. **Exact Go Type of $r$**: `int32` (signed 32-bit integer).
2. **Signed vs Unsigned**: $r$ is signed (`int32`) during expression evaluation.
3. **Go Shift Behavior**:
   - Per Go Spec (Section: *Operators - Shift operators*):
     > "The shift operators shift the left operand by the shift count... They implement logical shifts if the left operand is an unsigned integer."
   - Casting `uint32(r)` converts $r$ to an unsigned 32-bit integer and guarantees a **logical right shift** in Go (`>> 30`).
4. **Valid Range of $r$**:
   - Minimum value (underflow): $0 - (2^{30} - 1) - 1 = -2^{30} = -1073741824 = \text{int32}(0xC0000000)$.
   - Maximum value (no underflow): $(2^{30} - 1) - 0 - 0 = 2^{30} - 1 = 1073741823 = \text{int32}(0x3FFFFFFF)$.
   - Therefore, $r \in [-2^{30}, 2^{30} - 1]$.
5. **Bit-Level Proof for JS `(r >>> 30) & 1` Equivalence**:
   - **Case A ($r \ge 0$, No Underflow)**:
     - $0 \le r \le 2^{30}-1 = 0x3FFFFFFF$.
     - Bit 31 = `0`, Bit 30 = `0`.
     - `uint32(r) >> 30` = `0`. `(0) & 1` = `0`. $\text{borrow} = 0$.
   - **Case B ($r < 0$, Underflow)**:
     - $-2^{30} \le r \le -1$.
     - Per Go Spec (Section: *Conversions*): Converting negative `int32` to `uint32` preserves its 32-bit two's complement representation.
     - For $r = -2^{30} = \text{int32}(0xC0000000)$, `uint32(r)` = `0xC0000000` (bits 31 and 30 are `11`).
     - For $r = -1 = \text{int32}(0xFFFFFFFF)$, `uint32(r)` = `0xFFFFFFFF` (bits 31 and 30 are `11`).
     - For all $r \in [-2^{30}, -1]$, bit 30 of `uint32(r)` is **always `1`**.
     - Logical right shift `uint32(r) >> 30` shifts bits 31..30 into bits 1..0, yielding `3` (`0b11`).
     - `(uint32(r) >> 30) & 1` evaluates to `3 & 1 = 1`. $\text{borrow} = 1$.
6. **Conclusion**: `borrow := (uint32(r) >> 30) & 1` produces **identical results** to JSBI's `(r >>> 30) & 1` for every possible value of $r \in [-2^{30}, 2^{30}-1]$, with zero inferences.

---

## 3. `absoluteAdd` Step-by-Step Multi-Limb Walkthrough

### Test Case: $(0x3FFFFFFF, 0x3FFFFFFF, 0x3FFFFFFF) + (1)$

```text
Operand X: [0x3FFFFFFF, 0x3FFFFFFF, 0x3FFFFFFF] (3 limbs)
Operand Y: [0x00000001]                         (1 limb)
Initial resultLength = x.length = 3
CLZ Check: clzmsd(X) = 0 -> resultLength allocated = 4
```

1. **Limb Index 0**:
   - $r = X[0] + Y[0] + \text{carry}_0 = 0x3FFFFFFF + 0x00000001 + 0 = 0x40000000$.
   - $\text{carry}_1 = \text{uint32}(0x40000000) \gg 30 = 1$.
   - $\text{result}[0] = 0x40000000 \ \& \ 0x3FFFFFFF = 0x00000000$.
2. **Limb Index 1**:
   - $r = X[1] + 0 + \text{carry}_1 = 0x3FFFFFFF + 0 + 1 = 0x40000000$.
   - $\text{carry}_2 = \text{uint32}(0x40000000) \gg 30 = 1$.
   - $\text{result}[1] = 0x40000000 \ \& \ 0x3FFFFFFF = 0x00000000$.
3. **Limb Index 2**:
   - $r = X[2] + 0 + \text{carry}_2 = 0x3FFFFFFF + 0 + 1 = 0x40000000$.
   - $\text{carry}_3 = \text{uint32}(0x40000000) \gg 30 = 1$.
   - $\text{result}[2] = 0x40000000 \ \& \ 0x3FFFFFFF = 0x00000000$.
4. **Limb Index 3 (Carry Flush)**:
   - $\text{result}[3] = \text{carry}_3 = 1$.
5. **Final Result & `Trim()`**:
   - Pre-trim digits: `[0x00000000, 0x00000000, 0x00000000, 0x00000001]`.
   - `Trim()` checks MSD `result[3] = 1 != 0` -> No limbs popped.
   - Final `BigInt`: `len = 4, digits = [0, 0, 0, 1]`.

---

## 4. `absoluteSub` Step-by-Step Borrow Chain Walkthrough

### Test Case: $(1, 0, 0, 0) - (1)$

```text
Operand X: [0x00000001, 0x00000000, 0x00000000, 0x00000001] (4 limbs)
Operand Y: [0x00000002]                                     (1 limb)
Initial resultLength = x.length = 4
```

1. **Limb Index 0**:
   - $r = \text{int32}(X[0] - Y[0] - \text{borrow}_0) = 1 - 2 - 0 = -1 = \text{int32}(0xFFFFFFFF)$.
   - $\text{borrow}_1 = (\text{uint32}(-1) \gg 30) \ \& \ 1 = (0xFFFFFFFF \gg 30) \ \& \ 1 = 3 \ \& \ 1 = 1$.
   - $\text{result}[0] = -1 \ \& \ 0x3FFFFFFF = 0x3FFFFFFF$.
2. **Limb Index 1**:
   - $r = \text{int32}(X[1] - 0 - \text{borrow}_1) = 0 - 0 - 1 = -1$.
   - $\text{borrow}_2 = 1$.
   - $\text{result}[1] = 0x3FFFFFFF$.
3. **Limb Index 2**:
   - $r = \text{int32}(X[2] - 0 - \text{borrow}_2) = 0 - 0 - 1 = -1$.
   - $\text{borrow}_3 = 1$.
   - $\text{result}[2] = 0x3FFFFFFF$.
4. **Limb Index 3**:
   - $r = \text{int32}(X[3] - 0 - \text{borrow}_3) = 1 - 0 - 1 = 0$.
   - $\text{borrow}_4 = (\text{uint32}(0) \gg 30) \ \& \ 1 = 0$.
   - $\text{result}[3] = 0$.
5. **Final `Trim()` Execution**:
   - Pre-trim digits: `[0x3FFFFFFF, 0x3FFFFFFF, 0x3FFFFFFF, 0x00000000]`.
   - `Trim()` sees MSD `result[3] == 0` -> pops `result[3]`.
   - MSD `result[2] == 0x3FFFFFFF != 0` -> loop terminates.
   - Final `BigInt`: `len = 3, digits = [0x3FFFFFFF, 0x3FFFFFFF, 0x3FFFFFFF]`.

#### Proof of Final Borrow Zero Guarantee
`absoluteSub` is called only when $|X| \ge |Y|$ (guaranteed by `absoluteCompare` check in `Add` and `Subtract`). Because $X \ge Y$, the total magnitude subtraction $X - Y \ge 0$, so net underflow cannot occur across the full length of $X$. Thus, after the last limb $i = \text{len}(X) - 1$, the final borrow bit $\text{borrow}_{\text{final}}$ **must always be 0**.

---

## 5. `UnaryMinus` Semantics & Canonical Zero

### 5.1 JSBI Source Reference (`jsbi/lib/jsbi.ts` lines 151–156, 611–620)
```typescript
151:   static unaryMinus(x: JSBI): JSBI {
152:     if (x.length === 0) return x;
153:     const result = x.__copy();
154:     result.sign = !x.sign;
155:     return result;
156:   }

611:   __trim(): this {
612:     let newLength = this.length;
613:     let last = this[newLength - 1];
614:     while (last === 0) {
615:       newLength--;
616:       last = this[newLength - 1];
617:       this.pop();
618:     }
619:     if (newLength === 0) this.sign = false;
620:     return this;
621:   }
```

### 5.2 Canonical Zero Questions & Answers
- **Does `UnaryMinus(0)` return canonical zero?**
  - **YES**. Sourced from line 152: `if (x.length === 0) return x;`. Since $0$ is represented as `length = 0, sign = false`, `UnaryMinus` returns `x` directly without toggling `sign`.
- **Is negative zero (`-0`) ever representable in memory?**
  - **NO**. Canonical zero strictly requires `sign = false` and `len = 0`. In `__trim()` (line 619), `if (newLength === 0) this.sign = false;` guarantees that whenever an arithmetic operation reduces a BigInt to zero digits, its sign is reset to `false`. Negative zero cannot exist.

---

## 6. Allocation Behavior
- **GO IMPLEMENTATION GOAL (Not a JSBI-Sourced Invariant)**: Addition and subtraction operations produce new `*BigInt` allocations. Re-using scratch buffer space where safe is allowed internally, but public API calls (`Add(x, y)`, `Subtract(x, y)`) return newly allocated instances.
- **Benchmarking Plan**: Allocation performance and heap allocations per operation will be empirically measured and verified via `go test -bench=BenchmarkAdd -benchmem ./tests/port/...` after code implementation.

---

## 7. Dedicated Arithmetic Invariants Section

| Invariant Name | Source | Why It Exists | What Breaks If Violated |
| :--- | :--- | :--- | :--- |
| **Carry Invariant** | JSBI line 1228 | $0 \le \text{carry} \le 1$ during `absoluteAdd`. | Bit 31 overflow corrupts unsigned limb math. |
| **Borrow Invariant** | JSBI line 1250 | $0 \le \text{borrow} \le 1$ during `absoluteSub`. | Bitwise sign-extension corrupts borrow subtraction. |
| **Limb Invariant** | JSBI line 1991 | Every limb $d_i$ satisfies $0 \le d_i \le 0x3FFFFFFF$. | Un-masked top bits break 30-bit digit boundary assumptions. |
| **Canonical-Zero Invariant**| JSBI lines 593, 619 | Zero is strictly `sign = false` and `len = 0`. | `0 == -0` comparison fails; invalid limb indexing on zero. |
| **Trim Invariant** | JSBI lines 611–620 | MSD $d_{\text{len}-1} \neq 0$ for all non-zero `BigInt`s. | Incorrect bit length calculations; flawed magnitude comparisons. |
| **Sign Invariant** | JSBI line 271 | Sign of `absoluteAdd` / `absoluteSub` matches operational sign rules. | Addition/subtraction results have inverted signs. |
| **Result-Length Invariant**| JSBI line 1219 | $\text{len}(\text{Add}) \le \max(\text{len}(x), \text{len}(y)) + 1$. | Buffer overflow during carry propagation. |

---

## 8. Differential Fuzzing Plan for Cluster 3

- **Harness File**: `fuzz/harness/fuzz_cluster3.go`
- **Oracle Helper**: `fuzz/harness/oracle_cluster3.mjs` (invoking `JSBI.add(x, y)` and `JSBI.subtract(x, y)`).
- **Target Duration**: 65+ continuous seconds.
- **Generators & Vectors**:
  1. **Random Operands**: 1 to 500 limbs.
  2. **Sign Combinations**: $(+X)+(+Y)$, $(-X)+(-Y)$, $(+X)+(-Y)$, $(-X)+(+Y)$, $(+X)-(-Y)$, $(-X)-(-Y)$.
  3. **Equal Operands**: $X + X$, $X - X \rightarrow 0$.
  4. **Special Numbers**: $0$, $1$, $-1$, Max Limb ($0x3FFFFFFF$).
  5. **Alternating Patterns**: `[0x3FFFFFFF, 0, 0x3FFFFFFF, 0]`.
  6. **Carry Chains**: `[0x3FFFFFFF, 0x3FFFFFFF, 0x3FFFFFFF] + 1`.
  7. **Borrow Chains**: `[1, 0, 0, 0] - 1`.
  8. **Algebraic Inverses**:
     - $\text{Subtract}(\text{Add}(X, Y), Y) == X$
     - $\text{Add}(\text{Subtract}(X, Y), Y) == X$

---

## 9. Cluster Interaction & Downstream Bug Propagation

| Future Cluster | Dependency on Cluster 3 | How an Arithmetic Bug Propagates |
| :--- | :--- | :--- |
| **Cluster 4: Multiply** | Uses `absoluteAdd` / `multiplyAccumulate` for product column summation. | Unhandled carry in `absoluteAdd` causes silent product corruption in large multiplications. |
| **Cluster 5: Division** | Uses `inplaceSub` / `inplaceAdd` for Knuth D quotient correction steps. | Incorrect borrow propagation causes quotient digit overflow or infinite division loops. |
| **Cluster 7: Bitwise** | Uses `absoluteAddOne` and `absoluteSubOne` to convert negative numbers to/from two's complement. | Flawed borrow/carry during $\pm 1$ corrupts `bitwiseAnd`, `bitwiseOr`, `bitwiseXor` for negative inputs. |
| **Cluster 8: asIntN / asUintN** | Uses `truncateAndSubFromPowerOfTwo` for wrap-around. | Borrow error in power-of-two subtraction produces invalid two's complement truncation. |

---

## 10. Self Review

### What implementation detail is most likely to be subtly wrong despite passing unit tests?
**Equal magnitude subtraction sign assignment in `Subtract`**.
When subtracting $X - Y$ where $|X| == |Y|$ and signs differ in input, `absoluteSub(x, y, sign)` returns $0$. If `Trim()` fails to reset `result.sign = false` when `len == 0`, the output becomes $-0$ (`sign = true, len = 0`), which passes standard limb loops but violates the **Canonical-Zero Invariant** and fails subtle equality checks in downstream clusters.

---

## 11. Truth Contract & Spec Labeling
- **JSBI Source Property**: Carry formula, borrow formula logic, sign dispatch rules, digit trimming (`trim()`), canonical zero representation (`len=0, sign=false`).
- **GO IMPLEMENTATION SPECIFICATION**:
  - `r` type is `int32` (signed 32-bit integer).
  - Expression `borrow := (uint32(r) >> 30) & 1` explicitly casts `r` to `uint32` to enforce a **logical right shift** per Go Language Specification (Section *Operators - Shift operators*), proving exact $1$-to-$1$ parity with JS `(r >>> 30) & 1` for all $r \in [-2^{30}, 2^{30}-1]$.
- **GO IMPLEMENTATION GOAL**: Allocation behavior ($0 \text{ allocs/op}$ on pure helpers vs new allocations for public operations).
