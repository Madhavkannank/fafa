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

type OracleReq struct {
	X     string `json:"x"`
	Radix int    `json:"radix"`
}

type OracleRes struct {
	Str string `json:"str"`
	Err string `json:"err"`
}

func main() {
	fmt.Println("=== Cluster 9 Differential Fuzzing (ToString) ===")

	logFile, err := os.OpenFile("fuzz/log.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Warning: could not open fuzz/log.txt: %v\n", err)
	} else {
		defer logFile.Close()
	}

	cmd := exec.Command("node", "fuzz/harness/oracle_cluster9.mjs")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start oracle: %v\n", err)
		os.Exit(1)
	}
	defer cmd.Process.Kill()

	scanner := bufio.NewScanner(stdout)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	startTime := time.Now()
	total := 0
	matched := 0

	// All radices 2-36 (exhaustive requirement from Design Review 09)
	allRadices := make([]int, 35)
	for i := range allRadices {
		allRadices[i] = i + 2
	}

	for time.Since(startTime) < 60*time.Second {
		for i := 0; i < 500; i++ {
			total++
			// Randomly pick any radix 2-36
			radix := allRadices[rng.Intn(len(allRadices))]
			xStr := genNumber(rng)

			req := OracleReq{X: xStr, Radix: radix}
			b, _ := json.Marshal(req)
			stdin.Write(append(b, '\n'))

			if !scanner.Scan() {
				fmt.Printf("Oracle pipe closed at case %d\n", total)
				os.Exit(1)
			}
			var res OracleRes
			if err := json.Unmarshal(scanner.Bytes(), &res); err != nil {
				fmt.Printf("Parse error at case %d: %v\n", total, err)
				os.Exit(1)
			}

			goBig, errParse := jsbi.FromString(xStr, 10)
			if errParse != nil {
				fmt.Printf("Go parse error at case %d: x=%s err=%v\n", total, xStr, errParse)
				os.Exit(1)
			}
			goStr, goErr := jsbi.ToString(goBig, radix)

			if res.Err != "" {
				if goErr == nil {
					fmt.Printf("MISMATCH case %d (radix=%d): x=%s | oracle error, Go returned %q\n", total, radix, xStr, goStr)
					os.Exit(1)
				}
				matched++
				continue
			}
			if goErr != nil {
				fmt.Printf("MISMATCH case %d (radix=%d): x=%s | Go error %v, oracle returned %q\n", total, radix, xStr, goErr, res.Str)
				os.Exit(1)
			}
			if goStr != res.Str {
				fmt.Printf("MISMATCH case %d (radix=%d): x=%s\n  Go:     %q\n  Oracle: %q\n", total, radix, xStr, goStr, res.Str)
				os.Exit(1)
			}
			matched++
		}

		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("[%s] [Cluster 9 Fuzz] Progress: %d cases cleanly (100%% equiv), elapsed: %.1fs\n",
			time.Now().Format("2006-01-02 15:04:05"), total, elapsed)
	}

	dur := time.Since(startTime).Seconds()
	logLine := fmt.Sprintf("[%s] Cluster 9 Fuzz Run: PASS | Elapsed: %.2fs | Total: %d | Matched: %d | Mismatched: 0 | radices: all 2-36\n",
		time.Now().Format("2006-01-02 15:04:05"), dur, total, matched)
	fmt.Print(logLine)
	if logFile != nil {
		logFile.WriteString(logLine)
	}
}

func genNumber(rng *rand.Rand) string {
	sign := ""
	if rng.Float64() < 0.4 {
		sign = "-"
	}
	// Distribution: 1 limb, 2 limbs, 3+ limbs, and boundary values
	r := rng.Float64()
	var numDigits int
	switch {
	case r < 0.1:
		// Boundary: 0, 1, small numbers
		return []string{"0", "1", "-1", "2", "1073741823", "1073741824", "-1073741824"}[rng.Intn(7)]
	case r < 0.35:
		numDigits = rng.Intn(9) + 1 // 1-9 digits (1 limb range)
	case r < 0.65:
		numDigits = rng.Intn(9) + 10 // 10-18 digits (2 limb range)
	default:
		numDigits = rng.Intn(12) + 19 // 19-30 digits (multi-limb)
	}
	digits := make([]byte, numDigits)
	digits[0] = byte('1' + rng.Intn(9))
	for i := 1; i < numDigits; i++ {
		digits[i] = byte('0' + rng.Intn(10))
	}
	return sign + string(digits)
}
