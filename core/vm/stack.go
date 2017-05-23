// Copyright 2014 The go-ethereum Authors
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

package vm

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

// stack is an object for basic stack operations. Items popped to the stack are
// expected to be changed and modified. stack does not take care of adding newly
// initialised objects.
type Stack struct {
	data []*big.Int
}

func newstack() *Stack {
	return &Stack{}
}

func (st *Stack) Data() []*big.Int {
	return st.data
}

func (st *Stack) push(d *big.Int) {
	// NOTE push limit (1024) is checked in baseCheck
	//stackItem := new(big.Int).Set(d)
	//st.data = append(st.data, stackItem)
	st.data = append(st.data, d)
}
func (st *Stack) pushN(ds ...*big.Int) {
	st.data = append(st.data, ds...)
}

func (st *Stack) pop() (ret *big.Int) {
	ret = st.data[len(st.data)-1]
	st.data = st.data[:len(st.data)-1]
	return
}
func (st *Stack) Add() {
	x, y := st.data[len(st.data)-1], st.data[len(st.data)-2]
	st.data[len(st.data)-2] = y.Add(x, y)
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) Sub() {
	x, y := st.data[len(st.data)-1], st.data[len(st.data)-2]
	st.data[len(st.data)-2] = y.Sub(x, y)
	st.data = st.data[:len(st.data)-1]
}

func (st *Stack) Mul() {
	x, y := st.data[len(st.data)-1], st.data[len(st.data)-2]
	y.Mul(y, x)
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) Div() {
	x, y := st.data[len(st.data)-1], st.data[len(st.data)-2]
	if y.Sign() != 0 {
		math.U256(y.Div(x, y))
	} else {
		y.SetUint64(0)
	}
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) Sdiv() {
	y := math.S256(st.data[len(st.data)-2])
	if y.Sign() == 0 {
		y.SetUint64(0)
	} else {
		x := math.S256(st.data[len(st.data)-1])
		if x.Sign() == y.Sign() {
			x.Div(x.Abs(x), y.Abs(y))
		} else {
			x.Div(x.Abs(x), y.Abs(y))
			y.Neg(x)
		}
	}
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) Mod() {
	x, y := st.data[len(st.data)-1], st.data[len(st.data)-2]
	if y.Sign() == 0 {
		y.SetUint64(0)
	} else {
		math.U256(y.Mod(x, x))
		//st.data[len(st.data)-2].SetUint64(1)
	}
	st.data[len(st.data)-2] = math.U256(y.Div(x, y))
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) Smod() {
	y := st.data[len(st.data)-2]
	if y.Sign() == 0 {
		y.SetUint64(0)
	} else {
		x := st.data[len(st.data)-1]
		if x.Sign() < 0 {
			y.Mod(x.Abs(x), y.Abs(y))
			y.Neg(y)
		} else {
			y.Mod(x.Abs(x), y.Abs(y))
		}
	}
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) Exp() {
	base, exp := st.data[len(st.data)-1], st.data[len(st.data)-2]
	exp.Set(math.Exp(base, exp))
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) SignExtend() {
	back := st.data[len(st.data)-1]
	if back.Cmp(big31) < 0 {
		bit := uint(back.Uint64()*8 + 7)
		num := st.data[len(st.data)-2]
		mask := back.Lsh(common.Big1, bit)
		mask.Sub(mask, common.Big1)
		if num.Bit(int(bit)) > 0 {
			num.Or(num, mask.Not(mask))
		} else {
			num.And(num, mask)
		}
		math.U256(num)
	}
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) Not() {
	math.U256(st.data[len(st.data)-1].Not(st.data[len(st.data)-1]))
}
func (st *Stack) Lt() {
	x, y := st.data[len(st.data)-1], st.data[len(st.data)-2]
	if x.Cmp(y) < 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) Gt() {
	x, y := st.data[len(st.data)-1], st.data[len(st.data)-2]
	if x.Cmp(y) > 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) Slt() {
	x, y := math.S256(st.data[len(st.data)-1]), math.S256(st.data[len(st.data)-2])
	if x.Cmp(y) < 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) Sgt() {
	x, y := math.S256(st.data[len(st.data)-1]), math.S256(st.data[len(st.data)-2])
	if x.Cmp(y) > 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) Eq() {
	x, y := st.data[len(st.data)-1], st.data[len(st.data)-2]
	if x.Cmp(y) == 0 {
		y.SetUint64(1)
	} else {
		y.SetUint64(0)
	}
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) isZero() {
	x := st.data[len(st.data)-1]
	if x.Sign() > 0 {
		st.data[len(st.data)-1] = x.SetUint64(0)
	} else {
		st.data[len(st.data)-1] = x.SetUint64(1)
	}
}

func (st *Stack) And() {
	y := st.data[len(st.data)-2]
	y.And(st.data[len(st.data)-1], y)
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) Or() {
	y := st.data[len(st.data)-2]
	y.Or(st.data[len(st.data)-1], y)
	st.data = st.data[:len(st.data)-1]
}
func (st *Stack) Xor() {
	y := st.data[len(st.data)-2]
	y.Xor(st.data[len(st.data)-1], y)
	st.data = st.data[:len(st.data)-1]
}

func (st *Stack) Byte() {
	th, val := st.data[len(st.data)-1], st.data[len(st.data)-2]
	if th.Cmp(big32) < 0 {
		val.SetInt64(int64(math.PaddedBigBytes(val, 32)[th.Int64()]))
	} else {
		val.SetUint64(0)
	}
}

func (st *Stack) Addmod() {
	x, y, z := st.data[len(st.data)-1], st.data[len(st.data)-2], st.data[len(st.data)-3]
	if z.Cmp(bigZero) > 0 {
		x.Add(x, y)
		math.U256(z.Mod(x, z))
	} else {
		z.SetUint64(0)
	}
	st.data = st.data[:len(st.data)-2]
}

func (st *Stack) Mulmod() {
	x, y, z := st.data[len(st.data)-1], st.data[len(st.data)-2], st.data[len(st.data)-3]
	if z.Cmp(bigZero) > 0 {
		x.Mul(x, y)
		math.U256(z.Mod(x, z))
	} else {
		z.SetUint64(0)
	}
	st.data = st.data[:len(st.data)-2]
}

func (st *Stack) pop2() (a *big.Int, b *big.Int) {
	a, b = st.data[len(st.data)-1], st.data[len(st.data)-2]
	st.data = st.data[:len(st.data)-2]
	return
}
func (st *Stack) pop3() (a *big.Int, b *big.Int, c *big.Int) {
	a, b, c = st.data[len(st.data)-1], st.data[len(st.data)-2], st.data[len(st.data)-3]
	st.data = st.data[:len(st.data)-3]
	return
}

func (st *Stack) len() int {
	return len(st.data)
}

func (st *Stack) swap(n int) {
	st.data[st.len()-n], st.data[st.len()-1] = st.data[st.len()-1], st.data[st.len()-n]
}

func (st *Stack) dup(n int) {
	st.push(new(big.Int).Set(st.data[st.len()-n]))
}

func (st *Stack) peek() *big.Int {
	return st.data[st.len()-1]
}

// Back returns the n'th item in stack
func (st *Stack) Back(n int) *big.Int {
	return st.data[st.len()-n-1]
}

func (st *Stack) require(n int) error {
	if st.len() < n {
		return fmt.Errorf("stack underflow (%d <=> %d)", len(st.data), n)
	}
	return nil
}

func (st *Stack) Print() {
	fmt.Println("### stack ###")
	if len(st.data) > 0 {
		for i, val := range st.data {
			fmt.Printf("%-3d  %v\n", i, val)
		}
	} else {
		fmt.Println("-- empty --")
	}
	fmt.Println("#############")
}
