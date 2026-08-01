package port_test

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
	"testing"
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

func TestPureBigIntComparisons(t *testing.T) {
	zero := jsbi.Zero()
	one := jsbi.FromInt64(1)
	negOne := jsbi.FromInt64(-1)
	five := jsbi.FromInt64(5)
	ten := jsbi.FromInt64(10)
	negTen := jsbi.FromInt64(-10)

	if !jsbi.Equal(zero, zero) {
		t.Errorf("Equal(0, 0) failed")
	}
	if jsbi.Equal(zero, one) {
		t.Errorf("Equal(0, 1) returned true")
	}
	if !jsbi.NotEqual(ten, negTen) {
		t.Errorf("NotEqual(10, -10) failed")
	}

	if !jsbi.LessThan(negTen, ten) {
		t.Errorf("LessThan(-10, 10) failed")
	}
	if !jsbi.LessThan(five, ten) {
		t.Errorf("LessThan(5, 10) failed")
	}
	if jsbi.LessThan(ten, five) {
		t.Errorf("LessThan(10, 5) returned true")
	}
	if !jsbi.LessThanOrEqual(five, five) {
		t.Errorf("LessThanOrEqual(5, 5) failed")
	}
	if !jsbi.LessThanOrEqual(negOne, zero) {
		t.Errorf("LessThanOrEqual(-1, 0) failed")
	}

	if !jsbi.GreaterThan(ten, five) {
		t.Errorf("GreaterThan(10, 5) failed")
	}
	if !jsbi.GreaterThan(ten, negTen) {
		t.Errorf("GreaterThan(10, -10) failed")
	}
	if !jsbi.GreaterThanOrEqual(five, five) {
		t.Errorf("GreaterThanOrEqual(5, 5) failed")
	}
}

func TestMultiLimbComparisons(t *testing.T) {
	val1 := jsbi.FromInt64(0x40000000)
	val2 := jsbi.FromInt64(0x40000001)
	negVal1 := jsbi.FromInt64(-0x40000000)
	negVal2 := jsbi.FromInt64(-0x40000001)

	if !jsbi.LessThan(val1, val2) {
		t.Errorf("LessThan(2^30, 2^30+1) failed")
	}
	if !jsbi.GreaterThan(val2, val1) {
		t.Errorf("GreaterThan(2^30+1, 2^30) failed")
	}
	if !jsbi.GreaterThan(negVal1, negVal2) {
		t.Errorf("GreaterThan(-2^30, -2^30-1) failed")
	}
	if !jsbi.LessThan(negVal2, negVal1) {
		t.Errorf("LessThan(-2^30-1, -2^30) failed")
	}
}

func TestCompareToFloat64Targeted(t *testing.T) {
	zero := jsbi.Zero()
	posBig := jsbi.FromInt64(1 << 53)
	posBigPlusOne, _ := jsbi.FromString("9007199254740993", 10)
	posBigMinusOne, _ := jsbi.FromString("9007199254740991", 10)
	multiLimbBig, _ := jsbi.FromString("1000000000000000000000000", 10)

	if _, isNaN := jsbi.CompareToFloat64(zero, math.NaN()); !isNaN {
		t.Errorf("CompareToFloat64(0, NaN) isNaN = false, want true")
	}
	if cmp, isNaN := jsbi.CompareToFloat64(zero, math.Inf(1)); isNaN || cmp != -1 {
		t.Errorf("CompareToFloat64(0, +Inf) = %d (isNaN=%v), want -1 (isNaN=false)", cmp, isNaN)
	}
	if cmp, isNaN := jsbi.CompareToFloat64(zero, math.Inf(-1)); isNaN || cmp != 1 {
		t.Errorf("CompareToFloat64(0, -Inf) = %d (isNaN=%v), want 1 (isNaN=false)", cmp, isNaN)
	}
	if cmp, isNaN := jsbi.CompareToFloat64(zero, 0.0); isNaN || cmp != 0 {
		t.Errorf("CompareToFloat64(0, 0.0) = %d (isNaN=%v), want 0 (isNaN=false)", cmp, isNaN)
	}
	if cmp, isNaN := jsbi.CompareToFloat64(zero, -0.0); isNaN || cmp != 0 {
		t.Errorf("CompareToFloat64(0, -0.0) = %d (isNaN=%v), want 0 (isNaN=false)", cmp, isNaN)
	}
	subnormal := math.Float64frombits(1)
	if cmp, isNaN := jsbi.CompareToFloat64(zero, subnormal); isNaN || cmp != -1 {
		t.Errorf("CompareToFloat64(0, subnormal) = %d (isNaN=%v), want -1 (isNaN=false)", cmp, isNaN)
	}

	f2_53 := float64(1 << 53)
	if cmp, isNaN := jsbi.CompareToFloat64(posBig, f2_53); isNaN || cmp != 0 {
		t.Errorf("CompareToFloat64(2^53, 2^53 float) = %d (isNaN=%v), want 0", cmp, isNaN)
	}
	if cmp, isNaN := jsbi.CompareToFloat64(posBigPlusOne, f2_53); isNaN || cmp != 1 {
		t.Errorf("CompareToFloat64(2^53 + 1, 2^53 float) = %d (isNaN=%v), want 1", cmp, isNaN)
	}
	if cmp, isNaN := jsbi.CompareToFloat64(posBigMinusOne, f2_53); isNaN || cmp != -1 {
		t.Errorf("CompareToFloat64(2^53 - 1, 2^53 float) = %d (isNaN=%v), want -1", cmp, isNaN)
	}

	if cmp, isNaN := jsbi.CompareToFloat64(multiLimbBig, 1e20); isNaN || cmp != 1 {
		t.Errorf("CompareToFloat64(10^24, 10^20 float) = %d (isNaN=%v), want 1", cmp, isNaN)
	}
}

func TestNaNRelationalComparisons(t *testing.T) {
	val := jsbi.FromInt64(42)
	nan := math.NaN()

	// All 6 Typed Float Predicates against NaN
	if jsbi.EqualToFloat64(val, nan) != false {
		t.Errorf("EqualToFloat64(42, NaN) = true, want false")
	}
	if jsbi.NotEqualToFloat64(val, nan) != true {
		t.Errorf("NotEqualToFloat64(42, NaN) = false, want true")
	}
	if jsbi.LessThanFloat64(val, nan) != false {
		t.Errorf("LessThanFloat64(42, NaN) = true, want false")
	}
	if jsbi.LessThanOrEqualFloat64(val, nan) != false {
		t.Errorf("LessThanOrEqualFloat64(42, NaN) = true, want false")
	}
	if jsbi.GreaterThanFloat64(val, nan) != false {
		t.Errorf("GreaterThanFloat64(42, NaN) = true, want false")
	}
	if jsbi.GreaterThanOrEqualFloat64(val, nan) != false {
		t.Errorf("GreaterThanOrEqualFloat64(42, NaN) = true, want false")
	}

	// All 6 Dynamic Operators against NaN
	if jsbi.EQ(val, nan) != false {
		t.Errorf("EQ(42, NaN) = true, want false")
	}
	if jsbi.NE(val, nan) != true {
		t.Errorf("NE(42, NaN) = false, want true")
	}
	if jsbi.LT(val, nan) != false {
		t.Errorf("LT(42, NaN) = true, want false")
	}
	if jsbi.LE(val, nan) != false {
		t.Errorf("LE(42, NaN) = true, want false")
	}
	if jsbi.GT(val, nan) != false {
		t.Errorf("GT(42, NaN) = true, want false")
	}
	if jsbi.GE(val, nan) != false {
		t.Errorf("GE(42, NaN) = true, want false")
	}
}

func TestDifferentialFuzzCluster2(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping differential fuzz test in short mode")
	}

	os.MkdirAll("fuzz", 0755)
	os.MkdirAll("tmp", 0755)

	logFile, err := os.OpenFile("fuzz/log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
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
	if err != nil || func() bool { _, e := os.Stat(oraclePath); return os.IsNotExist(e) }() {
		oraclePath, _ = filepath.Abs("../../fuzz/harness/oracle_cluster2.mjs")
	}

	tmpInputPath, err := filepath.Abs("tmp/fuzz_cluster2_input.json")
	if err != nil || func() bool { _, e := os.Stat(filepath.Dir(tmpInputPath)); return os.IsNotExist(e) }() {
		os.MkdirAll("../../tmp", 0755)
		tmpInputPath, _ = filepath.Abs("../../tmp/fuzz_cluster2_input.json")
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
		math.Float64frombits(1),
	}

	nodeBinary := "C:\\Program Files\\nodejs\\node.exe"

	for time.Since(startTime) < targetDuration {
		batch := make([]TestCaseCluster2, 0, batchSize)
		goXs := make([]*jsbi.BigInt, 0, batchSize)
		goYs := make([]interface{}, 0, batchSize)

		for i := 0; i < batchSize; i++ {
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

		inputJSON, err := json.Marshal(batch)
		if err != nil {
			t.Fatalf("Failed to marshal batch JSON: %v", err)
		}
		if err := os.WriteFile(tmpInputPath, inputJSON, 0644); err != nil {
			t.Fatalf("Failed to write tmp input file: %v", err)
		}

		cmd := exec.Command(nodeBinary, "--no-warnings", oraclePath, tmpInputPath)
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf

		if err := cmd.Run(); err != nil {
			t.Fatalf("Node oracle process failed: %v, stderr: %s, stdout: %s", err, errBuf.String(), outBuf.String())
		}

		var oracleResults []OracleResultCluster2
		if err := json.Unmarshal(outBuf.Bytes(), &oracleResults); err != nil {
			t.Fatalf("Failed to unmarshal oracle response: %v, raw stdout: %q, raw stderr: %q", err, outBuf.String(), errBuf.String())
		}

		for idx, tc := range batch {
			totalCases++
			oracleRes := oracleResults[idx]
			goX := goXs[idx]
			goY := goYs[idx]

			if oracleRes.Status == "OK" {
				if yBig, ok := goY.(*jsbi.BigInt); ok {
					goComp := jsbi.Compare(goX, yBig)
					if goComp != oracleRes.Comp {
						t.Fatalf("MISMATCH on case %v: Compare mismatch (Oracle=%d, Go=%d)", tc, oracleRes.Comp, goComp)
					}
					if jsbi.Equal(goX, yBig) != oracleRes.EQ {
						t.Fatalf("MISMATCH on case %v: Equal mismatch (Oracle=%v, Go=%v)", tc, oracleRes.EQ, jsbi.Equal(goX, yBig))
					}
					if jsbi.NotEqual(goX, yBig) != oracleRes.NE {
						t.Fatalf("MISMATCH on case %v: NotEqual mismatch (Oracle=%v, Go=%v)", tc, oracleRes.NE, jsbi.NotEqual(goX, yBig))
					}
					if jsbi.LessThan(goX, yBig) != oracleRes.LT {
						t.Fatalf("MISMATCH on case %v: LessThan mismatch (Oracle=%v, Go=%v)", tc, oracleRes.LT, jsbi.LessThan(goX, yBig))
					}
					if jsbi.LessThanOrEqual(goX, yBig) != oracleRes.LE {
						t.Fatalf("MISMATCH on case %v: LessThanOrEqual mismatch (Oracle=%v, Go=%v)", tc, oracleRes.LE, jsbi.LessThanOrEqual(goX, yBig))
					}
					if jsbi.GreaterThan(goX, yBig) != oracleRes.GT {
						t.Fatalf("MISMATCH on case %v: GreaterThan mismatch (Oracle=%v, Go=%v)", tc, oracleRes.GT, jsbi.GreaterThan(goX, yBig))
					}
					if jsbi.GreaterThanOrEqual(goX, yBig) != oracleRes.GE {
						t.Fatalf("MISMATCH on case %v: GreaterThanOrEqual mismatch (Oracle=%v, Go=%v)", tc, oracleRes.GE, jsbi.GreaterThanOrEqual(goX, yBig))
					}
				} else if yFloat, ok := goY.(float64); ok {
					goComp, isNaN := jsbi.CompareToFloat64(goX, yFloat)
					if math.IsNaN(yFloat) {
						if !isNaN {
							t.Fatalf("MISMATCH on NaN case %v: CompareToFloat64 expected isNaN=true", tc)
						}
					} else {
						if isNaN || goComp != oracleRes.Comp {
							t.Fatalf("MISMATCH on float case %v: CompareToFloat64 mismatch (Oracle=%d, Go=%d, isNaN=%v)", tc, oracleRes.Comp, goComp, isNaN)
						}
					}
					if jsbi.EqualToFloat64(goX, yFloat) != oracleRes.EQ {
						t.Fatalf("MISMATCH on float case %v: EqualToFloat64 mismatch (Oracle=%v, Go=%v)", tc, oracleRes.EQ, jsbi.EqualToFloat64(goX, yFloat))
					}
					if jsbi.NotEqualToFloat64(goX, yFloat) != oracleRes.NE {
						t.Fatalf("MISMATCH on float case %v: NotEqualToFloat64 mismatch (Oracle=%v, Go=%v)", tc, oracleRes.NE, jsbi.NotEqualToFloat64(goX, yFloat))
					}
					if jsbi.LessThanFloat64(goX, yFloat) != oracleRes.LT {
						t.Fatalf("MISMATCH on float case %v: LessThanFloat64 mismatch (Oracle=%v, Go=%v)", tc, oracleRes.LT, jsbi.LessThanFloat64(goX, yFloat))
					}
					if jsbi.LessThanOrEqualFloat64(goX, yFloat) != oracleRes.LE {
						t.Fatalf("MISMATCH on float case %v: LessThanOrEqualFloat64 mismatch (Oracle=%v, Go=%v)", tc, oracleRes.LE, jsbi.LessThanOrEqualFloat64(goX, yFloat))
					}
					if jsbi.GreaterThanFloat64(goX, yFloat) != oracleRes.GT {
						t.Fatalf("MISMATCH on float case %v: GreaterThanFloat64 mismatch (Oracle=%v, Go=%v)", tc, oracleRes.GT, jsbi.GreaterThanFloat64(goX, yFloat))
					}
					if jsbi.GreaterThanOrEqualFloat64(goX, yFloat) != oracleRes.GE {
						t.Fatalf("MISMATCH on float case %v: GreaterThanOrEqualFloat64 mismatch (Oracle=%v, Go=%v)", tc, oracleRes.GE, jsbi.GreaterThanOrEqualFloat64(goX, yFloat))
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

func BenchmarkComparePure(b *testing.B) {
	x := jsbi.FromInt64(123456789)
	y := jsbi.FromInt64(987654321)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = jsbi.Compare(x, y)
	}
}

func BenchmarkEqualPure(b *testing.B) {
	x := jsbi.FromInt64(123456789)
	y := jsbi.FromInt64(123456789)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = jsbi.Equal(x, y)
	}
}
