package cliquepcr

import (
	"errors"
	"math"
	"math/big"
)

// The standard WhitePaper computation
type WPRewardComputation struct {
}

func (wp *WPRewardComputation) GetAlgorithmId() int {
	return 0
}

func (wp *WPRewardComputation) CalculateAcceptNewSealersReward(nbNodes *big.Int) (*big.Int, error) {
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
func (wp *WPRewardComputation) CalculateGlobalInflationControlFactor(M *big.Int) (*big.Rat, error) {
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
func (wp *WPRewardComputation) CalculateCarbonFootprintRewardCollection(nodesFootprint []*big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {
	if footprint.Cmp(zero) <= 0 {
		return nil, errors.New("cannot proceed with zero or negative footprint")
	}
	// size of the array
	n := len(nodesFootprint)
	n_big := big.NewInt(int64(n))
	// declaring a variable
	// to store the sum
	var sumCF = big.NewInt(0)
	// traversing through the
	// array using for loop
	for i := 0; i < n; i++ {
		sumCF = sumCF.Add(sumCF, nodesFootprint[i])
	}
	if sumCF.Cmp(zero) <= 0 {
		return nil, errors.New("cannot proceed with zero or negative total footprint")
	}
	reward, errorReward := wp.CalculateCarbonFootprintReward(n_big, sumCF, footprint, totalCryptoAmount)
	return reward, errorReward

}
func (wp *WPRewardComputation) CalculateCarbonFootprintReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {
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
