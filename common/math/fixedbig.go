// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package math provides integer math utilities.

package math

import (
	"fmt"
	"math/big"
	"math/bits"
)

type Fixed256bit struct {
	a uint64 // Most significant
	b uint64
	c uint64
	d uint64 // Least significant
}

// newFixedFromBig is a convenience-constructor from big.Int. Not optimized for speed, mainly for easy testing
func NewFixedFromBig(int *big.Int) (*Fixed256bit, bool) {
	// Let's not ruin the argument
	z := &Fixed256bit{}
	overflow := z.Set(int)
	return z, overflow
}

func NewFixed() *Fixed256bit {
	return &Fixed256bit{}
}

// Set is a convenience-setter from big.Int. Not optimized for speed, mainly for easy testing
func (z *Fixed256bit) Set(int *big.Int) bool {
	// Let's not ruin the argument
	x := new(big.Int).Set(int)

	for x.Cmp(new(big.Int)) < 0 {
		// Below 0
		x.Add(tt256, x)
	}
	z.d = x.Uint64()
	z.c = x.Rsh(x, 64).Uint64()
	z.b = x.Rsh(x, 64).Uint64()
	z.a = x.Rsh(x, 64).Uint64()
	x.Rsh(x, 64).Uint64()
	return len(x.Bits()) != 0
}
func (z *Fixed256bit) Clone() *Fixed256bit {
	return &Fixed256bit{z.a, z.b, z.c, z.d}
}
func add64(a uint64, b uint64, carry uint64) (uint64, uint64) {

	var (
		q, sum uint64
	)
	sum = carry + (a & 0x00000000ffffffff) + (b & 0x00000000ffffffff)
	q = sum & 0x00000000ffffffff
	carry = sum >> 32
	sum = carry + (a >> 32) + (b >> 32)
	q |= (sum & 0x00000000ffffffff) << 32
	carry = sum >> 32
	return q, carry
}

// Add sets z to the sum x+y and returns whether overflow occurred
func (z *Fixed256bit) Add(x, y *Fixed256bit) bool {

	var (
		q, carry, sum uint64
	)
	q = x.d + y.d
	if y.d > 0 && x.d > 0 {
		sum = carry + (y.d & 0x00000000ffffffff) + (x.d & 0x00000000ffffffff)
		carry = (sum >> 32)
		sum = carry + (y.d >> 32) + (x.d >> 32)
		carry = (sum >> 32)
	}
	z.d = q
	z.c, carry = add64(y.c, x.c, carry)

	if carry == 0 && y.b == 0 && y.a == 0 && x.b == 0 && x.a == 0 {
		z.b = 0
		z.a = 0
		return false
	}
	z.b, carry = add64(y.b, x.b, carry)

	if carry == 0 && y.a == 0 && x.a == 0 {
		z.a = 0
		return false
	}
	z.a, carry = add64(y.a, x.a, carry)
	return (carry != 0)
}

// Sub sets z to the difference x-y and returns z.
func (z *Fixed256bit) Sub(x, y *Fixed256bit) bool {

	var (
		underflow bool
		q         uint64
	)

	q = x.d - y.d
	underflow = (q > x.d) // underflow
	z.d = q

	q = x.c - y.c
	if q > x.c { // underflow again
		if underflow {
			q--
		}
		underflow = true
	} else if underflow {
		// No underflow, we can decrement it
		q--
		// May cause another underflow
		underflow = (q > x.c)
	}
	z.c = q

	q = x.b - y.b
	if q > x.b { // underflow again
		if underflow {
			q--
		}
		underflow = true
	} else if underflow {
		// No underflow, we can decrement it
		q--
		// May cause another underflow
		underflow = (q > x.b)
	}
	z.b = q

	q = x.a - y.a
	if q > x.a { // underflow again
		if underflow {
			q--
		}
		underflow = true
	} else if underflow {
		// No underflow, we can decrement it
		q--
		// May cause another underflow
		underflow = (q > x.a)
	}
	z.a = q
	return underflow
}

// mul64 multiplies two 64-bit uints and sets the result in x. The parameter y
// is used as a buffer, and will be overwritten (does not have to be cleared prior
// to usage.
func (x *Fixed256bit) mul64(a, b uint64, y *Fixed256bit) *Fixed256bit {

	if a == 0 || b == 0 {
		return x.Clear()
	}
	low_a := a & 0x00000000ffffffff
	low_b := b & 0x00000000ffffffff
	high_a := a >> 32
	high_b := b >> 32

	d2 := low_a * high_b // Needs up 32
	d3 := high_a * low_b // Needs up 32

	x.a, x.b, x.c, x.d = 0, 0, high_a*high_b, low_a*low_b

	y.a, y.b = 0, 0
	y.c = d2 >> 32
	y.d = (d2 & 0x00000000ffffffff) << 32

	x.Add(x, y)

	y.a, y.b = 0, 0
	y.c = d3 >> 32
	y.d = (d3 & 0x00000000ffffffff) << 32

	x.Add(x, y)
	return x
}

// Mul sets z to the sum x*y
func (z *Fixed256bit) Mul(x, y *Fixed256bit) {

	var (
		alfa  = &Fixed256bit{} // Aggregate results
		beta  = &Fixed256bit{} // Calculate intermediate
		gamma = &Fixed256bit{} // Throwaway buffer
	)
	// The numbers are internally represented as [ a, b, c, d ]
	// We do the following operations
	//
	// d1 * d2
	// d1 * c2 (upshift 64)
	// d1 * b2 (upshift 128)
	// d1 * a2 (upshift 192)
	//
	// c1 * d2 (upshift 64)
	// c1 * c2 (upshift 128)
	// c1 * b2 (upshift 192)
	//
	// b1 * d2 (upshift 128)
	// b1 * c2 (upshift 192)
	//
	// a1 * d2 (upshift 192)
	//
	// And we aggregate results into 'alfa'

	// One optimization, however, is reordering.
	// For these ones, we don't care about if they overflow, thus we can use native multiplication
	// and set the result immediately into `a` of the result.
	// b1 * c2 (upshift 192)
	// a1 * d2 (upshift 192)
	// d1 * a2 (upshift 192)
	// c1 * b2 (upshift 192)

	// Remaining ops:
	//
	// d1 * d2
	// d1 * c2 (upshift 64)
	// d1 * b2 (upshift 128)
	//
	// c1 * d2 (upshift 64)
	// c1 * c2 (upshift 128)
	//
	// b1 * d2 (upshift 128)

	alfa.mul64(x.d, y.d, beta)
	alfa.a = x.d*y.a + x.c*y.b + x.b*y.c + x.a*y.d // Top ones, ignore overflow

	beta.mul64(x.d, y.c, gamma).lsh64(beta)

	alfa.Add(alfa, beta)

	beta.mul64(x.d, y.b, gamma).lsh128(beta)

	alfa.Add(alfa, beta)

	beta.mul64(x.c, y.d, gamma).lsh64(beta)

	alfa.Add(alfa, beta)
	beta.mul64(x.c, y.c, gamma).lsh128(beta)
	alfa.Add(alfa, beta)

	beta.mul64(x.b, y.d, gamma).lsh128(beta)
	z.Add(alfa, beta)

}

/*
// Div sets z to the quotient x/y for y != 0 and returns z.
// If y == 0, z is set to 0
// Div implements Euclidean division (unlike Go); see DivMod for more details.
func (z *Fixed256bit) Div(x, y *Fixed256bit) *Fixed256bit {
	// Shortcut some cases
	if y.IsZero() || y.Gt(x) {
		return z.Clear()
	}
	if y.Eq(x) {
		z.a, z.b, z.c, z.d = 0, 0, 0, 1
		return z
	}
	// At this point, we know
	// x/y ; x > y > 0

	// The rest is a pretty un-optimized implementation of "Long division"
	// from https://en.wikipedia.org/wiki/Division_algorithm.
	// Could probably be improved upon
	xbitlen := x.Bitlen()

	R := &Fixed256bit{}
	Q := &Fixed256bit{}
	N := x
	D := y
	for i:= xbitlen -1; i > 0 ; i--{
		R.Rsh(1)

	}
}
*/
func (x *Fixed256bit) Bitlen() int {
	switch {
	case x.a != 0:
		return 192 + bits.Len64(x.a)
	case x.b != 0:
		return 128 + bits.Len64(x.b)
	case x.c != 0:
		return 64 + bits.Len64(x.c)
	default:
		return bits.Len64(x.d)
	}
}

// Mod sets z to the modulus x%y for y != 0 and returns z.
// If y == 0, z is set to 0 (OBS: differs from the big.Int)
// Mod implements Euclidean modulus (unlike Go); see DivMod for more details.
func (z *Fixed256bit) Mod(x, y *Fixed256bit) *Fixed256bit {
	if y.IsZero() {
		return z.Clear()
	}
	panic("TODO! Implement me")
	return z
}

func (z *Fixed256bit) lsh64(x *Fixed256bit) *Fixed256bit {
	z.a = x.b
	z.b = x.c
	z.c = x.d
	z.d = 0
	return z
}
func (z *Fixed256bit) lsh128(x *Fixed256bit) *Fixed256bit {
	z.a = x.c
	z.b = x.d
	z.c, z.d = 0, 0
	return z
}
func (z *Fixed256bit) lsh192(x *Fixed256bit) *Fixed256bit {
	z.a, z.b, z.c, z.d = x.d, 0, 0, 0
	return z
}
func (z *Fixed256bit) rsh128(x *Fixed256bit) *Fixed256bit {
	z.d = x.b
	z.c = x.a
	z.b = 0
	z.a = 0
	return z
}
func (z *Fixed256bit) rsh64(x *Fixed256bit) *Fixed256bit {
	z.d = x.c
	z.c = x.b
	z.b = x.a
	z.a = 0
	return z
}
func (z *Fixed256bit) rsh192(x *Fixed256bit) *Fixed256bit {
	z.d, z.c, z.b, z.a = x.a, 0, 0, 0
	return z
}

// Not sets z = ^x and returns z.
func (z *Fixed256bit) Not() *Fixed256bit {
	z.a, z.b, z.c, z.d = ^z.a, ^z.b, ^z.c, ^z.d
	return z
}

// Gt returns true if f > g
func (f *Fixed256bit) Gt(g *Fixed256bit) bool {
	return (f.a > g.a) || (f.b > g.b) || (f.c > g.c) || (f.d > g.d)
}
// SetIfGt sets f to 1 if f > g
func (f *Fixed256bit) SetIfGt(g *Fixed256bit) {
	if (f.a > g.a) || (f.b > g.b) || (f.c > g.c) || (f.d > g.d){
		f.SetOne()
	}else{
		f.Clear()
	}
}

// Lt returns true if l < g
func (f *Fixed256bit) Lt(g *Fixed256bit) bool {
	return (f.a < g.a) || (f.b < g.b) || (f.c < g.c) || (f.d < g.d)
}

// SetIfLt sets f to 1 if f < g
func (f *Fixed256bit) SetIfLt(g *Fixed256bit) bool {
	if (f.a < g.a) || (f.b < g.b) || (f.c < g.c) || (f.d < g.d){
		f.SetOne()
	}else{
		f.Clear()
	}
}
func (f *Fixed256bit) SetUint64(a uint64) *Fixed256bit{
	f.a, f.b, f.c, f.d = 0,0,0,a
	return f
}

// Eq returns true if f == g
func (f *Fixed256bit) Eq(g *Fixed256bit) bool {
	return (f.a == g.a) && (f.b == g.b) && (f.c == g.c) && (f.d == g.d)
}
// Eq returns true if f == g
func (f *Fixed256bit) SetIfEq(g *Fixed256bit) {
	if (f.a == g.a) && (f.b == g.b) && (f.c == g.c) && (f.d == g.d){
		f.SetOne()
	}else{
		f.Clear()
	}
}
// Cmp compares x and y and returns:
//
//   -1 if x <  y
//    0 if x == y
//   +1 if x >  y
//
func (x *Fixed256bit) Cmp(y *Fixed256bit) (r int) {
	if x.Gt(y) {
		return 1
	}
	if x.Lt(y) {
		return -1
	}
	return 0
}

// ltsmall can be used to check if x is smaller than n
func (x *Fixed256bit) ltSmall(n uint64) bool {
	return x.a == 0 && x.b == 0 && x.c == 0 && x.d < n
}

// IsUint64 reports whether x can be represented as a uint64.
func (x *Fixed256bit) IsUint64() bool {
	return (x.a == 0) && (x.b == 0) && (x.c == 0)
}

// IsZero returns true if f == 0
func (f *Fixed256bit) IsZero() bool {
	return (f.a == 0) && (f.b == 0) && (f.c == 0) && (f.d == 0)
}

// IsOne returns true if f == 1
func (f *Fixed256bit) IsOne() bool {
	return f.a == 0 && f.b == 0 && f.c == 0 && f.d == 1
}

// Clear sets z to 0
func (z *Fixed256bit) Clear() *Fixed256bit {
	z.a, z.b, z.c, z.d = 0, 0, 0, 0
	return z
}

// SetOne sets z to 1
func (z *Fixed256bit) SetOne() *Fixed256bit {
	z.a, z.b, z.c, z.d = 0, 0, 0, 1
	return z
}
// Lsh sets z = x << n and returns z.
func (z *Fixed256bit) Lsh(x *Fixed256bit, n uint) *Fixed256bit {

	// Big swaps first
	switch {
	case n >= 256:
		return z.Clear()
	case n >= 192:
		z.lsh192(x)
		n -= 192
	case n >= 128:
		z.lsh128(x)
		n -= 128
	case n >= 64:
		z.lsh64(x)
		n -= 64
	default:
		z.Copy(x)
	}
	if n == 0 {
		return z
	}
	// remaining shifts
	var (
		a, b uint64
	)
	a = z.d >> (64 - n)
	z.d = z.d << n

	b = z.c >> (64 - n)
	z.c = (z.c << n) | a

	a = z.b >> (64 - n)
	z.b = (z.b << n) | b

	b = z.a >> (64 - n)
	z.a = (z.a << n) | a
	return z
}

// Rsh sets z = x >> n and returns z.
func (z *Fixed256bit) Rsh(x *Fixed256bit, n uint) *Fixed256bit {

	// Big swaps first
	switch {
	case n >= 256:
		return z.Clear()
	case n >= 192:
		z.rsh192(x)
		n -= 192
	case n >= 128:
		z.rsh128(x)
		n -= 128
	case n >= 64:
		z.rsh64(x)
		n -= 64
	default:
		z.Copy(x)
	}
	if n == 0 {
		return z
	}

	// remaining shifts
	var (
		a, b uint64
	)

	a = z.a << (64 - n)
	z.a = z.a >> n

	b = z.b << (64 - n)
	z.b = (z.b >> n) | a

	a = z.c << (64 - n)
	z.c = (z.c >> n) | b

	b = z.d << (64 - n)
	z.d = (z.d >> n) | a

	return z
}
func (z *Fixed256bit) Copy(x *Fixed256bit) *Fixed256bit {
	z.a, z.b, z.c, z.d = x.a, x.b, x.c, x.d
	return z
}

// Or sets z = x | y and returns z.
func (z *Fixed256bit) Or(x, y *Fixed256bit) *Fixed256bit {
	z.a = x.a | y.a
	z.b = x.b | y.b
	z.c = x.c | y.c
	z.d = x.d | y.d
	return z
}

// And sets z = x & y and returns z.
func (z *Fixed256bit) And(x, y *Fixed256bit) *Fixed256bit {
	z.a = x.a & y.a
	z.b = x.b & y.b
	z.c = x.c & y.c
	z.d = x.d & y.d
	return z
}

// Xor sets z = x ^ y and returns z.
func (z *Fixed256bit) Xor(x, y *Fixed256bit) *Fixed256bit {
	z.a = x.a ^ y.a
	z.b = x.b ^ y.b
	z.c = x.c ^ y.c
	z.d = x.d ^ y.d
	return z
}

// Byte sets f to the value of the byte at position n,
// Example: f = '5', n=31 => 5
func (f *Fixed256bit) Byte(n *Fixed256bit) *Fixed256bit {
	var number uint64
	if n.ltSmall(32) {
		if n.d > 24 {
			// f.d holds bytes [24 .. 31]
			number = f.d
		} else if n.d > 15 {
			// f.c holds bytes [16 .. 23]
			number = f.c
		} else if n.d > 7 {
			// f.b holds bytes [8 .. 15]
			number = f.b
		} else {
			// f.a holds MSB, bytes [0 .. 7]
			number = f.a
		}
		offset := 8*(n.d % 8)
		number = (number & (0xff00000000000000 >> offset)) >> (56 - offset)
	}

	f.a,f.b, f.c, f.d = 0,0,0, number
	return f
}

func (f *Fixed256bit) Hex() string {
	return fmt.Sprintf("%016x.%016x.%016x.%016x", f.a, f.b, f.c, f.d)
}

// Exp implements exponentiation by squaring.
// Exp returns a newly-allocated big integer and does not change
// base or exponent.
//
// Courtesy @karalabe and @chfast, with improvements by @holiman
func ExpF(base, exponent *Fixed256bit) *Fixed256bit {
	z := &Fixed256bit{a: 0, b: 0, c: 0, d: 1}
	// b^0 == 1
	if exponent.IsZero() || base.IsOne() {
		return z
	}
	// b^1 == 1
	if exponent.IsOne() {
		z.Copy(base)
		return z
	}
	var (
		word uint64
		bits int
	)
	exp_bitlen := exponent.Bitlen()

	word = exponent.d
	bits = 0
	for ; bits < exp_bitlen && bits < 64; bits++ {
		if word&1 == 1 {
			z.Mul(z, base)
		}
		base.Mul(base, base)
		word >>= 1
	}

	word = exponent.c
	for ; bits < exp_bitlen && bits < 128; bits++ {
		if word&1 == 1 {
			z.Mul(z, base)
		}
		base.Mul(base, base)
		word >>= 1
	}

	word = exponent.b
	for ; bits < exp_bitlen && bits < 192; bits++ {
		if word&1 == 1 {
			z.Mul(z, base)
		}
		base.Mul(base, base)
		word >>= 1
	}

	word = exponent.a
	for ; bits < exp_bitlen && bits < 256; bits++ {
		if word&1 == 1 {
			z.Mul(z, base)
		}
		base.Mul(base, base)
		word >>= 1
	}
	return z
}
