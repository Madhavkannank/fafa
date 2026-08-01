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

type TestCaseCluster5 struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type OracleResultCluster5 struct {
	Status    string   `json:"status"`
	DivSign   bool     `json:"divSign"`
	DivLen    int      `json:"divLen"`
	DivDigits []uint32 `json:"divDigits"`
	RemSign   bool     `json:"remSign"`
	RemLen    int      `json:"remLen"`
	RemDigits []uint32 `json:"remDigits"`
	ErrName   string   `json:"errName"`
}

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

func generateBoundaryDecimalString(r *rand.Rand) string {
	choice := r.Intn(15)
	switch choice {
	case 0:
		return "0"
	case 1:
		if r.Float64() < 0.5 {
			return "1"
		}
		return "-1"
	case 2:
		// 15-bit small divisor boundary: 0x7FFE, 0x7FFF
		if r.Float64() < 0.5 {
			return "32767" // 0x7FFF
		}
		return "32766" // 0x7FFE
	case 3:
		// Algorithm D threshold: 0x8000 (32768), 0x8001 (32769)
		if r.Float64() < 0.5 {
			return "32768"
		}
		return "32769"
	case 4:
		// 30-bit limb boundary: 0x3FFFFFFF (1073741823)
		if r.Float64() < 0.5 {
			return "1073741823"
		}
		return "-1073741823"
	case 5:
		// Power of two
		pow := uint(1 + r.Intn(60))
		if pow < 30 {
			return fmt.Sprintf("%d", int64(1)<<pow)
		}
		return generateNeutralDecimalString(r, 10+r.Intn(20))
	case 6:
		// Divisor boundary for normalization: 0x4000 (16384), 0x4001 (16385)
		if r.Float64() < 0.5 {
			return "16384"
		}
		return "16385"
	case 7:
		// Large multi-limb operand (10-100 digits)
		return generateNeutralDecimalString(r, 10+r.Intn(90))
	case 8:
		// Very large multi-limb operand (100-500 digits)
		return generateNeutralDecimalString(r, 100+r.Intn(400))
	default:
		return generateNeutralDecimalString(r, 1+r.Intn(20))
	}
}

func main() {
	os.MkdirAll("fuzz", 0755)
	os.MkdirAll("tmp", 0755)

	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))

	harnessDir, err := filepath.Abs("fuzz/harness")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve harness dir: %v\n", err)
		os.Exit(1)
	}

	oracleScript := filepath.Join(harnessDir, "oracle_cluster5.mjs")
	logPath := filepath.Join("fuzz", "log.txt")

	batchSize := 250
	totalCases := 0
	matchedCases := 0
	mismatchedCases := 0

	fmt.Println("Starting Cluster 5 (Divide & Remainder) Differential Fuzzing against Node.js JSBI Oracle...")
	startTime := time.Now()
	targetDuration := 65 * time.Second

	for time.Since(startTime) < targetDuration {
		var testCases []TestCaseCluster5
		for i := 0; i < batchSize; i++ {
			xStr := generateBoundaryDecimalString(r)
			yStr := generateBoundaryDecimalString(r)
			testCases = append(testCases, TestCaseCluster5{X: xStr, Y: yStr})
		}

		tmpInputFile := filepath.Join("tmp", fmt.Sprintf("input_c5_%d.json", time.Now().UnixNano()))
		inputBytes, err := json.Marshal(testCases)
		if err != nil {
			fmt.Fprintf(os.Stderr, "JSON marshal error: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(tmpInputFile, inputBytes, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write temp input: %v\n", err)
			os.Exit(1)
		}

		cmd := exec.Command("node", oracleScript, tmpInputFile)
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Node oracle execution failed: %v\nStderr: %s\n", err, errBuf.String())
			os.Remove(tmpInputFile)
			os.Exit(1)
		}
		os.Remove(tmpInputFile)

		var oracleResults []OracleResultCluster5
		if err := json.Unmarshal(outBuf.Bytes(), &oracleResults); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse oracle output: %v\nRaw output: %s\n", err, outBuf.String())
			os.Exit(1)
		}

		if len(oracleResults) != len(testCases) {
			fmt.Fprintf(os.Stderr, "Mismatched result count: got %d, want %d\n", len(oracleResults), len(testCases))
			os.Exit(1)
		}

		for i, tc := range testCases {
			totalCases++
			oracle := oracleResults[i]

			goX, errX := jsbi.FromString(tc.X, 10)
			goY, errY := jsbi.FromString(tc.Y, 10)

			if errX != nil || errY != nil {
				if oracle.Status == "ERR" {
					matchedCases++
					continue
				}
				fmt.Printf("MISMATCH on case #%d: X=%s, Y=%s -> Go FromString error vs Oracle OK\n", totalCases, tc.X, tc.Y)
				mismatchedCases++
				os.Exit(1)
			}

			goDiv, divErr := jsbi.Divide(goX, goY)
			goRem, remErr := jsbi.Remainder(goX, goY)
			goDivRemQ, goDivRemR, divRemErr := jsbi.DivRem(goX, goY)

			if divErr != nil || remErr != nil || divRemErr != nil {
				if oracle.Status == "ERR" {
					matchedCases++
					continue
				}
				fmt.Printf("MISMATCH on case #%d: X=%s, Y=%s -> Go returned error (%v) vs Oracle OK\n", totalCases, tc.X, tc.Y, divErr)
				mismatchedCases++
				os.Exit(1)
			}

			if oracle.Status == "ERR" {
				fmt.Printf("MISMATCH on case #%d: X=%s, Y=%s -> Go OK vs Oracle ERR (%s: %s)\n", totalCases, tc.X, tc.Y, oracle.ErrName)
				mismatchedCases++
				os.Exit(1)
			}

			// Verify Divide vs Oracle
			if goDiv.Sign() != oracle.DivSign {
				fmt.Printf("MISMATCH (Divide Sign) on case #%d: X=%s, Y=%s -> Go Sign=%v, Oracle Sign=%v\n", totalCases, tc.X, tc.Y, goDiv.Sign(), oracle.DivSign)
				mismatchedCases++
				os.Exit(1)
			}
			if goDiv.Length() != oracle.DivLen {
				fmt.Printf("MISMATCH (Divide Length) on case #%d: X=%s, Y=%s -> Go Len=%d, Oracle Len=%d\n", totalCases, tc.X, tc.Y, goDiv.Length(), oracle.DivLen)
				mismatchedCases++
				os.Exit(1)
			}
			for dIdx := 0; dIdx < goDiv.Length(); dIdx++ {
				if goDiv.Digit(dIdx) != oracle.DivDigits[dIdx] {
					fmt.Printf("MISMATCH (Divide Digit[%d]) on case #%d: X=%s, Y=%s -> Go Digit=%x, Oracle Digit=%x\n", dIdx, totalCases, tc.X, tc.Y, goDiv.Digit(dIdx), oracle.DivDigits[dIdx])
					mismatchedCases++
					os.Exit(1)
				}
			}
			// Canonical zero assertion for Divide
			if goDiv.Length() == 0 && goDiv.Sign() != false {
				fmt.Printf("MISMATCH (Divide Canonical Zero) on case #%d: Length is 0 but Sign is true\n", totalCases)
				mismatchedCases++
				os.Exit(1)
			}

			// Verify Remainder vs Oracle
			if goRem.Sign() != oracle.RemSign {
				fmt.Printf("MISMATCH (Remainder Sign) on case #%d: X=%s, Y=%s -> Go Sign=%v, Oracle Sign=%v\n", totalCases, tc.X, tc.Y, goRem.Sign(), oracle.RemSign)
				mismatchedCases++
				os.Exit(1)
			}
			if goRem.Length() != oracle.RemLen {
				fmt.Printf("MISMATCH (Remainder Length) on case #%d: X=%s, Y=%s -> Go Len=%d, Oracle Len=%d\n", totalCases, tc.X, tc.Y, goRem.Length(), oracle.RemLen)
				mismatchedCases++
				os.Exit(1)
			}
			for dIdx := 0; dIdx < goRem.Length(); dIdx++ {
				if goRem.Digit(dIdx) != oracle.RemDigits[dIdx] {
					fmt.Printf("MISMATCH (Remainder Digit[%d]) on case #%d: X=%s, Y=%s -> Go Digit=%x, Oracle Digit=%x\n", dIdx, totalCases, tc.X, tc.Y, goRem.Digit(dIdx), oracle.RemDigits[dIdx])
					mismatchedCases++
					os.Exit(1)
				}
			}
			// Canonical zero assertion for Remainder
			if goRem.Length() == 0 && goRem.Sign() != false {
				fmt.Printf("MISMATCH (Remainder Canonical Zero) on case #%d: Length is 0 but Sign is true\n", totalCases)
				mismatchedCases++
				os.Exit(1)
			}

			// Verify DivRem consistency vs Divide and Remainder
			if !jsbi.Equal(goDivRemQ, goDiv) || !jsbi.Equal(goDivRemR, goRem) {
				fmt.Printf("MISMATCH (DivRem Consistency) on case #%d: DivRem output differs from Divide/Remainder\n", totalCases)
				mismatchedCases++
				os.Exit(1)
			}

			matchedCases++
		}
	}

	elapsed := time.Since(startTime)
	summary := fmt.Sprintf("Cluster 5 Differential Fuzzing: PASS | Elapsed: %.2fs | Total Cases: %d | Matched: %d | Mismatched: %d\n",
		elapsed.Seconds(), totalCases, matchedCases, mismatchedCases)
	fmt.Print(summary)

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logEntry := fmt.Sprintf("[%s] Cluster 5 Fuzz Run: %s", timestamp, summary)
		f.WriteString(logEntry)
		f.Close()
	}
}
