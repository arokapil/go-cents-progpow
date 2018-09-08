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

package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
)

// VladVmTransition checks transaction processing without block context.
// See https://github.com/ethereum/EIPs/issues/176 for the test format specification.
type VladVmTransition struct {
	json vmTransiton
}

func (t *VladVmTransition) UnmarshalJSON(in []byte) error {
	return json.Unmarshal(in, &t.json)
}

type vmTransiton struct {
	Env stEnv              `json:"env"`
	Pre core.GenesisAlloc  `json:"pre"`
	Tx  types.Transactions `json:"transactions"`
}

type stPostState struct {
	Root common.UnprefixedHash `json:"hash"`
	Logs common.UnprefixedHash `json:"logs"`
}

//go:generate gencodec -type stEnv -field-override stEnvMarshaling -out gen_stenv.go

type stEnv struct {
	Coinbase   common.Address `json:"currentCoinbase"   gencodec:"required"`
	Difficulty *big.Int       `json:"currentDifficulty" gencodec:"required"`
	GasLimit   uint64         `json:"currentGasLimit"   gencodec:"required"`
	Number     uint64         `json:"currentNumber"     gencodec:"required"`
	Timestamp  uint64         `json:"currentTimestamp"  gencodec:"required"`
}

type stEnvMarshaling struct {
	Coinbase   common.UnprefixedAddress
	Difficulty *math.HexOrDecimal256
	GasLimit   math.HexOrDecimal64
	Number     math.HexOrDecimal64
	Timestamp  math.HexOrDecimal64
}

func blockHashGetter(n uint64) common.Hash {
	return common.BytesToHash(crypto.Keccak256([]byte(big.NewInt(int64(n)).String())))
}

// Run executes a specific subtest.
func (t *VladVmTransition) Run(vmconfig vm.Config) (*state.StateDB, []common.Hash, types.Receipts, error) {
	// Running on constantinople rules!
	config := &params.ChainConfig{
		ChainID:             big.NewInt(1),
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		DAOForkBlock:        big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
	}
	// Configure a signer with chainid 99
	signer := types.NewEIP155Signer(big.NewInt(99))

	block := t.genesis(config).ToBlock(nil)
	statedb := MakePreState(ethdb.NewMemDatabase(), t.json.Pre)
	gaspool := new(core.GasPool)
	gaspool.AddGas(block.GasLimit())
	var rejected []common.Hash
	gasUsed := uint64(0)
	var receipts types.Receipts
	for i, tx := range t.json.Tx {
		msg, err := tx.AsMessage(signer)
		if err != nil {
			fmt.Fprintf(os.Stderr, "rejected tx: 0x%x, could not recover sender: %v\n", tx.Hash(), err)
			rejected = append(rejected, tx.Hash())
			continue
		}
		statedb.Prepare(tx.Hash(), block.Hash(), i)
		context := core.NewEVMContext(msg, block.Header(), nil, &t.json.Env.Coinbase)
		context.GetHash = blockHashGetter
		evm := vm.NewEVM(context, statedb, config, vmconfig)
		snapshot := statedb.Snapshot()
		// (ret []byte, usedGas uint64, failed bool, err error)
		_, gas, failed, err := core.ApplyMessage(evm, msg, gaspool)
		if err != nil {
			statedb.RevertToSnapshot(snapshot)
			fmt.Fprintf(os.Stderr, "rejected tx: 0x%x from 0x%x: %v\n", tx.Hash(), msg.From(), err)
			rejected = append(rejected, tx.Hash())
		} else {
			gasUsed += gas
			// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
			// based on the eip phase, we're passing whether the root touch-delete accounts.
			var root []byte
			receipt := types.NewReceipt(root, failed, gasUsed)
			receipt.TxHash = tx.Hash()
			receipt.GasUsed = gas
			// if the transaction created a contract, store the creation address in the receipt.
			if msg.To() == nil {
				receipt.ContractAddress = crypto.CreateAddress(evm.Context.Origin, tx.Nonce())
			}
			// Set the receipt logs and create a bloom for filtering
			receipt.Logs = statedb.GetLogs(tx.Hash())
			receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
			receipts = append(receipts, receipt)
		}

	}
	// Commit block
	_, err := statedb.Commit(config.IsEIP158(block.Number()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not commit state: %v", err)
		return nil, nil, nil, err
	}
	return statedb, rejected, receipts, nil
}

func MakePreState(db ethdb.Database, accounts core.GenesisAlloc) *state.StateDB {
	sdb := state.NewDatabase(db)
	statedb, _ := state.New(common.Hash{}, sdb)
	for addr, a := range accounts {
		statedb.SetCode(addr, a.Code)
		statedb.SetNonce(addr, a.Nonce)
		statedb.SetBalance(addr, a.Balance)
		for k, v := range a.Storage {
			statedb.SetState(addr, k, v)
		}
	}
	// Commit and re-open to start with a clean state.
	root, _ := statedb.Commit(false)
	statedb, _ = state.New(root, sdb)
	return statedb
}

func (t *VladVmTransition) genesis(config *params.ChainConfig) *core.Genesis {
	return &core.Genesis{
		Config:     config,
		Coinbase:   t.json.Env.Coinbase,
		Difficulty: t.json.Env.Difficulty,
		GasLimit:   t.json.Env.GasLimit,
		Number:     t.json.Env.Number,
		Timestamp:  t.json.Env.Timestamp,
		Alloc:      t.json.Pre,
	}
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
