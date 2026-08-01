package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Madhavkannank/fafa/src"
)

type TestCaseCluster3 struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type OracleResultCluster3 struct {
	Status    string   `json:"status"`
	AddSign   bool     `json:"addSign"`
	AddLen    int      `json:"addLen"`
	AddDigits []uint32 `json:"addDigits"`
	SubSign   bool     `json:"subSign"`
	SubLen    int      `json:"subLen"`
	SubDigits []uint32 `json:"subDigits"`
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
	// First digit must be non-zero
	buf.WriteByte(byte('1' + r.Intn(9)))
	for i := 1; i < numDigits; i++ {
		buf.WriteByte(byte('0' + r.Intn(10)))
	}
	return buf.String()
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

	logMsg := func(msg string) {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		entry := fmt.Sprintf("[%s] [Cluster 3 Fuzz] %s\n", timestamp, msg)
		fmt.Print(entry)
		logFile.WriteString(entry)
	}

	logMsg("Starting Differential Fuzzing for Cluster 3 (Add / Subtract) [Target duration: 60s+]...")

	oraclePath, err := filepath.Abs("fuzz/harness/oracle_cluster3.mjs")
	if err != nil {
		logMsg(fmt.Sprintf("Failed to resolve oracle path: %v", err))
		os.Exit(1)
	}

	tmpInputPath, err := filepath.Abs("tmp/fuzz_cluster3_input.json")
	if err != nil {
		logMsg(fmt.Sprintf("Failed to resolve tmp input path: %v", err))
		os.Exit(1)
	}

	startTime := time.Now()
	targetDuration := 65 * time.Second
	totalCases := 0
	matchedCases := 0
	batchSize := 1000

	r := rand.New(rand.NewSource(2026))

	boundaryInts := []int64{
		0, 1, -1, 2, -2,
		0x3FFFFFFF, -0x3FFFFFFF,
		0x40000000, -0x40000000,
		1 << 31, -(1 << 31),
		1 << 32, -(1 << 32),
		1<<53 - 1, -(1<<53 - 1),
	}

	nodeBinary := "C:\\Program Files\\nodejs\\node.exe"

	for time.Since(startTime) < targetDuration {
		batch := make([]TestCaseCluster3, 0, batchSize)

		for i := 0; i < batchSize; i++ {
			var xStr, yStr string

			choice := r.Intn(4)
			switch choice {
			case 0:
				// Boundary integer strings
				xVal := boundaryInts[r.Intn(len(boundaryInts))]
				yVal := boundaryInts[r.Intn(len(boundaryInts))]
				xStr = strconv.FormatInt(xVal, 10)
				yStr = strconv.FormatInt(yVal, 10)
			case 1:
				// Large 10 to 100 digit neutral decimal strings
				xLen := r.Intn(90) + 10
				yLen := r.Intn(90) + 10
				xStr = generateNeutralDecimalString(r, xLen)
				yStr = generateNeutralDecimalString(r, yLen)
			case 2:
				// Chained carry/borrow strings (e.g. 999999999999999999 + 1)
				numNines := r.Intn(50) + 5
				var buf bytes.Buffer
				if r.Float64() < 0.5 {
					buf.WriteByte('-')
				}
				for k := 0; k < numNines; k++ {
					buf.WriteByte('9')
				}
				xStr = buf.String()
				yStr = "1"
			case 3:
				// Equal operand strings (X + X and X - X)
				xStr = generateNeutralDecimalString(r, r.Intn(40)+5)
				yStr = xStr
			}

			batch = append(batch, TestCaseCluster3{X: xStr, Y: yStr})
		}

		inputJSON, err := json.Marshal(batch)
		if err != nil {
			logMsg(fmt.Sprintf("ERROR: Failed to marshal batch JSON: %v", err))
			os.Exit(1)
		}
		if err := os.WriteFile(tmpInputPath, inputJSON, 0644); err != nil {
			logMsg(fmt.Sprintf("ERROR: Failed to write tmp input file: %v", err))
			os.Exit(1)
		}

		cmd := exec.Command(nodeBinary, "--no-warnings", oraclePath, tmpInputPath)
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf

		if err := cmd.Run(); err != nil {
			logMsg(fmt.Sprintf("ERROR: Node oracle process failed: %v, stderr: %s, stdout: %s", err, errBuf.String(), outBuf.String()))
			os.Exit(1)
		}

		var oracleResults []OracleResultCluster3
		if err := json.Unmarshal(outBuf.Bytes(), &oracleResults); err != nil {
			logMsg(fmt.Sprintf("ERROR: Failed to unmarshal oracle response: %v, raw stdout: %q, raw stderr: %q", err, outBuf.String(), errBuf.String()))
			os.Exit(1)
		}

		for idx, tc := range batch {
			totalCases++
			oracleRes := oracleResults[idx]

			// INDEPENDENT PARSING: Parse neutral string into Go BigInt
			goX, errX := jsbi.FromString(tc.X, 10)
			goY, errY := jsbi.FromString(tc.Y, 10)

			if errX != nil || errY != nil {
				logMsg(fmt.Sprintf("ERROR: Failed to parse input string in Go: X=%s, Y=%s", tc.X, tc.Y))
				os.Exit(1)
			}

			if oracleRes.Status == "OK" {
				// 1. Verify Add
				goAdd := jsbi.Add(goX, goY)
				if goAdd.Sign() != oracleRes.AddSign {
					logMsg(fmt.Sprintf("MISMATCH on Add case %v: Sign mismatch (Oracle=%v, Go=%v)", tc, oracleRes.AddSign, goAdd.Sign()))
					os.Exit(1)
				}
				if goAdd.Length() != oracleRes.AddLen {
					logMsg(fmt.Sprintf("MISMATCH on Add case %v: Length mismatch (Oracle=%d, Go=%d)", tc, oracleRes.AddLen, goAdd.Length()))
					os.Exit(1)
				}
				if goAdd.Length() == 0 && goAdd.Sign() != false {
					logMsg(fmt.Sprintf("MISMATCH on Add case %v: Non-canonical zero produced (len=0, sign=true)", tc))
					os.Exit(1)
				}
				for i := 0; i < goAdd.Length(); i++ {
					if goAdd.Digit(i) != oracleRes.AddDigits[i] {
						logMsg(fmt.Sprintf("MISMATCH on Add case %v: Limb mismatch at %d (Oracle=0x%X, Go=0x%X)", tc, i, oracleRes.AddDigits[i], goAdd.Digit(i)))
						os.Exit(1)
					}
				}

				// 2. Verify Subtract
				goSub := jsbi.Subtract(goX, goY)
				if goSub.Sign() != oracleRes.SubSign {
					logMsg(fmt.Sprintf("MISMATCH on Sub case %v: Sign mismatch (Oracle=%v, Go=%v)", tc, oracleRes.SubSign, goSub.Sign()))
					os.Exit(1)
				}
				if goSub.Length() != oracleRes.SubLen {
					logMsg(fmt.Sprintf("MISMATCH on Sub case %v: Length mismatch (Oracle=%d, Go=%d)", tc, oracleRes.SubLen, goSub.Length()))
					os.Exit(1)
				}
				if goSub.Length() == 0 && goSub.Sign() != false {
					logMsg(fmt.Sprintf("MISMATCH on Sub case %v: Non-canonical zero produced (len=0, sign=true)", tc))
					os.Exit(1)
				}
				for i := 0; i < goSub.Length(); i++ {
					if goSub.Digit(i) != oracleRes.SubDigits[i] {
						logMsg(fmt.Sprintf("MISMATCH on Sub case %v: Limb mismatch at %d (Oracle=0x%X, Go=0x%X)", tc, i, oracleRes.SubDigits[i], goSub.Digit(i)))
						os.Exit(1)
					}
				}

				// 3. Verify Algebraic Identity: (X + Y) - Y == X
				recoveredX := jsbi.Subtract(goAdd, goY)
				if !jsbi.Equal(recoveredX, goX) {
					logMsg(fmt.Sprintf("MISMATCH on Algebraic Identity (X+Y)-Y!=X on case %v", tc))
					os.Exit(1)
				}
			}
			matchedCases++
		}

		elapsed := time.Since(startTime)
		if totalCases%5000 == 0 {
			logMsg(fmt.Sprintf("Progress: %d cases executed cleanly (100%% equivalence), elapsed: %.1fs", totalCases, elapsed.Seconds()))
		}
	}

	totalDuration := time.Since(startTime)
	logMsg(fmt.Sprintf("SUCCESS: Differential Fuzzing for Cluster 3 COMPLETED. Total cases: %d, Matched: %d, Duration: %.2fs, Survival: 100%%", totalCases, matchedCases, totalDuration.Seconds()))
}
