# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]
### Added
- Implemented Cluster 1 (Construction & Parsing) in `src/`: `BigInt` core struct, `FromString`, `FromInt64`, `FromUint64`, `FromFloat64`, `FromBool`, `BigIntVal`, and error definitions (`ErrSyntax`, `ErrRange`, `ErrType`).
- Added Cluster 1 unit test suite in `tests/port/construction_test.go` (6/6 test suites passing).
- Built Cluster 1 differential fuzzing harness in `fuzz/harness/fuzz_cluster1.go` and Node JSBI oracle `fuzz/harness/oracle.mjs`.
- Completed 65.00s differential fuzzing run (516,000 cases executed, 100% equivalence survival).
- Initialized workspace structure, `.port-mortem.toml`, and `SDFGH` engineering workspace.
- Preserved original JSBI test suite under `tests/original/` (kickoff SHA256: `1309534b7a6d5d89f5340914b244d49eebb8c676d0040eb898570102dd973585`).

### Changed
- Locked architecture choice in `pp.md` and `DECISIONS.md` to **Option A: Faithful Limb-Based Go Representation**.
