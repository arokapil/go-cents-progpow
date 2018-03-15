// Copyright 2015 The go-ethereum Authors
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
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

var (

	errWriteProtection       = errors.New("evm: write protection")
	errReturnDataOutOfBounds = errors.New("evm: return data out of bounds")
	errExecutionReverted     = errors.New("evm: execution reverted")
	errMaxCodeSizeExceeded   = errors.New("evm: max code size exceeded")
)

func opAdd(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Add(x, y)
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opSub(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Sub(x, y)
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opMul(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Mul(x, y)
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opDiv(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Div(x, y)
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opSdiv(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Sdiv(x, y)
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opMod(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Mod(x, y)
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opSmod(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Smod(x, y)
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opExp(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	base, exponent := stack.pop(), stack.peek()
	exponent.Exp(base, exponent)
	evm.interpreter.intPool.put(base)
	return nil, nil
}

func opSignExtend(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	back, num := stack.pop(), stack.peek()
	num.SignExtend(back, num)
	evm.interpreter.intPool.put(back)
	return nil, nil
}

func opNot(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.peek().Not()
	return nil, nil
}

func opLt(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Lt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opGt(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Gt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opSlt(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Slt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opSgt(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	if x.Sgt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opEq(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.SetIfEq(x)
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opIszero(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := stack.peek()
	if x.IsZero() {
		x.SetOne()
	} else {
		x.Clear()
	}
	return nil, nil
}

func opAnd(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.And(x, y)
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opOr(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Or(x, y)
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opXor(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y := stack.pop(), stack.peek()
	y.Xor(x, y)
	evm.interpreter.intPool.put(x)
	return nil, nil
}

func opByte(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	th, val := stack.pop(), stack.peek()
	val.Byte(th)
	evm.interpreter.intPool.put(th)
	return nil, nil
}

func opAddmod(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y, z := stack.pop(), stack.pop(), stack.peek()
	if !z.IsZero() {
		// if x + y overflows, the mod becomes wrong.
		// Example:
		// x = ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
		// y = fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe
		// z = ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
		// ---
		// x + y        =  1fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd
		// in 256 bit repr: fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd
		//
		// As a hack, we use bigints here instead
		// !TODO
		bigx, bigy, bigz := x.ToBig(), y.ToBig(), z.ToBig()
		bigx.Add(bigx, bigy)
		bigz.Mod(bigx, bigz)
		z.SetBytes(bigz.Bytes())
	} else {
		z.Clear()
	}
	evm.interpreter.intPool.put(x, y)
	return nil, nil
}

func opMulmod(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x, y, z := stack.pop(), stack.pop(), stack.peek()
	if !z.IsZero() {
		// As a hack, we use bigints here instead
		// !TODO
		bigx, bigy, bigz := x.ToBig(), y.ToBig(), z.ToBig()
		bigx.Mul(bigx, bigy)
		bigz.Mod(bigx, bigz)
		z.SetBytes(bigz.Bytes())
	} else {
		z.Clear()
	}
	evm.interpreter.intPool.put(x, y)
	return nil, nil
}

// opSHL implements Shift Left
// The SHL instruction (shift left) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the left by arg1 number of bits.
func opSHL(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	shift, value := stack.pop(), stack.peek()
	if shift.LtUint64(256) {
		value.Lsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}
	return nil, nil
}

// opSHR implements Logical Shift Right
// The SHR instruction (logical shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with zero fill.
func opSHR(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	shift, value := stack.pop(), stack.peek()
	if shift.LtUint64(256) {
		value.Rsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}
	return nil, nil
}

// opSAR implements Arithmetic Shift Right
// The SAR instruction (arithmetic shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with sign extension.
func opSAR(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	shift, value := stack.pop(), stack.peek()
	defer evm.interpreter.intPool.put(shift) // First operand back into the pool

	if shift.GtUint64(256) {
		if value.Sign() > 0 {
			value.Clear()
		} else {
			value.Clear()
			value.Sub64(value, uint64(1))
		}
		return nil, nil
	}
	n := uint(shift.Uint64())
	value.Srsh(value, n)
	return nil, nil
}

func opSha3(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.peek()
	data := memory.Get(offset.Int64(), size.Int64())
	hash := crypto.Keccak256(data)

	if evm.vmConfig.EnablePreimageRecording {
		evm.StateDB.AddPreimage(common.BytesToHash(hash), data)
	}
	size.SetBytes(hash)
	evm.interpreter.intPool.put(offset)
	return nil, nil
}

func opAddress(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	a := evm.interpreter.intPool.get()
	a.SetBytes(contract.Address().Bytes())
	stack.push(a)
	return nil, nil
}

func opBalance(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := stack.peek()
	addr := common.BytesToAddress(x.Bytes())
	balance := evm.StateDB.GetBalance(addr)
	x.SetFromBig(balance)
	return nil, nil
}

func opOrigin(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := evm.interpreter.intPool.get()
	x.SetBytes(evm.Origin.Bytes())
	stack.push(x)
	return nil, nil
}

func opCaller(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := evm.interpreter.intPool.get()
	x.SetBytes(contract.Caller().Bytes())
	stack.push(x)
	return nil, nil
}

func opCallValue(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := evm.interpreter.intPool.get()
	x.SetFromBig(contract.value)
	stack.push(x)
	return nil, nil
}

func opCallDataLoad(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := stack.peek()
	if offset, overflow := x.Uint64WithOverflow(); !overflow {
		data := getData(contract.Input, offset, 32)
		x.SetBytes(data)
	} else {
		x.Clear()
	}
	return nil, nil
}

func opCallDataSize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.interpreter.intPool.get().SetUint64(uint64(len(contract.Input))))
	return nil, nil
}

func opCallDataCopy(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		dataOffset = stack.pop()
		length     = stack.pop()
	)

	uint64Offset, overflow := dataOffset.Uint64WithOverflow()
	if overflow {
		uint64Offset = 0xffffffffffffffff
	}

	// These values are already checked for overflow during
	// gas cost calculation
	memOffset64 := memOffset.Uint64()
	length64 := length.Uint64()
	memory.Set(memOffset64, length64, getData(contract.Input, uint64Offset, length64))

	defer evm.interpreter.intPool.put(memOffset, dataOffset, length)
	return nil, nil
}

func opReturnDataSize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.interpreter.intPool.get().SetUint64(uint64(len(evm.interpreter.returnData))))
	return nil, nil
}

func opReturnDataCopy(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		dataOffset = stack.pop()
		length     = stack.pop()
	)
	defer evm.interpreter.intPool.put(memOffset, dataOffset, length)
	end := evm.interpreter.intPool.get()
	overflow := end.AddOverflow(dataOffset, length)
	if overflow {
		return nil, errReturnDataOutOfBounds
	}
	if end.GtUint64(uint64(len(evm.interpreter.returnData))) {
		return nil, errReturnDataOutOfBounds
	}
	memory.Set(memOffset.Uint64(), length.Uint64(), evm.interpreter.returnData[dataOffset.Uint64():end.Uint64()])
	return nil, nil
}

func opExtCodeSize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	a := stack.peek()
	addr := common.BytesToAddress(a.Bytes())
	a.SetUint64(uint64(evm.StateDB.GetCodeSize(addr)))
	return nil, nil
}

func opCodeSize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	l := evm.interpreter.intPool.get().SetUint64(uint64(len(contract.Code)))
	stack.push(l)

	return nil, nil
}

func opCodeCopy(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	uint64Offset, overflow := codeOffset.Uint64WithOverflow()
	if overflow {
		uint64Offset = 0xffffffffffffffff
	}
	codeCopy := getData(contract.Code, uint64Offset, length.Uint64())

	memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	evm.interpreter.intPool.put(memOffset, codeOffset, length)
	return nil, nil
}

func opExtCodeCopy(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		a          = stack.pop()
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	addr := common.BytesToAddress(a.Bytes())

	uint64Offset, overflow := codeOffset.Uint64WithOverflow()
	if overflow {
		uint64Offset = 0xffffffffffffffff
	}

	codeCopy := getData(evm.StateDB.GetCode(addr), uint64Offset, length.Uint64())
	memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)

	evm.interpreter.intPool.put(a, memOffset, codeOffset, length)
	return nil, nil
}

func opGasprice(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := evm.interpreter.intPool.get()
	x.SetFromBig(evm.GasPrice)
	stack.push(x)
	return nil, nil
}

func opBlockhash(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	num := stack.peek()
	num64, overflow := num.Uint64WithOverflow()
	if overflow {
		num.Clear()
		return nil, nil
	}
	var (
		upperBlocknum uint64
		lowerBlocknum uint64
	)
	upperBlocknum = evm.BlockNumber.Uint64()
	if upperBlocknum < 257 {
		lowerBlocknum = 0
	} else {
		lowerBlocknum = upperBlocknum - 256
	}
	if num64 >= lowerBlocknum && num64 < upperBlocknum {
		num.SetBytes(evm.GetHash(num64).Bytes())
	} else {
		num.Clear()
	}
	return nil, nil
}

func opCoinbase(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := evm.interpreter.intPool.get()
	x.SetBytes(evm.Coinbase.Bytes())
	stack.push(x)
	return nil, nil
}

func opTimestamp(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := evm.interpreter.intPool.get()
	x.SetFromBig(evm.Time)
	stack.push(x)
	return nil, nil
}

func opNumber(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := evm.interpreter.intPool.get()
	x.SetFromBig(evm.BlockNumber)
	stack.push(x)
	return nil, nil
}

func opDifficulty(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := evm.interpreter.intPool.get()
	x.SetFromBig(evm.Difficulty)
	stack.push(x)
	return nil, nil
}

func opGasLimit(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := evm.interpreter.intPool.get()
	x.SetUint64(evm.GasLimit)
	stack.push(x)
	return nil, nil
}

func opPop(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	evm.interpreter.intPool.put(stack.pop())
	return nil, nil
}

func opMload(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	x := stack.peek()
	x.SetBytes(memory.Get(x.Int64(), 32))
	return nil, nil
}

func opMstore(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// pop value of the stack
	mStart, val := stack.pop(), stack.pop()
	memory.Set(mStart.Uint64(), 32, val.PaddedBytes(32))

	evm.interpreter.intPool.put(mStart, val)
	return nil, nil
}

func opMstore8(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	off, val := stack.pop(), stack.pop()
	memory.store[off.Int64()] = byte(val.Int64() & 0xff)

	evm.interpreter.intPool.put(off, val)
	return nil, nil
}

func opSload(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	loc := stack.peek()
	hash := common.BytesToHash(loc.Bytes())
	val := evm.StateDB.GetState(contract.Address(), hash).Bytes()
	loc.SetBytes(val)
	return nil, nil
}

func opSstore(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	loc, val := stack.pop(), stack.pop()
	hash := common.BytesToHash(loc.Bytes())
	evm.StateDB.SetState(contract.Address(), hash, common.BytesToHash(val.Bytes()))

	evm.interpreter.intPool.put(loc, val)
	return nil, nil
}

func opJump(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	pos := stack.pop()
	if !contract.jumpdests.has(contract.CodeHash, contract.Code, pos) {
		nop := contract.GetOp(pos.Uint64())
		return nil, fmt.Errorf("invalid jump destination (%v) %v", nop, pos)
	}
	*pc = pos.Uint64()

	evm.interpreter.intPool.put(pos)
	return nil, nil
}

func opJumpi(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	pos, cond := stack.pop(), stack.pop()
	if cond.Sign() != 0 {
		if !contract.jumpdests.has(contract.CodeHash, contract.Code, pos) {
			nop := contract.GetOp(pos.Uint64())
			return nil, fmt.Errorf("invalid jump destination (%v) %v", nop, pos)
		}
		*pc = pos.Uint64()
	} else {
		*pc++
	}

	evm.interpreter.intPool.put(pos, cond)
	return nil, nil
}

func opJumpdest(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	return nil, nil
}

func opPc(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.interpreter.intPool.get().SetUint64(*pc))
	return nil, nil
}

func opMsize(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.interpreter.intPool.get().SetUint64(uint64(memory.Len())))
	return nil, nil
}

func opGas(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	stack.push(evm.interpreter.intPool.get().SetUint64(contract.Gas))
	return nil, nil
}

func opCreate(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	var (
		value        = stack.pop()
		offset, size = stack.pop(), stack.pop()
		input        = memory.Get(offset.Int64(), size.Int64())
		gas          = contract.Gas
	)
	if evm.ChainConfig().IsEIP150(evm.BlockNumber) {
		gas -= gas / 64
	}

	stackvalue := evm.interpreter.intPool.get()

	contract.UseGas(gas)
	res, addr, returnGas, suberr := evm.Create(contract, input, gas, value.ToBig())
	// Push item on the stack based on the returned error. If the ruleset is
	// homestead we must check for CodeStoreOutOfGasError (homestead only
	// rule) and treat as an error, if the ruleset is frontier we must
	// ignore this error and pretend the operation was successful.
	if evm.ChainConfig().IsHomestead(evm.BlockNumber) && suberr == ErrCodeStoreOutOfGas {
		stackvalue.Clear()
	} else if suberr != nil && suberr != ErrCodeStoreOutOfGas {
		stackvalue.Clear()
	} else {
		stackvalue.SetBytes(addr.Bytes())
	}
	stack.push(stackvalue)
	contract.Gas += returnGas
	evm.interpreter.intPool.put(value, offset, size)

	if suberr == errExecutionReverted {
		return res, nil
	}
	return nil, nil
}

func opCall(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas in in evm.callGasTemp. Use as temp
	temporary := stack.pop()
	gas := evm.callGasTemp
	// Pop other call parameters.
	addr, _value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.BytesToAddress(addr.Bytes())
	value := _value.ToBig()
	// Get the arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if value.Sign() != 0 {
		gas += params.CallStipend
	}
	ret, returnGas, err := evm.Call(contract, toAddr, args, gas, value)
	if err != nil {
		temporary.Clear()
	} else {
		temporary.SetOne()
	}
	stack.push(temporary)
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Gas += returnGas

	evm.interpreter.intPool.put(addr, _value, inOffset, inSize, retOffset, retSize)
	return ret, nil
}

func opCallCode(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas in in evm.callGasTemp. Use as temp
	temporary := stack.pop()
	gas := evm.callGasTemp
	// Pop other call parameters.
	addr, _value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.BytesToAddress(addr.Bytes())
	value := _value.ToBig()
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	if value.Sign() != 0 {
		gas += params.CallStipend
	}
	ret, returnGas, err := evm.CallCode(contract, toAddr, args, gas, value)
	if err != nil {
		temporary.Clear()
	} else {
		temporary.SetOne()
	}
	stack.push(temporary)
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Gas += returnGas

	evm.interpreter.intPool.put(addr, _value, inOffset, inSize, retOffset, retSize)
	return ret, nil
}

func opDelegateCall(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas in in evm.callGasTemp. Use as temp
	temporary := stack.pop()
	gas := evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.BytesToAddress(addr.Bytes())
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	ret, returnGas, err := evm.DelegateCall(contract, toAddr, args, gas)
	if err != nil {
		temporary.Clear()
	} else {
		temporary.SetOne()
	}
	stack.push(temporary)
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Gas += returnGas

	evm.interpreter.intPool.put(addr, inOffset, inSize, retOffset, retSize)
	return ret, nil
}

func opStaticCall(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	// Pop gas. The actual gas in in evm.callGasTemp. Use as temp
	temporary := stack.pop()
	gas := evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := common.BytesToAddress(addr.Bytes())
	// Get arguments from the memory.
	args := memory.Get(inOffset.Int64(), inSize.Int64())

	ret, returnGas, err := evm.StaticCall(contract, toAddr, args, gas)
	if err != nil {
		temporary.Clear()
	} else {
		temporary.SetOne()
	}
	stack.push(temporary)
	if err == nil || err == errExecutionReverted {
		memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	contract.Gas += returnGas

	evm.interpreter.intPool.put(addr, inOffset, inSize, retOffset, retSize)
	return ret, nil
}

func opReturn(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.GetPtr(offset.Int64(), size.Int64())

	evm.interpreter.intPool.put(offset, size)
	return ret, nil
}

func opRevert(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	offset, size := stack.pop(), stack.pop()
	ret := memory.GetPtr(offset.Int64(), size.Int64())

	evm.interpreter.intPool.put(offset, size)
	return ret, nil
}

func opStop(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	return nil, nil
}

func opSuicide(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	balance := evm.StateDB.GetBalance(contract.Address())
	evm.StateDB.AddBalance(common.BytesToAddress(stack.pop().Bytes()), balance)

	evm.StateDB.Suicide(contract.Address())
	return nil, nil
}

// following functions are used by the instruction jump  table

// make log instruction function
func makeLog(size int) executionFunc {
	return func(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		topics := make([]common.Hash, size)
		mStart, mSize := stack.pop(), stack.pop()
		for i := 0; i < size; i++ {
			topics[i] = common.BytesToHash(stack.pop().Bytes())
		}

		d := memory.Get(mStart.Int64(), mSize.Int64())
		evm.StateDB.AddLog(&types.Log{
			Address: contract.Address(),
			Topics:  topics,
			Data:    d,
			// This is a non-consensus field, but assigned here because
			// core/state doesn't know the current block number.
			BlockNumber: evm.BlockNumber.Uint64(),
		})

		evm.interpreter.intPool.put(mStart, mSize)
		return nil, nil
	}
}

// make push instruction function
func makePush(size uint64, pushByteSize int) executionFunc {
	return func(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		codeLen := len(contract.Code)

		startMin := codeLen
		if int(*pc+1) < startMin {
			startMin = int(*pc + 1)
		}

		endMin := codeLen
		if startMin+pushByteSize < endMin {
			endMin = startMin + pushByteSize
		}

		integer := evm.interpreter.intPool.get()
		integer.SetBytes(common.RightPadBytes(contract.Code[startMin:endMin], pushByteSize))
		stack.push(integer)

		*pc += size
		return nil, nil
	}
}

// make push instruction function
func makeDup(size int64) executionFunc {
	return func(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		stack.dup(evm.interpreter.intPool, int(size))
		return nil, nil
	}
}

// make swap instruction function
func makeSwap(size int64) executionFunc {
	// switch n + 1 otherwise n would be swapped with n
	size += 1
	return func(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
		stack.swap(int(size))
		return nil, nil
	}
}
