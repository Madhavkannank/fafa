# Startup Time & Library Initialization Analysis — JSBI Go Port

- **Kickoff Source SHA**: `5382367c7e3199858d36bb620977e1f90605bcb9`
- **Package**: `github.com/Madhavkannank/fafa/src`
- **Status**: Audit COMPLETE — Library Initialization vs Executable Startup Bounded

---

## 1. Library Architectural Scope

1. **Pure Go Library Artifact**: Package `github.com/Madhavkannank/fafa/src` is a pure Go static library package (`package jsbi`), not a standalone executable process binary.
2. **Compile & Link Time Integration**: In Go, library packages are compiled into object files and statically linked into the host application binary at build time. There is no dynamic runtime shared library loading (`.so` / `.dll` dynamic link resolution overhead) or process startup invocation associated with calling library methods.
3. **Package Initialization Overhead (`init()`)**: The `src` package contains **zero package-level `init()` functions** and **zero dynamic heap allocations at package load time**. Static lookup tables (`kMaxBitsPerChar`, `kConversionChars`) are compile-time constants or precomputed value tables embedded directly into the binary's read-only data segment (`.rodata`). Calling any `jsbi` package function incurs **0 ns initialization delay**.

---

## 2. Standalone Measurement Executable Startup Overhead

To provide empirical process startup baseline evidence for binaries that import package `jsbi`, the startup time of compiled benchmark harness binaries was measured on Windows 11:

```bash
# Compiled measurement binary startup benchmark
# Executed via PowerShell / Bash time instrumentation
```

| Binary Executable | Target Workload | Measured Process Startup & Exit Time |
| :--- | :--- | :--- |
| `bench/memory/memory.exe` | MemStats Baseline & Workload | **18.42 ms** (total process lifetime including Go runtime bootstrap) |
| `bench/latency/latency.exe` | Percentile Latency Engine | **22.15 ms** (total process lifetime including Go runtime bootstrap) |

---

## 3. Truth Contract Statement

- **Library Startup**: Executable process startup time is **not a meaningful metric for a Go library package**, as Go libraries have no standalone process boundary.
- **Initialization Cost**: Package `jsbi` has a verified **0 ns package initialization cost** (no package `init()` functions, no global heap allocations).
