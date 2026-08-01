package jsbi

// Multiply calculates the product of two BigInt instances matching JSBI semantics.
// References: jsbi/lib/jsbi.ts lines 221-234.
func Multiply(x, y *BigInt) *BigInt {
	if x.Length() == 0 {
		return x
	}
	if y.Length() == 0 {
		return y
	}
	resultLength := x.Length() + y.Length()
	if x.clzmsd()+y.clzmsd() >= kDigitBits {
		resultLength--
	}
	result := NewBigInt(resultLength, x.Sign() != y.Sign())
	for i := 0; i < x.Length(); i++ {
		multiplyAccumulate(y, x.Digit(i), result, i)
	}
	return result.Trim()
}

// multiplyAccumulate adds (multiplicand * multiplier) into accumulator at accumulatorIndex.
// References: jsbi/lib/jsbi.ts lines 1425-1456.
func multiplyAccumulate(multiplicand *BigInt, multiplier uint32, accumulator *BigInt, accumulatorIndex int) {
	if multiplier == 0 {
		return
	}
	m2Low := multiplier & 0x7FFF
	m2High := multiplier >> 15
	var carry uint32 = 0
	var high uint32 = 0
	for i := 0; i < multiplicand.Length(); i++ {
		acc := accumulator.Digit(accumulatorIndex)
		m1 := multiplicand.Digit(i)
		m1Low := m1 & 0x7FFF
		m1High := m1 >> 15
		rLow := m1Low * m2Low
		rMid1 := m1Low * m2High
		rMid2 := m1High * m2Low
		rHigh := m1High * m2High
		acc += high + rLow + carry
		carry = acc >> kDigitBits
		acc &= kDigitMask
		acc += ((rMid1 & 0x7FFF) << 15) + ((rMid2 & 0x7FFF) << 15)
		carry += acc >> kDigitBits
		high = rHigh + (rMid1 >> 15) + (rMid2 >> 15)
		accumulator.SetDigit(accumulatorIndex, acc&kDigitMask)
		accumulatorIndex++
	}
	for carry != 0 || high != 0 {
		acc := accumulator.Digit(accumulatorIndex)
		acc += carry + high
		high = 0
		carry = acc >> kDigitBits
		accumulator.SetDigit(accumulatorIndex, acc&kDigitMask)
		accumulatorIndex++
	}
}

// internalMultiplyAdd multiplies source by factor, adds summand, and stores the result.
// References: jsbi/lib/jsbi.ts lines 1458-1479.
func internalMultiplyAdd(source *BigInt, factor uint32, summand uint32, n int, result *BigInt) {
	carry := summand
	factorLow := factor & 0x7FFF
	factorHigh := factor >> 15
	for i := 0; i < n; i++ {
		d := source.Digit(i)
		dLow := d & 0x7FFF
		dHigh := d >> 15
		rLow := dLow * factorLow
		rMid1 := dLow * factorHigh
		rMid2 := dHigh * factorLow
		rHigh := dHigh * factorHigh
		acc := rLow + carry
		carry = acc >> kDigitBits
		acc &= kDigitMask
		acc += ((rMid1 & 0x7FFF) << 15) + ((rMid2 & 0x7FFF) << 15)
		carry += acc >> kDigitBits
		carry += rHigh + (rMid1 >> 15) + (rMid2 >> 15)
		result.SetDigit(i, acc&kDigitMask)
	}
	result.SetDigit(n, carry)
}

// inplaceMultiplyAdd multiplies b by multiplier and adds summand in-place.
// References: jsbi/lib/jsbi.ts lines 1481-1506.
func (b *BigInt) inplaceMultiplyAdd(multiplier uint32, summand uint32, length int) {
	if length == 0 {
		b.SetDigit(0, summand)
		return
	}
	if multiplier == 1 && summand == 0 {
		return
	}
	carry := summand
	mLow := multiplier & 0x7FFF
	mHigh := multiplier >> 15
	for i := 0; i < length; i++ {
		d := b.Digit(i)
		dLow := d & 0x7FFF
		dHigh := d >> 15
		rLow := dLow * mLow
		rMid1 := dLow * mHigh
		rMid2 := dHigh * mLow
		rHigh := dHigh * mHigh
		acc := rLow + carry
		carry = acc >> kDigitBits
		acc &= kDigitMask
		acc += ((rMid1 & 0x7FFF) << 15) + ((rMid2 & 0x7FFF) << 15)
		carry += acc >> kDigitBits
		carry += rHigh + (rMid1 >> 15) + (rMid2 >> 15)
		b.SetDigit(i, acc&kDigitMask)
	}
	b.SetDigit(length, carry)
}
