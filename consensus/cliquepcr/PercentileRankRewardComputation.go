package cliquepcr
import (
	"errors"
	"math/big"
	"sort"
	"math"
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
/*
	Percentile Rank Formula:
	PR% = (L + ( 0.5 x S )) / N Where,
	L = Number of below rank,
	S = Number of same rank,
	N = Total numbers.
*/
func (wp *PercentileRankRewardComputation) CalculateCarbonFootprintRewardCollection(nodesFootprint []*big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {
	if footprint.Cmp(zero) <= 0 {
		return nil, errors.New("cannot proceed with zero or negative footprint")
	}
	sort.Slice(nodesFootprint, func(i, j int) bool {
		return nodesFootprint[i].Cmp(nodesFootprint[j]) > 0
	})
	var L int
	var S int
	N := len(nodesFootprint)
	for i := 0; i < N; i++ {
        if (nodesFootprint[i].Cmp(footprint)==-1) { 
			L++ 
		} else if (nodesFootprint[i].Cmp(footprint)==-0) { 
			S++ 
		}
    }
	var rank float64
	rank = (float64(L) + 0.5*float64(S))/float64(N)
	baseReward := ((2/(1+rank)))-1
	globalInflationFactor, errorGIF := wp.CalculateGlobalInflationControlFactor(totalCryptoAmount)
	if (errorGIF != nil) {
		return nil, errorGIF
	}
	gif_Float, isExact := globalInflationFactor.Float64()
	  _ = isExact
	reward := baseReward*float64(N)*gif_Float
	result := big.NewInt(int64(math.Round(reward)))
	return result,nil
}
func (wp *PercentileRankRewardComputation) CalculateCarbonFootprintReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {
	panic("CalculateCarbonFootprintReward not implemented")
}
