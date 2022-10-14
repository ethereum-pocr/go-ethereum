package cliquepcr
import (
	"math/big"
)

type IRewardComputation interface{
	// Methods
	GetAlgorithmId() int
	CalculateAcceptNewSealersReward(nbNodes *big.Int) (*big.Int, error) 
	CalculateGlobalInflationControlFactor(M *big.Int) (*big.Rat, error)
	CalculateCarbonFootprintRewardCollection(nodesFootprint []*big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) 
	CalculateCarbonFootprintReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) 
	}