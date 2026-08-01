package port_test

import (
	"math"
	"testing"

	"github.com/Madhavkannank/fafa/src"
)

func TestFromInt64(t *testing.T) {
	tests := []struct {
		input    int64
		wantSign bool
		wantLen  int
	}{
		{0, false, 0},
		{1, false, 1},
		{-1, true, 1},
		{0x3FFFFFFF, false, 1},
		{-0x3FFFFFFF, true, 1},
		{0x40000000, false, 2},
		{-0x40000000, true, 2},
		{math.MaxInt64, false, 3},
		{math.MinInt64, true, 3},
	}

	for _, tt := range tests {
		got := jsbi.FromInt64(tt.input)
		if got.Sign() != tt.wantSign {
			t.Errorf("FromInt64(%d).Sign() = %v, want %v", tt.input, got.Sign(), tt.wantSign)
		}
		if got.Length() != tt.wantLen {
			t.Errorf("FromInt64(%d).Length() = %d, want %d", tt.input, got.Length(), tt.wantLen)
		}
	}
}

func TestFromUint64(t *testing.T) {
	tests := []struct {
		input   uint64
		wantLen int
	}{
		{0, 0},
		{1, 1},
		{0x3FFFFFFF, 1},
		{0x40000000, 2},
		{math.MaxUint64, 3},
	}

	for _, tt := range tests {
		got := jsbi.FromUint64(tt.input)
		if got.Sign() {
			t.Errorf("FromUint64(%d).Sign() is true, want false", tt.input)
		}
		if got.Length() != tt.wantLen {
			t.Errorf("FromUint64(%d).Length() = %d, want %d", tt.input, got.Length(), tt.wantLen)
		}
	}
}

func TestFromBool(t *testing.T) {
	bTrue := jsbi.FromBool(true)
	if bTrue.Sign() || bTrue.Length() != 1 || bTrue.Digit(0) != 1 {
		t.Errorf("FromBool(true) unexpected result: sign=%v len=%d d0=%d", bTrue.Sign(), bTrue.Length(), bTrue.Digit(0))
	}

	bFalse := jsbi.FromBool(false)
	if bFalse.Sign() || bFalse.Length() != 0 {
		t.Errorf("FromBool(false) unexpected result: sign=%v len=%d", bFalse.Sign(), bFalse.Length())
	}
}

func TestFromFloat64(t *testing.T) {
	valid := []struct {
		input    float64
		wantSign bool
		wantLen  int
	}{
		{0.0, false, 0},
		{-0.0, false, 0},
		{1.0, false, 1},
		{-1.0, true, 1},
		{1e10, false, 2},
	}

	for _, tt := range valid {
		got, err := jsbi.FromFloat64(tt.input)
		if err != nil {
			t.Errorf("FromFloat64(%f) unexpected error: %v", tt.input, err)
			continue
		}
		if got.Sign() != tt.wantSign {
			t.Errorf("FromFloat64(%f).Sign() = %v, want %v", tt.input, got.Sign(), tt.wantSign)
		}
		if got.Length() != tt.wantLen {
			t.Errorf("FromFloat64(%f).Length() = %d, want %d", tt.input, got.Length(), tt.wantLen)
		}
	}

	invalid := []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
		1.5,
		-0.1,
	}

	for _, f := range invalid {
		_, err := jsbi.FromFloat64(f)
		if err != jsbi.ErrRange {
			t.Errorf("FromFloat64(%f) expected ErrRange, got: %v", f, err)
		}
	}
}

func TestFromString(t *testing.T) {
	valid := []struct {
		input    string
		radix    int
		wantSign bool
		wantLen  int
	}{
		{"0", 0, false, 0},
		{"0", 10, false, 0},
		{"12345", 10, false, 1},
		{"-12345", 10, true, 1},
		{"+12345", 10, false, 1},
		{"  \t 42 \n ", 0, false, 1},
		{"0x10", 0, false, 1},
		{"0XFF", 0, false, 1},
		{"0b1010", 0, false, 1},
		{"0o777", 0, false, 1},
		{"z", 36, false, 1},
	}

	for _, tt := range valid {
		got, err := jsbi.FromString(tt.input, tt.radix)
		if err != nil {
			t.Errorf("FromString(%q, %d) unexpected error: %v", tt.input, tt.radix, err)
			continue
		}
		if got.Sign() != tt.wantSign {
			t.Errorf("FromString(%q, %d).Sign() = %v, want %v", tt.input, tt.radix, got.Sign(), tt.wantSign)
		}
		if got.Length() != tt.wantLen {
			t.Errorf("FromString(%q, %d).Length() = %d, want %d", tt.input, tt.radix, got.Length(), tt.wantLen)
		}
	}

	syntaxErr := []struct {
		input string
		radix int
	}{
		{"12abc", 10},
		{"+0x10", 0},
		{"-0b10", 0},
		{"+", 0},
		{"-", 0},
		{"0x", 0},
	}

	for _, tt := range syntaxErr {
		_, err := jsbi.FromString(tt.input, tt.radix)
		if err != jsbi.ErrSyntax {
			t.Errorf("FromString(%q, %d) expected ErrSyntax, got: %v", tt.input, tt.radix, err)
		}
	}

	rangeErr := []struct {
		input string
		radix int
	}{
		{"10", 1},
		{"10", 37},
	}

	for _, tt := range rangeErr {
		_, err := jsbi.FromString(tt.input, tt.radix)
		if err != jsbi.ErrRange {
			t.Errorf("FromString(%q, %d) expected ErrRange, got: %v", tt.input, tt.radix, err)
		}
	}
}

func TestBigIntVal(t *testing.T) {
	b, err := jsbi.BigIntVal(int64(42))
	if err != nil || b.Length() != 1 || b.Digit(0) != 42 {
		t.Errorf("BigIntVal(int64(42)) failed: %v, b=%v", err, b)
	}

	bStr, err := jsbi.BigIntVal("100")
	if err != nil || bStr.Length() != 1 || bStr.Digit(0) != 100 {
		t.Errorf("BigIntVal(\"100\") failed: %v, b=%v", err, bStr)
	}

	_, errInvalid := jsbi.BigIntVal(struct{}{})
	if errInvalid != jsbi.ErrType {
		t.Errorf("BigIntVal(struct) expected ErrType, got: %v", errInvalid)
	}
}
