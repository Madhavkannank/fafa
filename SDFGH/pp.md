# Project Constitution (pp.md)

## Repository Metadata
- **Track**: Track C (JS/TS -> Go)
- **Source Repository**: `https://github.com/GoogleChromeLabs/jsbi`
- **Kickoff Commit**: `5382367c7e3199858d36bb620977e1f90605bcb9`
- **Kickoff Test Suite Hash**: `1309534b7a6d5d89f5340914b244d49eebb8c676d0040eb898570102dd973585`

## Scoring Priorities
- **Functionality & Reliability — 40%**: One-command build, passes original JSBI test suite unmodified.
- **Behavioral Equivalence — 30%**: Differential fuzz survival, honest benchmarks with stated methodology.
- **Code Quality — 20%**: Idiomatic Go, minimal/no `unsafe`, DECISIONS.md quality, native Go error handling.
- **Innovation — 10%**: Latent bugs found in the original via fuzzing, defensible architectural choices.
- **Target Time Split**: 40% port, 30% fuzz/property tests, 15% bench, 10% DECISIONS.md, 5% README/write-up.

## Truth Contract
Never fabricate: tests, benchmarks, fuzz results, coverage, progress percentages, logs, issues, commits, or statistics.

- If unknown -> say `UNKNOWN`.
- If not yet executed -> say `NOT EXECUTED`.
- If inferred rather than measured -> say `INFERENCE`, and say from what.
- Every number in README.md, DECISIONS.md, or any status report must trace back to real command output you just ran. No exceptions, no "should be around X."

The judges have said explicitly: an honest partial result beats a confident unverified claim. Treat that as binding.

## Coding Standards & Policies
- All port code written during the hackathon window.
- No source-language runtime: Node/JS is allowed ONLY as an oracle inside the fuzz harness (test-time only).
- Original test suite (`tests/original/`): Unmodified ideal.
- One-command build: `make` / `docker`.

## Architectural Principle Choice [PENDING USER DECISION]

Before Cluster 1 starts, the user must choose between Option A and Option B:

### Option A: Faithful Limb-Based Go Representation
*Description*: Direct port of JSBI's custom `BigInt` internal structure (sign flag + array of 32-bit/64-bit digit limbs) and algorithms (manual multi-precision addition, subtraction, Karatsuba/schoolbook multiplication, Knuth division, bitwise limb logic, radix conversion).

*Objective Tradeoffs*:
- **Pros**:
  - Direct structural and algorithmic fidelity to JSBI reference codebase.
  - High potential for "Innovation" score (10%) and "Bug Catcher" bonus (+3) by surfacing subtle edge-case bugs/divergences in JSBI's low-level digit array math.
  - No reliance on Go stdlib internal representation; explicit control over memory allocation and digit math.
- **Cons**:
  - Substantially higher code surface area and algorithm complexity across all 9 clusters.
  - Increased risk of manual implementation bugs within the hackathon time constraint.

### Option B: `math/big.Int`-Backed Wrapper
*Description*: Internal data model delegates underlying multi-precision arithmetic to Go standard library `math/big.Int`, while providing a complete JSBI-compatible API wrapper handling JSBI semantics, parameter conversions, string formatting, `asIntN`/`asUintN` bitwise operations, and exception/error contracts.

*Objective Tradeoffs*:
- **Pros**:
  - Leverages Go stdlib's battle-tested, assembly-optimized multi-precision math engine for high speed, reliability, and correctness.
  - Reduced implementation overhead for core math, allowing more hackathon time for exhaustive differential fuzzing, benchmarking, and documentation.
  - Clean, idiomatic Go wrapper API.
- **Cons**:
  - Does not mirror JSBI's internal digit array layout.
  - Lower probability of finding low-level limb bugs in JSBI itself during fuzzing.
  - Requires careful translation layer for JSBI-specific bitwise and two's complement mask operations (`asIntN`, `asUintN`) on `math/big.Int`.

*Status*: **Awaiting User Selection**. Once chosen, this policy will be locked in `pp.md` and `DECISIONS.md`.
