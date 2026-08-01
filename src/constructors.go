package jsbi

import (
	"math"
)

// BigIntVal constructs a BigInt from a Go interface{} value (int, float, string, bool, *BigInt).
func BigIntVal(arg interface{}) (*BigInt, error) {
	switch v := arg.(type) {
	case int:
		return FromInt64(int64(v)), nil
	case int64:
		return FromInt64(v), nil
	case int32:
		return FromInt64(int64(v)), nil
	case uint:
		return FromUint64(uint64(v)), nil
	case uint64:
		return FromUint64(v), nil
	case uint32:
		return FromUint64(uint64(v)), nil
	case float64:
		return FromFloat64(v)
	case float32:
		return FromFloat64(float64(v))
	case string:
		return FromString(v, 0)
	case bool:
		return FromBool(v), nil
	case *BigInt:
		if v == nil {
			return Zero(), nil
		}
		return v.Copy(), nil
	default:
		return nil, ErrType
	}
}

// FromInt64 creates a BigInt from a signed 64-bit integer.
func FromInt64(v int64) *BigInt {
	if v == 0 {
		return Zero()
	}
	sign := v < 0
	var u uint64
	if sign {
		u = uint64(-v)
	} else {
		u = uint64(v)
	}
	return fromUint64Internal(u, sign)
}

// FromUint64 creates a BigInt from an unsigned 64-bit integer.
func FromUint64(v uint64) *BigInt {
	if v == 0 {
		return Zero()
	}
	return fromUint64Internal(v, false)
}

func fromUint64Internal(u uint64, sign bool) *BigInt {
	if u <= kDigitMask {
		return OneDigit(uint32(u), sign)
	}
	res := &BigInt{sign: sign}
	res.SetDigit(0, uint32(u&kDigitMask))
	res.SetDigit(1, uint32((u>>30)&kDigitMask))
	if (u >> 60) > 0 {
		res.SetDigit(2, uint32(u>>60))
	}
	return res.Trim()
}

// FromBool creates a BigInt from a boolean value (true -> 1, false -> 0).
func FromBool(b bool) *BigInt {
	if b {
		return OneDigit(1, false)
	}
	return Zero()
}

// FromFloat64 converts a float64 into a BigInt, mirroring JSBI.__fromDouble.
func FromFloat64(value float64) (*BigInt, error) {
	if value == 0 {
		return Zero(), nil
	}
	if math.IsNaN(value) || math.IsInf(value, 0) || math.Floor(value) != value {
		return nil, ErrRange
	}
	sign := value < 0
	bits := math.Float64bits(value)
	rawExponent := int((bits >> 52) & 0x7FF)
	exponent := rawExponent - 0x3FF
	if exponent < 0 {
		return Zero(), nil
	}
	digitsNeeded := (exponent / 30) + 1
	result := &BigInt{
		sign:   sign,
		digits: make([]uint32, digitsNeeded),
	}

	const kHiddenBit uint32 = 0x00100000
	mantissaHigh := uint32((bits>>32)&0xFFFFF) | kHiddenBit
	mantissaLow := uint32(bits & 0xFFFFFFFF)

	const kMantissaHighTopBit = 20
	msdTopBit := exponent % 30
	remainingMantissaBits := 0
	var digit uint32

	if msdTopBit < kMantissaHighTopBit {
		shift := kMantissaHighTopBit - msdTopBit
		remainingMantissaBits = shift + 32
		digit = mantissaHigh >> shift
		mantissaHigh = (mantissaHigh << (32 - shift)) | (mantissaLow >> shift)
		mantissaLow = mantissaLow << (32 - shift)
	} else if msdTopBit == kMantissaHighTopBit {
		remainingMantissaBits = 32
		digit = mantissaHigh
		mantissaHigh = mantissaLow
		mantissaLow = 0
	} else {
		shift := msdTopBit - kMantissaHighTopBit
		remainingMantissaBits = 32 - shift
		digit = (mantissaHigh << shift) | (mantissaLow >> (32 - shift))
		mantissaHigh = mantissaLow << shift
		mantissaLow = 0
	}

	result.SetDigit(digitsNeeded-1, digit)

	for digitIndex := digitsNeeded - 2; digitIndex >= 0; digitIndex-- {
		if remainingMantissaBits > 0 {
			remainingMantissaBits -= 30
			digit = mantissaHigh >> 2
			mantissaHigh = (mantissaHigh << 30) | (mantissaLow >> 2)
			mantissaLow = mantissaLow << 30
		} else {
			digit = 0
		}
		result.SetDigit(digitIndex, digit)
	}

	return result.Trim(), nil
}
