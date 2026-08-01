package jsbi

import "errors"

var (
	// ErrSyntax represents ECMAScript SyntaxError for invalid BigInt string representation.
	ErrSyntax = errors.New("SyntaxError: invalid BigInt string")
	// ErrRange represents ECMAScript RangeError for values out of range or unrepresentable floats.
	ErrRange = errors.New("RangeError: invalid value or radix out of range")
	// ErrType represents ECMAScript TypeError for invalid types supplied to BigInt.
	ErrType = errors.New("TypeError: cannot convert value to BigInt")
)
