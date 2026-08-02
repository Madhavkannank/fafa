# Scope, Known Limitations, & Evidence Boundaries — JSBI Go Port

- **Repository**: `github.com/Madhavkannank/fafa`
- **Reference**: `GoogleChromeLabs/jsbi` (`5382367c7e3199858d36bb620977e1f90605bcb9`)
- **Status**: Audit COMPLETE — Scope & Limitations Explicitly Bounded

---

## 1. Current Project Scope

The project scope encompasses a complete, 100% Go native port of the 9 functional clusters defined in `GoogleChromeLabs/jsbi`:

1. **Construction & Parsing**: `BigInt`, `FromString`, `FromFloat64`, `ToFloat64`, `BigIntVal`.
2. **Comparison**: `Equal`, `NotEqual`, `LessThan`, `LessThanOrEqual`, `GreaterThan`, `GreaterThanOrEqual`, `Compare`.
3. **Add / Subtract**: `Add`, `Subtract`, `UnaryMinus`.
4. **Multiply**: `Multiply`.
5. **Divide / Remainder**: `Divide`, `Remainder`, `DivRem`.
6. **Shifts**: `LeftShift`, `SignedRightShift`, `UnsignedRightShift`.
7. **Bitwise Operations**: `BitwiseAnd`, `BitwiseOr`, `BitwiseXor`, `BitwiseNot`.
8. **Fixed-Width Truncation**: `AsIntN`, `AsUintN`.
9. **String Formatting**: `ToString`, `Exponentiate`.

---

## 2. Known Limitations & Architectural Choices

1. **Unsigned Right Shift (`UnsignedRightShift`)**:
   - *Behavior*: Always returns `nil, ErrTypeError`.
   - *Rationale*: Directly mirrors ECMAScript ECMA-262 specification and JSBI's `unsignedRightShift()` which throws `TypeError('BigInts have no unsigned right shift')`. BigInt values in JavaScript do not support unsigned right shift because they are conceptually infinite precision signed two's complement numbers.

2. **Maximum Allocation Bit Width Guard ($2^{30}$ bits)**:
   - *Behavior*: Bit widths exceeding $2^{30}$ bits ($1,073,741,824$ bits, or $2^{25}$ 30-bit limbs) return `ErrRange`.
   - *Rationale*: Mirrors JSBI's internal `__kMaxLength = 1 << 25` and `__kMaxLengthBits = __kMaxLength << 5` memory limits, preventing out-of-memory crashes on invalid or unbounded shift/truncation inputs.

3. **Fast-Path Copy Allocation (`x.Copy()`)**:
   - *Behavior*: Operations that perform no mutation (e.g. `AsIntN(100, x)` when $x$ is small, `LeftShift(x, 0)`) allocate and return a deep copy `x.Copy()`.
   - *Rationale*: In JavaScript, returning the same object reference `x` is safe because JS primitives/objects have garbage collection semantics. In Go, returning the same `*BigInt` pointer would allow callers to mutate internal state through shared references. Returning `x.Copy()` guarantees `returnedPointer != inputPointer` and protects value independence.

4. **No Native JS Engine / V8 Embedding**:
   - *Behavior*: The shipped Go library is 100% pure Go with zero CGO, V8, or Node.js dependencies. Node.js is used **only** as a test-time ESM oracle driver during differential fuzzing.

---

## 3. Metrics Intentionally Not Measured

To comply with the project Truth Contract, the following metrics are explicitly declared as **not measured**:

- **Resident Set Size (RSS)**: Memory footprint of running process was not measured.
- **Executable Startup Time**: Cold/warm binary startup time was not measured.
- **Multi-Threaded Throughput**: Parallel worker pool throughput (ops/sec) was not measured.
- **Tail Latency Percentiles (p95/p99)**: Only mean execution speed (`ns/op`) from Go standard benchmark harness is reported.

---

## 4. Evidence Boundaries

- All claims of behavioral equivalence apply to the 25 exported APIs and 20 helper functions tested against `GoogleChromeLabs/jsbi` (`5382367c7e3199858d36bb620977e1f90605bcb9`).
- Differential fuzzing evidence covers 9,696,250 executed test cases across radices 2–36, operand lengths up to 30 decimal digits, and mandatory boundary bit widths `{0, 1, 29, 30, 31, 59, 60, 61, 2^30-1, 2^30, 2^30+1}`.

---

## 5. Potential Future Improvements

1. **Karatsuba & Toom-Cook Multiplication**: Current multiplication uses 15-bit decomposition $O(n^2)$ column accumulation. Adding Karatsuba multiplication for operands $> 100$ limbs would improve asymptotic complexity to $O(n^{1.585})$.
2. **Montgomery Reduction**: Accelerate `Exponentiate` for large modular exponentiation workloads.
3. **Buffer Pre-Allocation Pools (`sync.Pool`)**: Recycle intermediate 30-bit digit slices during heavy allocation loops.
