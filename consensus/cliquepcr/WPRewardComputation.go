package cliquepcr
import (
	"math/big"
)
// The standard WhitePaper computation
type WPRewardComputation struct {
}

func (wp *WPRewardComputation) GetAlgorithmId() (int) {
	return 0;
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
func (wp *WPRewardComputation) CalculateGlobalInflationControlFactor(M *big.Int) (*big.Rat, error){
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
func (wp *WPRewardComputation) CalculateCarbonFootprintRewardCollection(nodesFootprint []*big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {
	panic("CalculateCarbonFootprintRewardCollection not implemented")
}
func (wp *WPRewardComputation) CalculateCarbonFootprintReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {
	panic("CalculateCarbonFootprintRewardCollection not implemented")
}
