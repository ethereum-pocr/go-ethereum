// Copyright 2019 The go-ethereum Authors
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

package cliquepcr

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// Clique proof-of-authority protocol constants.
var (
	extraSeal = crypto.SignatureLength // Fixed number of extra-data suffix bytes reserved for signer seal

	nonceAuthVote = hexutil.MustDecode("0xffffffffffffffff") // Magic nonce number to vote on adding a new signer
	nonceDropVote = hexutil.MustDecode("0x0000000000000000") // Magic nonce number to vote on removing a signer.

	uncleHash = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.

	diffInTurn = big.NewInt(2) // Block difficulty for in-turn signatures
	diffNoTurn = big.NewInt(1) // Block difficulty for out-of-turn signatures
)

// This test case is a repro of an annoying bug that took us forever to catch.
// In Clique PoA networks (Rinkeby, Görli, etc), consecutive blocks might have
// the same state root (no block subsidy, empty block). If a node crashes, the
// chain ends up losing the recent state and needs to regenerate it from blocks
// already in the database. The bug was that processing the block *prior* to an
// empty one **also completes** the empty one, ending up in a known-block error.
func TestReimportMirroredState(t *testing.T) {
	// Initialize a Clique chain with a single signer
	var (
		db     = rawdb.NewMemoryDatabase()
		key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		addr   = crypto.PubkeyToAddress(key.PublicKey)
		pocrAddr = common.HexToAddress("0x0000000000000000000000000000000000000100")
		pocr2Addr = common.HexToAddress("0x0000000000000000000000000000000000000101")
		engine = New(params.AllCliqueProtocolChanges.Clique, db)
		signer = new(types.HomesteadSigner)
	)
	genspec := &core.Genesis{
		ExtraData: make([]byte, extraVanity+common.AddressLength+extraSeal),
		Alloc: map[common.Address]core.GenesisAccount{
			addr: {Balance: big.NewInt(10000000000000000)},
			pocrAddr: {
				Balance: big.NewInt(0),
				Code: common.Hex2Bytes("608060405234801561001057600080fd5b506004361061004c5760003560e01c806303b2ec981461005157806346c556cc1461006f57806379f858161461008b578063b6c3dcf8146100bb575b600080fd5b6100596100d9565b60405161006691906103bb565b60405180910390f35b61008960048036038101906100849190610465565b6100df565b005b6100a560048036038101906100a091906104a5565b610384565b6040516100b291906103bb565b60405180910390f35b6100c361039c565b6040516100d091906103bb565b60405180910390f35b60015481565b8173ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141561014e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161014590610555565b60405180910390fd5b60008060008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205490506000811480156101a15750600082115b1561021c57816000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000208190555060018060008282546101fb91906105a4565b92505081905550816002600082825461021491906105a4565b925050819055505b60008111801561022c5750600082115b156102a757816000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550806002600082825461028691906105fa565b92505081905550816002600082825461029f91906105a4565b925050819055505b6000811180156102b75750600082145b156103315780600260008282546102ce91906105fa565b9250508190555060018060008282546102e791906105fa565b925050819055506000808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600090555b8273ffffffffffffffffffffffffffffffffffffffff167f4a4e48dea93c3be0cee4d331f816573cf58d3dabc07503caa878df70b04d84548360405161037791906103bb565b60405180910390a2505050565b60006020528060005260406000206000915090505481565b60025481565b6000819050919050565b6103b5816103a2565b82525050565b60006020820190506103d060008301846103ac565b92915050565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610406826103db565b9050919050565b610416816103fb565b811461042157600080fd5b50565b6000813590506104338161040d565b92915050565b610442816103a2565b811461044d57600080fd5b50565b60008135905061045f81610439565b92915050565b6000806040838503121561047c5761047b6103d6565b5b600061048a85828601610424565b925050602061049b85828601610450565b9150509250929050565b6000602082840312156104bb576104ba6103d6565b5b60006104c984828501610424565b91505092915050565b600082825260208201905092915050565b7f7468652061756469746f722063616e6e6f742073657420697473206f776e206660008201527f6f6f747072696e74000000000000000000000000000000000000000000000000602082015250565b600061053f6028836104d2565b915061054a826104e3565b604082019050919050565b6000602082019050818103600083015261056e81610532565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60006105af826103a2565b91506105ba836103a2565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156105ef576105ee610575565b5b828201905092915050565b6000610605826103a2565b9150610610836103a2565b92508282101561062357610622610575565b5b82820390509291505056fea26469706673582212202fe6ed98ca2ba2ebca013eb9bba9d0ed88a143db753f24e9934cccb8695f39ce64736f6c634300080c0033"),
			},
			pocr2Addr: {
				Balance: big.NewInt(0),
				Code: common.Hex2Bytes("608060"),
			},
		},
		BaseFee: big.NewInt(params.InitialBaseFee),
	}
	copy(genspec.ExtraData[extraVanity:], addr[:])
	genesis := genspec.MustCommit(db)

	// Generate a batch of blocks, each properly signed
	chain, _ := core.NewBlockChain(db, nil, params.AllCliqueProtocolChanges, engine, vm.Config{}, nil, nil)
	defer chain.Stop()

	blocks, _ := core.GenerateChain(params.AllCliqueProtocolChanges, genesis, engine, db, 3, func(i int, block *core.BlockGen) {
		// The chain maker doesn't have access to a chain, so the difficulty will be
		// lets unset (nil). Set it here to the correct value.
		block.SetDifficulty(diffInTurn)

		// We want to simulate an empty middle block, having the same state as the
		// first one. The last is needs a state change again to force a reorg.
		if i != 1 {
			tx, err := types.SignTx(types.NewTransaction(block.TxNonce(addr), common.Address{0x00}, new(big.Int), params.TxGas, block.BaseFee(), nil), signer, key)
			if err != nil {
				panic(err)
			}
			block.AddTxWithChain(chain, tx)
		}
	})
	for i, block := range blocks {
		header := block.Header()
		if i > 0 {
			header.ParentHash = blocks[i-1].Hash()
		}
		header.Extra = make([]byte, extraVanity+extraSeal)
		header.Difficulty = diffInTurn

		sig, _ := crypto.Sign(engine.SealHash(header).Bytes(), key)
		copy(header.Extra[len(header.Extra)-extraSeal:], sig)
		blocks[i] = block.WithSeal(header)
	}
	// Insert the first two blocks and make sure the chain is valid
	db = rawdb.NewMemoryDatabase()
	genspec.MustCommit(db)

	chain, _ = core.NewBlockChain(db, nil, params.AllCliqueProtocolChanges, engine, vm.Config{}, nil, nil)
	defer chain.Stop()

	if _, err := chain.InsertChain(blocks[:2]); err != nil {
		t.Fatalf("failed to insert initial blocks: %v", err)
	}
	if head := chain.CurrentBlock().NumberU64(); head != 2 {
		t.Fatalf("chain head mismatch: have %d, want %d", head, 2)
	}

	// Simulate a crash by creating a new chain on top of the database, without
	// flushing the dirty states out. Insert the last block, triggering a sidechain
	// reimport.
	chain, _ = core.NewBlockChain(db, nil, params.AllCliqueProtocolChanges, engine, vm.Config{}, nil, nil)
	defer chain.Stop()

	if _, err := chain.InsertChain(blocks[2:]); err != nil {
		t.Fatalf("failed to insert final block: %v", err)
	}
	if head := chain.CurrentBlock().NumberU64(); head != 3 {
		t.Fatalf("chain head mismatch: have %d, want %d", head, 3)
	}
}
