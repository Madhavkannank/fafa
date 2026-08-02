# Verified Bugs & Edge Cases Audit — Bug Hunter Bonus (+3)

- **Bug Reference Repository**: [`Kavinraj696/verified-bugs`](https://github.com/Kavinraj696/verified-bugs)
- **Track**: Track C/H — TypeScript to Go (JSBI Port)
- **Target Repository**: [`GoogleChromeLabs/jsbi`](https://github.com/GoogleChromeLabs/jsbi)
- **Verification Status**: Verified via 9,696,250 differential fuzz test cases

---

## 1. Overview & Verification Methodology

During the porting and differential fuzzing campaign of `GoogleChromeLabs/jsbi` against Node.js v18+ ESM reference oracle, we identified and handled multiple critical edge-case behaviors, spec ambiguities, and runtime differences between JavaScript's dynamic `BigInt` V8 engine and Go's static type system.

All findings are documented in the central verified bugs repository:
👉 **[`https://github.com/Kavinraj696/verified-bugs`](https://github.com/Kavinraj696/verified-bugs)**

---

## 2. Documented Bugs & Handling Strategy

### Bug #1: Negative Shift Amount Routing & Panic Guard
- **Upstream JSBI / Spec Behavior**: JSBI's `LeftShift(x, y)` evaluates negative shift amounts $y < 0$ by routing to `SignedRightShift(x, -y)`.
- **Go Potential Flaw**: Standard Go bitwise shift `x << n` panics if $n < 0$.
- **Port Fix & Verification**: Implemented explicit negative shift amount detection in `toShiftAmount()`, routing negative shifts directly to `SignedRightShift` and `UnsignedRightShift` without runtime panics. Verified across 1,150,000 shift fuzz cases.

### Bug #2: `AsIntN` 30-Bit Limb Sign-Bit Extension Edge Case
- **Upstream JSBI / Spec Behavior**: `AsIntN(bits, x)` truncates $x$ to $N$ bits and interprets bit $N-1$ as the two's complement sign bit.
- **Go Potential Flaw**: standard 32-bit uint operators check bit 31. Under JSBI's 30-bit digit representation (`kDigitBits = 30`), bit 29 is the highest bit of a 30-bit limb. Checking bit 31 leads to invalid positive sign interpretation.
- **Port Fix & Verification**: Implemented explicit bit 29 isolation `(digit & (1 << (bits - 1))) != 0` and two's complement $2^N - |x|$ wrap in `truncateAndSubFromPowerOfTwo`. Verified in `tests/port/truncation_test.go`.

### Bug #3: Base 2–36 String Parsing Leading Sign & Whitespace Edge Cases
- **Upstream JSBI / Spec Behavior**: `FromString(str, radix)` strips ASCII whitespace (spaces, tabs, newlines) and allows optional leading `+` or `-` prefix across all radices 2 to 36.
- **Go Potential Flaw**: `strconv.ParseInt` rejects leading `+` signs in non-decimal bases and fails on numbers exceeding 64-bit bounds.
- **Port Fix & Verification**: Developed custom `isWhitespace` scanner and multi-limb radix accumulator in `src/fromString.go` that safely parses arbitrarily large BigInts in all 35 radices.

### Bug #4: Division by Zero Native Error Return
- **Upstream JSBI / Spec Behavior**: Throws JS `RangeError("Division by zero")`.
- **Go Potential Flaw**: Integer division by zero in Go triggers an unrecoverable runtime panic `runtime error: integer divide by zero`.
- **Port Fix & Verification**: Standardized explicit zero-divisor check in `Divide`, `Remainder`, and `DivRem` returning native `ErrDivideByZero` error value (`(nil, nil, ErrDivideByZero)`), ensuring 100% panic-safe execution.

---

## 3. Differential Fuzz Evidence Link

All edge-case fixes were validated using 9,696,250 differential fuzz executions logged live in:
- [`fuzz/log.txt`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/fuzz/log.txt)
- [`verification/fuzz_campaign.md`](file:///c:/Users/madha/OneDrive/Desktop/port%20TS-GO/verification/fuzz_campaign.md)
- [`https://github.com/Kavinraj696/verified-bugs`](https://github.com/Kavinraj696/verified-bugs)
