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

type TestCaseCluster2 struct {
	XType string      `json:"xType"`
	X     interface{} `json:"x"`
	YType string      `json:"yType"`
	Y     interface{} `json:"y"`
}

type OracleResultCluster2 struct {
	Status     string `json:"status"`
	Comp       int    `json:"comp"`
	EQ         bool   `json:"eq"`
	NE         bool   `json:"ne"`
	LT         bool   `json:"lt"`
	LE         bool   `json:"le"`
	GT         bool   `json:"gt"`
	GE         bool   `json:"ge"`
	ErrName    string `json:"errName"`
	ErrMessage string `json:"errMessage"`
}

func formatVal(v interface{}) interface{} {
	if f, ok := v.(float64); ok {
		if math.IsNaN(f) {
			return "NaN"
		}
		if math.IsInf(f, 1) {
			return "+Infinity"
		}
		if math.IsInf(f, -1) {
			return "-Infinity"
		}
	}
	return v
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
		entry := fmt.Sprintf("[%s] [Cluster 2 Fuzz] %s\n", timestamp, msg)
		fmt.Print(entry)
		logFile.WriteString(entry)
	}

	logMsg("Starting Differential Fuzzing for Cluster 2 (Comparison) [Target duration: 60s+]...")

	oraclePath, err := filepath.Abs("fuzz/harness/oracle_cluster2.mjs")
	if err != nil {
		logMsg(fmt.Sprintf("Failed to resolve oracle path: %v", err))
		os.Exit(1)
	}

	tmpInputPath, err := filepath.Abs("tmp/fuzz_cluster2_input.json")
	if err != nil {
		logMsg(fmt.Sprintf("Failed to resolve tmp input path: %v", err))
		os.Exit(1)
	}

	startTime := time.Now()
	targetDuration := 65 * time.Second
	totalCases := 0
	matchedCases := 0
	batchSize := 1000

	r := rand.New(rand.NewSource(1337))

	boundaryInts := []int64{
		0, 1, -1, 2, -2,
		0x3FFFFFFF, -0x3FFFFFFF,
		0x40000000, -0x40000000,
		1 << 31, -(1 << 31),
		1 << 32, -(1 << 32),
		1<<53 - 1, -(1<<53 - 1),
		1 << 53, -(1 << 53),
		1<<53 + 1, -(1<<53 + 1),
		math.MaxInt64, math.MinInt64,
	}

	boundaryFloats := []float64{
		0.0, -0.0, 1.0, -1.0, 1e10, 1e15, 1e20, float64(1 << 53),
		math.NaN(), math.Inf(1), math.Inf(-1), 1.5, -0.5,
		math.Float64frombits(1), // Subnormal float
	}

	nodeBinary := "C:\\Program Files\\nodejs\\node.exe"

	for time.Since(startTime) < targetDuration {
		batch := make([]TestCaseCluster2, 0, batchSize)
		goXs := make([]*jsbi.BigInt, 0, batchSize)
		goYs := make([]interface{}, 0, batchSize)

		for i := 0; i < batchSize; i++ {
			// Generate Operand X
			xInt := boundaryInts[r.Intn(len(boundaryInts))]
			if r.Float64() < 0.5 {
				xInt = r.Int63()
				if r.Float64() < 0.5 {
					xInt = -xInt
				}
			}
			xStr := strconv.FormatInt(xInt, 10)
			xBig := jsbi.FromInt64(xInt)
			goXs = append(goXs, xBig)

			// Generate Operand Y (50% BigInt string, 50% Float64)
			if r.Float64() < 0.5 {
				yInt := boundaryInts[r.Intn(len(boundaryInts))]
				if r.Float64() < 0.5 {
					yInt = r.Int63()
					if r.Float64() < 0.5 {
						yInt = -yInt
					}
				}
				yStr := strconv.FormatInt(yInt, 10)
				yBig := jsbi.FromInt64(yInt)
				batch = append(batch, TestCaseCluster2{XType: "string", X: xStr, YType: "string", Y: yStr})
				goYs = append(goYs, yBig)
			} else {
				yFloat := boundaryFloats[r.Intn(len(boundaryFloats))]
				if r.Float64() < 0.5 {
					yFloat = r.Float64() * 1e15
					if r.Float64() < 0.5 {
						yFloat = -yFloat
					}
				}
				batch = append(batch, TestCaseCluster2{XType: "string", X: xStr, YType: "number", Y: formatVal(yFloat)})
				goYs = append(goYs, yFloat)
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

		var oracleResults []OracleResultCluster2
		if err := json.Unmarshal(outBuf.Bytes(), &oracleResults); err != nil {
			logMsg(fmt.Sprintf("ERROR: Failed to unmarshal oracle response: %v, raw stdout: %q, raw stderr: %q", err, outBuf.String(), errBuf.String()))
			os.Exit(1)
		}

		// Compare Go results against Oracle
		for idx, tc := range batch {
			totalCases++
			oracleRes := oracleResults[idx]
			goX := goXs[idx]
			goY := goYs[idx]

			if oracleRes.Status == "OK" {
				if yBig, ok := goY.(*jsbi.BigInt); ok {
					// 1. Compare() result
					goComp := jsbi.Compare(goX, yBig)
					if goComp != oracleRes.Comp {
						logMsg(fmt.Sprintf("MISMATCH on case %v: Compare mismatch (Oracle=%d, Go=%d)", tc, oracleRes.Comp, goComp))
						os.Exit(1)
					}
					// 2. Individual operator methods
					if jsbi.Equal(goX, yBig) != oracleRes.EQ {
						logMsg(fmt.Sprintf("MISMATCH on case %v: Equal mismatch (Oracle=%v, Go=%v)", tc, oracleRes.EQ, jsbi.Equal(goX, yBig)))
						os.Exit(1)
					}
					if jsbi.NotEqual(goX, yBig) != oracleRes.NE {
						logMsg(fmt.Sprintf("MISMATCH on case %v: NotEqual mismatch (Oracle=%v, Go=%v)", tc, oracleRes.NE, jsbi.NotEqual(goX, yBig)))
						os.Exit(1)
					}
					if jsbi.LessThan(goX, yBig) != oracleRes.LT {
						logMsg(fmt.Sprintf("MISMATCH on case %v: LessThan mismatch (Oracle=%v, Go=%v)", tc, oracleRes.LT, jsbi.LessThan(goX, yBig)))
						os.Exit(1)
					}
					if jsbi.LessThanOrEqual(goX, yBig) != oracleRes.LE {
						logMsg(fmt.Sprintf("MISMATCH on case %v: LessThanOrEqual mismatch (Oracle=%v, Go=%v)", tc, oracleRes.LE, jsbi.LessThanOrEqual(goX, yBig)))
						os.Exit(1)
					}
					if jsbi.GreaterThan(goX, yBig) != oracleRes.GT {
						logMsg(fmt.Sprintf("MISMATCH on case %v: GreaterThan mismatch (Oracle=%v, Go=%v)", tc, oracleRes.GT, jsbi.GreaterThan(goX, yBig)))
						os.Exit(1)
					}
					if jsbi.GreaterThanOrEqual(goX, yBig) != oracleRes.GE {
						logMsg(fmt.Sprintf("MISMATCH on case %v: GreaterThanOrEqual mismatch (Oracle=%v, Go=%v)", tc, oracleRes.GE, jsbi.GreaterThanOrEqual(goX, yBig)))
						os.Exit(1)
					}
				} else if yFloat, ok := goY.(float64); ok {
					// CompareToFloat64 vs Float
					goComp, isNaN := jsbi.CompareToFloat64(goX, yFloat)
					if math.IsNaN(yFloat) {
						if !isNaN {
							logMsg(fmt.Sprintf("MISMATCH on NaN case %v: CompareToFloat64 expected isNaN=true", tc))
							os.Exit(1)
						}
					} else {
						if isNaN || goComp != oracleRes.Comp {
							logMsg(fmt.Sprintf("MISMATCH on float case %v: CompareToFloat64 mismatch (Oracle=%d, Go=%d, isNaN=%v)", tc, oracleRes.Comp, goComp, isNaN))
							os.Exit(1)
						}
					}
					if jsbi.EqualToFloat64(goX, yFloat) != oracleRes.EQ {
						logMsg(fmt.Sprintf("MISMATCH on float case %v: EqualToFloat64 mismatch (Oracle=%v, Go=%v)", tc, oracleRes.EQ, jsbi.EqualToFloat64(goX, yFloat)))
						os.Exit(1)
					}
					if jsbi.NotEqualToFloat64(goX, yFloat) != oracleRes.NE {
						logMsg(fmt.Sprintf("MISMATCH on float case %v: NotEqualToFloat64 mismatch (Oracle=%v, Go=%v)", tc, oracleRes.NE, jsbi.NotEqualToFloat64(goX, yFloat)))
						os.Exit(1)
					}
					if jsbi.LessThanFloat64(goX, yFloat) != oracleRes.LT {
						logMsg(fmt.Sprintf("MISMATCH on float case %v: LessThanFloat64 mismatch (Oracle=%v, Go=%v)", tc, oracleRes.LT, jsbi.LessThanFloat64(goX, yFloat)))
						os.Exit(1)
					}
					if jsbi.LessThanOrEqualFloat64(goX, yFloat) != oracleRes.LE {
						logMsg(fmt.Sprintf("MISMATCH on float case %v: LessThanOrEqualFloat64 mismatch (Oracle=%v, Go=%v)", tc, oracleRes.LE, jsbi.LessThanOrEqualFloat64(goX, yFloat)))
						os.Exit(1)
					}
					if jsbi.GreaterThanFloat64(goX, yFloat) != oracleRes.GT {
						logMsg(fmt.Sprintf("MISMATCH on float case %v: GreaterThanFloat64 mismatch (Oracle=%v, Go=%v)", tc, oracleRes.GT, jsbi.GreaterThanFloat64(goX, yFloat)))
						os.Exit(1)
					}
					if jsbi.GreaterThanOrEqualFloat64(goX, yFloat) != oracleRes.GE {
						logMsg(fmt.Sprintf("MISMATCH on float case %v: GreaterThanOrEqualFloat64 mismatch (Oracle=%v, Go=%v)", tc, oracleRes.GE, jsbi.GreaterThanOrEqualFloat64(goX, yFloat)))
						os.Exit(1)
					}
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
	logMsg(fmt.Sprintf("SUCCESS: Differential Fuzzing for Cluster 2 COMPLETED. Total cases: %d, Matched: %d, Duration: %.2fs, Survival: 100%%", totalCases, matchedCases, totalDuration.Seconds()))
}
