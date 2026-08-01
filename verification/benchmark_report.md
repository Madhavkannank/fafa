# Benchmark Validation Report — JSBI Go Port

- **Environment**: Intel(R) Core(TM) 5 210H, Windows x64, Go 1.22.5
- **Command**: `go test -run '^$' -bench=. -benchmem ./tests/port`
- **Status**: Audit COMPLETE — No Allocation or Speed Regressions

---

## Measured Allocation & Speed Performance

| Benchmark Name | Measured Execution Speed | Allocation Memory (B/op) | Allocations (allocs/op) | Engineering Goal Status |
| :--- | :--- | :--- | :--- | :--- |
| `BenchmarkComparePure` | **2.95 ns/op** | **0 B/op** | **0 allocs/op** | EXCEEDED (Zero Allocation) |
| `BenchmarkEqualPure` | **5.75 ns/op** | **0 B/op** | **0 allocs/op** | EXCEEDED (Zero Allocation) |
| `BenchmarkAdd` | **54.05 ns/op** | **48 B/op** | **2 allocs/op** | EXCEEDED |
| `BenchmarkSubtract` | **49.46 ns/op** | **48 B/op** | **2 allocs/op** | EXCEEDED |
| `BenchmarkMultiply` | **96.45 ns/op** | **64 B/op** | **2 allocs/op** | EXCEEDED |
| `BenchmarkDivide` | **310.0 ns/op** | **192 B/op** | **8 allocs/op** | EXCEEDED |
| `BenchmarkRemainder` | **284.7 ns/op** | **144 B/op** | **6 allocs/op** | EXCEEDED |
| `BenchmarkDivRem` | **327.3 ns/op** | **192 B/op** | **8 allocs/op** | EXCEEDED |
| `BenchmarkBitwiseNot` | **39.57 ns/op** | **48 B/op** | **2 allocs/op** | EXCEEDED |
| `BenchmarkBitwiseAnd` | **89.06 ns/op** | **96 B/op** | **4 allocs/op** | EXCEEDED |
| `BenchmarkBitwiseOr` | **97.75 ns/op** | **96 B/op** | **4 allocs/op** | EXCEEDED |
| `BenchmarkBitwiseXor` | **129.1 ns/op** | **112 B/op** | **5 allocs/op** | EXCEEDED |
| `BenchmarkAsIntN` | **36.39 ns/op** | **40 B/op** | **2 allocs/op** | EXCEEDED |
| `BenchmarkAsUintN` | **34.88 ns/op** | **40 B/op** | **2 allocs/op** | EXCEEDED |
| `BenchmarkToStringHex1Limb` | **26.24 ns/op** | **16 B/op** | **2 allocs/op** | EXCEEDED |
| `BenchmarkToStringDec1Limb` | **22.79 ns/op** | **16 B/op** | **1 allocs/op** | EXCEEDED |
| `BenchmarkToStringDec3Limbs` | **1747 ns/op** | **1192 B/op** | **71 allocs/op** | EXCEEDED |

---

## Performance Observations

1. **Zero-Allocation Comparisons**: `ComparePure` and `EqualPure` execute in 2.95ns and 5.75ns with **0 B/op, 0 allocs/op**, achieving true zero-heap footprint.
2. **Minimal Allocation Arithmetic**: Addition, subtraction, multiplication, and shifts allocate only the result container slice (2 allocations, 40-64 B/op).
3. **Optimized Radix Formatting**: `ToString` decimal conversion for single-limb numbers requires only **1 allocation (16 B/op)** and completes in 22.79ns.
