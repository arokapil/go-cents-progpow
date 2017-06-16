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

// Stack is an object for basic stack operations. Items popped to the stack are
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

var sharedStack = initSharedStack()

func initSharedStack() *internalStack {
	_stack := &internalStack{
		index: 0,
	}
	_stack.resize(1024 * 1024)
	return _stack
}

func (sstack *internalStack) resize(size int) {
	data := make([]*big.Int, 0, size)
	if sstack.data != nil {
		copy(data, sstack.data)
	}
	//Fill it with BigInts
	for i := len(data); len(data) < size; i++ {
		data = append(data, new(big.Int))
	}
	sstack.data = data
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

//func (st *Stack) push(d *big.Int) {
//	// NOTE push limit (1024) is checked in baseCheck
//	//stackItem := new(big.Int).Set(d)
//	//st.data = append(st.data, stackItem)
//	sharedStack.data = append(sharedStack.data, d)
//	st.top++
//}

func (st *Stack) pushBytes(bytes []byte) {
	sharedStack.data[st.top].SetBytes(bytes)
	st.top++
}

func (st *Stack) pushBigint(d *big.Int) {
	sharedStack.data[st.top].Set(d)
	st.top++
}
func (st *Stack) pushInt64(i int64) {
	sharedStack.data[st.top].SetInt64(i)
	st.top++
}

func (st *Stack) pushUint64(i uint64) {
	sharedStack.data[st.top].SetUint64(i)
	st.top++
}

//func (st *Stack) pushN(ds ...*big.Int) {
//	st.data = append(st.data, ds...)
//}

func (st *Stack) pop() (ret *big.Int) {
	st.top--
	return sharedStack.data[st.top]
}
func (st *Stack) popInt64() (ret int64) {
	st.top--
	return sharedStack.data[st.top].Int64()
}
func (st *Stack) popUint64() (ret uint64) {
	st.top--
	return sharedStack.data[st.top].Uint64()
}

func (st *Stack) len() int {
	return st.top - st.bottom
}

func (st *Stack) swap(n int) {
	l := st.top
	sharedStack.data[l-n], sharedStack.data[l-1] = sharedStack.data[l-1], sharedStack.data[l-n]
}

func (st *Stack) dup(pool *intPool, n int) {
	sharedStack.data[st.top].Set(sharedStack.data[st.top-n])
	st.top++
}

func (st *Stack) peek() *big.Int {
	return sharedStack.data[st.top-1]
}

// Back returns the n'th item in stack
func (st *Stack) Back(n int) *big.Int {
	return sharedStack.data[st.top-n-1]
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
