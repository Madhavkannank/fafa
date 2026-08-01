package jsbi

import (
	"math/bits"
	"strconv"
)

// kConversionChars is the output character alphabet for radix conversion.
// Sourced from JSBI.__kConversionChars (jsbi.ts line 1967).
const kConversionChars = "0123456789abcdefghijklmnopqrstuvwxyz"

// Exponentiate computes x^y. Sourced from JSBI.exponentiate (jsbi.ts lines 167-219).
func Exponentiate(x, y *BigInt) (*BigInt, error) {
	if y.Sign() {
		return nil, ErrRange
	}
	if y.Length() == 0 {
		one := NewBigInt(1, false)
		one.SetDigit(0, 1)
		return one, nil
	}
	if x.Length() == 0 {
		return x.Copy(), nil
	}
	// x == 1 or x == -1 special cases
	if x.Length() == 1 && x.Digit(0) == 1 {
		if x.Sign() && (y.Digit(0)&1) == 0 {
			return UnaryMinus(x), nil
		}
		return x.Copy(), nil
	}
	// Guard: y must fit in one limb
	if y.Length() > 1 {
		return nil, ErrRange
	}
	expValue := y.Digit(0)
	if expValue == 1 {
		return x.Copy(), nil
	}
	if int(expValue) >= kMaxLengthBits {
		return nil, ErrRange
	}
	// Fast path: 2^n
	if x.Length() == 1 && x.Digit(0) == 2 {
		neededDigits := 1 + int(expValue)/30
		sign := x.Sign() && (expValue&1) != 0
		result := NewBigInt(neededDigits, sign)
		msd := uint32(1 << (expValue % 30))
		result.SetDigit(neededDigits-1, msd)
		return result, nil
	}
	// General: binary (square-and-multiply) exponentiation
	var result *BigInt
	runningSquare := x.Copy()
	if (expValue & 1) != 0 {
		result = x.Copy()
	}
	expValue >>= 1
	for ; expValue != 0; expValue >>= 1 {
		runningSquare = Multiply(runningSquare, runningSquare)
		if (expValue & 1) != 0 {
			if result == nil {
				result = runningSquare.Copy()
			} else {
				result = Multiply(result, runningSquare)
			}
		}
	}
	return result, nil
}

// ToString converts x to its string representation in the given radix.
// Sourced from JSBI.toString (jsbi.ts lines 67-77).
func ToString(x *BigInt, radix int) (string, error) {
	if radix < 2 || radix > 36 {
		return "", ErrRange
	}
	if x.Length() == 0 {
		return "0", nil
	}
	if (radix & (radix - 1)) == 0 {
		return toStringBasePowerOfTwo(x, radix)
	}
	return toStringGeneric(x, radix, false)
}

// toStringBasePowerOfTwo converts x to string for power-of-two radices.
// Sourced from JSBI.__toStringBasePowerOfTwo (jsbi.ts lines 916-958).
func toStringBasePowerOfTwo(x *BigInt, radix int) (string, error) {
	length := x.Length()
	bitsPerChar := bits.OnesCount32(uint32(radix - 1)) // popcount(radix-1)
	charMask := uint32(radix - 1)
	msd := x.Digit(length - 1)
	msdLeadingZeros := clz30(msd)
	bitLength := length*30 - msdLeadingZeros
	charsRequired := (bitLength + bitsPerChar - 1) / bitsPerChar
	if x.Sign() {
		charsRequired++
	}
	if charsRequired > (1 << 28) {
		return "", ErrRange
	}
	result := make([]byte, charsRequired)
	pos := charsRequired - 1
	var digit uint32 = 0
	availableBits := 0

	for i := 0; i < length-1; i++ {
		newDigit := x.Digit(i)
		current := (digit | (newDigit << availableBits)) & charMask
		result[pos] = kConversionChars[current]
		pos--
		consumedBits := bitsPerChar - availableBits
		digit = newDigit >> consumedBits
		availableBits = 30 - consumedBits
		for availableBits >= bitsPerChar {
			result[pos] = kConversionChars[digit&charMask]
			pos--
			digit >>= bitsPerChar
			availableBits -= bitsPerChar
		}
	}
	// Handle MSD
	current := (digit | (msd << availableBits)) & charMask
	result[pos] = kConversionChars[current]
	pos--
	digit = msd >> (bitsPerChar - availableBits)
	for digit != 0 {
		result[pos] = kConversionChars[digit&charMask]
		pos--
		digit >>= bitsPerChar
	}
	if x.Sign() {
		result[pos] = '-'
		pos--
	}
	if pos != -1 {
		return "", ErrRange // internal consistency check (mirrors JSBI's assert)
	}
	return string(result), nil
}

// toStringGeneric converts x to string for non-power-of-two radices.
// Sourced from JSBI.__toStringGeneric (jsbi.ts lines 960-1010).
func toStringGeneric(x *BigInt, radix int, isRecursiveCall bool) (string, error) {
	length := x.Length()
	if length == 0 {
		return "", nil
	}
	if length == 1 {
		result := strconv.FormatUint(uint64(x.Digit(0)), radix)
		if !isRecursiveCall && x.Sign() {
			result = "-" + result
		}
		return result, nil
	}

	bitLength := length*30 - clz30(x.Digit(length-1))
	// kMaxBitsPerChar is uint32; cast to int for arithmetic
	maxBitsPerChar := int(kMaxBitsPerChar[radix])
	minBitsPerChar := maxBitsPerChar - 1
	charsRequired := bitLength * kBitsPerCharTableMultiplier
	charsRequired += minBitsPerChar - 1
	charsRequired = charsRequired / minBitsPerChar
	secondHalfChars := (charsRequired + 1) >> 1

	// Compute conqueror = radix^secondHalfChars
	radixBig := NewBigInt(1, false)
	radixBig.SetDigit(0, uint32(radix))
	expBig := NewBigInt(1, false)
	expBig.SetDigit(0, uint32(secondHalfChars))
	conqueror, err := Exponentiate(radixBig, expBig)
	if err != nil {
		return "", err
	}

	var quotient *BigInt
	var secondHalf string

	divisor := conqueror.Digit(0)
	if conqueror.Length() == 1 && divisor <= 0x7FFF {
		// Fast half-digit division path
		quotient = NewBigInt(x.Length(), false)
		var remainder uint32 = 0
		for i := x.Length()*2 - 1; i >= 0; i-- {
			input := (remainder << 15) | x.halfDigit(i)
			quotient.setHalfDigit(i, input/divisor)
			remainder = input % divisor
		}
		secondHalf = strconv.FormatUint(uint64(remainder), radix)
	} else {
		// Full absoluteDivLarge path
		divResult, remResult := absoluteDivLarge(x, conqueror, true, true)
		quotient = divResult
		rem := remResult.Trim()
		secondHalf, err = toStringGeneric(rem, radix, true)
		if err != nil {
			return "", err
		}
	}
	quotient = quotient.Trim()
	firstHalf, err := toStringGeneric(quotient, radix, true)
	if err != nil {
		return "", err
	}
	// Pad secondHalf with leading zeros to exactly secondHalfChars
	for len(secondHalf) < secondHalfChars {
		secondHalf = "0" + secondHalf
	}
	if !isRecursiveCall && x.Sign() {
		firstHalf = "-" + firstHalf
	}
	return firstHalf + secondHalf, nil
}
