package cliquepcr
import (
	"math/big"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)
type IRewardComputation interface{
	// Methods
	CalculateAcceptNewSealersReward(nbNodes *big.Int) (*big.Int, error) 
	CalculateGlobalInflationControlFactor(M *big.Int) (*big.Rat, error)
	CalculatePoCRReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error)
	CalcCarbonFootprintReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int, address common.Address, config *params.ChainConfig, state *state.StateDB, header *types.Header) (*big.Int, error)
	CalculateCarbonFootprintReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int) (*big.Int, error) 
	}