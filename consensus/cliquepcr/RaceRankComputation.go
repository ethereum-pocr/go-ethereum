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

// On an annual basis, what is the minimimum amount of CRC tokens that have to be generated each year
var minRequiredInflation = 0.03

// Public function for auditing, but used internally only
func (wp *RaceRankComputation) CalculateGlobalInflationControlFactor(M *big.Int) (float64, error) {
	// L = M / (8 000 000 * 30 / 3) // as integer value
	// D = 2^L // The divisor : 2 at the power of L
	// GlobalInflationControl = 1/D // 1; 1/2; 1/4; 1/8 ....

	// If there is no crpto created, return 1
	if M.Cmp(zero) == 0 {
		return 1, nil
	}
	// L = TotalCRC / SimulationVariables().InflationDenominator
	// D = pow(SimulationVariables().alpha, L)
	// self.StandardWhitePaperComputation.CurrentGlobalInflation  = 1/D
	C := big.NewInt(100000)
	C = C.Mul(C, CTCUnit)
	L := new(big.Rat).SetFrac(M, C)

	L2 := new(big.Int).Quo(L.Num(), L.Denom()).Uint64()
	L3 := float64(L2)

	res := 1 / math.Pow(1.5, L3)
	return res, nil
}
func (wp *RaceRankComputation) CalculateCarbonFootprintRewardCollection(nodesFootprint []*big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {
	if footprint.Cmp(zero) <= 0 {
		return nil, errors.New("cannot proceed with zero or negative footprint")
	}
	if len(nodesFootprint) < 2 {
		return nil, errors.New("Not enough nodes carbon footprint to compute the reward")
	}
	sort.Slice(nodesFootprint, func(a, b int) bool {
		// sort direction high before low.
		return nodesFootprint[a].Cmp(nodesFootprint[b]) < 0
	})
	// NbItemsAbove is the number of items above the current footprint
	var NbItemsAbove int
	N := len(nodesFootprint)

	if N == 0 {
		return nil, errors.New("cannot average with zero node")
	}

	if footprint.Cmp(zero) <= 0 {
		return nil, errors.New("cannot proceed with zero or negative footprint")
	}

	for i := 0; i < N; i++ {
		if nodesFootprint[i].Cmp(footprint) == -1 {
			NbItemsAbove++
		}
		/*
			if i == 0 {
				if nodesFootprint[i].Cmp(footprint) == -1 {
					NbItemsAbove++
				}
			} else if nodesFootprint[i].Cmp(nodesFootprint[i-1]) != 0 {
				if nodesFootprint[i].Cmp(footprint) == -1 {
					NbItemsAbove++
				}
			}
		*/
	}
	reward := math.Pow(0.9, float64(NbItemsAbove))

	globalInflationFactor, errorGIF := wp.CalculateGlobalInflationControlFactor(totalCryptoAmount)
	if errorGIF != nil {
		return nil, errorGIF
	}
	a := new(big.Float).Mul(big.NewFloat(reward), big.NewFloat(globalInflationFactor))
	b := new(big.Float).Mul(a, big.NewFloat(float64(CTCUnit.Uint64())))
	rewardCRCUnit := new(big.Float).Mul(b, big.NewFloat(float64(N)))

	// minReward := float64(totalCryptoAmount.Int64()) * minRequiredInflation * 365 * 24 * 3600 / float64(4)
	// minRewardCRCUnit := new(big.Float).Mul(big.NewFloat(minReward), big.NewFloat(float64(CTCUnit.Uint64())))

	// Ignore at this step
	// if rewardCRCUnit.Cmp(minRewardCRCUnit) < 0 {
	//	rewardCRCUnit = minRewardCRCUnit
	// }

	// minRewardCRCUnit.Int()

	u, _ := rewardCRCUnit.Int(nil)

	return u, nil
}
