# 🎬 Energetic YouTube-Style 5-Minute Hackathon Demo Script

**Title:** Porting Google's JSBI to Pure Go | Hackathon Submission  
**Presenter:** Madhav Kannan  
**Format:** AI Avatar on Left (25% width), Terminal/Code Screen Recording on Right (75% width)  
**Background Music:** Light upbeat instrumental @ 20–25% volume  

---

## ⏱️ Timeline & Script Breakdown

### 0:00 – 0:20 | Introduction & Hook
- **🖥️ On Screen**: Project logo, `GoogleChromeLabs/jsbi` logo, GitHub repository homepage (`github.com/Madhavkannank/fafa`), AI avatar.
- **🎤 Speaking**:
  > *"Imagine taking a JavaScript BigInt library... removing JavaScript completely... rebuilding everything in pure Go... and then proving it actually behaves the same.*
  >
  > *That's exactly what this project does.*
  >
  > *Hi! I'm Madhav, and this is my Go port of Google's JSBI library."*

---

### 0:20 – 0:50 | Repository Overview & Goals
- **🖥️ On Screen**: Repository file tree zooming into `src/`, `tests/`, `fuzz/`, `bench/`, `verification/`.
- **🎤 Speaking**:
  > *"The goal wasn't just to translate code.*
  >
  > *It was to create an idiomatic Go implementation while preserving the observable behaviour of the original JSBI library.*
  >
  > *So instead of stopping at 'it compiles', I wanted evidence.*
  >
  > *Lots of evidence.*
  >
  > *Because computers don't care about promises. They care about proof."*

---

### 0:50 – 1:30 | Command-Line Reproducibility & Build
- **🖥️ On Screen**: Highlight `make build`, `make test`, `make verify`. Terminal displaying green `PASS` output.
- **🎤 Speaking**:
  > *"The repository is designed so anyone can verify it.*
  >
  > *Build it.*
  >
  > *Run the tests.*
  >
  > *Run the verification pipeline.*
  >
  > *No mysterious scripts. No magic.*
  >
  > *If something breaks... the terminal will happily expose me.*
  >
  > *(Small pause)*
  >
  > *And thankfully... it didn't."*

---

### 1:30 – 2:20 | Testing & 9.69M Differential Fuzzing
- **🖥️ On Screen**: Show `tests/original/` (5/5 original tests), `tests/port/`, and `fuzz/log.txt` showing `9,696,250 cases, 0 mismatches`. Animated callout numbers.
- **🎤 Speaking**:
  > *"One thing I wanted to preserve was compatibility with the upstream project.*
  >
  > *The original JSBI tests remain untouched.*
  >
  > *Then I added an entirely new Go test suite.*
  >
  > *But the real confidence comes from differential fuzzing.*
  >
  > *Random inputs are generated... sent to my Go implementation... sent to the original JavaScript implementation... and every output is compared.*
  >
  > *Nearly ten million test cases later... there were no observed behavioural differences.*
  >
  > *(Smile)*
  >
  > *My laptop probably needs therapy now."*

---

### 2:20 – 3:10 | Evidence Campaigns & Audit Trail
- **🖥️ On Screen**: Show `verification/` folder, animating icons for `benchstat`, `pprof`, `coverage`, `govulncheck`, `golangci-lint`, `staticcheck`, `gosec`.
- **🎤 Speaking**:
  > *"I also wanted this repository to be easy to audit.*
  >
  > *So performance isn't just a single benchmark screenshot.*
  >
  > *Everything is reproducible.*
  >
  > *Coverage. Benchmarks. CPU profiles. Memory profiles. Static analysis. Security scanning.*
  >
  > *Raw outputs are included so reviewers can verify every reported metric."*

---

### 3:10 – 4:10 | Benchmarks & Latency Metrics
- **🖥️ On Screen**: Benchmark table highlighting `Compare: 2.7 ns/op, 0 allocs, 0 bytes`. Sliding graphs for p99 latency, throughput, memory, coverage.
- **🎤 Speaking**:
  > *"I also measured more than average execution time.*
  >
  > *The project includes:*
  > - *Percentile latency.*
  > - *Operational throughput.*
  > - *Runtime memory statistics.*
  > - *Statistical benchmark campaigns.*
  > - *Coverage reports.*
  >
  > *Everything generated from repeatable executions.*
  >
  > *Rather than saying 'trust me bro', I tried to make the repository say 'here are the logs.'"*

---

### 4:10 – 4:45 | Engineering Process & Submission Metadata
- **🖥️ On Screen**: Open `DECISIONS.md`, `BUGS.md`, `Dockerfile`, `.port-mortem.toml`.
- **🎤 Speaking**:
  > *"Beyond code, I documented the engineering process.*
  >
  > *Implementation decisions. Known trade-offs. Bug investigations. Build instructions. Container setup. Submission metadata.*
  >
  > *The idea was simple—*
  >
  > *If someone reviews this project six months from now... they shouldn't need to reverse engineer my thinking."*

---

### 4:45 – 5:00 | Conclusion & Call to Action
- **🖥️ On Screen**: Zoom out to GitHub repo (`github.com/Madhavkannank/fafa`) with "Thank You" banner.
- **🎤 Speaking**:
  > *"So that's my Go port of Google's JSBI.*
  >
  > *A pure Go implementation... backed by testing... benchmarking... fuzzing... profiling... and reproducible verification.*
  >
  > *Thanks for watching, and I hope you enjoy exploring the repository!"*

---

## 🎬 Video Editing Checklist

- [x] **Layout**: AI Avatar on Left (25% width), Screen Recording on Right (75% width).
- [x] **Zooming**: Add subtle zoom cuts every 5–8 seconds to keep visual focus dynamic.
- [x] **Animated Callouts**:
  - **9,696,250 Differential Tests**
  - **0 Mismatches**
  - **88.7% Coverage**
  - **0 `unsafe`**
  - **0 Vulnerabilities**
  - **2.7 ns/op**
- [x] **Audio**: Background music at 20-25% volume; vocal narration clear and unclipped.
