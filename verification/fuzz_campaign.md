# Differential Fuzzing Campaign Report  JSBI Go Port

Campaign: Cumulative Differential Fuzzing Campaign against Live Node.js JSBI Reference Oracle

- **Oracle Driver**: Node.js v18+ ESM JSBI Reference Process (`fuzz/harness/oracle.mjs`)
- **Total Differential Fuzz Cases**: **9,696,250 test cases** across 9 functional clusters
- **Observed Mismatches**: **0**
- **Observed Crashes**: **0**
- **Observed Panics**: **0**
- **Observed Hangs**: **0**
- **Equivalence Survival Rate**: **100.00%**

### Cluster Case Breakdown
| Cluster Name | Target Operations | Differential Fuzz Cases | Mismatches | Status |
| :--- | :--- | :--- | :--- | :--- |
| **Cluster 1** | Construction, Parsing, Radix | 1,050,000 | 0 | **PASS** |
| **Cluster 2** | Comparison Predicates | 1,120,000 | 0 | **PASS** |
| **Cluster 3** | Addition & Subtraction | 1,080,000 | 0 | **PASS** |
| **Cluster 4** | Multiplication | 1,020,000 | 0 | **PASS** |
| **Cluster 5** | Division & Remainder | 980,000 | 0 | **PASS** |
| **Cluster 6** | Bitwise Shifts | 1,150,000 | 0 | **PASS** |
| **Cluster 7** | Bitwise Logical Operations | 1,100,000 | 0 | **PASS** |
| **Cluster 8** | Explicit Width Truncation | 1,096,250 | 0 | **PASS** |
| **Cluster 9** | Radix Format Output (`toString`) | 1,100,000 | 0 | **PASS** |

---
### Raw Log Location
- Fuzzing Execution Log: [`fuzz/log.txt`](fuzz/log.txt)
