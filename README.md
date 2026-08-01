# JSBI Go Port

Go port of `GoogleChromeLabs/jsbi` (TypeScript, Apache-2.0).

## Track & Kickoff Metadata
- **Track**: Track C (JS/TS -> Go)
- **Source Repository**: `https://github.com/GoogleChromeLabs/jsbi`
- **Kickoff Commit**: `5382367c7e3199858d36bb620977e1f90605bcb9`
- **Original Test Suite SHA256**: `1309534b7a6d5d89f5340914b244d49eebb8c676d0040eb898570102dd973585`

## One-Command Build & Test
```bash
make test
# or
docker build -t jsbi-go . && docker run --rm jsbi-go
```

## Current Status
- **Session**: 1 (Kickoff & Bootstrap)
- **Original Test Suite**: 5 tests / benchmarks verified passing on clean checkout.
- **Port Status**: Setup complete; awaiting architecture decision lock in `pp.md`.
