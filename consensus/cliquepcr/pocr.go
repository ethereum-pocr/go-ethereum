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
package cliquepcr

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	// "github.com/ethereum/go-ethereum/accounts/abi/bind"
	// "github.com/ethereum/go-ethereum/contracts/carbonfootprint"
)

// address of the PoCR smart contract, with the governance, the footprint, the auditors and the auditor's pledged amount
var proofOfCarbonReductionContractAddress = "0x0000000000000000000000000000000000000100"

// Use a separate address for collecting the total crypto generated because the smart contract also needs to hold auditor pledge
var totalCryptoGeneratedAddress = "0x0000000000000000000000000000000000000101"
var zero = big.NewInt(0)
var CTCUnit = big.NewInt(1e+18)

type PoCR struct {
	Clique
}

// New creates a PoCR proof-of-authority consensus engine with the initial
// It wrap the Clique consensus to only modify the reward
func NewPoCR(config *params.CliqueConfig, db ethdb.Database) *PoCR {

	cli := New(config, db)
	pocr := &PoCR{*cli}
	return pocr
}

// Finalize implements consensus.Engine, ensuring no uncles are set,
// Calculate and give the block reward according to the consensus.
func (c *PoCR) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header) {
	accumulateRewards(c, chain.Config(), state, header, uncles)
	// Finalize
	c.Clique.Finalize(chain, header, state, txs, uncles)
}

// FinalizeAndAssemble implements consensus.Engine, ensuring no uncles are set,
// Calculate and give the block reward according to the consensus.
func (c *PoCR) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	accumulateRewards(c, chain.Config(), state, header, uncles)
	// Finalize block
	return c.Clique.FinalizeAndAssemble(chain, header, state, txs, uncles, receipts)
}

// AccumulateRewards credits the coinbase of the given block with the mining
// reward. The total reward consists of the static block reward and rewards for
// included uncles. The coinbase of each uncle block is also rewarded.
func accumulateRewards(c *PoCR, config *params.ChainConfig, state *state.StateDB, header *types.Header, uncles []*types.Header) {
	// log.Info("AccumulateRewards", "blockNumber", header.Number.String())
	// Select the correct block reward based on chain progression
	author, err := c.Author(header)
	if err != nil {
		// log.Error("Fail getting the Author of the block")
		author = c.signer
	}

	blockReward, err := calcCarbonFootprintReward(author, config, state, header)
	// if it could not be calculated or if the calculation returned zero
	if err != nil || blockReward.Sign() == 0 {
		log.Info("No reward for signer", "node", author.String(), "error", err)
		return
	}
	// Accumulate the rewards for the miner and any included uncles
	// reward := new(big.Int).Set(blockReward)
	// log.Info("Accumulate Reward", author.Hex(), reward)
	state.AddBalance(author, blockReward)

	// TODO : AddBalance to a non accessible account to just accrue the total amount of crypto created a
	// and use this as a control of the monetary creation policy
	state.AddBalance(common.HexToAddress(totalCryptoGeneratedAddress), blockReward)
}

func calcCarbonFootprintReward(address common.Address, config *params.ChainConfig, state *state.StateDB, header *types.Header) (*big.Int, error) {
	// skip block 0
	if header.Number.Int64() <= 0 {
		return nil, errors.New("cannot support genesis block")
	}
	contract := NewCarbonFootPrintContract(address, config, state, header)
	nbNodes, err := contract.nbNodes()
	if err != nil {
		return nil, err
	}
	if nbNodes.Uint64() == 0 {
		return nil, errors.New("no node in PoCR smart contract")
	}
	totalFootprint, err := contract.totalFootprint()
	if err != nil {
		return nil, err
	}
	footprint, err := contract.footprint(address)
	if err != nil {
		return nil, err
	}
	if footprint.Uint64() == 0 {
		return nil, errors.New("no footprint for sealer")
	}

	totalCrypto, err := contract.getBalance()
	if err != nil {
		return nil, err
	}

	reward, err := CalculatePoCRReward(nbNodes, totalFootprint, footprint, totalCrypto)
	if err != nil {
		return nil, err
	}

	log.Info("Calculated reward based on footprint", "block", header.Number, "node", address.String(), "total", totalFootprint, "nb", nbNodes, "footprint", footprint, "reward", reward)
	return reward, nil
}

func CalculatePoCRReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {

	cf, err := CalculateCarbonFootprintReward(nbNodes, totalFootprint, footprint)
	if err != nil {
		return nil, err
	}

	// ns, err := CalculateAcceptNewSealersReward(nbNodes)
	// if err != nil {
	// 	return nil, err
	// }

	infl, err := CalculateGlobalInflationControlFactor(totalCryptoAmount)
	if err != nil {
		return nil, err
	}
	// Reward(n, b) = CarbonReduction(n) * N * GlobalInflationControl(b)
	rew := new(big.Rat).SetInt(cf)
	// rew = rew.Add(rew, new(big.Rat).SetInt(ns))
	rew = rew.Mul(rew, new(big.Rat).SetInt(nbNodes))
	rew = rew.Mul(rew, infl)

	rewI := new(big.Int).Quo(rew.Num(), rew.Denom())
	return rewI, nil
}

func CalculateCarbonFootprintReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int) (*big.Int, error) {
	if nbNodes.Cmp(zero) == 0 {
		return nil, errors.New("cannot average with zero node")
	}
	if totalFootprint.Cmp(zero) <= 0 {
		return nil, errors.New("cannot proceed with zero or negative total footprint")
	}
	if footprint.Cmp(zero) <= 0 {
		return nil, errors.New("cannot proceed with zero or negative footprint")
	}
	// average = totalFootprint / nbNodes
	average := new(big.Rat).SetFrac(totalFootprint, nbNodes)
	// ratio = nbNodes / totalFootprint
	ratio := new(big.Rat).Inv(average)
	// ratio = footprint * (nbNodes / totalFootprint) = X
	ratio = ratio.Mul(ratio, new(big.Rat).SetInt(footprint))
	// ratio = X + 0,2
	ratio = ratio.Add(ratio, big.NewRat(2, 10))
	// ratio = 1 / (X + 0,2)
	ratio = ratio.Inv(ratio)
	// ratio = 1 / (X + 0,2) - 0,5
	ratio = ratio.Sub(ratio, big.NewRat(5, 10))
	if ratio.Sign() <= 0 {
		return big.NewInt(0), nil
	}
	// reward = 1 CTC (10^18 Wei)
	reward := new(big.Rat).SetInt(CTCUnit)
	// reward = ratio * CTC unit
	reward = reward.Mul(reward, ratio)
	// convert to big.Int
	rewardI := new(big.Int).Quo(reward.Num(), reward.Denom())
	// cap to 2 CTC units
	cap := big.NewInt(2)
	cap = cap.Mul(cap, CTCUnit)
	if rewardI.Cmp(cap) > 0 {
		rewardI = cap
	}

	return rewardI, nil
}

func CalculateAcceptNewSealersReward(nbNodes *big.Int) (*big.Int, error) {
	// no additional reward when there is one node or less
	one := big.NewInt(1)
	if nbNodes.Cmp(one) <= 0 {
		return zero, nil
	}
	// N = nbNodes - 1
	N := new(big.Rat).SetInt(nbNodes)
	N = N.Sub(N, big.NewRat(1, 1))
	// reward = (N-1)/3
	rew := big.NewRat(1, 3)
	rew = rew.Mul(N, rew)
	// reward = (N-1)/3 * CTC Unit
	rew = rew.Mul(rew, new(big.Rat).SetInt(CTCUnit))
	// calculate the result rounding to the unit
	rewI := new(big.Int).Quo(rew.Num(), rew.Denom())
	return rewI, nil
}

// Implements the alternative as we have the total amount of crypto created available
func CalculateGlobalInflationControlFactor(M *big.Int) (*big.Rat, error) {
	// L = M / (8 000 000 * 30 / 3) // as integer value
	// D = 2^L // The divisor : 2 at the power of L
	// GlobalInflationControl = 1/D // 1; 1/2; 1/4; 1/8 ....

	// If there is no crpto created, return 1
	if M.Cmp(zero) == 0 {
		return big.NewRat(1, 1), nil
	}
	C := big.NewInt(8_000_000 * 30 / 3)
	C = C.Mul(C, CTCUnit)
	L := new(big.Rat).SetFrac(M, C)
	L2 := new(big.Int).Quo(L.Num(), L.Denom()).Uint64()
	// D = 2^L
	D := int64(1) << L2
	// log.Info("Trace CalculateGlobalInflationControlFactor", "M", M, "L2", L2, "D", D)
	if D == 0 { // The divisor has reached such a large amount (2^63) than the shift gave 0, So Dividing by a very large number is equivalent to 0
		return big.NewRat(0, 1), nil
	}
	return big.NewRat(1, D), nil
}

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

func (contract *CarbonFootprintContract) getBalance() (*big.Int, error) {
	return contract.RuntimeConfig.State.GetBalance(common.HexToAddress(totalCryptoGeneratedAddress)), nil
}

func (contract *CarbonFootprintContract) totalFootprint() (*big.Int, error) {
	input := common.Hex2Bytes("b6c3dcf8")
	result, _, err := runtime.Call(contract.ContractAddress, input, contract.RuntimeConfig)
	// log.Info("Result/Err", "Result", common.Bytes2Hex(result), "Err", err.Error())
	if err != nil {
		log.Error("Impossible to get the total carbon footprint", "err", err.Error(), "block", contract.RuntimeConfig.BlockNumber.Int64())
		return nil, err
	} else {
		// log.Info("Total Carbon footprint", "result", common.Bytes2Hex(result))
		return common.BytesToHash(result).Big(), nil
	}
}
func (contract *CarbonFootprintContract) nbNodes() (*big.Int, error) {
	input := common.Hex2Bytes("03b2ec98")
	result, _, err := runtime.Call(contract.ContractAddress, input, contract.RuntimeConfig)
	// log.Info("Result/Err", "Result", common.Bytes2Hex(result), "Err", err.Error())
	if err != nil {
		log.Error("Impossible to get the number of nodes in carbon footprint contract", "err", err.Error(), "block", contract.RuntimeConfig.BlockNumber.Int64())
		return nil, err
	} else {
		// log.Info("Carbon footprint nb nodes", "result", common.Bytes2Hex(result))
		return common.BytesToHash(result).Big(), nil
	}
}
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
		// log.Info("Carbon footprint node", "result", common.Bytes2Hex(result), "node", ofNode.String())
		return common.BytesToHash(result).Big(), nil
	}
}
