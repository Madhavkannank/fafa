package port_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	jsbi "github.com/Madhavkannank/fafa/src"
)

// --- PHASE 3: PROPERTY TESTING ---

func TestPropertyAlgebraicIdentities(t *testing.T) {
	rng := rand.New(rand.NewSource(12345))
	for i := 0; i < 2000; i++ {
		aStr := genRandomDecString(rng, 1, 30)
		bStr := genRandomDecString(rng, 1, 30)
		a, _ := jsbi.FromString(aStr, 10)
		b, _ := jsbi.FromString(bStr, 10)

		// 1. (a + b) - b == a
		sum := jsbi.Add(a, b)
		resSub := jsbi.Subtract(sum, b)
		if !jsbi.Equal(resSub, a) {
			strVal, _ := jsbi.ToString(resSub, 10)
			t.Fatalf("Property failure (a+b)-b==a: a=%s b=%s got=%s", aStr, bStr, strVal)
		}

		// 2. (a * b) / b == a (when b != 0)
		if b.Length() > 0 {
			prod := jsbi.Multiply(a, b)
			quot, err := jsbi.Divide(prod, b)
			if err != nil || !jsbi.Equal(quot, a) {
				t.Fatalf("Property failure (a*b)/b==a: a=%s b=%s got=%v err=%v", aStr, bStr, quot, err)
			}
		}

		// 3. Bitwise idempotency & self-identities
		// x ^ x == 0
		xorSelf := jsbi.BitwiseXor(a, a)
		if xorSelf.Length() != 0 {
			t.Fatalf("Property failure a^a==0: a=%s got=%v", aStr, xorSelf)
		}
		// x & x == x
		andSelf := jsbi.BitwiseAnd(a, a)
		if !jsbi.Equal(andSelf, a) {
			t.Fatalf("Property failure a&a==a: a=%s got=%v", aStr, andSelf)
		}
		// x | x == x
		orSelf := jsbi.BitwiseOr(a, a)
		if !jsbi.Equal(orSelf, a) {
			t.Fatalf("Property failure a|a==a: a=%s got=%v", aStr, orSelf)
		}
		// ~~x == x
		notNot := jsbi.BitwiseNot(jsbi.BitwiseNot(a))
		if !jsbi.Equal(notNot, a) {
			t.Fatalf("Property failure ~~a==a: a=%s got=%v", aStr, notNot)
		}

		// 4. Truncation Idempotency
		n := rng.Intn(120) + 1
		u1, _ := jsbi.AsUintN(n, a)
		u2, _ := jsbi.AsUintN(n, u1)
		if !jsbi.Equal(u1, u2) {
			t.Fatalf("Property failure AsUintN idempotency: n=%d a=%s", n, aStr)
		}

		s1, _ := jsbi.AsIntN(n, a)
		s2, _ := jsbi.AsIntN(n, s1)
		if !jsbi.Equal(s1, s2) {
			t.Fatalf("Property failure AsIntN idempotency: n=%d a=%s", n, aStr)
		}

		// 5. String parsing round-trip
		// JSBI spec: explicit non-10 radix parsing only supports non-signed numbers (line 784)
		// Test radix 10 roundtrip for signed/unsigned, and radices 2,8,16,36 for magnitude
		sDec, _ := jsbi.ToString(a, 10)
		parsedDec, errDec := jsbi.FromString(sDec, 10)
		if errDec != nil || !jsbi.Equal(parsedDec, a) {
			t.Fatalf("Decimal parse round-trip failure: a=%s str=%s parsed=%v", aStr, sDec, parsedDec)
		}

		// Magnitude parse test across radices
		aMag := a.Copy()
		aMag.SetSign(false)
		for _, radix := range []int{2, 8, 10, 16, 36} {
			s, errStr := jsbi.ToString(aMag, radix)
			if errStr != nil {
				t.Fatalf("ToString err: %v", errStr)
			}
			parsed, errParse := jsbi.FromString(s, radix)
			if errParse != nil || !jsbi.Equal(parsed, aMag) {
				t.Fatalf("Parse round-trip failure: aMag=%v radix=%d str=%s parsed=%v err=%v", aMag, radix, s, parsed, errParse)
			}
		}
	}
}

// --- PHASE 4: STRESS TESTING ---

func TestStressLargeOperands(t *testing.T) {
	limbCounts := []int{100, 250, 500, 1000}
	rng := rand.New(rand.NewSource(9999))

	for _, limbs := range limbCounts {
		t.Run(fmt.Sprintf("%d_Limbs", limbs), func(t *testing.T) {
			a := genLimbBigInt(rng, limbs, false)
			b := genLimbBigInt(rng, limbs/2, false)

			start := time.Now()

			// Addition
			addRes := jsbi.Add(a, b)
			assertCanonicalZero(t, addRes, "Add")

			// Subtraction
			subRes := jsbi.Subtract(a, b)
			assertCanonicalZero(t, subRes, "Subtract")

			// Multiplication (limited size for 1000 limbs to maintain test speed)
			if limbs <= 250 {
				multRes := jsbi.Multiply(a, b)
				assertCanonicalZero(t, multRes, "Multiply")
			}

			// Division
			divRes, errDiv := jsbi.Divide(a, b)
			if errDiv != nil {
				t.Fatalf("Stress divide error: %v", errDiv)
			}
			assertCanonicalZero(t, divRes, "Divide")

			// Bitwise
			andRes := jsbi.BitwiseAnd(a, b)
			assertCanonicalZero(t, andRes, "BitwiseAnd")

			// Shift
			fifteen, _ := jsbi.BigIntVal(15)
			shiftRes, errShift := jsbi.LeftShift(a, fifteen)
			if errShift != nil {
				t.Fatalf("Stress shift error: %v", errShift)
			}
			assertCanonicalZero(t, shiftRes, "LeftShift")

			// Truncation
			asInt, _ := jsbi.AsIntN(limbs*15, a)
			assertCanonicalZero(t, asInt, "AsIntN")

			asUint, _ := jsbi.AsUintN(limbs*15, a)
			assertCanonicalZero(t, asUint, "AsUintN")

			// ToString
			toStrHex, _ := jsbi.ToString(b, 16)
			if len(toStrHex) == 0 {
				t.Fatalf("Stress ToString returned empty string")
			}

			t.Logf("Limb stress test %d limbs completed in %v", limbs, time.Since(start))
		})
	}
}

// --- PHASE 6: IMMUTABILITY & VALUE INDEPENDENCE AUDIT ---

func TestImmutabilityAudit(t *testing.T) {
	rng := rand.New(rand.NewSource(777))
	for i := 0; i < 500; i++ {
		aStr := genRandomDecString(rng, 1, 20)
		bStr := genRandomDecString(rng, 1, 20)
		a, _ := jsbi.FromString(aStr, 10)
		b, _ := jsbi.FromString(bStr, 10)

		// Snapshot a
		aSnapLen := a.Length()
		aSnapSign := a.Sign()
		aSnapDigits := make([]uint32, a.Length())
		for d := 0; d < a.Length(); d++ {
			aSnapDigits[d] = a.Digit(d)
		}

		// Operations to test (fast and slow paths)
		ops := []struct {
			name string
			fn   func() *jsbi.BigInt
		}{
			{"Add", func() *jsbi.BigInt { return jsbi.Add(a, b) }},
			{"Subtract", func() *jsbi.BigInt { return jsbi.Subtract(a, b) }},
			{"Multiply", func() *jsbi.BigInt { return jsbi.Multiply(a, b) }},
			{"Divide", func() *jsbi.BigInt { res, _ := jsbi.Divide(a, b); return res }},
			{"BitwiseAnd", func() *jsbi.BigInt { return jsbi.BitwiseAnd(a, b) }},
			{"BitwiseNot", func() *jsbi.BigInt { return jsbi.BitwiseNot(a) }},
			{"LeftShiftFast", func() *jsbi.BigInt { res, _ := jsbi.LeftShift(a, jsbi.Zero()); return res }},
			{"AsIntNFast", func() *jsbi.BigInt { res, _ := jsbi.AsIntN(1000, a); return res }},
			{"AsUintNFast", func() *jsbi.BigInt { res, _ := jsbi.AsUintN(1000, a); return res }},
			{"UnaryMinus", func() *jsbi.BigInt { return jsbi.UnaryMinus(a) }},
		}

		for _, op := range ops {
			res := op.fn()

			// Mutate returned object
			if res != nil && res.Length() > 0 {
				res.SetDigit(0, res.Digit(0)^0x3FFFFFFF)
				res.SetSign(!res.Sign())
			}

			// Assert original a is unmodified
			if a.Length() != aSnapLen || a.Sign() != aSnapSign {
				t.Fatalf("Immutability violation on %s: metadata altered! aLen=%d snapLen=%d", op.name, a.Length(), aSnapLen)
			}
			for d := 0; d < a.Length(); d++ {
				if a.Digit(d) != aSnapDigits[d] {
					t.Fatalf("Immutability violation on %s: digit %d mutated! got=0x%x want=0x%x", op.name, d, a.Digit(d), aSnapDigits[d])
				}
			}
		}
	}
}

// --- PHASE 7: CANONICAL ZERO AUDIT ---

func TestCanonicalZeroAudit(t *testing.T) {
	valFive, _ := jsbi.BigIntVal(5)
	valNegFive, _ := jsbi.BigIntVal(-5)
	valTen, _ := jsbi.BigIntVal(10)
	val9999, _ := jsbi.BigIntVal(9999)
	valFF, _ := jsbi.BigIntVal(0xFF)
	val1234, _ := jsbi.BigIntVal(1234)

	zeroInputs := []*jsbi.BigInt{
		jsbi.Zero(),
		jsbi.Add(valFive, valNegFive),
		jsbi.Subtract(valTen, valTen),
		jsbi.Multiply(jsbi.Zero(), val9999),
		jsbi.BitwiseAnd(valFF, jsbi.Zero()),
		jsbi.BitwiseXor(val1234, val1234),
	}

	for i, z := range zeroInputs {
		if z.Length() != 0 {
			t.Fatalf("Expected zero length for case %d, got %d", i, z.Length())
		}
		if z.Sign() {
			t.Fatalf("Canonical Zero Violation! Length()==0 but Sign()==true for case %d", i)
		}
	}
}

// --- Helpers ---

func assertCanonicalZero(t *testing.T, x *jsbi.BigInt, opName string) {
	if x != nil && x.Length() == 0 && x.Sign() {
		t.Fatalf("Canonical Zero Violation in %s: Length()==0 but Sign()==true", opName)
	}
}

func genRandomDecString(rng *rand.Rand, minDigits, maxDigits int) string {
	numDigits := rng.Intn(maxDigits-minDigits+1) + minDigits
	digits := make([]byte, numDigits)
	digits[0] = byte('1' + rng.Intn(9))
	for i := 1; i < numDigits; i++ {
		digits[i] = byte('0' + rng.Intn(10))
	}
	sign := ""
	if rng.Float64() < 0.5 {
		sign = "-"
	}
	return sign + string(digits)
}

func genLimbBigInt(rng *rand.Rand, numLimbs int, sign bool) *jsbi.BigInt {
	res := jsbi.NewBigInt(numLimbs, sign)
	for i := 0; i < numLimbs; i++ {
		res.SetDigit(i, uint32(rng.Int31n(0x3FFFFFFF)))
	}
	// Ensure MSD is non-zero
	res.SetDigit(numLimbs-1, uint32(rng.Int31n(0x3FFFFFFE))+1)
	return res
}
