package port_test

import (
	"testing"

	jsbi "github.com/Madhavkannank/fafa/src"
)

func TestBitwiseNot(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0", "-1"},
		{"5", "-6"},
		{"-6", "5"},
		{"1073741823", "-1073741824"},  // 0x3FFFFFFF -> -0x40000000
		{"-1073741824", "1073741823"}, // -0x40000000 -> 0x3FFFFFFF
		{"4294967295", "-4294967296"},
	}

	for _, tt := range tests {
		x, _ := jsbi.FromString(tt.input, 10)
		got := jsbi.BitwiseNot(x)
		want, _ := jsbi.FromString(tt.want, 10)

		if !jsbi.Equal(got, want) {
			t.Errorf("BitwiseNot(%s) = %v, want %v", tt.input, got, want)
		}
	}
}

func TestBitwiseAndSignCombinations(t *testing.T) {
	tests := []struct {
		xStr string
		yStr string
		want string
	}{
		// (+,+)
		{"12", "10", "8"},
		{"4294967301", "4294967299", "4294967297"}, // 0x100000005 & 0x100000003 = 0x100000001
		// (-,-)
		{"-12", "-10", "-12"},
		{"-5", "-5", "-5"},
		// (+,-)
		{"12", "-5", "8"},
		// (-,+)
		{"-5", "12", "8"},
		{"0", "-100", "0"},
	}

	for _, tt := range tests {
		x, _ := jsbi.FromString(tt.xStr, 10)
		y, _ := jsbi.FromString(tt.yStr, 10)
		got := jsbi.BitwiseAnd(x, y)
		want, _ := jsbi.FromString(tt.want, 10)

		if !jsbi.Equal(got, want) {
			t.Errorf("BitwiseAnd(%s, %s) = %v, want %v", tt.xStr, tt.yStr, got, want)
		}
	}
}

func TestBitwiseOrSignCombinations(t *testing.T) {
	tests := []struct {
		xStr string
		yStr string
		want string
	}{
		// (+,+)
		{"12", "10", "14"},
		{"0", "5", "5"},
		// (-,-)
		{"-12", "-10", "-10"},
		// (+,-)
		{"12", "-5", "-1"},
		// (-,+)
		{"-5", "12", "-1"},
		{"0", "-5", "-5"},
	}

	for _, tt := range tests {
		x, _ := jsbi.FromString(tt.xStr, 10)
		y, _ := jsbi.FromString(tt.yStr, 10)
		got := jsbi.BitwiseOr(x, y)
		want, _ := jsbi.FromString(tt.want, 10)

		if !jsbi.Equal(got, want) {
			t.Errorf("BitwiseOr(%s, %s) = %v, want %v", tt.xStr, tt.yStr, got, want)
		}
	}
}

func TestBitwiseXorSignCombinations(t *testing.T) {
	tests := []struct {
		xStr string
		yStr string
		want string
	}{
		// (+,+)
		{"12", "10", "6"},
		{"4294967301", "4294967299", "6"},
		// (-,-)
		{"-12", "-10", "2"},
		{"-1073741825", "-1073741825", "0"}, // -0x40000001 ^ -0x40000001 = 0
		// (+,-)
		{"12", "-5", "-9"},
		// (-,+)
		{"-5", "12", "-9"},
		{"0", "-1", "-1"},
	}

	for _, tt := range tests {
		x, _ := jsbi.FromString(tt.xStr, 10)
		y, _ := jsbi.FromString(tt.yStr, 10)
		got := jsbi.BitwiseXor(x, y)
		want, _ := jsbi.FromString(tt.want, 10)

		if !jsbi.Equal(got, want) {
			t.Errorf("BitwiseXor(%s, %s) = %v, want %v", tt.xStr, tt.yStr, got, want)
		}
	}
}

func TestBitwiseLargeMultiLimbVectors(t *testing.T) {
	// 8 limbs (all ones: 0x3FFFFFFF in each limb)
	x8 := jsbi.NewBigInt(8, false)
	y8 := jsbi.NewBigInt(8, false)
	for i := 0; i < 8; i++ {
		x8.SetDigit(i, 0x3FFFFFFF)
		y8.SetDigit(i, 0x15555555) // alternating mask
	}

	and8 := jsbi.BitwiseAnd(x8, y8)
	if and8.Length() != 8 {
		t.Errorf("BitwiseAnd 8-limb failed: expected len 8, got %d", and8.Length())
	}
	for i := 0; i < 8; i++ {
		if and8.Digit(i) != 0x15555555 {
			t.Errorf("BitwiseAnd 8-limb digit[%d] = %X, want 0x15555555", i, and8.Digit(i))
		}
	}

	// 16 limbs
	x16 := jsbi.NewBigInt(16, false)
	y16 := jsbi.NewBigInt(16, false)
	for i := 0; i < 16; i++ {
		x16.SetDigit(i, 0x3FFFFFFF)
		y16.SetDigit(i, 0x2AAAAAAA)
	}
	or16 := jsbi.BitwiseOr(x16, y16)
	if or16.Length() != 16 {
		t.Errorf("BitwiseOr 16-limb failed: expected len 16, got %d", or16.Length())
	}

	// 32 limbs
	x32 := jsbi.NewBigInt(32, false)
	y32 := jsbi.NewBigInt(32, false)
	for i := 0; i < 32; i++ {
		x32.SetDigit(i, 0x15555555)
		y32.SetDigit(i, 0x2AAAAAAA)
	}
	xor32 := jsbi.BitwiseXor(x32, y32)
	if xor32.Length() != 32 {
		t.Errorf("BitwiseXor 32-limb failed: expected len 32, got %d", xor32.Length())
	}
	for i := 0; i < 32; i++ {
		if xor32.Digit(i) != 0x3FFFFFFF {
			t.Errorf("BitwiseXor 32-limb digit[%d] = %X, want 0x3FFFFFFF", i, xor32.Digit(i))
		}
	}

	// Carry propagation across 8 full limbs for BitwiseNot
	not8 := jsbi.BitwiseNot(x8)
	if !not8.Sign() || not8.Length() != 9 {
		t.Errorf("BitwiseNot carry propagation failed: expected negative len 9, got sign=%v len=%d", not8.Sign(), not8.Length())
	}

	// Borrow propagation across 8 zero limbs
	z8 := jsbi.NewBigInt(8, true) // -0x40000000000000000000000000000000...
	z8.SetDigit(7, 1)
	notZ8 := jsbi.BitwiseNot(z8)
	if notZ8.Sign() || notZ8.Length() != 7 {
		// ~(-x) = x - 1: borrow propagates across limbs 0..6
		for i := 0; i < 7; i++ {
			if notZ8.Digit(i) != 0x3FFFFFFF {
				t.Errorf("Borrow propagation failed at digit[%d]: %X, want 0x3FFFFFFF", i, notZ8.Digit(i))
			}
		}
	}
}

func TestBitwiseCanonicalZero(t *testing.T) {
	x, _ := jsbi.FromString("12345", 10)
	got := jsbi.BitwiseXor(x, x)

	if got.Length() != 0 || got.Sign() != false {
		t.Errorf("Canonical Zero invariant violated: Length()=%d, Sign()=%v", got.Length(), got.Sign())
	}
}

func TestBitwiseValueIndependenceByteForByte(t *testing.T) {
	x, _ := jsbi.FromString("12345678901234567890", 10)
	y, _ := jsbi.FromString("-98765432109876543210", 10)

	// Snapshot original digit values and signs
	xLen, xSign := x.Length(), x.Sign()
	xDigits := make([]uint32, xLen)
	for i := 0; i < xLen; i++ {
		xDigits[i] = x.Digit(i)
	}

	yLen, ySign := y.Length(), y.Sign()
	yDigits := make([]uint32, yLen)
	for i := 0; i < yLen; i++ {
		yDigits[i] = y.Digit(i)
	}

	resAnd := jsbi.BitwiseAnd(x, y)
	resOr := jsbi.BitwiseOr(x, y)
	resXor := jsbi.BitwiseXor(x, y)
	resNot := jsbi.BitwiseNot(x)

	// Mutate results
	resAnd.SetDigit(0, 99999)
	resOr.SetDigit(0, 99999)
	resXor.SetDigit(0, 99999)
	resNot.SetDigit(0, 99999)

	// Assert byte-for-byte immutability of input x
	if x.Length() != xLen || x.Sign() != xSign {
		t.Errorf("Byte-for-byte immutability of x header violated")
	}
	for i := 0; i < xLen; i++ {
		if x.Digit(i) != xDigits[i] {
			t.Errorf("Byte-for-byte immutability of x digit[%d] violated: %d != %d", i, x.Digit(i), xDigits[i])
		}
	}

	// Assert byte-for-byte immutability of input y
	if y.Length() != yLen || y.Sign() != ySign {
		t.Errorf("Byte-for-byte immutability of y header violated")
	}
	for i := 0; i < yLen; i++ {
		if y.Digit(i) != yDigits[i] {
			t.Errorf("Byte-for-byte immutability of y digit[%d] violated: %d != %d", i, y.Digit(i), yDigits[i])
		}
	}
}

func BenchmarkBitwiseNot(b *testing.B) {
	x, _ := jsbi.FromString("12345678901234567890", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = jsbi.BitwiseNot(x)
	}
}

func BenchmarkBitwiseAnd(b *testing.B) {
	x, _ := jsbi.FromString("12345678901234567890", 10)
	y, _ := jsbi.FromString("-98765432109876543210", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = jsbi.BitwiseAnd(x, y)
	}
}

func BenchmarkBitwiseOr(b *testing.B) {
	x, _ := jsbi.FromString("12345678901234567890", 10)
	y, _ := jsbi.FromString("-98765432109876543210", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = jsbi.BitwiseOr(x, y)
	}
}

func BenchmarkBitwiseXor(b *testing.B) {
	x, _ := jsbi.FromString("12345678901234567890", 10)
	y, _ := jsbi.FromString("-98765432109876543210", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = jsbi.BitwiseXor(x, y)
	}
}
