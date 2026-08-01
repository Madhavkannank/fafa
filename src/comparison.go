package jsbi

import (
	"math"
)

// Compare compares x and y and returns:
//
//	-1 if x < y
//	 0 if x == y
//	+1 if x > y
func Compare(x, y *BigInt) int {
	if x == nil {
		x = Zero()
	}
	if y == nil {
		y = Zero()
	}
	xSign := x.Sign()
	ySign := y.Sign()
	if xSign != ySign {
		if xSign {
			return -1
		}
		return 1
	}
	absResult := absoluteCompare(x, y)
	if absResult > 0 {
		if xSign {
			return -1
		}
		return 1
	}
	if absResult < 0 {
		if xSign {
			return 1
		}
		return -1
	}
	return 0
}

// Equal returns true if x == y.
func Equal(x, y *BigInt) bool {
	if x == nil {
		x = Zero()
	}
	if y == nil {
		y = Zero()
	}
	if x.Sign() != y.Sign() {
		return false
	}
	if x.Length() != y.Length() {
		return false
	}
	for i := 0; i < x.Length(); i++ {
		if x.Digit(i) != y.Digit(i) {
			return false
		}
	}
	return true
}

// NotEqual returns true if x != y.
func NotEqual(x, y *BigInt) bool {
	return !Equal(x, y)
}

// LessThan returns true if x < y.
func LessThan(x, y *BigInt) bool {
	return Compare(x, y) < 0
}

// LessThanOrEqual returns true if x <= y.
func LessThanOrEqual(x, y *BigInt) bool {
	return Compare(x, y) <= 0
}

// GreaterThan returns true if x > y.
func GreaterThan(x, y *BigInt) bool {
	return Compare(x, y) > 0
}

// GreaterThanOrEqual returns true if x >= y.
func GreaterThanOrEqual(x, y *BigInt) bool {
	return Compare(x, y) >= 0
}

// absoluteCompare compares the magnitude of x and y ignoring sign.
func absoluteCompare(x, y *BigInt) int {
	diff := x.Length() - y.Length()
	if diff != 0 {
		return diff
	}
	for i := x.Length() - 1; i >= 0; i-- {
		xd := x.Digit(i)
		yd := y.Digit(i)
		if xd > yd {
			return 1
		}
		if xd < yd {
			return -1
		}
	}
	return 0
}

// CompareToFloat64 compares a BigInt x with a float64 y without precision loss.
// Returns (cmp, isNaN). If isNaN is true, y is NaN and no ordering exists.
// Sourced line-for-line from JSBI.__compareToDouble (jsbi.ts lines 1051-1140).
func CompareToFloat64(x *BigInt, y float64) (cmp int, isNaN bool) {
	if x == nil {
		x = Zero()
	}
	if math.IsNaN(y) {
		return 0, true
	}
	if math.IsInf(y, 1) {
		return -1, false
	}
	if math.IsInf(y, -1) {
		return 1, false
	}
	xSign := x.Sign()
	ySign := y < 0
	if xSign != ySign {
		if xSign {
			return -1, false
		}
		return 1, false
	}
	if y == 0 {
		if x.Length() == 0 {
			return 0, false
		}
		if xSign {
			return -1, false
		}
		return 1, false
	}
	if x.Length() == 0 {
		if ySign {
			return 1, false
		}
		return -1, false
	}

	bits := math.Float64bits(y)
	rawExponent := int((bits >> 52) & 0x7FF)
	exponent := rawExponent - 0x3FF

	if exponent < 0 {
		if xSign {
			return -1, false
		}
		return 1, false
	}

	xLength := x.Length()
	xMsd := x.Digit(xLength - 1)
	msdLeadingZeros := clz30(xMsd)
	xBitLength := xLength*30 - msdLeadingZeros
	yBitLength := exponent + 1

	if xBitLength < yBitLength {
		if xSign {
			return 1, false
		}
		return -1, false
	}
	if xBitLength > yBitLength {
		if xSign {
			return -1, false
		}
		return 1, false
	}

	// Same sign, same integer bit length. Compare mantissa bit for bit.
	const kHiddenBit uint32 = 0x00100000
	mantissaHigh := uint32((bits>>32)&0xFFFFF) | kHiddenBit
	mantissaLow := uint32(bits & 0xFFFFFFFF)

	const kMantissaHighTopBit = 20
	msdTopBit := 29 - msdLeadingZeros

	var compareMantissa uint32
	remainingMantissaBits := 0

	if msdTopBit < kMantissaHighTopBit {
		shift := kMantissaHighTopBit - msdTopBit
		remainingMantissaBits = shift + 32
		compareMantissa = mantissaHigh >> shift
		mantissaHigh = (mantissaHigh << (32 - shift)) | (mantissaLow >> shift)
		mantissaLow = mantissaLow << (32 - shift)
	} else if msdTopBit == kMantissaHighTopBit {
		remainingMantissaBits = 32
		compareMantissa = mantissaHigh
		mantissaHigh = mantissaLow
		mantissaLow = 0
	} else {
		shift := msdTopBit - kMantissaHighTopBit
		remainingMantissaBits = 32 - shift
		compareMantissa = (mantissaHigh << shift) | (mantissaLow >> (32 - shift))
		mantissaHigh = mantissaLow << shift
		mantissaLow = 0
	}

	if xMsd > compareMantissa {
		if xSign {
			return -1, false
		}
		return 1, false
	}
	if xMsd < compareMantissa {
		if xSign {
			return 1, false
		}
		return -1, false
	}

	for digitIndex := xLength - 2; digitIndex >= 0; digitIndex-- {
		if remainingMantissaBits > 0 {
			remainingMantissaBits -= 30
			compareMantissa = mantissaHigh >> 2
			mantissaHigh = (mantissaHigh << 30) | (mantissaLow >> 2)
			mantissaLow = mantissaLow << 30
		} else {
			compareMantissa = 0
		}
		digit := x.Digit(digitIndex)
		if digit > compareMantissa {
			if xSign {
				return -1, false
			}
			return 1, false
		}
		if digit < compareMantissa {
			if xSign {
				return 1, false
			}
			return -1, false
		}
	}

	if mantissaHigh != 0 || mantissaLow != 0 {
		if xSign {
			return 1, false
		}
		return -1, false
	}

	return 0, false
}

// Float64 Relational Predicates

// EqualToFloat64 returns true if BigInt x is numerically equal to float64 y.
func EqualToFloat64(x *BigInt, y float64) bool {
	cmp, isNaN := CompareToFloat64(x, y)
	return !isNaN && cmp == 0
}

// NotEqualToFloat64 returns true if BigInt x is not equal to float64 y.
func NotEqualToFloat64(x *BigInt, y float64) bool {
	return !EqualToFloat64(x, y)
}

// LessThanFloat64 returns true if BigInt x is strictly less than float64 y.
func LessThanFloat64(x *BigInt, y float64) bool {
	cmp, isNaN := CompareToFloat64(x, y)
	return !isNaN && cmp < 0
}

// LessThanOrEqualFloat64 returns true if BigInt x is less than or equal to float64 y.
func LessThanOrEqualFloat64(x *BigInt, y float64) bool {
	cmp, isNaN := CompareToFloat64(x, y)
	return !isNaN && cmp <= 0
}

// GreaterThanFloat64 returns true if BigInt x is strictly greater than float64 y.
func GreaterThanFloat64(x *BigInt, y float64) bool {
	cmp, isNaN := CompareToFloat64(x, y)
	return !isNaN && cmp > 0
}

// GreaterThanOrEqualFloat64 returns true if BigInt x is greater than or equal to float64 y.
func GreaterThanOrEqualFloat64(x *BigInt, y float64) bool {
	cmp, isNaN := CompareToFloat64(x, y)
	return !isNaN && cmp >= 0
}

// CompareToInt64 compares a BigInt x with a signed 64-bit integer y.
func CompareToInt64(x *BigInt, y int64) int {
	return Compare(x, FromInt64(y))
}

// Dynamic Comparison Operators (ECMAScript Operators Emulation)

// EQ returns true if x == y (abstract equality).
func EQ(x, y interface{}) bool {
	bx, errX := BigIntVal(x)
	by, errY := BigIntVal(y)
	if errX == nil && errY == nil {
		return Equal(bx, by)
	}
	if fx, ok := x.(float64); ok && errY == nil {
		return EqualToFloat64(by, fx)
	}
	if fy, ok := y.(float64); ok && errX == nil {
		return EqualToFloat64(bx, fy)
	}
	return false
}

// NE returns true if x != y.
func NE(x, y interface{}) bool {
	return !EQ(x, y)
}

// LT returns true if x < y.
func LT(x, y interface{}) bool {
	bx, errX := BigIntVal(x)
	by, errY := BigIntVal(y)
	if errX == nil && errY == nil {
		return LessThan(bx, by)
	}
	if fx, ok := x.(float64); ok && errY == nil {
		return GreaterThanFloat64(by, fx)
	}
	if fy, ok := y.(float64); ok && errX == nil {
		return LessThanFloat64(bx, fy)
	}
	return false
}

// LE returns true if x <= y.
func LE(x, y interface{}) bool {
	bx, errX := BigIntVal(x)
	by, errY := BigIntVal(y)
	if errX == nil && errY == nil {
		return LessThanOrEqual(bx, by)
	}
	if fx, ok := x.(float64); ok && errY == nil {
		return GreaterThanOrEqualFloat64(by, fx)
	}
	if fy, ok := y.(float64); ok && errX == nil {
		return LessThanOrEqualFloat64(bx, fy)
	}
	return false
}

// GT returns true if x > y.
func GT(x, y interface{}) bool {
	bx, errX := BigIntVal(x)
	by, errY := BigIntVal(y)
	if errX == nil && errY == nil {
		return GreaterThan(bx, by)
	}
	if fx, ok := x.(float64); ok && errY == nil {
		return LessThanFloat64(by, fx)
	}
	if fy, ok := y.(float64); ok && errX == nil {
		return GreaterThanFloat64(bx, fy)
	}
	return false
}

// GE returns true if x >= y.
func GE(x, y interface{}) bool {
	bx, errX := BigIntVal(x)
	by, errY := BigIntVal(y)
	if errX == nil && errY == nil {
		return GreaterThanOrEqual(bx, by)
	}
	if fx, ok := x.(float64); ok && errY == nil {
		return LessThanOrEqualFloat64(by, fx)
	}
	if fy, ok := y.(float64); ok && errX == nil {
		return GreaterThanOrEqualFloat64(bx, fy)
	}
	return false
}
