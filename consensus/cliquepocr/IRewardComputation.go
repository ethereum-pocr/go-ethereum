package cliquepocr

import (
	"math/big"
)

type IRewardComputation interface {
	// Methods
	GetAlgorithmId() int
	CalculateRanking(footprint *big.Int, nodesFootprint []*big.Int) (rank *big.Rat, nbNodes int, err error)
	// CalculateGlobalInflationControlFactor(M *big.Int) (*big.Rat, error)
	CalculateCarbonFootprintReward(rank *big.Rat, nbNodes int, totalCryptoAmount *big.Int) (*big.Int, error)
}
