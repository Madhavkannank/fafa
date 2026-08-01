package jsbi

// absoluteAnd computes magnitude bitwise AND (x & y).
// Sourced from JSBI.__absoluteAnd (jsbi.ts lines 1297-1324).
func absoluteAnd(x, y, result *BigInt) *BigInt {
	xLength := x.Length()
	yLength := y.Length()
	numPairs := yLength
	if xLength < yLength {
		numPairs = xLength
		x, y = y, x
		xLength, yLength = yLength, xLength
	}
	resultLength := numPairs
	if result == nil {
		result = NewBigInt(resultLength, false)
	} else {
		resultLength = result.Length()
	}
	i := 0
	for ; i < numPairs; i++ {
		result.SetDigit(i, x.Digit(i)&y.Digit(i))
	}
	for ; i < resultLength; i++ {
		result.SetDigit(i, 0)
	}
	return result
}

// absoluteAndNot computes magnitude bitwise AND NOT (x & ~y).
// Sourced from JSBI.__absoluteAndNot (jsbi.ts lines 1326-1350).
func absoluteAndNot(x, y, result *BigInt) *BigInt {
	xLength := x.Length()
	yLength := y.Length()
	numPairs := yLength
	if xLength < yLength {
		numPairs = xLength
	}
	resultLength := xLength
	if result == nil {
		result = NewBigInt(resultLength, false)
	} else {
		resultLength = result.Length()
	}
	i := 0
	for ; i < numPairs; i++ {
		result.SetDigit(i, x.Digit(i)&^y.Digit(i))
	}
	for ; i < xLength; i++ {
		result.SetDigit(i, x.Digit(i))
	}
	for ; i < resultLength; i++ {
		result.SetDigit(i, 0)
	}
	return result
}

// absoluteOr computes magnitude bitwise OR (x | y).
// Sourced from JSBI.__absoluteOr (jsbi.ts lines 1352-1382).
func absoluteOr(x, y, result *BigInt) *BigInt {
	xLength := x.Length()
	yLength := y.Length()
	numPairs := yLength
	if xLength < yLength {
		numPairs = xLength
		x, y = y, x
		xLength, yLength = yLength, xLength
	}
	resultLength := xLength
	if result == nil {
		result = NewBigInt(resultLength, false)
	} else {
		resultLength = result.Length()
	}
	i := 0
	for ; i < numPairs; i++ {
		result.SetDigit(i, x.Digit(i)|y.Digit(i))
	}
	for ; i < xLength; i++ {
		result.SetDigit(i, x.Digit(i))
	}
	for ; i < resultLength; i++ {
		result.SetDigit(i, 0)
	}
	return result
}

// absoluteXor computes magnitude bitwise XOR (x ^ y).
// Sourced from JSBI.__absoluteXor (jsbi.ts lines 1384-1410).
func absoluteXor(x, y, result *BigInt) *BigInt {
	xLength := x.Length()
	yLength := y.Length()
	numPairs := yLength
	if xLength < yLength {
		numPairs = xLength
		x, y = y, x
		xLength, yLength = yLength, xLength
	}
	resultLength := xLength
	if result == nil {
		result = NewBigInt(resultLength, false)
	} else {
		resultLength = result.Length()
	}
	i := 0
	for ; i < numPairs; i++ {
		result.SetDigit(i, x.Digit(i)^y.Digit(i))
	}
	for ; i < xLength; i++ {
		result.SetDigit(i, x.Digit(i))
	}
	for ; i < resultLength; i++ {
		result.SetDigit(i, 0)
	}
	return result
}

// BitwiseNot returns the bitwise NOT of x (~x).
// Sourced from JSBI.bitwiseNot (jsbi.ts lines 158-165).
func BitwiseNot(x *BigInt) *BigInt {
	if x == nil {
		x = Zero()
	}
	if x.Sign() {
		// ~(-x) == ~(~(x-1)) == x-1
		return absoluteSubOne(x).Trim()
	}
	// ~x == -x-1 == -(x+1)
	return absoluteAddOne(x, true)
}

// BitwiseAnd returns the bitwise AND of x and y (x & y).
// Sourced from JSBI.bitwiseAnd (jsbi.ts lines 345-363).
func BitwiseAnd(x, y *BigInt) *BigInt {
	if x == nil {
		x = Zero()
	}
	if y == nil {
		y = Zero()
	}
	if !x.Sign() && !y.Sign() {
		return absoluteAnd(x, y, nil).Trim()
	} else if x.Sign() && y.Sign() {
		resultLength := maxInt(x.Length(), y.Length()) + 1
		// (-x) & (-y) == ~(x-1) & ~(y-1) == ~((x-1) | (y-1))
		// == -(((x-1) | (y-1)) + 1)
		result := absoluteSubOne(x)
		if result.Length() < resultLength {
			resExpanded := NewBigInt(resultLength, false)
			for i := 0; i < result.Length(); i++ {
				resExpanded.SetDigit(i, result.Digit(i))
			}
			result = resExpanded
		}
		y1 := absoluteSubOne(y)
		result = absoluteOr(result, y1, result)
		return absoluteAddOne(result, true).Trim()
	}
	// Assume x is positive.
	if x.Sign() {
		x, y = y, x
	}
	// x & (-y) == x & ~(y-1) == x &~ (y-1)
	return absoluteAndNot(x, absoluteSubOne(y), nil).Trim()
}

// BitwiseOr returns the bitwise OR of x and y (x | y).
// Sourced from JSBI.bitwiseOr (jsbi.ts lines 386-406).
func BitwiseOr(x, y *BigInt) *BigInt {
	if x == nil {
		x = Zero()
	}
	if y == nil {
		y = Zero()
	}
	resultLength := maxInt(x.Length(), y.Length())
	if !x.Sign() && !y.Sign() {
		return absoluteOr(x, y, nil).Trim()
	} else if x.Sign() && y.Sign() {
		// (-x) | (-y) == ~(x-1) | ~(y-1) == ~((x-1) & (y-1))
		// == -(((x-1) & (y-1)) + 1)
		result := absoluteSubOne(x)
		if result.Length() < resultLength {
			resExpanded := NewBigInt(resultLength, false)
			for i := 0; i < result.Length(); i++ {
				resExpanded.SetDigit(i, result.Digit(i))
			}
			result = resExpanded
		}
		y1 := absoluteSubOne(y)
		result = absoluteAnd(result, y1, result)
		return absoluteAddOne(result, true).Trim()
	}
	// Assume x is positive.
	if x.Sign() {
		x, y = y, x
	}
	// x | (-y) == x | ~(y-1) == ~((y-1) &~ x) == -(((y-1) ~& x) + 1)
	result := absoluteSubOne(y)
	if result.Length() < resultLength {
		resExpanded := NewBigInt(resultLength, false)
		for i := 0; i < result.Length(); i++ {
			resExpanded.SetDigit(i, result.Digit(i))
		}
		result = resExpanded
	}
	result = absoluteAndNot(result, x, result)
	return absoluteAddOne(result, true).Trim()
}

// BitwiseXor returns the bitwise XOR of x and y (x ^ y).
// Sourced from JSBI.bitwiseXor (jsbi.ts lines 365-384).
func BitwiseXor(x, y *BigInt) *BigInt {
	if x == nil {
		x = Zero()
	}
	if y == nil {
		y = Zero()
	}
	if !x.Sign() && !y.Sign() {
		return absoluteXor(x, y, nil).Trim()
	} else if x.Sign() && y.Sign() {
		// (-x) ^ (-y) == ~(x-1) ^ ~(y-1) == (x-1) ^ (y-1)
		resultLength := maxInt(x.Length(), y.Length())
		result := absoluteSubOne(x)
		if result.Length() < resultLength {
			resExpanded := NewBigInt(resultLength, false)
			for i := 0; i < result.Length(); i++ {
				resExpanded.SetDigit(i, result.Digit(i))
			}
			result = resExpanded
		}
		y1 := absoluteSubOne(y)
		return absoluteXor(result, y1, result).Trim()
	}
	resultLength := maxInt(x.Length(), y.Length()) + 1
	// Assume x is positive.
	if x.Sign() {
		x, y = y, x
	}
	// x ^ (-y) == x ^ ~(y-1) == ~(x ^ (y-1)) == -((x ^ (y-1)) + 1)
	result := absoluteSubOne(y)
	if result.Length() < resultLength {
		resExpanded := NewBigInt(resultLength, false)
		for i := 0; i < result.Length(); i++ {
			resExpanded.SetDigit(i, result.Digit(i))
		}
		result = resExpanded
	}
	result = absoluteXor(result, x, result)
	return absoluteAddOne(result, true).Trim()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
