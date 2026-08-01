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

## Standing Regression Policy [LOCKED POLICY]
- **Unit Test Regression**: Before any future cluster is marked complete, its full unit test suite must pass, AND the full unit test suite of every previously completed cluster must still pass.
- **Fuzzing Regression Triggers**: A full differential fuzz re-run of a prior cluster is required whenever a new cluster's implementation plausibly touches that prior cluster's invariants (e.g., shared limb arithmetic, shared overflow/carry logic, digit trimming). This requirement must be explicitly noted in the design review for the new cluster. Otherwise, unit-test regression is sufficient.
- **Zero Regression Rule**: No behavioral regression is permitted in any case.

## Architectural Principle Choice [LOCKED POLICY]

**Selected Architecture: Option A — Faithful Limb-Based Go Representation**

*Locked Policy*:
- The Go implementation will mirror JSBI's custom internal representation (`sign` boolean + 30-bit digit slice `[]uint32`).
- All multi-precision arithmetic, shifts, bitwise operations, radix parsing, and base conversion algorithms will be ported faithfully from JSBI's algorithms directly in Go without relying on Go stdlib `math/big.Int` as an internal backend.
- Memory allocation, digit trimming, carry/borrow propagation, and sign semantics will match JSBI line-for-line.
