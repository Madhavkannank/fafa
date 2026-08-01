# Architectural & Implementation Decisions Log

This document records every non-trivial design choice, divergence from JSBI TS reference, or trade-off made during the Go port.

## Decision 0: Kickoff Architecture Decision [PENDING USER CHOICE]
- **Options**:
  1. Faithful limb-based Go representation (mirroring JSBI's internal sign + 32-bit/64-bit digit-array structure).
  2. `math/big.Int`-backed wrapper implementing JSBI's API surface.
- **Status**: Pending explicit user lock in `pp.md`.
