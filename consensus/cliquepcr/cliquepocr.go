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
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/clique"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	lru "github.com/hashicorp/golang-lru"
	// "github.com/ethereum/go-ethereum/cliquepcr/rewardcomputation"
)

// We put new constants to be able to override default clique values
const (
	checkpointInterval = 1024 // Number of blocks after which to save the vote snapshot to the database
	inmemorySnapshots  = 128  // Number of recent vote snapshots to keep in memory
	inmemorySignatures = 4096 // Number of recent block signatures to keep in memory

	wiggleTime  = 500 * time.Millisecond // Random delay (per signer) to allow concurrent signers
	extraVanity = 32
)

// Clique proof-of-authority protocol constants.
var (
	epochLength = uint64(30000) // Default number of blocks after which to checkpoint and reset the pending votes, that could be overrided from default Clique
)

// address of the PoCR smart contract, with the governance, the footprint, the auditors and the auditor's pledged amount
var proofOfCarbonReductionContractAddress = "0x0000000000000000000000000000000000000100"

// Use a separate address for collecting the total crypto generated because the smart contract also needs to hold auditor pledge
var sessionVariablesContractAddress = "0x0000000000000000000000000000000000000101"
var sessionVariableTotalPocRCoins = "GeneratedPocRTotal"
var zero = big.NewInt(0)
var CTCUnit = big.NewInt(1e+18)
var rewardComputation0 WPRewardComputation
var rewardComputation1 ZScoreRewardComputation
var rewardComputation2 PercentileRankRewardComputation
var chosenRewardAlgorithm = 0
var errUnknownBlock = errors.New("unknown block")

type CliquePoCR struct {
	config *params.CliqueConfig // Consensus engine configuration parameters
	db     ethdb.Database       // Database to store and retrieve snapshot checkpoints

	recents    *lru.ARCCache // Snapshots for recent block to speed up reorgs
	signatures *lru.ARCCache // Signatures of recent blocks to speed up mining

	proposals map[common.Address]bool // Current list of proposals we are pushing

	signer common.Address  // Ethereum address of the signing key
	signFn clique.SignerFn // Signer function to authorize hashes with
	lock   sync.RWMutex    // Protects the signer fields

	// The fields below are for testing only
	fakeDiff             bool // Skip difficulty verifications
	EngineInstance       *clique.Clique
	signersList          []common.Address
	signersListLastBlock uint64
}

type CarbonFootprintContract struct {
	ContractAddress common.Address
	RuntimeConfig   *runtime.Config
}

func New(config *params.CliqueConfig, db ethdb.Database) *CliquePoCR {
	conf := *config
	if conf.Epoch == 0 {
		conf.Epoch = epochLength
	}
	// Allocate the snapshot caches and create the engine
	recents, _ := lru.NewARC(inmemorySnapshots)
	signatures, _ := lru.NewARC(inmemorySignatures)
	return &CliquePoCR{
		config:         &conf,
		db:             db,
		recents:        recents,
		signatures:     signatures,
		proposals:      make(map[common.Address]bool),
		EngineInstance: clique.New(config, db)}
}

func SetSessionVariable(key string, value *big.Int, state *state.StateDB) {
	state.SetState(common.HexToAddress(sessionVariablesContractAddress), common.BytesToHash(crypto.Keccak256([]byte(key))), common.BigToHash(value))
}
func ReadSessionVariable(key string, state *state.StateDB) *big.Int {
	return state.GetState(common.HexToAddress(sessionVariablesContractAddress), common.BytesToHash(crypto.Keccak256([]byte(key)))).Big()
}

func (c *CliquePoCR) Author(header *types.Header) (common.Address, error) {
	return c.EngineInstance.Author(header)
}

// VerifyHeader checks whether a header conforms to the consensus rules of a
// given EngineInstance. Verifying the seal may be done optionally here, or explicitly
// via the VerifySeal method.
func (c *CliquePoCR) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header, seal bool) error {

	return c.EngineInstance.VerifyHeader(chain, header, seal)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications (the order is that of
// the input slice).

func (c *CliquePoCR) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	go func() {
		for i, header := range headers {
			c.buildSealersList(chain, header, headers[:i])
		}
	}()
	return c.EngineInstance.VerifyHeaders(chain, headers, seals)
}

func (c *CliquePoCR) buildSealersList(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	if header.Number == nil {
		return errUnknownBlock
	}
	number := header.Number.Uint64()
	if c.signersListLastBlock != number {
		c.signersList, _ = c.getSigners(chain, header, parents)
		c.signersListLastBlock = number
	}
	return nil
}

// VerifyUncles verifies that the given block's uncles conform to the consensus
// rules of a given EngineInstance.

func (c *CliquePoCR) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	return c.EngineInstance.VerifyUncles(chain, block)
}

// Prepare initializes the consensus fields of a block header according to the
// rules of a particular EngineInstance. The changes are executed inline.

func (c *CliquePoCR) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	return c.EngineInstance.Prepare(chain, header)
}

// Finalize runs any post-transaction state modifications (e.g. block rewards)
// but does not assemble the block.
//
// Note: The block header and state database might be updated to reflect any
// consensus rules that happen at finalization (e.g. block rewards).
func (c *CliquePoCR) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header) {
	accumulateRewards(c, chain.Config(), state, header, uncles)
	// Finalize
	c.EngineInstance.Finalize(chain, header, state, txs, uncles)
}

// FinalizeAndAssemble runs any post-transaction state modifications (e.g. block
// rewards) and assembles the final block.
//
// Note: The block header and state database might be updated to reflect any
// consensus rules that happen at finalization (e.g. block rewards).

func (c *CliquePoCR) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	accumulateRewards(c, chain.Config(), state, header, uncles)
	// Finalize block
	return c.EngineInstance.FinalizeAndAssemble(chain, header, state, txs, uncles, receipts)
}

// Seal generates a new sealing request for the given input block and pushes
// the result into the given channel.
//
// Note, the method returns immediately and will send the result async. More
// than one result may also be returned depending on the consensus algorithm.

func (c *CliquePoCR) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	return c.EngineInstance.Seal(chain, block, results, stop)
}

// SealHash returns the hash of a block prior to it being sealed.
func (c *CliquePoCR) SealHash(header *types.Header) common.Hash {
	return c.EngineInstance.SealHash(header)
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have.

func (c *CliquePoCR) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	return c.EngineInstance.CalcDifficulty(chain, time, parent)
}

// APIs returns the RPC APIs this consensus engine provides.
func (c *CliquePoCR) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	return c.EngineInstance.APIs(chain)
}

// Close terminates any background threads maintained by the consensus EngineInstance.
func (c *CliquePoCR) Close() error {
	return c.EngineInstance.Close()
}

// Authorize injects a private key into the consensus engine to mint new blocks
// with.

func (c *CliquePoCR) Authorize(signer common.Address, signFn clique.SignerFn) {
	c.EngineInstance.Authorize(signer, signFn)
}

// AccumulateRewards credits the coinbase of the given block with the mining
// reward. The total reward consists of the static block reward and rewards for
// included uncles. The coinbase of each uncle block is also rewarded.
func accumulateRewards(c *CliquePoCR, config *params.ChainConfig, state *state.StateDB, header *types.Header, uncles []*types.Header) {
	// log.Info("AccumulateRewards", "blockNumber", header.Number.String())
	// Select the correct block reward based on chain progression
	author, err := c.Author(header)
	if err != nil {
		log.Error("Fail getting the Author of the block")
		author = c.EngineInstance.Signer
	}

	blockReward, err := calcCarbonFootprintReward(c, author, config, state, header)
	// if it could not be calculated or if the calculation returned zero
	if err != nil || blockReward.Sign() == 0 {
		log.Info("No reward for signer", "node", author.String(), "error", err)
		return
	}
	// Accumulate the rewards for the miner and any included uncles
	// reward := new(big.Int).Set(blockReward)
	log.Info("Accumulate Reward", author.Hex(), blockReward)
	state.AddBalance(author, blockReward)

	// AddBalance to a non accessible account storage to just accrue the total amount of crypto created a
	// and use this as a control of the monetary creation policy
	addTotalCryptoBalance(state, blockReward)
}

func getTotalCryptoBalance(state *state.StateDB) *big.Int {
	return ReadSessionVariable(sessionVariableTotalPocRCoins, state)
}

func addTotalCryptoBalance(state *state.StateDB, reward *big.Int) *big.Int {
	// state.CreateAccount(common.HexToAddress(totalCryptoGeneratedAddress))
	currentTotal := ReadSessionVariable(sessionVariableTotalPocRCoins, state)
	newTotal := big.NewInt(0).Add(currentTotal, reward)
	SetSessionVariable(sessionVariableTotalPocRCoins, newTotal, state)
	log.Info("Increasing the total crypto", "from", currentTotal.String(), "to", newTotal.String())
	return newTotal
}

func calcCarbonFootprintReward(c *CliquePoCR, address common.Address, config *params.ChainConfig, state *state.StateDB, header *types.Header) (*big.Int, error) {
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

	// totalCrypto, err := contract.getBalance()
	totalCrypto := getTotalCryptoBalance(state)
	if err != nil {
		return nil, err
	}
	var a []*big.Int
	var reward *big.Int
	_ = reward
	var rewardError error
	_ = rewardError
	switch chosenRewardAlgorithm {
	case 0:
		reward, rewardError = rewardComputation0.CalculateCarbonFootprintRewardCollection(a, footprint, totalCrypto)
	case 1:

		for i, signerAddress := range c.signersList {
			a[i], _ = contract.footprint(signerAddress)
		}
		reward, rewardError = rewardComputation1.CalculateCarbonFootprintRewardCollection(a, footprint, totalCrypto)
	case 2:
		for i, signerAddress := range c.signersList {
			a[i], _ = contract.footprint(signerAddress)
		}
		reward, rewardError = rewardComputation2.CalculateCarbonFootprintRewardCollection(a, footprint, totalCrypto)
	default:
		reward, rewardError = rewardComputation0.CalculateCarbonFootprintRewardCollection(a, footprint, totalCrypto)
	}
	if rewardError != nil {
		return nil, rewardError
	}
	log.Info("Calculated reward based on footprint", "block", header.Number, "node", address.String(), "nb", nbNodes, "footprint", footprint, "reward", reward)
	return reward, nil
}

func (c *CliquePoCR) getSigners(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) ([]common.Address, error) {
	number := header.Number.Uint64()

	// Retrieve the snapshot needed to verify this header and cache it
	snap, err := c.EngineInstance.Snapshot(chain, number-1, header.ParentHash, parents)
	// If the block is a checkpoint block, verify the signer list
	if number%c.config.Epoch == 0 {
		signersArray := snap.GetSigners()
		return signersArray, err
	}
	return nil, errors.New("Invalid Epoch when getting Signers list")
}

// func (contract *CarbonFootprintContract) getBalance() (*big.Int, error) {
// 	return contract.RuntimeConfig.State.GetBalance(common.HexToAddress(totalCryptoGeneratedAddress)), nil
// }
