// Copyright 2017 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"
)

var stateTestCommand = cli.Command{
	Action:    stateTestCmd,
	Name:      "apply",
	Usage:     "executes the given state transitin",
	ArgsUsage: "<file>",
}

// ExecutionResult contains the execution status after running a state test, any
// error that might have occurred and a dump of the final state if requested.
type ExecutionResult struct {
	Error    string         `json:"error,omitempty"`
	State    *state.Dump    `json:"state,omitempty"`
	Receipts types.Receipts `json:"receipts,omitempty"`
	Rejected []common.Hash  `json:"rejected,omitempty"`
}

func stateTestCmd(ctx *cli.Context) error {
	if len(ctx.Args().First()) == 0 {
		return errors.New("path-to-test argument required")
	}
	// Configure the go-ethereum logger
	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.Lvl(ctx.GlobalInt(VerbosityFlag.Name)))
	log.Root().SetHandler(glogger)

	// Configure the EVM logger
	config := &vm.LogConfig{
		DisableMemory: ctx.GlobalBool(DisableMemoryFlag.Name),
		DisableStack:  ctx.GlobalBool(DisableStackFlag.Name),
	}
	var (
		tracer   vm.Tracer
		debugger *vm.StructLogger
	)
	switch {
	case ctx.GlobalBool(MachineFlag.Name):
		tracer = NewJSONLogger(config, os.Stderr)

	case ctx.GlobalBool(DebugFlag.Name):
		debugger = vm.NewStructLogger(config)
		tracer = debugger

	default:
		debugger = vm.NewStructLogger(config)
	}
	// Load the test content from the input file
	src, err := ioutil.ReadFile(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("Failed reading file: %v", err)
	}
	var test VladVmTransition
	if err = json.Unmarshal(src, &test); err != nil {
		return fmt.Errorf("Failed unmarshaling file: %v", err)
	}
	// Iterate over all the tests, run them and aggregate the results
	cfg := vm.Config{
		Tracer: tracer,
		Debug:  ctx.GlobalBool(DebugFlag.Name) || ctx.GlobalBool(MachineFlag.Name),
	}
	var result ExecutionResult
	// Run the test and aggregate the result
	state, rejected, receipts, err := test.Run(cfg)
	if err != nil {
		result = ExecutionResult{
			Error: fmt.Sprintf("Error: %v", err),
		}
	} else {
		dump := state.RawDump()
		result = ExecutionResult{
			Receipts: receipts,
			Rejected: rejected,
			State:    &dump,
		}
	}
	// print state root for evmlab tracing (already committed above, so no need to delete objects again
	if ctx.GlobalBool(MachineFlag.Name) && state != nil {
		fmt.Fprintf(os.Stderr, "{\"stateRoot\": \"%x\"}\n", state.IntermediateRoot(false))
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
	return nil
}
