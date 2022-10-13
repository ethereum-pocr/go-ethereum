package cliquepcr
import (
	"errors"
	"math/big"
	"sort"
)
// The standard WhitePaper computation
type PercentileRankRewardComputation struct {
}

func (wp *PercentileRankRewardComputation) GetAlgorithmId() (int) {
	return 1;
}

func (wp *PercentileRankRewardComputation) CalculateAcceptNewSealersReward(nbNodes *big.Int) (*big.Int, error) {
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
func (wp *PercentileRankRewardComputation) CalculateGlobalInflationControlFactor(M *big.Int) (*big.Rat, error){
	// L = M / (8 000 000 * 30 / 3) // as integer value
	// D = 2^L // The divisor : 2 at the power of L
	// GlobalInflationControl = 1/D // 1; 1/2; 1/4; 1/8 ....

	// If there is no crpto created, return 1
	if M.Cmp(zero) == 0 {
		return big.NewRat(1, 1), nil
	}
	C := big.NewInt(8_000_000 * 30 / 3)
	C = C.Mul(C, CTCUnit)
	L := new(big.Rat).SetFrac(M, C)
	L2 := new(big.Int).Quo(L.Num(), L.Denom()).Uint64()
	// D = 2^L
	D := int64(1) << L2
	// log.Info("Trace CalculateGlobalInflationControlFactor", "M", M, "L2", L2, "D", D)
	if D == 0 { // The divisor has reached such a large amount (2^63) than the shift gave 0, So Dividing by a very large number is equivalent to 0
		return big.NewRat(0, 1), nil
	}
	return big.NewRat(1, D), nil
}
func (wp *PercentileRankRewardComputation) CalculateCarbonFootprintRewardCollection(nodesFootprint []*big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {
	
	sort.Slice(nodesFootprint, func(i, j int) bool {
		return nodesFootprint[i].Cmp(nodesFootprint[j]) > 0
	})
	return nil,nil
	// sort.Big(nodesFootprint)
}
func (wp *PercentileRankRewardComputation) CalculateCarbonFootprintReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {
	rewardI:= big.NewInt(0)
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
	// reward = 1 CTC (10^18 Wei)
	reward := new(big.Rat).SetInt(CTCUnit)
	// reward = ratio * CTC unit
	reward = reward.Mul(reward, ratio)
	// convert to big.Int
	if ratio.Sign() <= 0 {
		rewardI =  big.NewInt(0)
	} else
	{
		rewardI = new(big.Int).Quo(reward.Num(), reward.Denom())
	}
	// cap to 2 CTC units
	cap := big.NewInt(2)
	cap = cap.Mul(cap, CTCUnit)
	if rewardI.Cmp(cap) > 0 {
		rewardI = cap
	}
	infl, err := wp.CalculateGlobalInflationControlFactor(totalCryptoAmount)
	if err != nil {
		return nil, err
	}
	// Reward(n, b) = CarbonReduction(n) * N * GlobalInflationControl(b)
	rew := new(big.Rat).SetInt(rewardI)
	rew = rew.Mul(rew, new(big.Rat).SetInt(nbNodes))
	rew = rew.Mul(rew, infl)
	rewI := new(big.Int).Quo(rew.Num(), rew.Denom())
	return rewI, nil
}
