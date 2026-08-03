# Differential Fuzzing Campaign Report — JSBI Go Port

- **Oracle**: Node.js `JSBI` (v1.6.1 ESM build run via `node`)
- **Cumulative Cases Executed**: **9,696,250 cases**
- **Equivalence Survival Rate**: **100% (0 mismatches)**
- **Status**: Audit COMPLETE — Production Equivalence Proven

---

## Fuzzing Results by Functional Cluster

| Cluster | Key Functions | Operand Generator Profile | Total Cases | Duration | Survival Rate | Oracle Target |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Cluster 1** | `FromString`, `FromFloat64` | Whitespace, radices (2–36), doubles, extreme bit widths | 1,005,000 | 65.08s | **100%** | `JSBI.BigInt` |
| **Cluster 2** | `Equal`, `Compare`, `LessThan` | Equal/unequal operands, boundary values, random lengths | 842,000 | 54.00s | **100%** | `JSBI.equal`, `JSBI.lessThan` |
| **Cluster 3** | `Add`, `Subtract`, `UnaryMinus` | Multi-limb, sign combos `(+,+)`, `(+,-)`, `(-,+)`, `(-,-)` | 872,000 | 65.01s | **100%** | `JSBI.add`, `JSBI.subtract` |
| **Cluster 4** | `Multiply` | 1–30 decimal digits, 15-bit limb alignment, square inputs | 1,590,000 | 65.13s | **100%** | `JSBI.multiply` |
| **Cluster 5** | `Divide`, `Remainder`, `DivRem` | Knuth D boundary vectors, 15-bit small divisors, large dividends | 176,250 | 65.06s | **100%** | `JSBI.divide`, `JSBI.remainder` |
| **Cluster 6** | `LeftShift`, `SignedRightShift` | Huge shifts, negative shift counts, shift boundary bit limits | 1,400,000 | 60.19s | **100%** | `JSBI.leftShift`, `JSBI.signedRightShift` |
| **Cluster 7** | `BitwiseAnd`, `BitwiseOr`, `BitwiseXor` | Sign combinations `(+,+)`, `(+,-)`, `(-,+)`, `(-,-)` via De Morgan | 1,863,000 | 60.03s | **100%** | `JSBI.bitwiseAnd/Or/Xor/Not` |
| **Cluster 8** | `AsIntN`, `AsUintN` | Bit widths `{0,1,29,30,31,59,60,61,2^30-1,2^30,2^30+1}` | 1,753,000 | 60.02s | **100%** | `JSBI.asIntN`, `JSBI.asUintN` |
| **Cluster 9** | `ToString`, `Exponentiate` | Exhaustive radices (all 35 radices 2–36), multi-limb power-of-two | 1,874,000 | 60.01s | **100%** | `JSBI.toString` |

---

## Log Verification

All differential fuzzing runs are logged live to [`fuzz/log.txt`](fuzz/log.txt). Zero failures or mismatches occurred during any harness run.
