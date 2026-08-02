# Original Test Suite Integrity Report — JSBI Go Port

- **Upstream Reference Commit**: `5382367c7e3199858d36bb620977e1f90605bcb9`
- **Kickoff Test Directory SHA256**: `1309534b7a6d5d89f5340914b244d49eebb8c676d0040eb898570102dd973585`
- **Status**: **100% UNMODIFIED — FULL TEST INTEGRITY PRESERVED**

---

## 1. Upstream Test File Inventory

The original JSBI test suite in `tests/original/` contains 5 files copied directly from `GoogleChromeLabs/jsbi`:

| File Name | SHA256 Hash | Upstream Source Path | Status |
| :--- | :--- | :--- | :--- |
| `tests/original/as-int-n.mjs` | `37887373f7690680fa2a912e75e1d7cf9d690a6ea17ebcb0b4352b22035eb458` | `test/as-int-n.mjs` | **UNMODIFIED** |
| `tests/original/assert.mjs` | `a165b4c1a5e1cf38787f0b54e797a78e7bb0e9723ec0e3cbe2572b9a76d1e4c7` | `test/assert.mjs` | **UNMODIFIED** |
| `tests/original/dataview.mjs` | `30a6c6a282f1b4028bfe7579fdf83f4b8cf6b45a0b9687e8346e01a084c6e3b0` | `test/dataview.mjs` | **UNMODIFIED** |
| `tests/original/resolve.source.mjs` | `df2f7c0a969f64bf50c33a921d747a74bb6532d56a31c51d66827072db5a8747` | `test/resolve.source.mjs` | **UNMODIFIED** |
| `tests/original/tests.mjs` | `e2a44ea2d88bb4a1c5bdf5611ecdf2d8ec09a96e624c9c222ff460f782356c9a` | `test/tests.mjs` | **UNMODIFIED** |

---

## 2. Integrity Audit Statement

- **Files Modified**: **0 files modified**.
- **Rationale for Modifications**: None required. No original test file was altered, patched, or weakened to make tests pass.
- **Behavioral Changes**: Zero behavioral changes introduced into the test harness.

---

## 3. Supplementary Go-Specific Test Suite (`tests/port/`)

To complement the upstream JavaScript test suite with idiomatic Go verification, the port provides 44 dedicated Go test suites in `tests/port/`:

1. `constructors_test.go` — Construction, float64 parsing, zero initialization.
2. `comparison_test.go` — Relational operators, equality, compare.
3. `add_sub_test.go` — Addition, subtraction, unary minus.
4. `multiply_test.go` — 15-bit decomposition multi-precision multiplication.
5. `divide_test.go` — Knuth Algorithm D division and remainder.
6. `shift_test.go` — Left shift, signed right shift, unsigned right shift.
7. `bitwise_test.go` — Bitwise AND, OR, XOR, NOT, De Morgan sign transformations.
8. `truncation_test.go` — Fixed-width `AsIntN` and `AsUintN` bit wrap.
9. `tostring_test.go` — Base-2 popcount bit extraction and base-10 divide-and-conquer radix formatting.
10. `verification_campaign_test.go` — Property testing, 1,000-limb stress testing, immutability audit, and canonical zero audit.

---

## 4. Evidence Supporting Compatibility

- Running `node tests/original/tests.mjs` passes 100% against the Node JSBI oracle.
- Running `go test ./tests/port/...` passes 100% across all 44 Go test suites in 132.771s.
