package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Madhavkannank/fafa/src"
)

type TestCaseCluster4 struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type OracleResultCluster4 struct {
	Status    string   `json:"status"`
	MulSign   bool     `json:"mulSign"`
	MulLen    int      `json:"mulLen"`
	MulDigits []uint32 `json:"mulDigits"`
	ErrName   string   `json:"errName"`
}

// Neutral string generator — generates random decimal strings completely independent of Go or JSBI
func generateNeutralDecimalString(r *rand.Rand, numDigits int) string {
	if numDigits <= 0 {
		return "0"
	}
	var buf bytes.Buffer
	if r.Float64() < 0.5 {
		buf.WriteByte('-')
	}
	buf.WriteByte(byte('1' + r.Intn(9)))
	for i := 1; i < numDigits; i++ {
		buf.WriteByte(byte('0' + r.Intn(10)))
	}
	return buf.String()
}

// Generate boundary / worst-case test vector decimal strings
func generateBoundaryDecimalString(r *rand.Rand) string {
	choice := r.Intn(6)
	switch choice {
	case 0:
		return "0"
	case 1:
		if r.Float64() < 0.5 {
			return "1"
		}
		return "-1"
	case 2:
		// 30-bit limb max boundaries (0x3FFFFFFF = 1073741823)
		if r.Float64() < 0.5 {
			return "1073741823"
		}
		return "-1073741823"
	case 3:
		// Power of two (2^30 = 1073741824, 2^60 = 1152921504606846976)
		if r.Float64() < 0.5 {
			return "1073741824"
		}
		return "1152921504606846976"
	case 4:
		// Alternating limb representation or large multi-limb string
		length := 10 + r.Intn(100)
		return generateNeutralDecimalString(r, length)
	default:
		length := 1 + r.Intn(30)
		return generateNeutralDecimalString(r, length)
	}
}

func main() {
	os.MkdirAll("fuzz", 0755)
	os.MkdirAll("tmp", 0755)

	logFile, err := os.OpenFile("fuzz/log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	oracleScript := filepath.Join("fuzz", "harness", "oracle_cluster4.mjs")
	if _, err := os.Stat(oracleScript); err != nil {
		fmt.Printf("Oracle script not found at %s: %v\n", oracleScript, err)
		os.Exit(1)
	}

	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))

	fmt.Printf("[%s] [Cluster 4 Fuzz] Starting Differential Fuzzing for Cluster 4 (Multiply) [Target duration: 65s+]...\n",
		time.Now().Format("2006-01-02 15:04:05"))

	startTime := time.Now()
	targetDuration := 65 * time.Second
	batchSize := 5000

	totalCases := 0
	matchedCases := 0

	for time.Since(startTime) < targetDuration {
		var cases []TestCaseCluster4
		for i := 0; i < batchSize; i++ {
			xStr := generateBoundaryDecimalString(r)
			yStr := generateBoundaryDecimalString(r)
			cases = append(cases, TestCaseCluster4{X: xStr, Y: yStr})
		}

		tmpInput := filepath.Join("tmp", fmt.Sprintf("fuzz_c4_input_%d.json", time.Now().UnixNano()))
		inputBytes, err := json.Marshal(cases)
		if err != nil {
			fmt.Printf("Failed to marshal input cases: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(tmpInput, inputBytes, 0644); err != nil {
			fmt.Printf("Failed to write input file: %v\n", err)
			os.Exit(1)
		}

		cmd := exec.Command("node", oracleScript, tmpInput)
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf

		if err := cmd.Run(); err != nil {
			os.Remove(tmpInput)
			fmt.Printf("Node oracle process failed: %v, stderr: %s\n", err, errBuf.String())
			os.Exit(1)
		}

		var oracleResults []OracleResultCluster4
		if err := json.Unmarshal(outBuf.Bytes(), &oracleResults); err != nil {
			os.Remove(tmpInput)
			fmt.Printf("Failed to unmarshal oracle response: %v\n", err)
			os.Exit(1)
		}
		os.Remove(tmpInput)

		if len(oracleResults) != len(cases) {
			fmt.Printf("Batch count mismatch: sent %d, got %d\n", len(cases), len(oracleResults))
			os.Exit(1)
		}

		for i, c := range cases {
			oracleRes := oracleResults[i]
			totalCases++

			goX, errX := jsbi.FromString(c.X, 10)
			goY, errY := jsbi.FromString(c.Y, 10)

			if errX != nil || errY != nil {
				if oracleRes.Status != "ERR" {
					fmt.Printf("MISMATCH case %d: Go parser error but Oracle OK (x=%s, y=%s)\n", totalCases, c.X, c.Y)
					os.Exit(1)
				}
				matchedCases++
				continue
			}

			if oracleRes.Status != "OK" {
				fmt.Printf("MISMATCH case %d: Go parsed OK but Oracle errored (x=%s, y=%s, err=%s)\n", totalCases, c.X, c.Y, oracleRes.ErrName)
				os.Exit(1)
			}

			// Compute multiplication in Go port
			goMul := jsbi.Multiply(goX, goY)

			// 1. Sign match
			if goMul.Sign() != oracleRes.MulSign {
				fmt.Printf("MISMATCH case %d (Multiply Sign): x=%s, y=%s | Go Sign=%v, Oracle Sign=%v\n",
					totalCases, c.X, c.Y, goMul.Sign(), oracleRes.MulSign)
				os.Exit(1)
			}

			// 2. Length match
			if goMul.Length() != oracleRes.MulLen {
				fmt.Printf("MISMATCH case %d (Multiply Length): x=%s, y=%s | Go Len=%d, Oracle Len=%d\n",
					totalCases, c.X, c.Y, goMul.Length(), oracleRes.MulLen)
				os.Exit(1)
			}

			// 3. Canonical Zero assertion
			if goMul.Length() == 0 && goMul.Sign() != false {
				fmt.Printf("MISMATCH case %d (Canonical Zero): x=%s, y=%s | Go returned negative zero!\n",
					totalCases, c.X, c.Y)
				os.Exit(1)
			}

			// 4. Element-by-element 30-bit digit array match
			for dIdx := 0; dIdx < oracleRes.MulLen; dIdx++ {
				goDigit := goMul.Digit(dIdx)
				oracleDigit := oracleRes.MulDigits[dIdx]
				if goDigit != oracleDigit {
					fmt.Printf("MISMATCH case %d (Multiply Digit[%d]): x=%s, y=%s | Go Digit=0x%X (%d), Oracle Digit=0x%X (%d)\n",
						totalCases, dIdx, c.X, c.Y, goDigit, goDigit, oracleDigit, oracleDigit)
					os.Exit(1)
				}
			}

			matchedCases++
		}

		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("[%s] [Cluster 4 Fuzz] Progress: %d cases executed cleanly (100%% equivalence), elapsed: %.1fs\n",
			time.Now().Format("2006-01-02 15:04:05"), totalCases, elapsed)
	}

	durationSec := time.Since(startTime).Seconds()
	logEntry := fmt.Sprintf("[%s] [Cluster 4 Fuzz] SUCCESS: Differential Fuzzing for Cluster 4 (Multiply) COMPLETED. Total cases: %d, Matched: %d, Duration: %.2fs, Survival: 100%%\n",
		time.Now().Format("2006-01-02 15:04:05"), totalCases, matchedCases, durationSec)

	fmt.Print(logEntry)
	logFile.WriteString(logEntry)
}
