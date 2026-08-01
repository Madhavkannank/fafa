package jsbi

var kMaxBitsPerChar = []uint32{
	0, 0, 32, 51, 64, 75, 83, 90, 96, // 0..8
	102, 107, 111, 115, 119, 122, 126, 128, // 9..16
	131, 134, 136, 139, 141, 143, 145, 147, // 17..24
	149, 151, 153, 154, 156, 158, 159, 160, // 25..32
	162, 163, 165, 166, // 33..36
}

const (
	kBitsPerCharTableShift      = 5
	kBitsPerCharTableMultiplier = 1 << kBitsPerCharTableShift
)

// isWhitespace checks ECMAScript compliant whitespace runes.
func isWhitespace(c rune) bool {
	if c >= 0x09 && c <= 0x0D {
		return true
	}
	if c == 0x20 || c == 0xA0 || c == 0x1680 || c == 0xFEFF {
		return true
	}
	if c >= 0x2000 && c <= 0x200A {
		return true
	}
	if c == 0x2028 || c == 0x2029 || c == 0x202F || c == 0x205F || c == 0x3000 {
		return true
	}
	return false
}

// FromString parses a string into a BigInt with an optional radix (2 to 36, or 0 for auto-detect).
func FromString(s string, radix int) (*BigInt, error) {
	if radix != 0 && (radix < 2 || radix > 36) {
		return nil, ErrRange
	}
	runes := []rune(s)
	length := len(runes)
	cursor := 0

	if cursor == length {
		return Zero(), nil
	}

	// Skip leading whitespace.
	for cursor < length && isWhitespace(runes[cursor]) {
		cursor++
	}
	if cursor == length {
		return Zero(), nil
	}

	sign := 0
	current := runes[cursor]

	// Parse sign prefix.
	if current == '+' {
		cursor++
		if cursor == length {
			return nil, ErrSyntax
		}
		current = runes[cursor]
		sign = 1
	} else if current == '-' {
		cursor++
		if cursor == length {
			return nil, ErrSyntax
		}
		current = runes[cursor]
		sign = -1
	}

	// Detect radix if 0.
	if radix == 0 {
		radix = 10
		if current == '0' {
			cursor++
			if cursor == length {
				return Zero(), nil
			}
			current = runes[cursor]
			if current == 'X' || current == 'x' {
				radix = 16
				cursor++
				if cursor == length {
					return nil, ErrSyntax
				}
				current = runes[cursor]
			} else if current == 'O' || current == 'o' {
				radix = 8
				cursor++
				if cursor == length {
					return nil, ErrSyntax
				}
				current = runes[cursor]
			} else if current == 'B' || current == 'b' {
				radix = 2
				cursor++
				if cursor == length {
					return nil, ErrSyntax
				}
				current = runes[cursor]
			}
		}
	} else if radix == 16 {
		if current == '0' {
			cursor++
			if cursor == length {
				return Zero(), nil
			}
			current = runes[cursor]
			if current == 'X' || current == 'x' {
				cursor++
				if cursor == length {
					return nil, ErrSyntax
				}
				current = runes[cursor]
			}
		}
	}

	if sign != 0 && radix != 10 {
		return nil, ErrSyntax
	}

	// Skip leading zeros.
	for current == '0' {
		cursor++
		if cursor == length {
			return Zero(), nil
		}
		current = runes[cursor]
	}

	chars := length - cursor
	bitsPerChar := kMaxBitsPerChar[radix]
	roundup := uint32(kBitsPerCharTableMultiplier - 1)
	if uint32(chars) > ((1 << 30) / bitsPerChar) {
		return nil, ErrSyntax
	}
	bitsMin := (bitsPerChar*uint32(chars) + roundup) >> kBitsPerCharTableShift
	resultLength := int((bitsMin + 29) / 30)
	if resultLength < 1 {
		resultLength = 1
	}
	result := &BigInt{sign: false, digits: make([]uint32, resultLength)}

	limDigit := uint32(radix)
	if limDigit > 10 {
		limDigit = 10
	}
	var limAlpha uint32
	if radix > 10 {
		limAlpha = uint32(radix - 10)
	}

	if (radix & (radix - 1)) == 0 {
		// Power of two radix fast path.
		bitsPerChar >>= kBitsPerCharTableShift
		var parts []uint32
		var partsBits []uint32
		done := false
		for !done {
			var part uint32
			var bits uint32
			for {
				var d uint32
				c := uint32(current)
				if c-48 < limDigit {
					d = c - 48
				} else if (c|32)-97 < limAlpha {
					d = (c | 32) - 87
				} else {
					done = true
					break
				}
				bits += bitsPerChar
				part = (part << bitsPerChar) | d
				cursor++
				if cursor == length {
					done = true
					break
				}
				current = runes[cursor]
				if bits+bitsPerChar > 30 {
					break
				}
			}
			parts = append(parts, part)
			partsBits = append(partsBits, bits)
		}
		fillFromParts(result, parts, partsBits)
	} else {
		// Generic radix path.
		done := false
		charsSoFar := uint32(0)
		for !done {
			var part uint32
			multiplier := uint32(1)
			for {
				var d uint32
				c := uint32(current)
				if c-48 < limDigit {
					d = c - 48
				} else if (c|32)-97 < limAlpha {
					d = (c | 32) - 87
				} else {
					done = true
					break
				}
				m := multiplier * uint32(radix)
				if m > 0x3FFFFFFF {
					break
				}
				multiplier = m
				part = part*uint32(radix) + d
				charsSoFar++
				cursor++
				if cursor == length {
					done = true
					break
				}
				current = runes[cursor]
			}
			roundup = kBitsPerCharTableMultiplier*30 - 1
			digitsSoFar := int(((bitsPerChar*charsSoFar + roundup) >> kBitsPerCharTableShift) / 30)
			inplaceMultiplyAdd(result, multiplier, part, digitsSoFar)
		}
	}

	if cursor < length {
		if !isWhitespace(current) {
			return nil, ErrSyntax
		}
		for cursor++; cursor < length; cursor++ {
			if !isWhitespace(runes[cursor]) {
				return nil, ErrSyntax
			}
		}
	}

	result.sign = (sign == -1)
	return result.Trim(), nil
}

func fillFromParts(result *BigInt, parts []uint32, partsBits []uint32) {
	digitIndex := 0
	var digit uint32
	var bitsInDigit uint32
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		partBits := partsBits[i]
		digit |= (part << bitsInDigit)
		bitsInDigit += partBits
		if bitsInDigit == 30 {
			result.SetDigit(digitIndex, digit)
			digitIndex++
			bitsInDigit = 0
			digit = 0
		} else if bitsInDigit > 30 {
			result.SetDigit(digitIndex, digit&kDigitMask)
			digitIndex++
			bitsInDigit -= 30
			digit = part >> (partBits - bitsInDigit)
		}
	}
	if digit != 0 {
		result.SetDigit(digitIndex, digit)
		digitIndex++
	}
	for digitIndex < len(result.digits) {
		result.SetDigit(digitIndex, 0)
		digitIndex++
	}
}

func inplaceMultiplyAdd(result *BigInt, multiplier uint32, addend uint32, digitsSoFar int) {
	if multiplier == 1 && addend == 0 {
		return
	}
	var carry uint64 = uint64(addend)
	for i := 0; i < digitsSoFar || carry != 0; i++ {
		var product uint64 = uint64(result.Digit(i))*uint64(multiplier) + carry
		result.SetDigit(i, uint32(product&uint64(kDigitMask)))
		carry = product >> 30
	}
}
