package jsbi

// Divide computes the quotient of x and y (truncating toward zero) matching JSBI semantics.
// References: jsbi/lib/jsbi.ts lines 236-250.
func Divide(x, y *BigInt) (*BigInt, error) {
	if y.Length() == 0 {
		return nil, ErrRange
	}
	if absoluteCompare(x, y) < 0 {
		return Zero(), nil
	}
	resultSign := x.Sign() != y.Sign()
	divisor := y.UnsignedDigit(0)
	if y.Length() == 1 && divisor <= 0x7FFF {
		if divisor == 1 {
			res := x.Copy()
			res.SetSign(resultSign)
			return res, nil
		}
		quotient := absoluteDivSmall(x, divisor, nil)
		quotient.SetSign(resultSign)
		return quotient.Trim(), nil
	}
	quotient, _ := absoluteDivLarge(x, y, true, false)
	quotient.SetSign(resultSign)
	return quotient.Trim(), nil
}

// Remainder computes the remainder of x divided by y matching JSBI semantics.
// References: jsbi/lib/jsbi.ts lines 252-267.
func Remainder(x, y *BigInt) (*BigInt, error) {
	if y.Length() == 0 {
		return nil, ErrRange
	}
	if absoluteCompare(x, y) < 0 {
		return x.Copy(), nil
	}
	divisor := y.UnsignedDigit(0)
	if y.Length() == 1 && divisor <= 0x7FFF {
		if divisor == 1 {
			return Zero(), nil
		}
		remainderDigit := absoluteModSmall(x, divisor)
		if remainderDigit == 0 {
			return Zero(), nil
		}
		return OneDigit(remainderDigit, x.Sign()), nil
	}
	_, remainder := absoluteDivLarge(x, y, false, true)
	remainder.SetSign(x.Sign())
	return remainder.Trim(), nil
}

// DivRem computes both the quotient and remainder of x divided by y.
// Extension to JSBI providing single-pass Algorithm D execution for large divisors.
func DivRem(x, y *BigInt) (quotient *BigInt, remainder *BigInt, err error) {
	if y.Length() == 0 {
		return nil, nil, ErrRange
	}
	if absoluteCompare(x, y) < 0 {
		return Zero(), x.Copy(), nil
	}
	resultSign := x.Sign() != y.Sign()
	divisor := y.UnsignedDigit(0)
	if y.Length() == 1 && divisor <= 0x7FFF {
		if divisor == 1 {
			q := x.Copy()
			q.SetSign(resultSign)
			return q, Zero(), nil
		}
		q := absoluteDivSmall(x, divisor, nil)
		q.SetSign(resultSign)
		q.Trim()
		remDigit := absoluteModSmall(x, divisor)
		var r *BigInt
		if remDigit == 0 {
			r = Zero()
		} else {
			r = OneDigit(remDigit, x.Sign())
		}
		return q, r, nil
	}
	q, r := absoluteDivLarge(x, y, true, true)
	q.SetSign(resultSign)
	q.Trim()
	r.SetSign(x.Sign())
	r.Trim()
	return q, r, nil
}

// absoluteModSmall computes |x| mod divisor for divisor <= 0x7FFF.
// References: jsbi/lib/jsbi.ts lines 1525-1532.
func absoluteModSmall(x *BigInt, divisor uint32) uint32 {
	var remainder uint32 = 0
	for i := x.Length()*2 - 1; i >= 0; i-- {
		input := (remainder << 15) | x.halfDigit(i)
		remainder = input % divisor
	}
	return remainder
}

// absoluteDivSmall divides |x| by a small divisor <= 0x7FFF.
// References: jsbi/lib/jsbi.ts lines 1509-1523.
func absoluteDivSmall(x *BigInt, divisor uint32, quotient *BigInt) *BigInt {
	if quotient == nil {
		quotient = NewBigInt(x.Length(), false)
	}
	var remainder uint32 = 0
	for i := x.Length()*2 - 1; i >= 0; i-- {
		input := (remainder << 15) | x.halfDigit(i)
		q := input / divisor
		remainder = input % divisor
		quotient.setHalfDigit(i, q)
	}
	return quotient
}

// absoluteDivLarge implements Knuth Algorithm D for division.
// References: jsbi/lib/jsbi.ts lines 1540-1605.
func absoluteDivLarge(dividend, divisor *BigInt, wantQuotient, wantRemainder bool) (quotient, remainder *BigInt) {
	n := divisor.halfDigitLength()
	n2 := divisor.Length()
	m := dividend.halfDigitLength() - n
	var q *BigInt
	if wantQuotient {
		q = NewBigInt((m+2)>>1, false)
	}
	qhatv := NewBigInt((n+2)>>1, false)

	// D1. Normalization
	shift := clz15(divisor.halfDigit(n - 1))
	normalizedDivisor := divisor
	if shift > 0 {
		normalizedDivisor = specialLeftShift(divisor, shift, 0)
	}
	u := specialLeftShift(dividend, shift, 1)

	// D2. Loop initialization
	vn1 := normalizedDivisor.halfDigit(n - 1)
	var halfDigitBuffer uint32 = 0
	for j := m; j >= 0; j-- {
		// D3. Trial quotient
		var qhat uint32 = 0x7FFF
		ujn := u.halfDigit(j + n)
		if ujn != vn1 {
			input := (ujn << 15) | u.halfDigit(j+n-1)
			qhat = input / vn1
			rhat := input % vn1
			vn2 := normalizedDivisor.halfDigit(n - 2)
			ujn2 := u.halfDigit(j + n - 2)
			for (qhat * vn2) > ((rhat << 16) | ujn2) {
				qhat--
				rhat += vn1
				if rhat > 0x7FFF {
					break
				}
			}
		}

		// D4. Multiply & subtract
		internalMultiplyAdd(normalizedDivisor, qhat, 0, n2, qhatv)
		c := u.inplaceSub(qhatv, j, n+1)
		if c != 0 {
			c = u.inplaceAdd(normalizedDivisor, j, n)
			u.setHalfDigit(j+n, (u.halfDigit(j+n)+c)&0x7FFF)
			qhat--
		}

		if wantQuotient {
			if (j & 1) != 0 {
				halfDigitBuffer = qhat << 15
			} else {
				q.SetDigit(j>>1, halfDigitBuffer|qhat)
			}
		}
	}

	if wantRemainder {
		u.inplaceRightShift(shift)
		remainder = u
	}
	if wantQuotient {
		quotient = q
	}
	return quotient, remainder
}

// halfDigitLength returns the number of 15-bit half-digits in x.
// References: jsbi/lib/jsbi.ts lines 1922-1926.
func (x *BigInt) halfDigitLength() int {
	len := x.Length()
	if len == 0 {
		return 0
	}
	if x.Digit(len-1) <= 0x7FFF {
		return len*2 - 1
	}
	return len * 2
}

// halfDigit returns the 15-bit half-digit at index i.
// References: jsbi/lib/jsbi.ts lines 1927-1929.
func (x *BigInt) halfDigit(i int) uint32 {
	if i < 0 || i >= x.Length()*2 {
		return 0
	}
	return (x.Digit(i>>1) >> (uint(i&1) * 15)) & 0x7FFF
}

// setHalfDigit sets the 15-bit half-digit at index i.
// References: jsbi/lib/jsbi.ts lines 1930-1936.
func (x *BigInt) setHalfDigit(i int, value uint32) {
	digitIndex := i >> 1
	previous := x.Digit(digitIndex)
	var updated uint32
	if (i & 1) != 0 {
		updated = (previous & 0x7FFF) | ((value & 0x7FFF) << 15)
	} else {
		updated = (previous & 0x3FFF8000) | (value & 0x7FFF)
	}
	x.SetDigit(digitIndex, updated)
}

// inplaceAdd adds summand to b starting at half-digit offset startIndex over halfDigits.
// References: jsbi/lib/jsbi.ts lines 1612-1622.
func (b *BigInt) inplaceAdd(summand *BigInt, startIndex int, halfDigits int) uint32 {
	var carry uint32 = 0
	for i := 0; i < halfDigits; i++ {
		sum := b.halfDigit(startIndex+i) + summand.halfDigit(i) + carry
		carry = sum >> 15
		b.setHalfDigit(startIndex+i, sum&0x7FFF)
	}
	return carry
}

// inplaceSub subtracts subtrahend from b starting at half-digit offset startIndex over halfDigits.
// References: jsbi/lib/jsbi.ts lines 1624-1684.
func (b *BigInt) inplaceSub(subtrahend *BigInt, startIndex int, halfDigits int) uint32 {
	fullSteps := (halfDigits - 1) >> 1
	subLen := (halfDigits + 1) >> 1
	var borrow uint32 = 0
	if (startIndex & 1) != 0 {
		startIndex >>= 1
		current := b.Digit(startIndex)
		r0 := current & 0x7FFF
		i := 0
		for ; i < fullSteps; i++ {
			sub := subtrahend.Digit(i)
			r15 := (current >> 15) - (sub & 0x7FFF) - borrow
			borrow = (r15 >> 15) & 1
			b.SetDigit(startIndex+i, ((r15 & 0x7FFF) << 15) | (r0 & 0x7FFF))
			current = b.Digit(startIndex + i + 1)
			r0 = (current & 0x7FFF) - (sub >> 15) - borrow
			borrow = (r0 >> 15) & 1
		}
		sub := subtrahend.Digit(i)
		r15 := (current >> 15) - (sub & 0x7FFF) - borrow
		borrow = (r15 >> 15) & 1
		b.SetDigit(startIndex+i, ((r15 & 0x7FFF) << 15) | (r0 & 0x7FFF))
		subTop := sub >> 15
		if (halfDigits & 1) == 0 {
			current = b.Digit(startIndex + i + 1)
			r0 = (current & 0x7FFF) - subTop - borrow
			borrow = (r0 >> 15) & 1
			b.SetDigit(startIndex+subLen, (current & 0x3FFF8000) | (r0 & 0x7FFF))
		}
	} else {
		startIndex >>= 1
		i := 0
		for ; i < subLen-1; i++ {
			current := b.Digit(startIndex + i)
			sub := subtrahend.Digit(i)
			r0 := (current & 0x7FFF) - (sub & 0x7FFF) - borrow
			borrow = (r0 >> 15) & 1
			r15 := (current >> 15) - (sub >> 15) - borrow
			borrow = (r15 >> 15) & 1
			b.SetDigit(startIndex+i, ((r15 & 0x7FFF) << 15) | (r0 & 0x7FFF))
		}
		current := b.Digit(startIndex + i)
		sub := subtrahend.Digit(i)
		r0 := (current & 0x7FFF) - (sub & 0x7FFF) - borrow
		borrow = (r0 >> 15) & 1
		var r15 uint32 = 0
		if (halfDigits & 1) == 0 {
			r15 = (current >> 15) - (sub >> 15) - borrow
			borrow = (r15 >> 15) & 1
		}
		b.SetDigit(startIndex+i, ((r15 & 0x7FFF) << 15) | (r0 & 0x7FFF))
	}
	return borrow
}

// inplaceRightShift shifts b right in-place by shift bits within 30-bit limbs.
// References: jsbi/lib/jsbi.ts lines 1686-1696.
func (b *BigInt) inplaceRightShift(shift int) {
	if shift == 0 {
		return
	}
	carry := b.Digit(0) >> uint(shift)
	last := b.Length() - 1
	for i := 0; i < last; i++ {
		d := b.Digit(i + 1)
		b.SetDigit(i, ((d<<uint(30-shift))&kDigitMask)|carry)
		carry = d >> uint(shift)
	}
	b.SetDigit(last, carry)
}

// specialLeftShift shifts x left by shift bits within 30-bit limbs, returning a new BigInt.
// References: jsbi/lib/jsbi.ts lines 1698-1717.
func specialLeftShift(x *BigInt, shift int, addDigit int) *BigInt {
	n := x.Length()
	resultLength := n + addDigit
	result := NewBigInt(resultLength, false)
	if shift == 0 {
		for i := 0; i < n; i++ {
			result.SetDigit(i, x.Digit(i))
		}
		if addDigit > 0 {
			result.SetDigit(n, 0)
		}
		return result
	}
	var carry uint32 = 0
	for i := 0; i < n; i++ {
		d := x.Digit(i)
		result.SetDigit(i, ((d<<uint(shift))&kDigitMask)|carry)
		carry = d >> uint(30-shift)
	}
	if addDigit > 0 {
		result.SetDigit(n, carry)
	}
	return result
}

// clz15 returns the number of leading zeros in a 15-bit half-digit.
// References: jsbi/lib/jsbi.ts line 1608.
func clz15(value uint32) int {
	return clz30(value) - 15
}
