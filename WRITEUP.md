# I Thought Porting Google's JSBI to Go Would Take a Weekend. I Was Wrong.

When first encountering Google's JSBI library, the initial thought was simple:

"It's just a BigInt library. Port it to Go."

Famous last words.

How hard could it be?

Very hard.

Not because Go lacks big integers - Go already has `math/big`. I chose Go because its simplicity makes correctness easier to reason about, but that also meant recreating functionality that `math/big` already provides instead of relying on it.

The difficult part was something else entirely: building the exact same BigInt library. Every observable behavior, every edge case, every strange JavaScript quirk - without writing a single line of JavaScript in production.

That became **FaFa**, a pure Go port of `GoogleChromeLabs/jsbi`.

---

## The Temptation to Cheat

The obvious solution looked like this:

```go
type BigInt struct {
    *big.Int
}
```

Done. Problem solved in 10 lines.

Except... it wouldn't actually be JSBI anymore. It would be `math/big` wearing a fake mustache and pretending to be JSBI. It would simply be Go's standard library wearing a JavaScript API mask. That completely defeats the purpose of a behavioral port.

So FaFa started over from scratch.

Instead of wrapping `math/big`, the engine was rebuilt using JSBI's 30-bit limb representation (`type BigInt struct { sign bool; digits []uint32 }`), preserving its internal algorithms line-for-line.

We also set a strict rule: **Zero `unsafe` pointers and zero CGO dependencies.** Pure, safe Go.

That meant implementing parsing, multi-precision arithmetic, division, bitwise operations, and radix conversion from the ground up.

---

## Then the Bugs Arrived

The first serious bug looked completely harmless.

JavaScript uses unsigned right shift:
```ts
r >>> 30
```

Go uses:
```go
r >> 30
```

They look almost identical. They aren't.

During multi-limb subtraction, borrow propagation depends on logical right shifts. In Go, right-shifting a signed `int32` performs an **arithmetic right shift**, filling the top bits with 1s when `r < 0`. `-1 >> 30` equals `-1`, corrupting the borrow bit across limb boundaries!

```go
// NAIVE GO PORT (BROKEN)
r := int32(x.digit(i) - y.digit(i) - borrow)
borrow = uint32((r >> 30) & 1) // Sign extension corrupted borrow!

// THE FIX
borrow = (uint32(r) >> 30) & 1 // Explicit cast forces logical right shift per Go spec
```

Unit tests barely noticed. The differential fuzzer caught it instantly. One `uint32` cast fixed thousands of subtle arithmetic failures. Nearly six hours were spent staring at what looked like a perfectly correct line of code.

---

## Then Came Bit 29

JSBI stores numbers in 30-bit limbs (`kDigitMask = 0x3FFFFFFF`), not 32. That tiny detail caused another wonderful trap. Those are the bugs that convince you counting to 30 is somehow harder than counting to 32.

Standard 32-bit integer code checks bit 31 for the sign. But in JSBI's 30-bit limb structure, the highest bit of a single limb sits at bit 29 (`1 << 29` or `0x20000000`).

The initial `AsIntN` truncation logic checked bit 31 instead of bit 29:

```go
// BROKEN: Checked 32-bit sign instead of 30-bit limb sign
if (digit & (1 << 31)) != 0 { ... }

// THE FIX: Dynamic bit index inside the 30-bit limb boundary
signBit := uint32(1) << ((bits - 1) % 30)
if (digit & signBit) != 0 { ... }
```

Everything worked... until values crossed that 30-bit boundary. `BigInt.asIntN()` started disagreeing with JavaScript for a tiny set of values. Those are the bugs that make you question your career choices.

---

## The Bug Nobody Expects: `NaN` Comparisons

Comparing a BigInt with `NaN` - nobody writes that in production code.

Except differential fuzzers do.

JavaScript specifies:
```js
10n == NaN // false
10n <  NaN // false
10n >  NaN // false
10n != NaN // true
```

The early Go implementation returned `0` for `CompareToFloat64(x, math.NaN())`. The equality function interpreted `0` as "equal", so `Equal(10n, NaN)` returned `true`!

The fix was updating internal float comparison helpers to return a `(cmp int, isNaN bool)` tuple:

```go
func CompareToFloat64(x *BigInt, y float64) (int, bool) {
    if math.IsNaN(y) {
        return 0, true
    }
    // ...
}
```

Finding the bug took hours; fixing it took 4 lines of code.

---

## Unit Tests Weren't Enough

"All tests passed" wasn't enough. Tests only prove the scenarios you thought to write. We wanted to test the scenarios we couldn’t imagine.

So we built a differential fuzzing harness. Every generated random input was piped simultaneously to:
1. The original JSBI running inside Node.js ESM.
2. The FaFa Go implementation.

Then every output was compared bit-for-bit.

Not approximately. Exactly.

Over 9 cluster runs, we executed **nearly 9.7 million differential test cases**.

The number of observed behavioral differences? **Zero.**

That was the first moment we actually trusted the port. My laptop, however, deserved a vacation.

---

## The Mistake I'd Take Back: 15-Bit Multiplication

If I were starting today, I'd take back one major design decision: preserving JSBI's 15-bit multiplication decomposition.

In JSBI's TypeScript source, multiplication decomposes 30-bit limbs into two 15-bit half-limbs ($m = m_{\text{high}} \times 2^{15} + m_{\text{low}}$). This was done because JavaScript numbers historically lacked native 64-bit integer multiplication without precision loss.

Go has native 64-bit `uint64` arithmetic. Future Me keeps asking why Past Me made this decision.

Preserving JSBI's 15-bit decomposition made step-by-step verification easier during early porting. But once correctness was established, keeping it meant performing twice as many loop passes as a native `uint64` carry chain would require.

If I were starting today, I wouldn't preserve that 15-bit split. It made verification easier, but once correctness was proven, Go's native 64-bit arithmetic would likely have been the better long-term engineering choice.

---

## Proof Over Promises

Writing the code wasn't the longest part of this hackathon - proving the code was.

FaFa ended up with reproducible documentation for every claim:
- Millions of differential fuzz cases logged
- Full unit test suites passing
- Original V8 test suites running 100% unmodified
- Zero `unsafe` imports and zero CGO dependencies
- Standard `benchstat` benchmark outputs

Instead of asking judges or users to "trust me", every single claim traces back to reproducible command outputs in the repo.

---

## What I Learned

This project changed how I think about software engineering.

Writing code isn't the difficult part anymore.

Proving that the code is correct is.

That's ultimately what FaFa became - not just a Go port of JSBI, but an experiment in how much evidence you can collect before you're willing to say, "Yes, this behaves the same."

If someone finds something the tests, fuzzers, benchmarks, and audits missed, we'd genuinely love to investigate it.

Preferably after we've had some sleep.

---

## Repository & Open Source Evidence

GitHub: **github.com/Madhavkannank/fafa**

If you're curious, everything is fully reproducible:
- Pure Go implementation (`package jsbi`)
- Original V8 test suites
- Differential fuzzing harness
- Benchmark campaigns & raw outputs

*Thanks to @HackathonRaptors for creating a challenge that rewarded not just building software, but proving that it works.*

**Tag**: @HackathonRaptors #HackathonRaptors #Golang #TypeScript #SoftwareEngineering #OpenSource #Fuzzing
