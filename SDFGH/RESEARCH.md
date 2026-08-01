# Research Notes

## Overview
Porting `GoogleChromeLabs/jsbi` (TypeScript) to Go.

### Reference Repository
- Source: `https://github.com/GoogleChromeLabs/jsbi`
- Commit: `5382367c7e3199858d36bb620977e1f90605bcb9`
- License: Apache-2.0

### Key Clusters to Port
1. Construction & parsing (string -> BigInt, radix input)
2. Comparison (EQ, LT, LE, GT, GE, NE)
3. Add / Subtract (addition, subtraction, sign handling)
4. Multiply (multiplication algorithms)
5. Divide / Remainder (division, modulo, sign handling)
6. Shifts (left shift, signed right shift)
7. Bitwise (AND, OR, XOR, NOT)
8. asIntN / asUintN (bit-width masking/clipping)
9. toString / radix conversion (output formatting)
