package cliquepcr

import (
	"math/big"
)

type IRewardComputation interface {
	// Methods
	GetAlgorithmId() int

	CalculateGlobalInflationControlFactor(M *big.Int) (float64, error)
	CalculateCarbonFootprintRewardCollection(nodesFootprint []*big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error)
}
