package jsbi

// truncateToNBits truncates x to n bits.
// Sourced from JSBI.__truncateToNBits (jsbi.ts lines 1860-1874).
func truncateToNBits(n int, x *BigInt) *BigInt {
	neededDigits := (n + 29) / 30
	result := NewBigInt(neededDigits, x.Sign())
	last := neededDigits - 1
	for i := 0; i < last; i++ {
		result.SetDigit(i, x.Digit(i))
	}
	var msd uint32
	if last < x.Length() {
		msd = x.Digit(last)
	}
	if n%30 != 0 {
		drop := 32 - (n % 30)
		msd = (msd << drop) >> drop
	}
	result.SetDigit(last, msd)
	return result.Trim()
}

// truncateAndSubFromPowerOfTwo simulates two's complement subtraction 2^n - |x|.
// Sourced from JSBI.__truncateAndSubFromPowerOfTwo (jsbi.ts lines 1876-1907).
func truncateAndSubFromPowerOfTwo(n int, x *BigInt, resultSign bool) *BigInt {
	neededDigits := (n + 29) / 30
	result := NewBigInt(neededDigits, resultSign)
	i := 0
	last := neededDigits - 1
	var borrow uint32 = 0
	limit := last
	if x.Length() < limit {
		limit = x.Length()
	}
	for ; i < limit; i++ {
		r := 0 - int64(x.Digit(i)) - int64(borrow)
		borrow = uint32((r >> 30) & 1)
		result.SetDigit(i, uint32(r)&0x3FFFFFFF)
	}
	for ; i < last; i++ {
		result.SetDigit(i, uint32(-int64(borrow)&0x3FFFFFFF))
	}
	var msd uint32 = 0
	if last < x.Length() {
		msd = x.Digit(last)
	}
	msdBitsConsumed := n % 30
	var resultMsd uint32
	if msdBitsConsumed == 0 {
		rMsd := 0 - int64(msd) - int64(borrow)
		resultMsd = uint32(rMsd) & 0x3FFFFFFF
	} else {
		drop := 32 - msdBitsConsumed
		msd = (msd << drop) >> drop
		minuendMsd := uint32(1 << (32 - drop))
		rMsd := int64(minuendMsd) - int64(msd) - int64(borrow)
		resultMsd = uint32(rMsd) & (minuendMsd - 1)
	}
	result.SetDigit(last, resultMsd)
	return result.Trim()
}

// AsIntN wraps x to a signed n-bit integer.
// Sourced from JSBI.asIntN (jsbi.ts lines 408-437).
func AsIntN(n int, x *BigInt) (*BigInt, error) {
	if x == nil || x.Length() == 0 {
		return Zero(), nil
	}
	if n < 0 {
		return nil, ErrRange
	}
	if n == 0 {
		return Zero(), nil
	}
	// If x has less than n bits, return an independent copy directly.
	if n >= kMaxLengthBits {
		return x.Copy(), nil
	}
	neededLength := (n + 29) / 30
	if x.Length() < neededLength {
		return x.Copy(), nil
	}
	topDigit := x.Digit(neededLength - 1)
	compareDigit := uint32(1 << ((n - 1) % 30))
	if x.Length() == neededLength && topDigit < compareDigit {
		return x.Copy(), nil
	}
	hasBit := (topDigit & compareDigit) == compareDigit
	if !hasBit {
		return truncateToNBits(n, x), nil
	}
	if !x.Sign() {
		return truncateAndSubFromPowerOfTwo(n, x, true), nil
	}
	if (topDigit & (compareDigit - 1)) == 0 {
		for i := neededLength - 2; i >= 0; i-- {
			if x.Digit(i) != 0 {
				return truncateAndSubFromPowerOfTwo(n, x, false), nil
			}
		}
		if x.Length() == neededLength && topDigit == compareDigit {
			return x.Copy(), nil
		}
		return truncateToNBits(n, x), nil
	}
	return truncateAndSubFromPowerOfTwo(n, x, false), nil
}

// AsUintN wraps x to an unsigned n-bit integer.
// Sourced from JSBI.asUintN (jsbi.ts lines 439-466).
func AsUintN(n int, x *BigInt) (*BigInt, error) {
	if x == nil || x.Length() == 0 {
		return Zero(), nil
	}
	if n < 0 {
		return nil, ErrRange
	}
	if n == 0 {
		return Zero(), nil
	}
	// If x is negative, simulate two's complement representation.
	if x.Sign() {
		if (n+29)/30 > kMaxLength {
			return nil, ErrRange
		}
		return truncateAndSubFromPowerOfTwo(n, x, false), nil
	}
	// If x is positive and has up to n bits, return an independent copy.
	if n >= kMaxLengthBits {
		return x.Copy(), nil
	}
	neededLength := (n + 29) / 30
	if x.Length() < neededLength {
		return x.Copy(), nil
	}
	bitsInTopDigit := n % 30
	if x.Length() == neededLength {
		if bitsInTopDigit == 0 {
			return x.Copy(), nil
		}
		topDigit := x.Digit(neededLength - 1)
		if (topDigit >> bitsInTopDigit) == 0 {
			return x.Copy(), nil
		}
	}
	return truncateToNBits(n, x), nil
}
