package jsbi

// toShiftAmount converts a BigInt shift count into an int bit count.
// If x has more than 1 digit or exceeds max allowed bits (1 << 30),
// it returns -1 as an internal sentinel.
func toShiftAmount(x *BigInt) int {
	if x.Length() > 1 {
		return -1
	}
	val := x.UnsignedDigit(0)
	if val > kMaxLengthBits {
		return -1
	}
	return int(val)
}

// rightShiftByMaximum returns canonical -1 if sign is true, else canonical zero.
func rightShiftByMaximum(sign bool) *BigInt {
	if sign {
		return OneDigit(1, true)
	}
	return Zero()
}

// leftShiftByAbsolute shifts x left by abs(y) bits.
// Returns ErrRange if the shift amount exceeds max bit length.
func leftShiftByAbsolute(x, y *BigInt) (*BigInt, error) {
	shift := toShiftAmount(y)
	if shift < 0 {
		return nil, ErrRange
	}
	digitShift := shift / 30
	bitsShift := shift % 30
	length := x.Length()

	grow := false
	if bitsShift != 0 && (x.Digit(length-1)>>uint(30-bitsShift)) != 0 {
		grow = true
	}

	resultLength := length + digitShift
	if grow {
		resultLength++
	}

	result := NewBigInt(resultLength, x.Sign())

	if bitsShift == 0 {
		for i := 0; i < digitShift; i++ {
			result.SetDigit(i, 0)
		}
		for i := digitShift; i < resultLength; i++ {
			result.SetDigit(i, x.Digit(i-digitShift))
		}
	} else {
		var carry uint32 = 0
		for i := 0; i < digitShift; i++ {
			result.SetDigit(i, 0)
		}
		for i := 0; i < length; i++ {
			d := x.Digit(i)
			result.SetDigit(i+digitShift, ((d<<uint(bitsShift))&0x3FFFFFFF)|carry)
			carry = d >> uint(30-bitsShift)
		}
		if grow {
			result.SetDigit(length+digitShift, carry)
		}
	}

	return result.Trim(), nil
}

// rightShiftByAbsolute shifts x right by abs(y) bits.
func rightShiftByAbsolute(x, y *BigInt) *BigInt {
	length := x.Length()
	sign := x.Sign()
	shift := toShiftAmount(y)
	if shift < 0 {
		return rightShiftByMaximum(sign)
	}
	digitShift := shift / 30
	bitsShift := shift % 30
	resultLength := length - digitShift
	if resultLength <= 0 {
		return rightShiftByMaximum(sign)
	}

	mustRoundDown := false
	if sign {
		mask := uint32((1 << uint(bitsShift)) - 1)
		if (x.Digit(digitShift) & mask) != 0 {
			mustRoundDown = true
		} else {
			for i := 0; i < digitShift; i++ {
				if x.Digit(i) != 0 {
					mustRoundDown = true
					break
				}
			}
		}
	}

	if mustRoundDown && bitsShift == 0 {
		msd := x.Digit(length - 1)
		roundingCanOverflow := (^msd) == 0
		if roundingCanOverflow {
			resultLength++
		}
	}

	result := NewBigInt(resultLength, sign)

	if bitsShift == 0 {
		result.SetDigit(resultLength-1, 0)
		for i := digitShift; i < length; i++ {
			result.SetDigit(i-digitShift, x.Digit(i))
		}
	} else {
		carry := x.Digit(digitShift) >> uint(bitsShift)
		last := length - digitShift - 1
		for i := 0; i < last; i++ {
			d := x.Digit(i + digitShift + 1)
			result.SetDigit(i, ((d<<uint(30-bitsShift))&0x3FFFFFFF)|carry)
			carry = d >> uint(bitsShift)
		}
		result.SetDigit(last, carry)
	}

	if mustRoundDown {
		result = absoluteAddOne(result, true)
	}

	return result.Trim()
}

// LeftShift shifts x left by y bits (or right if y is negative).
func LeftShift(x, y *BigInt) (*BigInt, error) {
	if y.Length() == 0 || x.Length() == 0 {
		return x.Copy(), nil
	}
	if y.Sign() {
		return rightShiftByAbsolute(x, y), nil
	}
	return leftShiftByAbsolute(x, y)
}

// SignedRightShift shifts x right by y bits with sign preservation (floor division toward -inf).
func SignedRightShift(x, y *BigInt) (*BigInt, error) {
	if y.Length() == 0 || x.Length() == 0 {
		return x.Copy(), nil
	}
	if y.Sign() {
		return leftShiftByAbsolute(x, y)
	}
	return rightShiftByAbsolute(x, y), nil
}

// UnsignedRightShift returns ErrType because BigInts do not support unsigned right shift.
func UnsignedRightShift(x, y *BigInt) (*BigInt, error) {
	return nil, ErrType
}
