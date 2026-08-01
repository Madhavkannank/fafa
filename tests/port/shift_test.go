package port_test

import (
	"math/big"
	"testing"

	jsbi "github.com/Madhavkannank/fafa/src"
)

func TestShiftZeroOperand(t *testing.T) {
	x, _ := jsbi.FromString("0", 10)
	y, _ := jsbi.FromString("5", 10)

	gotL, errL := jsbi.LeftShift(x, y)
	if errL != nil || gotL.Length() != 0 || gotL.Sign() != false {
		t.Errorf("LeftShift(0, 5) failed: got len=%d sign=%v", gotL.Length(), gotL.Sign())
	}

	gotR, errR := jsbi.SignedRightShift(x, y)
	if errR != nil || gotR.Length() != 0 || gotR.Sign() != false {
		t.Errorf("SignedRightShift(0, 5) failed: got len=%d sign=%v", gotR.Length(), gotR.Sign())
	}
}

func TestShiftZeroShift(t *testing.T) {
	x, _ := jsbi.FromString("12345", 10)
	y, _ := jsbi.FromString("0", 10)

	gotL, errL := jsbi.LeftShift(x, y)
	if errL != nil || !jsbi.Equal(gotL, x) {
		t.Errorf("LeftShift(12345, 0) failed")
	}

	gotR, errR := jsbi.SignedRightShift(x, y)
	if errR != nil || !jsbi.Equal(gotR, x) {
		t.Errorf("SignedRightShift(12345, 0) failed")
	}
}

func TestShiftPositiveShift(t *testing.T) {
	x, _ := jsbi.FromString("13", 10)
	y, _ := jsbi.FromString("3", 10)

	got, err := jsbi.LeftShift(x, y)
	want, _ := jsbi.FromString("104", 10)
	if err != nil || !jsbi.Equal(got, want) {
		t.Errorf("LeftShift(13, 3) failed")
	}
}

func TestShiftNegativeShiftDispatch(t *testing.T) {
	x, _ := jsbi.FromString("13", 10)
	negY, _ := jsbi.FromString("-2", 10)

	// LeftShift(13, -2) should dispatch to SignedRightShift(13, 2) = 3
	gotL, errL := jsbi.LeftShift(x, negY)
	wantL, _ := jsbi.FromString("3", 10)
	if errL != nil || !jsbi.Equal(gotL, wantL) {
		t.Errorf("LeftShift(13, -2) dispatch failed")
	}

	negY3, _ := jsbi.FromString("-3", 10)
	// SignedRightShift(13, -3) should dispatch to LeftShift(13, 3) = 104
	gotR, errR := jsbi.SignedRightShift(x, negY3)
	wantR, _ := jsbi.FromString("104", 10)
	if errR != nil || !jsbi.Equal(gotR, wantR) {
		t.Errorf("SignedRightShift(13, -3) dispatch failed")
	}
}

func TestShiftExactly30Bits(t *testing.T) {
	x, _ := jsbi.FromString("1", 10)
	y, _ := jsbi.FromString("30", 10)

	got, err := jsbi.LeftShift(x, y)
	want, _ := jsbi.FromString("1073741824", 10) // 1 << 30
	if err != nil || !jsbi.Equal(got, want) {
		t.Errorf("LeftShift(1, 30) failed")
	}
}

func TestShiftGreaterThan30Bits(t *testing.T) {
	x, _ := jsbi.FromString("1", 10)
	y, _ := jsbi.FromString("35", 10)

	got, err := jsbi.LeftShift(x, y)
	want, _ := jsbi.FromString("34359738368", 10) // 1 << 35
	if err != nil || !jsbi.Equal(got, want) {
		t.Errorf("LeftShift(1, 35) failed")
	}
}

func TestShiftMultiLimbLeft(t *testing.T) {
	x, _ := jsbi.FromString("123456789012345678901234567890", 10)
	y, _ := jsbi.FromString("45", 10)

	got, err := jsbi.LeftShift(x, y)

	bx, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	wantBig := new(big.Int).Lsh(bx, 45)
	want, _ := jsbi.FromString(wantBig.String(), 10)

	if err != nil || !jsbi.Equal(got, want) {
		t.Errorf("LeftShift multi-limb failed")
	}
}

func TestShiftMultiLimbRight(t *testing.T) {
	x, _ := jsbi.FromString("123456789012345678901234567890", 10)
	y, _ := jsbi.FromString("45", 10)

	got, err := jsbi.SignedRightShift(x, y)

	bx, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	wantBig := new(big.Int).Rsh(bx, 45)
	want, _ := jsbi.FromString(wantBig.String(), 10)

	if err != nil || !jsbi.Equal(got, want) {
		t.Errorf("SignedRightShift multi-limb failed")
	}
}

func TestShiftRemovingAllBits(t *testing.T) {
	x, _ := jsbi.FromString("13", 10)
	y, _ := jsbi.FromString("100", 10)

	got, err := jsbi.SignedRightShift(x, y)
	want := jsbi.Zero()
	if err != nil || !jsbi.Equal(got, want) {
		t.Errorf("SignedRightShift removing all bits failed")
	}
}

func TestRightShiftByMaximumZero(t *testing.T) {
	x, _ := jsbi.FromString("100", 10)
	y, _ := jsbi.FromString("1000", 10)

	got, err := jsbi.SignedRightShift(x, y)
	if err != nil || got.Length() != 0 || got.Sign() != false {
		t.Errorf("rightShiftByMaximum(false) failed: expected 0, got len=%d sign=%v", got.Length(), got.Sign())
	}
}

func TestRightShiftByMaximumMinusOne(t *testing.T) {
	x, _ := jsbi.FromString("-100", 10)
	y, _ := jsbi.FromString("1000", 10)

	got, err := jsbi.SignedRightShift(x, y)
	want, _ := jsbi.FromString("-1", 10)
	if err != nil || !jsbi.Equal(got, want) {
		t.Errorf("rightShiftByMaximum(true) failed: expected -1")
	}
}

func TestToShiftAmountOverflow(t *testing.T) {
	x, _ := jsbi.FromString("1", 10)
	// Shift count > 1 limb
	hugeY, _ := jsbi.FromString("100000000000000000000", 10)

	_, err := jsbi.LeftShift(x, hugeY)
	if err != jsbi.ErrRange {
		t.Errorf("LeftShift with huge shift expected ErrRange, got: %v", err)
	}
}

func TestUnsignedRightShiftErrType(t *testing.T) {
	x, _ := jsbi.FromString("100", 10)
	y, _ := jsbi.FromString("2", 10)

	_, err := jsbi.UnsignedRightShift(x, y)
	if err != jsbi.ErrType {
		t.Errorf("UnsignedRightShift expected ErrType, got: %v", err)
	}
}

func TestShiftCanonicalZero(t *testing.T) {
	x, _ := jsbi.FromString("0", 10)
	y, _ := jsbi.FromString("10", 10)

	got, _ := jsbi.LeftShift(x, y)
	if got.Sign() != false || got.Length() != 0 {
		t.Errorf("Canonical zero assertion failed for LeftShift(0, 10)")
	}
}

func TestShiftValueIndependence(t *testing.T) {
	x, _ := jsbi.FromString("13", 10)
	y, _ := jsbi.FromString("3", 10)

	got, _ := jsbi.LeftShift(x, y)
	// Mutate got
	got.SetDigit(0, 999)

	// Verify input x was not mutated
	if x.Digit(0) != 13 {
		t.Errorf("Value independence violated: mutating result mutated input x")
	}
}

func TestShiftFloorDivisionVectors(t *testing.T) {
	vectors := []struct {
		xStr string
		yStr string
		want string
	}{
		{"-3", "1", "-2"},
		{"-5", "1", "-3"},
		{"-7", "2", "-2"},
		{"-9", "3", "-2"},
		{"-17", "4", "-2"},
	}

	for _, v := range vectors {
		x, _ := jsbi.FromString(v.xStr, 10)
		y, _ := jsbi.FromString(v.yStr, 10)

		got, err := jsbi.SignedRightShift(x, y)
		want, _ := jsbi.FromString(v.want, 10)

		if err != nil || !jsbi.Equal(got, want) {
			t.Errorf("SignedRightShift(%s, %s) = %v, want %s", v.xStr, v.yStr, got, v.want)
		}
	}
}
