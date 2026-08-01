package port_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Madhavkannank/fafa/src"
)

func TestUnaryMinus(t *testing.T) {
	zero := jsbi.Zero()
	five := jsbi.FromInt64(5)
	negFive := jsbi.FromInt64(-5)

	negZero := jsbi.UnaryMinus(zero)
	if negZero.Sign() != false || negZero.Length() != 0 {
		t.Errorf("UnaryMinus(0) did not return canonical zero: sign=%v len=%d", negZero.Sign(), negZero.Length())
	}

	neg5 := jsbi.UnaryMinus(five)
	if !jsbi.Equal(neg5, negFive) {
		t.Errorf("UnaryMinus(5) = %v, want -5", neg5)
	}

	pos5 := jsbi.UnaryMinus(negFive)
	if !jsbi.Equal(pos5, five) {
		t.Errorf("UnaryMinus(-5) = %v, want 5", pos5)
	}
}

func TestAddBasic(t *testing.T) {
	five := jsbi.FromInt64(5)
	ten := jsbi.FromInt64(10)
	negFive := jsbi.FromInt64(-5)
	negTen := jsbi.FromInt64(-10)

	// 5 + 10 = 15
	fifteen := jsbi.Add(five, ten)
	if !jsbi.Equal(fifteen, jsbi.FromInt64(15)) {
		t.Errorf("Add(5, 10) failed")
	}

	// -5 + -10 = -15
	negFifteen := jsbi.Add(negFive, negTen)
	if !jsbi.Equal(negFifteen, jsbi.FromInt64(-15)) {
		t.Errorf("Add(-5, -10) failed")
	}

	// 10 + -5 = 5
	res1 := jsbi.Add(ten, negFive)
	if !jsbi.Equal(res1, five) {
		t.Errorf("Add(10, -5) failed")
	}

	// 5 + -10 = -5
	res2 := jsbi.Add(five, negTen)
	if !jsbi.Equal(res2, negFive) {
		t.Errorf("Add(5, -10) failed")
	}
}

func TestSubtractBasic(t *testing.T) {
	five := jsbi.FromInt64(5)
	ten := jsbi.FromInt64(10)
	negFive := jsbi.FromInt64(-5)

	// 10 - 5 = 5
	res1 := jsbi.Subtract(ten, five)
	if !jsbi.Equal(res1, five) {
		t.Errorf("Subtract(10, 5) failed")
	}

	// 5 - 10 = -5
	res2 := jsbi.Subtract(five, ten)
	if !jsbi.Equal(res2, negFive) {
		t.Errorf("Subtract(5, 10) failed")
	}

	// 5 - 5 = 0 (Canonical Zero check)
	zero := jsbi.Subtract(five, five)
	if zero.Sign() != false || zero.Length() != 0 {
		t.Errorf("Subtract(5, 5) produced non-canonical zero: sign=%v len=%d", zero.Sign(), zero.Length())
	}
}

func TestCarryPropagation(t *testing.T) {
	// (0x3FFFFFFF, 0x3FFFFFFF, 0x3FFFFFFF) + 1
	op, _ := jsbi.FromString("68719476735", 10) // 2^36 - 1 = 68719476735
	one := jsbi.FromInt64(1)

	res := jsbi.Add(op, one)
	want, _ := jsbi.FromString("68719476736", 10) // 2^36
	if !jsbi.Equal(res, want) {
		t.Errorf("Add carry propagation failed")
	}
}

func TestBorrowPropagation(t *testing.T) {
	// (2^36) - 1 = 2^36 - 1
	op, _ := jsbi.FromString("68719476736", 10)
	one := jsbi.FromInt64(1)

	res := jsbi.Subtract(op, one)
	want, _ := jsbi.FromString("68719476735", 10)
	if !jsbi.Equal(res, want) {
		t.Errorf("Subtract borrow propagation failed")
	}
}

func TestAlgebraicIdentities(t *testing.T) {
	x, _ := jsbi.FromString("12345678901234567890", 10)
	y, _ := jsbi.FromString("98765432109876543210", 10)

	// (X + Y) - Y == X
	sum := jsbi.Add(x, y)
	diff := jsbi.Subtract(sum, y)
	if !jsbi.Equal(diff, x) {
		t.Errorf("(X + Y) - Y != X")
	}

	// (X - Y) + Y == X
	diff2 := jsbi.Subtract(x, y)
	sum2 := jsbi.Add(diff2, y)
	if !jsbi.Equal(sum2, x) {
		t.Errorf("(X - Y) + Y != X")
	}
}

func TestDifferentialFuzzCluster3(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping 60s+ differential fuzzing in short mode")
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Root directory is either cwd or parent of tests/port
	rootDir := cwd
	for {
		if _, err := os.Stat(filepath.Join(rootDir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(rootDir)
		if parent == rootDir {
			t.Fatalf("Could not locate workspace root (go.mod) from %s", cwd)
		}
		rootDir = parent
	}

	harnessPath := filepath.Join(rootDir, "fuzz", "harness", "fuzz_cluster3.go")
	goBin := filepath.Join(rootDir, "go_sdk", "go", "bin", "go.exe")
	if _, err := os.Stat(goBin); err != nil {
		goBin = "go"
	}

	cmd := exec.Command(goBin, "run", harnessPath)
	cmd.Dir = rootDir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	t.Logf("Fuzz output:\n%s", outBuf.String())
	if err != nil {
		t.Fatalf("Differential fuzzing failed: %v, stderr: %s", err, errBuf.String())
	}
}

func BenchmarkAdd(b *testing.B) {
	x, _ := jsbi.FromString("12345678901234567890", 10)
	y, _ := jsbi.FromString("98765432109876543210", 10)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = jsbi.Add(x, y)
	}
}

func BenchmarkSubtract(b *testing.B) {
	x, _ := jsbi.FromString("98765432109876543210", 10)
	y, _ := jsbi.FromString("12345678901234567890", 10)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = jsbi.Subtract(x, y)
	}
}
