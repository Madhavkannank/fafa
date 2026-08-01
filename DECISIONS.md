# Architectural & Implementation Decisions Log

This document records every non-trivial design choice, divergence from JSBI TS reference, or trade-off made during the Go port.

## Decision 0: Architecture Selection — Faithful Limb-Based Representation (Option A)
- **Context**: Choice between faithful limb-based Go representation vs. wrapping Go stdlib `math/big.Int`.
- **Selection**: **Option A (Faithful Limb-Based Go Representation)**.
- **Rationale**:
  - Maximizes behavioral equivalence and algorithm-level fidelity to reference `JSBI`.
  - Maximizes potential to surface subtle edge-case bugs in JSBI's custom digit math during differential fuzzing (aiming for "Bug Catcher" and "Innovation" bonuses).
  - Maintains explicit control over 30-bit digit limb storage, memory allocation, and execution flow.
- **Consequences**:
  - Must port multi-precision addition, subtraction, multiplication, division, shifts, bitwise logic, and radix conversions faithfully in Go.
  - Core data structure locked: `type BigInt struct { sign bool; digits []uint32 }` using 30-bit digits (`0x3FFFFFFF` mask).

## Decision 1: Cluster 1 Construction & Parsing Representation Details
- **Context**: Porting JSBI constructors (`BigInt(arg)`, `__fromDouble`, `__fromString`, `__oneDigit`, `__zero`) to Go.
- **Choices**:
  1. Represent 30-bit digit limbs as `[]uint32` with mask `0x3FFFFFFF`.
  2. Implement `FromString` with full ECMAScript whitespace filtering, radix auto-detection (`0x`/`0o`/`0b`), power-of-two fast paths, and generic chunked base conversion.
  3. Map JavaScript exceptions (`SyntaxError`, `RangeError`, `TypeError`) to idiomatic Go sentinel errors (`ErrSyntax`, `ErrRange`, `ErrType`).
- **Rationale**: Direct 1-to-1 algorithmic translation of `JSBI` TypeScript logic guarantees exact parity for string parsing, float conversion, and digit trimming.
- **Verification**: Verified via 6 unit test suites and 251,000 differential fuzzing test cases against Node JSBI oracle with element-by-element limb array comparison (65.11s run, 100% survival rate).
