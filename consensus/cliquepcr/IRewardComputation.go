package cliquepcr
import (
	"math/big"
)
type IRewardComputation interface{
	// Methods
	CalculateAcceptNewSealersReward(nbNodes *big.Int) (*big.Int, error) 
	CalculateGlobalInflationControlFactor(M *big.Int) (*big.Rat, error)
	// CalculatePoCRReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error)
	// CalcCarbonFootprintReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int, address common.Address, config *params.ChainConfig, state *state.StateDB, header *types.Header) (*big.Int, error)
	CalculateCarbonFootprintReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) 
	}