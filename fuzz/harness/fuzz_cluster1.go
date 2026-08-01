package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Madhavkannank/fafa/src"
)

type TestCase struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
	Radix int         `json:"radix,omitempty"`
}

type OracleResult struct {
	Status     string   `json:"status"`
	Sign       bool     `json:"sign"`
	Len        int      `json:"len"`
	Digits     []uint32 `json:"digits"`
	ErrName    string   `json:"errName"`
	ErrMessage string   `json:"errMessage"`
}

func formatFloat(v float64) interface{} {
	if math.IsNaN(v) {
		return "NaN"
	}
	if math.IsInf(v, 1) {
		return "+Infinity"
	}
	if math.IsInf(v, -1) {
		return "-Infinity"
	}
	return v
}

func main() {
	os.MkdirAll("fuzz", 0755)
	os.MkdirAll("tmp", 0755)

	logFile, err := os.OpenFile("fuzz/log.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	logMsg := func(msg string) {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		entry := fmt.Sprintf("[%s] [Cluster 1 Fuzz] %s\n", timestamp, msg)
		fmt.Print(entry)
		logFile.WriteString(entry)
	}

	logMsg("Starting Differential Fuzzing for Cluster 1 (Construction & Parsing) [Target duration: 60s+]...")

	oraclePath, err := filepath.Abs("fuzz/harness/oracle.mjs")
	if err != nil {
		logMsg(fmt.Sprintf("Failed to resolve oracle path: %v", err))
		os.Exit(1)
	}

	tmpInputPath, err := filepath.Abs("tmp/fuzz_input.json")
	if err != nil {
		logMsg(fmt.Sprintf("Failed to resolve tmp input path: %v", err))
		os.Exit(1)
	}

	startTime := time.Now()
	targetDuration := 65 * time.Second
	totalCases := 0
	matchedCases := 0
	batchSize := 1000

	r := rand.New(rand.NewSource(42))

	boundaryInts := []int64{
		0, 1, -1, 2, -2,
		0x3FFFFFFF, -0x3FFFFFFF,
		0x40000000, -0x40000000,
		1 << 31, -(1 << 31),
		1 << 32, -(1 << 32),
		1<<53 - 1, -(1<<53 - 1),
		math.MaxInt64, math.MinInt64,
	}

	boundaryFloats := []float64{
		0.0, -0.0, 1.0, -1.0, 1e10, 1e15, 1e20,
		math.NaN(), math.Inf(1), math.Inf(-1), 1.5, -0.5,
	}

	radices := []int{0, 2, 8, 10, 16, 32, 36}

	nodeBinary := "C:\\Program Files\\nodejs\\node.exe"

	for time.Since(startTime) < targetDuration {
		batch := make([]TestCase, 0, batchSize)
		goFloats := make([]float64, 0, batchSize)

		for i := 0; i < batchSize; i++ {
			choice := r.Intn(4)
			switch choice {
			case 0:
				// Int test case
				v := boundaryInts[r.Intn(len(boundaryInts))]
				if r.Float64() < 0.5 {
					v = r.Int63()
					if r.Float64() < 0.5 {
						v = -v
					}
				}
				fVal := float64(v)
				batch = append(batch, TestCase{Type: "number", Value: formatFloat(fVal)})
				goFloats = append(goFloats, fVal)

			case 1:
				// Float test case
				v := boundaryFloats[r.Intn(len(boundaryFloats))]
				batch = append(batch, TestCase{Type: "number", Value: formatFloat(v)})
				goFloats = append(goFloats, v)

			case 2:
				// Boolean test case
				batch = append(batch, TestCase{Type: "boolean", Value: r.Float64() < 0.5})
				goFloats = append(goFloats, 0)

			case 3:
				// String test case
				rad := radices[r.Intn(len(radices))]
				var strVal string

				strChoice := r.Intn(5)
				switch strChoice {
				case 0:
					// Boundary int formatted to base
					val := boundaryInts[r.Intn(len(boundaryInts))]
					if rad == 0 || rad == 10 {
						strVal = strconv.FormatInt(val, 10)
					} else {
						if val < 0 {
							strVal = "-" + strconv.FormatInt(-val, rad)
						} else {
							strVal = strconv.FormatInt(val, rad)
						}
					}
				case 1:
					// Prefix string
					prefixes := []string{"0x", "0X", "0b", "0B", "0o", "0O", "+", "-"}
					p := prefixes[r.Intn(len(prefixes))]
					numStr := strconv.FormatInt(r.Int63n(100000), 16)
					strVal = p + numStr
				case 2:
					// Whitespace padded
					spaces := []string{"  ", "\t", "\n", " \t "}
					sp1 := spaces[r.Intn(len(spaces))]
					sp2 := spaces[r.Intn(len(spaces))]
					strVal = sp1 + strconv.FormatInt(r.Int63n(1000000), 10) + sp2
				case 3:
					// Invalid syntax noise
					strVal = "invalid_" + strconv.Itoa(r.Intn(100))
				case 4:
					// Empty or whitespace
					strVal = "   "
				}

				batch = append(batch, TestCase{Type: "string", Value: strVal, Radix: rad})
				goFloats = append(goFloats, 0)
			}
		}

		// Write batch JSON to tmp file
		inputJSON, err := json.Marshal(batch)
		if err != nil {
			logMsg(fmt.Sprintf("ERROR: Failed to marshal batch JSON: %v", err))
			os.Exit(1)
		}
		if err := os.WriteFile(tmpInputPath, inputJSON, 0644); err != nil {
			logMsg(fmt.Sprintf("ERROR: Failed to write tmp input file: %v", err))
			os.Exit(1)
		}

		// Run Node oracle on batch file
		cmd := exec.Command(nodeBinary, "--no-warnings", oraclePath, tmpInputPath)
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf

		if err := cmd.Run(); err != nil {
			logMsg(fmt.Sprintf("ERROR: Node oracle process failed: %v, stderr: %s, stdout: %s", err, errBuf.String(), outBuf.String()))
			os.Exit(1)
		}

		var oracleResults []OracleResult
		if err := json.Unmarshal(outBuf.Bytes(), &oracleResults); err != nil {
			logMsg(fmt.Sprintf("ERROR: Failed to unmarshal oracle response: %v, raw stdout: %q, raw stderr: %q", err, outBuf.String(), errBuf.String()))
			os.Exit(1)
		}

		// Compare Go results against Oracle
		for idx, tc := range batch {
			totalCases++
			oracleRes := oracleResults[idx]

			var goRes *jsbi.BigInt
			var goErr error

			switch tc.Type {
			case "string":
				strVal := tc.Value.(string)
				goRes, goErr = jsbi.FromString(strVal, tc.Radix)
			case "number":
				goRes, goErr = jsbi.FromFloat64(goFloats[idx])
			case "boolean":
				boolVal := tc.Value.(bool)
				goRes = jsbi.FromBool(boolVal)
			}

			if oracleRes.Status == "OK" {
				if goErr != nil {
					logMsg(fmt.Sprintf("MISMATCH on case %v: Oracle OK, Go error: %v", tc, goErr))
					os.Exit(1)
				}
				if goRes.Sign() != oracleRes.Sign {
					logMsg(fmt.Sprintf("MISMATCH on case %v: Sign mismatch (Oracle=%v, Go=%v)", tc, oracleRes.Sign, goRes.Sign()))
					os.Exit(1)
				}
				if goRes.Length() != oracleRes.Len {
					logMsg(fmt.Sprintf("MISMATCH on case %v: Length mismatch (Oracle=%d, Go=%d)", tc, oracleRes.Len, goRes.Length()))
					os.Exit(1)
				}
				for i := 0; i < goRes.Length(); i++ {
					if goRes.Digit(i) != oracleRes.Digits[i] {
						logMsg(fmt.Sprintf("MISMATCH on case %v: Limb mismatch at index %d (Oracle limb=0x%X, Go limb=0x%X)", tc, i, oracleRes.Digits[i], goRes.Digit(i)))
						os.Exit(1)
					}
				}
			} else {
				if goErr == nil {
					logMsg(fmt.Sprintf("MISMATCH on case %v: Oracle ERR (%s), Go returned no error (res len=%d)", tc, oracleRes.ErrName, goRes.Length()))
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
	logMsg(fmt.Sprintf("SUCCESS: Differential Fuzzing for Cluster 1 COMPLETED. Total cases: %d, Matched: %d, Duration: %.2fs, Survival: 100%%", totalCases, matchedCases, totalDuration.Seconds()))
}
