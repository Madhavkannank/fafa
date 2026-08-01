package port_test

import (
	"math/big"
	"testing"

	"github.com/Madhavkannank/fafa/src"
)

func TestMultiplyBasic(t *testing.T) {
	zero := jsbi.Zero()
	five := jsbi.FromInt64(5)
	ten := jsbi.FromInt64(10)
	negFive := jsbi.FromInt64(-5)
	negTen := jsbi.FromInt64(-10)

	// 0 * 5 = 0
	if !jsbi.Equal(jsbi.Multiply(zero, five), zero) {
		t.Errorf("Multiply(0, 5) failed")
	}

	// (-5) * 0 = 0 (canonical zero verification)
	resZero := jsbi.Multiply(negFive, zero)
	if resZero.Sign() != false || resZero.Length() != 0 {
		t.Errorf("Multiply(-5, 0) did not return canonical zero")
	}

	// 5 * 10 = 50
	fifty := jsbi.Multiply(five, ten)
	if !jsbi.Equal(fifty, jsbi.FromInt64(50)) {
		t.Errorf("Multiply(5, 10) = %v, want 50", fifty)
	}

	// (-5) * 10 = -50
	negFifty := jsbi.Multiply(negFive, ten)
	if !jsbi.Equal(negFifty, jsbi.FromInt64(-50)) {
		t.Errorf("Multiply(-5, 10) = %v, want -50", negFifty)
	}

	// (-5) * (-10) = 50
	posFifty := jsbi.Multiply(negFive, negTen)
	if !jsbi.Equal(posFifty, jsbi.FromInt64(50)) {
		t.Errorf("Multiply(-5, -10) = %v, want 50", posFifty)
	}
}

func TestMultiplyWorstCaseVectors(t *testing.T) {
	// Vector 1: Max 30-bit single limbs (2^30 - 1) * (2^30 - 1)
	maxLimbStr := "1073741823" // 0x3FFFFFFF
	x1, _ := jsbi.FromString(maxLimbStr, 10)
	y1, _ := jsbi.FromString(maxLimbStr, 10)

	prod1 := jsbi.Multiply(x1, y1)
	want1Str := "1152921502459363329"
	want1, _ := jsbi.FromString(want1Str, 10)
	if !jsbi.Equal(prod1, want1) {
		t.Errorf("Multiply(maxLimb, maxLimb) failed")
	}

	// Vector 2: Powers of two (2^30) * (2^30) = 2^60
	pow2Str := "1073741824" // 2^30
	x2, _ := jsbi.FromString(pow2Str, 10)
	y2, _ := jsbi.FromString(pow2Str, 10)

	prod2 := jsbi.Multiply(x2, y2)
	want2Str := "1152921504606846976" // 2^60
	want2, _ := jsbi.FromString(want2Str, 10)
	if !jsbi.Equal(prod2, want2) {
		t.Errorf("Multiply(2^30, 2^30) failed")
	}

	// Vector 3: Large multi-limb multiplication compared against math/big oracle
	strA := "9876543210987654321098765432109876543210"
	strB := "1234567890123456789012345678901234567890"
	bigA, _ := jsbi.FromString(strA, 10)
	bigB, _ := jsbi.FromString(strB, 10)

	prodAB := jsbi.Multiply(bigA, bigB)

	mbA, _ := new(big.Int).SetString(strA, 10)
	mbB, _ := new(big.Int).SetString(strB, 10)
	wantABStr := new(big.Int).Mul(mbA, mbB).String()
	wantAB, _ := jsbi.FromString(wantABStr, 10)

	if !jsbi.Equal(prodAB, wantAB) {
		t.Errorf("Multiply(largeA, largeB) failed")
	}
}

func BenchmarkMultiply(b *testing.B) {
	x, _ := jsbi.FromString("123456789012345678901234567890", 10)
	y, _ := jsbi.FromString("987654321098765432109876543210", 10)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = jsbi.Multiply(x, y)
	}
}
