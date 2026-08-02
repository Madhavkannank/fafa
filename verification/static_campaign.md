# Static Analysis & Security Campaign Report — JSBI Go Port

Campaign: Multi-Tool Static Analysis & Vulnerability Audit Pipeline

| Static Analysis Tool | Version | Raw Log Location | Warnings | Errors / Vulnerabilities | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **`go vet`** | Go 1.22.5 | `verification/raw/static/go_vet.txt` | 0 | 0 | **PASS** |
| **`staticcheck`** | v0.7.0 | `verification/raw/static/staticcheck.txt` | 0 | 0 | **PASS** |
| **`golangci-lint`** | v1.64.6 | `verification/raw/static/golangci_lint.txt` | 0 | 0 | **PASS** |
| **`govulncheck`** | v1.6.0 | `verification/raw/static/govulncheck.txt` | 0 | 0 | **PASS** |
| **`gosec`** Security | v2.28.0 | `verification/raw/static/gosec.txt` | 31 (G115 bitwise) | 0 Critical | **PASS** |
