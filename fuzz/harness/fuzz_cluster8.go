package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"

	jsbi "github.com/Madhavkannank/fafa/src"
)

type OracleRequest struct {
	Op   string `json:"op"`
	Bits int    `json:"bits"`
	X    string `json:"x"`
}

type OracleResponse struct {
	Sign   bool   `json:"sign"`
	Length int    `json:"length"`
	Digits []int  `json:"digits"`
	Str    string `json:"str"`
	Err    string `json:"err"`
}

func main() {
	fmt.Println("=== Cluster 8 Differential Fuzzing (AsIntN & AsUintN) ===")

	logFile, err := os.OpenFile("fuzz/log.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Warning: failed to open fuzz/log.txt: %v\n", err)
	} else {
		defer logFile.Close()
	}

	cmd := exec.Command("node", "fuzz/harness/oracle_cluster8.mjs")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("Failed to get stdin pipe for Node oracle: %v\n", err)
		os.Exit(1)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Failed to get stdout pipe for Node oracle: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start Node oracle process: %v\n", err)
		os.Exit(1)
	}
	defer cmd.Process.Kill()

	oracleScanner := bufio.NewScanner(stdout)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	startTime := time.Now()
	targetDuration := 60 * time.Second
	totalCases := 0
	matchedCases := 0

	mandatoryBits := []int{0, 1, 29, 30, 31, 59, 60, 61, 1<<30 - 1, 1 << 30, 1<<30 + 1}
	ops := []string{"asIntN", "asUintN"}

	for time.Since(startTime) < targetDuration {
		batchSize := 1000
		for i := 0; i < batchSize; i++ {
			totalCases++
			op := ops[rng.Intn(len(ops))]
			bits := 0
			if rng.Float64() < 0.4 {
				bits = mandatoryBits[rng.Intn(len(mandatoryBits))]
			} else {
				bits = rng.Intn(120)
			}
			xStr := generateRandomNumberStr(rng)

			req := OracleRequest{
				Op:   op,
				Bits: bits,
				X:    xStr,
			}

			reqBytes, _ := json.Marshal(req)
			stdin.Write(append(reqBytes, '\n'))

			if !oracleScanner.Scan() {
				fmt.Printf("Oracle pipe closed unexpectedly at case %d\n", totalCases)
				os.Exit(1)
			}

			var oracleRes OracleResponse
			if err := json.Unmarshal(oracleScanner.Bytes(), &oracleRes); err != nil {
				fmt.Printf("Failed to parse oracle response: %v\n", err)
				os.Exit(1)
			}

			goX, errX := jsbi.FromString(xStr, 10)
			if errX != nil {
				fmt.Printf("Failed to parse Go input x=%s\n", xStr)
				os.Exit(1)
			}

			var goResult *jsbi.BigInt
			var goErr error

			if op == "asIntN" {
				goResult, goErr = jsbi.AsIntN(bits, goX)
			} else {
				goResult, goErr = jsbi.AsUintN(bits, goX)
			}

			// Error check
			if oracleRes.Err != "" {
				if goErr == nil {
					fmt.Printf("MISMATCH case %d (%s bits=%d): x=%s | Go expected error, got nil\n",
						totalCases, op, bits, xStr)
					os.Exit(1)
				}
				matchedCases++
				continue
			}

			if goErr != nil {
				fmt.Printf("MISMATCH case %d (%s bits=%d): x=%s | Go unexpected error: %v\n",
					totalCases, op, bits, xStr, goErr)
				os.Exit(1)
			}

			// 1. Sign match
			if goResult.Sign() != oracleRes.Sign {
				fmt.Printf("MISMATCH case %d (%s bits=%d Sign): x=%s | Go Sign=%v, Oracle Sign=%v\n",
					totalCases, op, bits, xStr, goResult.Sign(), oracleRes.Sign)
				os.Exit(1)
			}

			// 2. Length match
			if goResult.Length() != oracleRes.Length {
				fmt.Printf("MISMATCH case %d (%s bits=%d Length): x=%s | Go Len=%d, Oracle Len=%d\n",
					totalCases, op, bits, xStr, goResult.Length(), oracleRes.Length)
				os.Exit(1)
			}

			// 3. Canonical Zero assertion
			if goResult.Length() == 0 && goResult.Sign() != false {
				fmt.Printf("MISMATCH case %d (%s bits=%d Canonical Zero): x=%s | Go returned negative zero!\n",
					totalCases, op, bits, xStr)
				os.Exit(1)
			}

			// 4. Element-by-element 30-bit digit array match
			for dIdx := 0; dIdx < oracleRes.Length; dIdx++ {
				goDigit := goResult.Digit(dIdx)
				oracleDigit := uint32(oracleRes.Digits[dIdx])
				if goDigit != oracleDigit {
					fmt.Printf("MISMATCH case %d (%s bits=%d Digit[%d]): x=%s | Go Digit=0x%X (%d), Oracle Digit=0x%X (%d)\n",
						totalCases, op, bits, dIdx, xStr, goDigit, goDigit, oracleDigit, oracleDigit)
					os.Exit(1)
				}
			}

			matchedCases++
		}

		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("[%s] [Cluster 8 Fuzz] Progress: %d cases executed cleanly (100%% equivalence), elapsed: %.1fs\n",
			time.Now().Format("2006-01-02 15:04:05"), totalCases, elapsed)
	}

	durationSec := time.Since(startTime).Seconds()
	logEntry := fmt.Sprintf("[%s] Cluster 8 Fuzz Run: Cluster 8 Differential Fuzzing: PASS | Elapsed: %.2fs | Total Cases: %d | Matched: %d | Mismatched: 0\n",
		time.Now().Format("2006-01-02 15:04:05"), durationSec, totalCases, matchedCases)

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

	length := rng.Intn(30) + 1
	digits := make([]byte, length)
	digits[0] = byte('1' + rng.Intn(9))
	for i := 1; i < length; i++ {
		digits[i] = byte('0' + rng.Intn(10))
	}
	return signStr + string(digits)
}
