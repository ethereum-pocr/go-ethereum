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

// Package clique implements the proof-of-authority consensus engine.
package cliquepocr

import (
	// "errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/crypto"
)

// address of the PoCR smart contract, with the governance, the footprint, the auditors and the auditor's pledged amount
var (
	proofOfCarbonReductionContractAddress = "0x0000000000000000000000000000000000000100"
	slotNbNodes = uint(0)
	slotSealers = uint(1)
	slotIsSealer = uint(2)
)

type CarbonFootprintContract struct {
	ContractAddress common.Address
	RuntimeConfig   *runtime.Config
}

func NewCarbonFootPrintContract(nodeAddress common.Address, config *params.ChainConfig, state *state.StateDB, header *types.Header) CarbonFootprintContract {
	contract := CarbonFootprintContract{}
	contract.ContractAddress = common.HexToAddress(proofOfCarbonReductionContractAddress)
	block := big.NewInt(0).Sub(header.Number, big.NewInt(1))
	stateCopy := state.Copy() // necessary to work on the copy of the state when performing a call
	cfg := runtime.Config{ChainConfig: config, Origin: nodeAddress, GasLimit: 1000000, State: stateCopy, BlockNumber: block}
	contract.RuntimeConfig = &cfg
	return contract
}

func NewCarbonFootPrintContractForUpdate(nodeAddress common.Address, config *params.ChainConfig, state *state.StateDB, header *types.Header) CarbonFootprintContract {
	contract := CarbonFootprintContract{}
	contract.ContractAddress = common.HexToAddress(proofOfCarbonReductionContractAddress)
	cfg := runtime.Config{ChainConfig: config, Origin: nodeAddress, GasLimit: 1000000, State: state, BlockNumber: header.Number}
	contract.RuntimeConfig = &cfg
	return contract
}

// func (contract *CarbonFootprintContract) totalFootprint() (*big.Int, error) {
// 	input := common.Hex2Bytes("b6c3dcf8")
// 	result, _, err := runtime.Call(contract.ContractAddress, input, contract.RuntimeConfig)
// 	// log.Info("Result/Err", "Result", common.Bytes2Hex(result), "Err", err.Error())
// 	if err != nil {
// 		log.Error("Impossible to get the total carbon footprint", "err", err.Error(), "block", contract.RuntimeConfig.BlockNumber.Int64())
// 		return nil, err
// 	} else {
// 		// log.Info("Total Carbon footprint", "result", common.Bytes2Hex(result))
// 		return common.BytesToHash(result).Big(), nil
// 	}
// }

// func (contract *CarbonFootprintContract) nbNodes() (*big.Int, error) {
// 	input := common.Hex2Bytes("03b2ec98")
// 	result, _, err := runtime.Call(contract.ContractAddress, input, contract.RuntimeConfig)
// 	// log.Info("Result/Err", "Result", common.Bytes2Hex(result), "Err", err.Error())
// 	if err != nil {
// 		log.Error("Impossible to get the number of nodes in carbon footprint contract", "err", err.Error(), "block", contract.RuntimeConfig.BlockNumber.Int64())
// 		return nil, err
// 	} else {
// 		// log.Info("Carbon footprint nb nodes", "result", common.Bytes2Hex(result))
// 		return common.BytesToHash(result).Big(), nil
// 	}
// }
func (contract *CarbonFootprintContract) footprint(ofNode common.Address) (*big.Int, error) {
	addressString := ofNode.String()
	addressString = addressString[2:]

	input := common.Hex2Bytes("79f85816000000000000000000000000" + addressString)
	result, _, err := runtime.Call(contract.ContractAddress, input, contract.RuntimeConfig)

	// log.Info("Result/Err", "Result", common.Bytes2Hex(result), "Err", err.Error())
	if err != nil {
		log.Error("Impossible to get the carbon footprint", "err", err.Error(), "node", ofNode.String(), "block", contract.RuntimeConfig.BlockNumber.Int64())
		return nil, err
	} else {
		return common.BytesToHash(result).Big(), nil
	}
}

func mappingLocation(slot uint, key common.Hash) (common.Hash) { 
	_slot := common.BigToHash(big.NewInt(int64(slot))).Bytes()
	_key := key.Bytes()
	_location := common.BytesToHash(crypto.Keccak256(_key, _slot))
	return _location
}

func (contract *CarbonFootprintContract) getMapping(slot uint, key common.Hash) (common.Hash) {
	_location := mappingLocation(slot, key)
	v := contract.RuntimeConfig.State.GetState(contract.ContractAddress, _location)
	// log.Info("getMapping", "slot", slot, "key", key, "location", _location, "value", v)
	return v
} 
	
func (contract *CarbonFootprintContract) setMapping(slot uint, key common.Hash, value common.Hash) {
	_location := mappingLocation(slot, key)
	// log.Info("setMapping", "slot", slot, "key", key, "location", _location, "value", value)
	contract.RuntimeConfig.State.SetState(contract.ContractAddress, _location, value)
} 

func (contract *CarbonFootprintContract) getSealerAt(index int64) (common.Address) {
	v := contract.getMapping(slotSealers, common.BigToHash(big.NewInt(index)))
	return common.BytesToAddress(v.Big().Bytes())
} 

func (contract *CarbonFootprintContract) setSealerAt(index int64, sealer common.Address) {
	contract.setMapping(slotSealers, common.BigToHash(big.NewInt(index)), common.BytesToHash(sealer.Bytes()))
}

func (contract *CarbonFootprintContract) getIsSealerOf(sealer common.Address) (bool) {
	v := contract.getMapping(slotIsSealer, common.BytesToHash(sealer.Bytes()))
	return v.Big().Cmp(zero) != 0
} 

func (contract *CarbonFootprintContract) setIsSealerOf(sealer common.Address, isSealer bool) {
	_b := big.NewInt(0)
	if isSealer {_b = big.NewInt(1)}
	contract.setMapping(slotIsSealer, common.BytesToHash(sealer.Bytes()), common.BytesToHash(_b.Bytes()))
} 

func (contract *CarbonFootprintContract) getNbNodes() (uint64) {
	_slot := common.BigToHash(big.NewInt(int64(slotNbNodes)))
	v := contract.RuntimeConfig.State.GetState(contract.ContractAddress, _slot)
	return v.Big().Uint64()
} 

func (contract *CarbonFootprintContract) setNbNodes(nbNodes int64) {
	_slot := common.BigToHash(big.NewInt(int64(slotNbNodes)))
	v := big.NewInt(int64(nbNodes))
	contract.RuntimeConfig.State.SetState(contract.ContractAddress, _slot, common.BigToHash(v))
} 



// func (contract *CarbonFootprintContract) clearFootprint(ofNode common.Address) (*big.Int, error) {
// 	addressString := ofNode.String()
// 	addressString = addressString[2:]

// 	input := common.Hex2Bytes("79f85816000000000000000000000000" + addressString)
// 	result, _, err := runtime.Call(contract.ContractAddress, input, contract.RuntimeConfig)
// 	// log.Info("Result/Err", "Result", common.Bytes2Hex(result), "Err", err.Error())
// 	if err != nil {
// 		log.Error("Impossible to get the carbon footprint", "err", err.Error(), "node", ofNode.String(), "block", contract.RuntimeConfig.BlockNumber.Int64())
// 		return nil, err
// 	} else {
// 		// log.Info("Carbon footprint node", "result", common.Bytes2Hex(result), "node", ofNode.String())
// 		return common.BytesToHash(result).Big(), nil
// 	}
// }
