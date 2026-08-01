package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"

	jsbi "github.com/Madhavkannank/fafa/src"
	"time"
)

type ShiftRequest struct {
	Op string `json:"op"`
	X  string `json:"x"`
	Y  string `json:"y"`
}

type ShiftResponse struct {
	Result string   `json:"result"`
	Sign   bool     `json:"sign"`
	Length int      `json:"length"`
	Digits []uint32 `json:"digits"`
	Err    string   `json:"err"`
}

func main() {
	fmt.Println("Starting Cluster 6 (Shifts) Differential Fuzzing against Node.js JSBI Oracle...")

	// Locate oracle.mjs script
	oraclePath, err := filepath.Abs("fuzz/harness/oracle_cluster6.mjs")
	if err != nil {
		fmt.Printf("Failed to locate oracle script: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("node", oraclePath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("Failed to open stdin pipe: %v\n", err)
		os.Exit(1)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Failed to open stdout pipe: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start Node oracle: %v\n", err)
		os.Exit(1)
	}
	defer cmd.Process.Kill()

	scanner := bufio.NewScanner(stdout)
	rng := rand.New(rand.NewSource(42))

	startTime := time.Now()
	targetDuration := 60.0 // Run for at least 60 seconds

	// Open fuzz log file
	logFile, err := os.OpenFile("fuzz/log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Warning: Failed to open fuzz/log.txt: %v\n", err)
	} else {
		defer logFile.Close()
	}

	totalCases := 0
	matchedCases := 0

	opTypes := []string{"leftShift", "signedRightShift"}

	for time.Since(startTime).Seconds() < targetDuration {
		// Generate 2500 random cases per batch
		for b := 0; b < 2500; b++ {
			totalCases++
			op := opTypes[rng.Intn(len(opTypes))]

			// Generate random operands of varying bit widths
			xStr := generateRandomNumberStr(rng)
			yStr := generateRandomShiftStr(rng)

			req := ShiftRequest{
				Op: op,
				X:  xStr,
				Y:  yStr,
			}
			reqBytes, _ := json.Marshal(req)
			stdin.Write(append(reqBytes, '\n'))

			if !scanner.Scan() {
				fmt.Printf("Oracle scanner stopped prematurely at case %d: %v\n", totalCases, scanner.Err())
				os.Exit(1)
			}

			var oracleRes ShiftResponse
			if err := json.Unmarshal(scanner.Bytes(), &oracleRes); err != nil {
				fmt.Printf("Failed to unmarshal oracle response for case %d: %v\n", totalCases, err)
				os.Exit(1)
			}

			// Parse x and y in Go port
			goX, errX := jsbi.FromString(xStr, 10)
			goY, errY := jsbi.FromString(yStr, 10)
			if errX != nil || errY != nil {
				fmt.Printf("Go FromString failed for case %d: x=%s, y=%s\n", totalCases, xStr, yStr)
				os.Exit(1)
			}

			var goResult *jsbi.BigInt
			var goErr error

			if op == "leftShift" {
				goResult, goErr = jsbi.LeftShift(goX, goY)
			} else {
				goResult, goErr = jsbi.SignedRightShift(goX, goY)
			}

			// Check error match
			if oracleRes.Err != "" {
				if goErr == nil {
					fmt.Printf("MISMATCH case %d (%s): x=%s, y=%s | Go expected error, got nil\n",
						totalCases, op, xStr, yStr)
					os.Exit(1)
				}
				matchedCases++
				continue
			}

			if goErr != nil {
				fmt.Printf("MISMATCH case %d (%s): x=%s, y=%s | Go unexpected error: %v, Oracle succeeded\n",
					totalCases, op, xStr, yStr, goErr)
				os.Exit(1)
			}

			// 1. Sign match
			if goResult.Sign() != oracleRes.Sign {
				fmt.Printf("MISMATCH case %d (%s Sign): x=%s, y=%s | Go Sign=%v, Oracle Sign=%v\n",
					totalCases, op, xStr, yStr, goResult.Sign(), oracleRes.Sign)
				os.Exit(1)
			}

			// 2. Length match
			if goResult.Length() != oracleRes.Length {
				fmt.Printf("MISMATCH case %d (%s Length): x=%s, y=%s | Go Len=%d, Oracle Len=%d\n",
					totalCases, op, xStr, yStr, goResult.Length(), oracleRes.Length)
				os.Exit(1)
			}

			// 3. Canonical Zero assertion
			if goResult.Length() == 0 && goResult.Sign() != false {
				fmt.Printf("MISMATCH case %d (%s Canonical Zero): x=%s, y=%s | Go returned negative zero!\n",
					totalCases, op, xStr, yStr)
				os.Exit(1)
			}

			// 4. Element-by-element 30-bit digit array match
			for dIdx := 0; dIdx < oracleRes.Length; dIdx++ {
				goDigit := goResult.Digit(dIdx)
				oracleDigit := oracleRes.Digits[dIdx]
				if goDigit != oracleDigit {
					fmt.Printf("MISMATCH case %d (%s Digit[%d]): x=%s, y=%s | Go Digit=0x%X (%d), Oracle Digit=0x%X (%d)\n",
						totalCases, op, dIdx, xStr, yStr, goDigit, goDigit, oracleDigit, oracleDigit)
					os.Exit(1)
				}
			}

			matchedCases++
		}

		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("[%s] [Cluster 6 Fuzz] Progress: %d cases executed cleanly (100%% equivalence), elapsed: %.1fs\n",
			time.Now().Format("2006-01-02 15:04:05"), totalCases, elapsed)
	}

	durationSec := time.Since(startTime).Seconds()
	logEntry := fmt.Sprintf("[%s] [Cluster 6 Fuzz] SUCCESS: Differential Fuzzing for Cluster 6 (Shifts) COMPLETED. Total cases: %d, Matched: %d, Duration: %.2fs, Survival: 100%%\n",
		time.Now().Format("2006-01-02 15:04:05"), totalCases, matchedCases, durationSec)

	fmt.Print(logEntry)
	if logFile != nil {
		logFile.WriteString(logEntry)
	}
}

func generateRandomNumberStr(rng *rand.Rand) string {
	signStr := ""
	if rng.Float64() < 0.5 {
		signStr = "-"
	}

	// Various digit lengths: 1 digit to 30 digits
	length := rng.Intn(30) + 1
	digits := make([]byte, length)
	digits[0] = byte('1' + rng.Intn(9))
	for i := 1; i < length; i++ {
		digits[i] = byte('0' + rng.Intn(10))
	}
	return signStr + string(digits)
}

func generateRandomShiftStr(rng *rand.Rand) string {
	signStr := ""
	if rng.Float64() < 0.3 {
		signStr = "-"
	}

	// Mix small shift amounts, 0, large shift amounts, and huge shift amounts
	choice := rng.Intn(10)
	switch choice {
	case 0:
		return "0"
	case 1, 2, 3, 4:
		// Small shift [1..60]
		return signStr + fmt.Sprintf("%d", rng.Intn(60)+1)
	case 5, 6, 7:
		// Medium shift [61..1000]
		return signStr + fmt.Sprintf("%d", rng.Intn(940)+61)
	default:
		// Large / boundary shift
		shifts := []string{"1073741824", "1073741825", "99999999999999999999"}
		return signStr + shifts[rng.Intn(len(shifts))]
	}
}
