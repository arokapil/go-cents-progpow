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
)

// stack is an object for basic stack operations. Items popped to the stack are
// expected to be changed and modified. stack does not take care of adding newly
// initialised objects.
type Stack struct {
	bottom int
	top    int
}

type internalStack struct {
	data  []*big.Int
	index int
	size  int
}

var sharedStack = &internalStack{
	data:  make([]*big.Int, 0, 1024),
	index: 0,
	size:  1024,
}

func newstack() *Stack {
	return &Stack{bottom: sharedStack.index, top: sharedStack.index}
}

func (st *Stack) Remove() {
	sharedStack.index = st.bottom
}

func (st *Stack) Data() []*big.Int {
	return sharedStack.data[st.bottom:st.top]
}

func (st *Stack) push(d *big.Int) {
	// NOTE push limit (1024) is checked in baseCheck
	//stackItem := new(big.Int).Set(d)
	//st.data = append(st.data, stackItem)
	sharedStack.data = append(sharedStack.data, d)
	st.top++
}

//func (st *Stack) pushN(ds ...*big.Int) {
//	st.data = append(st.data, ds...)
//}

func (st *Stack) pop() (ret *big.Int) {
	ret = sharedStack.data[len(sharedStack.data)-1]
	st.top--
	//	st.data = st.data[:len(st.data)-1]
	return
}

func (st *Stack) len() int {
	return st.top - st.bottom
}

func (st *Stack) swap(n int) {
	l := len(sharedStack.data)
	sharedStack.data[l-n], sharedStack.data[l-1] = sharedStack.data[l-1], sharedStack.data[l-n]
}

func (st *Stack) dup(pool *intPool, n int) {
	p := pool.get().Set(sharedStack.data[len(sharedStack.data)-n])
	st.push(p)
}

func (st *Stack) peek() *big.Int {
	return sharedStack.data[len(sharedStack.data)-1]
}

// Back returns the n'th item in stack
func (st *Stack) Back(n int) *big.Int {
	return sharedStack.data[len(sharedStack.data)-n-1]
}

func (st *Stack) require(n int) error {
	if st.len() < n {
		return fmt.Errorf("stack underflow (%d <=> %d)", st.len(), n)
	}
	return nil
}

func (st *Stack) Print() {
	fmt.Println("### stack ###")
	if len(st.Data()) > 0 {
		for i, val := range st.Data() {
			fmt.Printf("%-3d  %v\n", i, val)
		}
	} else {
		fmt.Println("-- empty --")
	}
	fmt.Println("#############")
}
