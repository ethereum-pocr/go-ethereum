package cliquepcr

import (
	"errors"
	"math"
	"math/big"
	"sort"
)

// The standard WhitePaper computation
type RaceRankComputation struct {
}

func (wp *RaceRankComputation) GetAlgorithmId() int {
	return 3
}

func (wp *RaceRankComputation) CalculateAcceptNewSealersReward(nbNodes *big.Int) (*big.Int, error) {
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

// Public function for auditing, but used internally only
func (wp *RaceRankComputation) CalculateGlobalInflationControlFactor(M *big.Int) (*big.Rat, error) {
	// L = M / (8 000 000 * 30 / 3) // as integer value
	// D = 2^L // The divisor : 2 at the power of L
	// GlobalInflationControl = 1/D // 1; 1/2; 1/4; 1/8 ....

	// If there is no crpto created, return 1
	if M.Cmp(zero) == 0 {
		return big.NewRat(1, 1), nil
	}
	// L = TotalCRC / SimulationVariables().InflationDenominator
	// D = pow(SimulationVariables().alpha, L)
	// self.StandardWhitePaperComputation.CurrentGlobalInflation  = 1/D
	C := big.NewInt(100000)
	// C = C.Mul(C, CTCUnit)
	L := new(big.Rat).SetFrac(M, C)

	L2 := new(big.Int).Quo(L.Num(), L.Denom()).Uint64()

	res := math.Pow(1.5, float64(L2))
	return big.NewRat(1, int64(res)), nil
}
func (wp *RaceRankComputation) CalculateCarbonFootprintRewardCollection(nodesFootprint []*big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Float, error) {
	if footprint.Cmp(zero) <= 0 {
		return nil, errors.New("cannot proceed with zero or negative footprint")
	}
	sort.Slice(nodesFootprint, func(a, b int) bool {
		// sort direction high before low.
		return nodesFootprint[a].Cmp(nodesFootprint[b]) < 0
	})
	// NbItemsAbove is the number of items above the current footprint
	var NbItemsAbove int
	N := len(nodesFootprint)
	for i := 0; i < N; i++ {
		if i == 0 {
			if nodesFootprint[i].Cmp(footprint) == -1 {
				NbItemsAbove++
			}
		} else if nodesFootprint[i].Cmp(nodesFootprint[i-1]) != 0 {
			if nodesFootprint[i].Cmp(footprint) == -1 {
				NbItemsAbove++
			}
		}
	}

	reward := math.Pow(0.9, float64(NbItemsAbove))
	result := big.NewFloat(reward)

	return result, nil

}
func (wp *RaceRankComputation) CalculateCarbonFootprintReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {
	panic("CalculateCarbonFootprintRewardCollection not implemented")
}
