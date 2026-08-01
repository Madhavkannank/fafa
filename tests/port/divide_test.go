package port_test

import (
	"math/big"
	"testing"

	"github.com/Madhavkannank/fafa/src"
)

func TestDivideByZero(t *testing.T) {
	zero := jsbi.Zero()
	five := jsbi.FromInt64(5)

	if _, err := jsbi.Divide(five, zero); err != jsbi.ErrRange {
		t.Errorf("Divide by zero want ErrRange, got %v", err)
	}
	if _, err := jsbi.Remainder(five, zero); err != jsbi.ErrRange {
		t.Errorf("Remainder by zero want ErrRange, got %v", err)
	}
	if _, _, err := jsbi.DivRem(five, zero); err != jsbi.ErrRange {
		t.Errorf("DivRem by zero want ErrRange, got %v", err)
	}
}

func TestDivideTruncationAndSigns(t *testing.T) {
	cases := []struct {
		x, y, wantQ, wantR int64
	}{
		{7, 3, 2, 1},
		{-7, 3, -2, -1},
		{7, -3, -2, 1},
		{-7, -3, 2, -1},
		{10, 2, 5, 0},
		{-10, 2, -5, 0},
		{10, -2, -5, 0},
		{-10, -2, 5, 0},
	}

	for _, c := range cases {
		bx := jsbi.FromInt64(c.x)
		by := jsbi.FromInt64(c.y)
		wantQ := jsbi.FromInt64(c.wantQ)
		wantR := jsbi.FromInt64(c.wantR)

		q, err := jsbi.Divide(bx, by)
		if err != nil || !jsbi.Equal(q, wantQ) {
			t.Errorf("Divide(%d, %d) = %v, err %v; want %d", c.x, c.y, q, err, c.wantQ)
		}

		r, err := jsbi.Remainder(bx, by)
		if err != nil || !jsbi.Equal(r, wantR) {
			t.Errorf("Remainder(%d, %d) = %v, err %v; want %d", c.x, c.y, r, err, c.wantR)
		}

		dq, dr, err := jsbi.DivRem(bx, by)
		if err != nil || !jsbi.Equal(dq, wantQ) || !jsbi.Equal(dr, wantR) {
			t.Errorf("DivRem(%d, %d) = (%v, %v), err %v; want (%d, %d)", c.x, c.y, dq, dr, err, c.wantQ, c.wantR)
		}
	}
}

func TestDivideDivisorGreaterThanDividend(t *testing.T) {
	x := jsbi.FromInt64(3)
	y := jsbi.FromInt64(7)

	q, err := jsbi.Divide(x, y)
	if err != nil || !jsbi.Equal(q, jsbi.Zero()) {
		t.Errorf("Divide(3, 7) want 0, got %v", q)
	}
	r, err := jsbi.Remainder(x, y)
	if err != nil || !jsbi.Equal(r, x) {
		t.Errorf("Remainder(3, 7) want 3, got %v", r)
	}

	dq, dr, err := jsbi.DivRem(x, y)
	if err != nil || !jsbi.Equal(dq, jsbi.Zero()) || !jsbi.Equal(dr, x) {
		t.Errorf("DivRem(3, 7) want (0, 3), got (%v, %v)", dq, dr)
	}
}

func TestDivideByOne(t *testing.T) {
	x := jsbi.FromInt64(123456789)
	one := jsbi.FromInt64(1)
	negOne := jsbi.FromInt64(-1)

	q1, _ := jsbi.Divide(x, one)
	if !jsbi.Equal(q1, x) {
		t.Errorf("Divide(x, 1) failed")
	}

	r1, _ := jsbi.Remainder(x, one)
	if !jsbi.Equal(r1, jsbi.Zero()) {
		t.Errorf("Remainder(x, 1) failed")
	}

	qNeg, _ := jsbi.Divide(x, negOne)
	if !jsbi.Equal(qNeg, jsbi.FromInt64(-123456789)) {
		t.Errorf("Divide(x, -1) failed")
	}
}

func TestDivideAlgorithmD(t *testing.T) {
	// Operands that force Knuth Algorithm D (divisor > 0x7FFF or multi-limb)
	// Example: 0xC0000001 / 0x8001
	strX := "3221225473" // 0xC0000001
	strY := "32769"      // 0x8001 (> 0x7FFF)

	x, _ := jsbi.FromString(strX, 10)
	y, _ := jsbi.FromString(strY, 10)

	q, err := jsbi.Divide(x, y)
	if err != nil {
		t.Fatalf("Divide error: %v", err)
	}
	r, err := jsbi.Remainder(x, y)
	if err != nil {
		t.Fatalf("Remainder error: %v", err)
	}

	// Verify against math/big QuoRem
	bigX, _ := new(big.Int).SetString(strX, 10)
	bigY, _ := new(big.Int).SetString(strY, 10)
	wantQ, wantR := new(big.Int), new(big.Int)
	wantQ.QuoRem(bigX, bigY, wantR)

	expectedQ, _ := jsbi.FromString(wantQ.String(), 10)
	expectedR, _ := jsbi.FromString(wantR.String(), 10)

	if !jsbi.Equal(q, expectedQ) {
		t.Errorf("Algorithm D Quotient failed, got %v, want %v", q, expectedQ)
	}
	if !jsbi.Equal(r, expectedR) {
		t.Errorf("Algorithm D Remainder failed, got %v, want %v", r, expectedR)
	}
}

func TestDivideLargeMultiLimb(t *testing.T) {
	strX := "98765432109876543210987654321098765432109876543210"
	strY := "123456789012345678901234"

	x, _ := jsbi.FromString(strX, 10)
	y, _ := jsbi.FromString(strY, 10)

	q, r, err := jsbi.DivRem(x, y)
	if err != nil {
		t.Fatalf("DivRem error: %v", err)
	}

	bigX, _ := new(big.Int).SetString(strX, 10)
	bigY, _ := new(big.Int).SetString(strY, 10)
	wantQ, wantR := new(big.Int), new(big.Int)
	wantQ.QuoRem(bigX, bigY, wantR)

	expectedQ, _ := jsbi.FromString(wantQ.String(), 10)
	expectedR, _ := jsbi.FromString(wantR.String(), 10)

	if !jsbi.Equal(q, expectedQ) {
		t.Errorf("Large DivRem Q failed")
	}
	if !jsbi.Equal(r, expectedR) {
		t.Errorf("Large DivRem R failed")
	}
}

func BenchmarkDivide(b *testing.B) {
	x, _ := jsbi.FromString("9876543210987654321098765432109876543210", 10)
	y, _ := jsbi.FromString("12345678901234567890", 10)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = jsbi.Divide(x, y)
	}
}

func BenchmarkRemainder(b *testing.B) {
	x, _ := jsbi.FromString("9876543210987654321098765432109876543210", 10)
	y, _ := jsbi.FromString("12345678901234567890", 10)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = jsbi.Remainder(x, y)
	}
}

func BenchmarkDivRem(b *testing.B) {
	x, _ := jsbi.FromString("9876543210987654321098765432109876543210", 10)
	y, _ := jsbi.FromString("12345678901234567890", 10)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = jsbi.DivRem(x, y)
	}
}
