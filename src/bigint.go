package jsbi

import (
	"math/bits"
)

const (
	kDigitBits     = 30
	kDigitMask     = 0x3FFFFFFF
	kMaxLength     = 1 << 25
	kMaxLengthBits = kMaxLength << 5
)

// BigInt represents an arbitrary-precision integer mirroring JSBI's internal representation.
type BigInt struct {
	sign   bool     // true if negative, false if positive or zero
	digits []uint32 // 30-bit digits (least-significant digit at index 0)
}

// Zero returns a canonical zero BigInt (sign = false, digits = nil).
func Zero() *BigInt {
	return &BigInt{sign: false, digits: nil}
}

// OneDigit returns a single-digit BigInt with the specified sign.
func OneDigit(value uint32, sign bool) *BigInt {
	value &= kDigitMask
	if value == 0 {
		return Zero()
	}
	return &BigInt{sign: sign, digits: []uint32{value}}
}

// Sign returns true if negative, false if positive or zero.
func (x *BigInt) Sign() bool {
	if x == nil {
		return false
	}
	return x.sign
}

// Length returns the number of 30-bit digit limbs.
func (x *BigInt) Length() int {
	if x == nil {
		return 0
	}
	return len(x.digits)
}

// Digit returns the 30-bit digit at 0-indexed position i.
func (x *BigInt) Digit(i int) uint32 {
	if i < 0 || i >= len(x.digits) {
		return 0
	}
	return x.digits[i] & kDigitMask
}

// UnsignedDigit returns the 30-bit digit at 0-indexed position i.
func (x *BigInt) UnsignedDigit(i int) uint32 {
	return x.Digit(i)
}

// SetDigit sets the 30-bit digit at 0-indexed position i.
func (x *BigInt) SetDigit(i int, digit uint32) {
	if i < 0 {
		return
	}
	for len(x.digits) <= i {
		x.digits = append(x.digits, 0)
	}
	x.digits[i] = digit & kDigitMask
}

// Trim removes most-significant zero digits and canonicalizes zero.
func (x *BigInt) Trim() *BigInt {
	if x == nil {
		return Zero()
	}
	newLen := len(x.digits)
	for newLen > 0 && (x.digits[newLen-1]&kDigitMask) == 0 {
		newLen--
	}
	if newLen == 0 {
		x.digits = nil
		x.sign = false
	} else if newLen < len(x.digits) {
		x.digits = x.digits[:newLen]
	}
	return x
}

// Copy creates a deep copy of the BigInt instance.
func (x *BigInt) Copy() *BigInt {
	if x == nil {
		return Zero()
	}
	d := make([]uint32, len(x.digits))
	copy(d, x.digits)
	return &BigInt{sign: x.sign, digits: d}
}

// clz30 returns leading zero count in 30-bit representation.
func clz30(x uint32) int {
	x &= kDigitMask
	if x == 0 {
		return 30
	}
	return bits.LeadingZeros32(x) - 2
}

// isOneDigitInt returns true if abs(x) fits within a single 30-bit limb.
func isOneDigitInt(x int64) bool {
	if x < 0 {
		x = -x
	}
	return (uint64(x) & uint64(kDigitMask)) == uint64(x)
}
