package port_test

import (
	"testing"

	jsbi "github.com/Madhavkannank/fafa/src"
)

func TestAsIntN(t *testing.T) {
	tests := []struct {
		bits int
		xStr string
		want string
	}{
		{4, "7", "7"},
		{4, "9", "-7"},
		{4, "-9", "7"},
		{4, "-7", "-7"},
		{0, "12345", "0"},
		{1, "1", "-1"},
		{1, "0", "0"},
		// 2^30-1 = 0x3FFFFFFF: bit 29 is the sign bit in 30-bit two's complement -> -1
		{30, "1073741823", "-1"},
		// 2^30 exceeds 30-bit range; truncated to bit-30 width -> 0
		{30, "1073741824", "0"},
		// 2^60-1: bit 59 is sign bit in 60-bit two's complement -> -1
		{60, "1152921504606846975", "-1"},
	}

	for _, tt := range tests {
		x, _ := jsbi.FromString(tt.xStr, 10)
		got, err := jsbi.AsIntN(tt.bits, x)
		want, _ := jsbi.FromString(tt.want, 10)

		if err != nil || !jsbi.Equal(got, want) {
			t.Errorf("AsIntN(%d, %s) = %v, err=%v, want %s", tt.bits, tt.xStr, got, err, tt.want)
		}
	}
}

func TestAsUintN(t *testing.T) {
	tests := []struct {
		bits int
		xStr string
		want string
	}{
		{4, "7", "7"},
		{4, "9", "9"},
		{4, "25", "9"}, // 25 % 16 = 9
		{4, "-3", "13"}, // 16 - 3 = 13
		{0, "12345", "0"},
		{1, "-1", "1"},
		{30, "1073741823", "1073741823"},
		{30, "-1", "1073741823"}, // 2^30 - 1
		{60, "-1", "1152921504606846975"}, // 2^60 - 1
	}

	for _, tt := range tests {
		x, _ := jsbi.FromString(tt.xStr, 10)
		got, err := jsbi.AsUintN(tt.bits, x)
		want, _ := jsbi.FromString(tt.want, 10)

		if err != nil || !jsbi.Equal(got, want) {
			t.Errorf("AsUintN(%d, %s) = %v, err=%v, want %s", tt.bits, tt.xStr, got, err, tt.want)
		}
	}
}

func TestAsIntNAsUintNErrors(t *testing.T) {
	x, _ := jsbi.FromString("100", 10)

	_, err1 := jsbi.AsIntN(-1, x)
	if err1 != jsbi.ErrRange {
		t.Errorf("AsIntN(-1) expected ErrRange, got %v", err1)
	}

	_, err2 := jsbi.AsUintN(-5, x)
	if err2 != jsbi.ErrRange {
		t.Errorf("AsUintN(-5) expected ErrRange, got %v", err2)
	}

	negX, _ := jsbi.FromString("-100", 10)
	_, err3 := jsbi.AsUintN(1<<30+1, negX)
	if err3 != jsbi.ErrRange {
		t.Errorf("AsUintN(huge, negative) expected ErrRange, got %v", err3)
	}
}

func TestAsUintNMultiLimbVector(t *testing.T) {
	// AsUintN(65, -0x4000000000000001) where -0x4000000000000001 = -4611686018427387905
	// JSBI oracle: 2^65 - (2^62+1) = 32281802128991715327
	// digits: [0x3FFFFFFF, 0x3FFFFFFF, 0x1B] (verified against Node.js JSBI)
	x, _ := jsbi.FromString("-4611686018427387905", 10)
	got, err := jsbi.AsUintN(65, x)
	want, _ := jsbi.FromString("32281802128991715327", 10)

	if err != nil || !jsbi.Equal(got, want) {
		t.Errorf("AsUintN(65, -4611686018427387905) = %v, err=%v, want %v", got, err, want)
	}
}

func TestTruncationValueIndependenceFastPath(t *testing.T) {
	x, _ := jsbi.FromString("12345", 10)

	// Fast path: n >= kMaxLengthBits
	got, err := jsbi.AsIntN(1000000000, x)
	if err != nil {
		t.Fatalf("AsIntN fast path error: %v", err)
	}

	// Assert returned pointer is NOT equal to input pointer
	if got == x {
		t.Errorf("Fast path value independence violated: returned pointer equals input pointer")
	}

	// Mutate returned object
	got.SetDigit(0, 99999)
	if x.Digit(0) != 12345 {
		t.Errorf("Fast path value independence violated: mutating result modified input x")
	}
}

func BenchmarkAsIntN(b *testing.B) {
	x, _ := jsbi.FromString("12345678901234567890", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = jsbi.AsIntN(32, x)
	}
}

func BenchmarkAsUintN(b *testing.B) {
	x, _ := jsbi.FromString("12345678901234567890", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = jsbi.AsUintN(32, x)
	}
}
