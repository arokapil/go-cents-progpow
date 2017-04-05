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

package state

import (
	"encoding/json"
	"fmt"
	"bytes"
//	"io/ioutil"
	"os"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/crypto"
)

type DumpAccount struct {
	Balance  string            `json:"balance"`
	Nonce    uint64            `json:"nonce"`
	Root     string            `json:"root"`
	CodeHash string            `json:"codeHash"`
	Code     string            `json:"code"`
	Storage  map[string]string `json:"storage"`
}

type Dump struct {
	Root     string                 `json:"root"`
	Accounts map[string]DumpAccount `json:"accounts"`
}

// Dumps the code from all accounts

func (self *StateDB) CodeDump() string {
	var emptyCodeHash = crypto.Keccak256(nil)
	it := self.trie.Iterator()
	var count = 0
	f, err := os.Create("/data/code_dump.dat")
	if err != nil{
			fmt.Printf("Aborted: %v", err)
			return "Error"
	}
	defer f.Close()

	for it.Next() {
		count ++;
		addr := self.trie.GetKey(it.Key)
		var data Account
		if err := rlp.DecodeBytes(it.Value, &data); err != nil {
			panic(err)
		}
		obj := newObject(nil, common.BytesToAddress(addr), data, nil)
		if(!bytes.Equal(data.CodeHash , emptyCodeHash)){		
			code := common.Bytes2Hex(obj.Code(self.db))
			//filename := fmt.Sprintf("/data/code_dump/%x.dat",common.BytesToAddress(it.Key))
			
			if _, err := f.WriteString( fmt.Sprintf("%x %v\n", common.BytesToAddress(it.Key),code) ); err != nil{
				fmt.Printf("Aborted: %v", err)
				return "Error"				
			}
			f.Sync()
			//ioutil.WriteFile(filename, []byte( code ), 0644)
			fmt.Printf("Wrote [@ %v]\n" , count )
			//fmt.Printf("Account    %x\n", common.BytesToAddress(addr))
			//fmt.Printf("Key    %x\n", common.BytesToAddress(it.Key))
			//fmt.Printf("Code: %v\n", code)
		}

	}
	return "All ok"
}

func (self *StateDB) RawDump() Dump {
	dump := Dump{
		Root:     common.Bytes2Hex(self.trie.Root()),
		Accounts: make(map[string]DumpAccount),
	}

	it := self.trie.Iterator()
	for it.Next() {
		addr := self.trie.GetKey(it.Key)
		var data Account
		if err := rlp.DecodeBytes(it.Value, &data); err != nil {
			panic(err)
		}

		obj := newObject(nil, common.BytesToAddress(addr), data, nil)
		account := DumpAccount{
			Balance:  data.Balance.String(),
			Nonce:    data.Nonce,
			Root:     common.Bytes2Hex(data.Root[:]),
			CodeHash: common.Bytes2Hex(data.CodeHash),
			Code:     common.Bytes2Hex(obj.Code(self.db)),
			Storage:  make(map[string]string),
		}
		storageIt := obj.getTrie(self.db).Iterator()
		for storageIt.Next() {
			account.Storage[common.Bytes2Hex(self.trie.GetKey(storageIt.Key))] = common.Bytes2Hex(storageIt.Value)
		}
		dump.Accounts[common.Bytes2Hex(addr)] = account
	}
	return dump
}

func (self *StateDB) Dump() []byte {
	json, err := json.MarshalIndent(self.RawDump(), "", "    ")
	if err != nil {
		fmt.Println("dump err", err)
	}

	return json
}
