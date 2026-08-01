package jsbi

// UnaryMinus returns the negation of x (-x).
// Sourced from JSBI.unaryMinus (jsbi.ts lines 151-156).
func UnaryMinus(x *BigInt) *BigInt {
	if x == nil || x.Length() == 0 {
		return Zero()
	}
	result := x.Copy()
	result.SetSign(!x.Sign())
	return result
}

// Add returns the sum of x and y (x + y).
// Sourced from JSBI.add (jsbi.ts lines 269-282).
func Add(x, y *BigInt) *BigInt {
	if x == nil {
		x = Zero()
	}
	if y == nil {
		y = Zero()
	}
	sign := x.Sign()
	if sign == y.Sign() {
		// x + y == x + y
		// -x + -y == -(x + y)
		return absoluteAdd(x, y, sign)
	}
	// x + -y == x - y == -(y - x)
	// -x + y == y - x == -(x - y)
	if absoluteCompare(x, y) >= 0 {
		return absoluteSub(x, y, sign)
	}
	return absoluteSub(y, x, !sign)
}

// Subtract returns the difference of x and y (x - y).
// Sourced from JSBI.subtract (jsbi.ts lines 284-297).
func Subtract(x, y *BigInt) *BigInt {
	if x == nil {
		x = Zero()
	}
	if y == nil {
		y = Zero()
	}
	sign := x.Sign()
	if sign != y.Sign() {
		// x - (-y) == x + y
		// (-x) - y == -(x + y)
		return absoluteAdd(x, y, sign)
	}
	// x - y == -(y - x)
	// (-x) - (-y) == y - x == -(x - y)
	if absoluteCompare(x, y) >= 0 {
		return absoluteSub(x, y, sign)
	}
	return absoluteSub(y, x, !sign)
}

// absoluteAdd adds magnitudes of x and y with resultSign.
// Sourced from JSBI.__absoluteAdd (jsbi.ts lines 1215-1240).
func absoluteAdd(x, y *BigInt, resultSign bool) *BigInt {
	if x.Length() < y.Length() {
		return absoluteAdd(y, x, resultSign)
	}
	if x.Length() == 0 {
		return Zero()
	}
	if y.Length() == 0 {
		if x.Sign() == resultSign {
			return x.Copy()
		}
		return UnaryMinus(x)
	}
	resultLength := x.Length()
	if x.clzmsd() == 0 || (y.Length() == x.Length() && y.clzmsd() == 0) {
		resultLength++
	}
	result := NewBigInt(resultLength, resultSign)
	var carry uint32 = 0
	i := 0
	for ; i < y.Length(); i++ {
		r := int32(x.Digit(i)) + int32(y.Digit(i)) + int32(carry)
		carry = uint32(r) >> 30
		result.SetDigit(i, uint32(r)&kDigitMask)
	}
	for ; i < x.Length(); i++ {
		r := int32(x.Digit(i)) + int32(carry)
		carry = uint32(r) >> 30
		result.SetDigit(i, uint32(r)&kDigitMask)
	}
	if i < result.Length() {
		result.SetDigit(i, carry)
	}
	return result.Trim()
}

// absoluteSub subtracts magnitude of y from x with resultSign. Assumes |x| >= |y|.
// Sourced from JSBI.__absoluteSub (jsbi.ts lines 1242-1259).
func absoluteSub(x, y *BigInt, resultSign bool) *BigInt {
	if x.Length() == 0 {
		return Zero()
	}
	if y.Length() == 0 {
		if x.Sign() == resultSign {
			return x.Copy()
		}
		return UnaryMinus(x)
	}
	result := NewBigInt(x.Length(), resultSign)
	var borrow uint32 = 0
	i := 0
	for ; i < y.Length(); i++ {
		r := int32(x.Digit(i)) - int32(y.Digit(i)) - int32(borrow)
		borrow = (uint32(r) >> 30) & 1
		result.SetDigit(i, uint32(r)&kDigitMask)
	}
	for ; i < x.Length(); i++ {
		r := int32(x.Digit(i)) - int32(borrow)
		borrow = (uint32(r) >> 30) & 1
		result.SetDigit(i, uint32(r)&kDigitMask)
	}
	return result.Trim()
}

// absoluteAddOne adds 1 to magnitude of x.
// Sourced from JSBI.__absoluteAddOne (jsbi.ts lines 1261-1278).
func absoluteAddOne(x *BigInt, resultSign bool) *BigInt {
	inputLength := x.Length()
	result := NewBigInt(inputLength, resultSign)
	var carry uint32 = 1
	for i := 0; i < inputLength; i++ {
		r := int32(x.Digit(i)) + int32(carry)
		carry = uint32(r) >> 30
		result.SetDigit(i, uint32(r)&kDigitMask)
	}
	if carry != 0 {
		// Allocate extra digit for carry expansion
		expanded := NewBigInt(inputLength+1, resultSign)
		for i := 0; i < inputLength; i++ {
			expanded.SetDigit(i, result.Digit(i))
		}
		expanded.SetDigit(inputLength, carry)
		return expanded.Trim()
	}
	return result.Trim()
}

// absoluteSubOne subtracts 1 from magnitude of x.
// Sourced from JSBI.__absoluteSubOne (jsbi.ts lines 1280-1295).
func absoluteSubOne(x *BigInt) *BigInt {
	length := x.Length()
	if length == 0 {
		return Zero()
	}
	result := NewBigInt(length, false)
	var borrow uint32 = 1
	for i := 0; i < length; i++ {
		r := int32(x.Digit(i)) - int32(borrow)
		borrow = (uint32(r) >> 30) & 1
		result.SetDigit(i, uint32(r)&kDigitMask)
	}
	return result.Trim()
}

// Helper method on BigInt for CLZ of MSD.
func (b *BigInt) clzmsd() int {
	if b.Length() == 0 {
		return 30
	}
	return clz30(b.Digit(b.Length() - 1))
}
