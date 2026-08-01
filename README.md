# JSBI Go Port

Go port of `GoogleChromeLabs/jsbi` (TypeScript, Apache-2.0).

## Track & Kickoff Metadata
- **Track**: Track C (JS/TS -> Go)
- **Source Repository**: `https://github.com/GoogleChromeLabs/jsbi`
- **Kickoff Commit**: `5382367c7e3199858d36bb620977e1f90605bcb9`
- **Original Test Suite SHA256**: `1309534b7a6d5d89f5340914b244d49eebb8c676d0040eb898570102dd973585`

## Architecture Policy
- **Selected Strategy**: Option A (Faithful Limb-Based Go Representation).
- **Limb Data Model**: 30-bit digit slice (`[]uint32` with mask `0x3FFFFFFF`) + boolean `sign`.

## One-Command Build & Test
```bash
./go_sdk/go/bin/go.exe test -v ./tests/port/...
# or via Docker
docker build -t jsbi-go . && docker run --rm jsbi-go
```

## Status & Progress

| Cluster | Status | Unit Tests | Differential Fuzzing | Fuzz Duration | Survival Rate |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **1. Construction & Parsing** | COMPLETE | 6/6 PASS | 251,000 cases (Element-by-Element Limb Match) | 65.11s | 100% |
| **2. Comparison** | PENDING | - | - | - | - |
| **3. Add / Subtract** | PENDING | - | - | - | - |
| **4. Multiply** | PENDING | - | - | - | - |
| **5. Divide / Remainder** | PENDING | - | - | - | - |
| **6. Shifts** | PENDING | - | - | - | - |
| **7. Bitwise** | PENDING | - | - | - | - |
| **8. asIntN / asUintN** | PENDING | - | - | - | - |
| **9. toString / Radix** | PENDING | - | - | - | - |

- **Original JSBI Test Suite**: 5 files verified passing unmodified on clean checkout.
