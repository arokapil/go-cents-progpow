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

package vm

import "math/big"

var checkVal = big.NewInt(-42)

const poolLimit = 256

type realStack struct {
	data []*big.Int
}

// intPool is a pool of big integers that
// can be reused for all big.Int operations.
type intPool struct {
	pool *realStack
}

func newRealstack() *realStack {
	return &realStack{data: make([]*big.Int, 0, poolLimit)}
}
func newIntPool() *intPool {
	return &intPool{pool: newRealstack()}
}
func (st *realStack) push(d *big.Int) {
	// NOTE push limit (1024) is checked in baseCheck
	//stackItem := new(big.Int).Set(d)
	//st.data = append(st.data, stackItem)
	st.data = append(st.data, d)
}
func (st *realStack) pushN(ds ...*big.Int) {
	st.data = append(st.data, ds...)
}

func (st *realStack) pop() (ret *big.Int) {
	ret = st.data[len(st.data)-1]
	st.data = st.data[:len(st.data)-1]
	return
}

func (st *realStack) len() int {
	return len(st.data)
}
func (p *intPool) get() *big.Int {
	if p.pool.len() > 0 {
		return p.pool.pop()
	}
	return new(big.Int)
}
func (p *intPool) put(is ...*big.Int) {
	if len(p.pool.data) >= poolLimit {
		return
	}

	for _, i := range is {
		// verifyPool is a build flag. Pool verification makes sure the integrity
		// of the integer pool by comparing values to a default value.
		if verifyPool {
			i.Set(checkVal)
		}

		p.pool.push(i)
	}
}
