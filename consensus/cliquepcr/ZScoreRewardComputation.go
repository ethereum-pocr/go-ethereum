package cliquepcr
import (
	"errors"
	"math/big"
	"math"
)
// The standard WhitePaper computation
type ZScoreRewardComputation struct {
}

func (wp *ZScoreRewardComputation) GetAlgorithmId() (int) {
	return 2;
}

func (wp *ZScoreRewardComputation) CalculateAcceptNewSealersReward(nbNodes *big.Int) (*big.Int, error) {
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
func (wp *ZScoreRewardComputation) CalculateGlobalInflationControlFactor(M *big.Int) (*big.Rat, error) {
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


// There is no special convention for BigInt units for digits after comma
func (wp *ZScoreRewardComputation) CalculateCarbonFootprintRewardCollection(nodesFootprint []*big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {
	  if footprint.Cmp(zero) <= 0 {
	      return nil, errors.New("cannot proceed with zero or negative footprint")
	  }
	  // size of the array
	  n := len(nodesFootprint)
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
	  avg := new(big.Int).Div(sumCF, new(big.Int).SetUint64(uint64(n))) 
	  var diffSquared []*big.Int
	  var sumDiff *big.Int

	  for i := 0; i < n; i++ {
		var diffAvg = new(big.Int)
		diffAvg.Sub(nodesFootprint[i], avg)
		var diffAvgSquare,e = new(big.Int), big.NewInt(2)
		diffAvgSquare.Exp(diffAvg,e, nil)
		diffSquared[i] = diffAvgSquare
		sumDiff = sumDiff.Add(sumDiff, diffAvgSquare)
		_ = sumDiff
	  }


	  f_sumDiff := new(big.Float).SetInt(sumDiff)
	  f_strDev :=  new(big.Float).Quo(f_sumDiff, new(big.Float).SetInt(big.NewInt(int64(n-1))))

	  // f_strDev := math.Sqrt((f_sumDiff / new(big.Float).SetInt(n-1)))

	  f_footprint := new(big.Float).SetInt(footprint)
	  
	  f_zscore := new(big.Float).Quo(new(big.Float).Sub(f_footprint, new(big.Float).SetInt(avg)), f_strDev)
	  

	  f_zptile_func_inside := big.NewFloat(0).Sqrt(new(big.Float).Quo(f_zscore,new(big.Float).SetInt(big.NewInt(int64(2)))))

	  f_zptile_func_inside_float64, accuracy1 := f_zptile_func_inside.Float64()
	  _ = accuracy1

	  // Only at this step we need to conver

	  f_zptile := 0.5*(math.Erf(f_zptile_func_inside_float64)+1)

	  globalInflationFactor, errorGIF := wp.CalculateGlobalInflationControlFactor(totalCryptoAmount)
	  if (errorGIF != nil) { return nil, errorGIF }
	  gif_Float, isExact := globalInflationFactor.Float64()
	  _ = isExact
	  reward := (1-f_zptile)*float64(n)*gif_Float

	  result := big.NewInt(int64(math.Round(reward)))
	  return result,nil
	}
	   
func (wp *ZScoreRewardComputation) CalculateCarbonFootprintReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {
	panic("CalculateCarbonFootprintRewardCollection not implemented")
}
