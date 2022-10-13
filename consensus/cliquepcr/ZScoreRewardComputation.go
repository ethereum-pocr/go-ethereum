package cliquepcr
import (
	"math/big"
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
func (wp *ZScoreRewardComputation) CalculateGlobalInflationControlFactor(M *big.Int) (*big.Rat, error){
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
func (wp *ZScoreRewardComputation) CalculateCarbonFootprintRewardCollection(nodesFootprint []*big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {
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
	  _ = f_zscore

	  // f_zptile := new(big.Float).Quo(f_zscore,

	 /* 
	  zscore = (self.CFoot - mean) / standard_deviation
        # zptile is a function that turns a zscoore into a percentile 
        zptile = .5 * (math.erf(zscore / 2 ** .5) + 1)
        self.ZScoreRankComputation.CFRankRatio = zptile
        TotalCRC = sum(np.array(agentList.ZScoreRankComputation.AccumulatedMiningReward))
        self.ZScoreRankComputation.TotalCRCGenerated = TotalCRC
        L = TotalCRC / SimulationVariables().InflationDenominator
        D = pow(SimulationVariables().alpha, L)
        self.ZScoreRankComputation.CurrentGlobalInflation  = 1/D
	  */
	  // bigInt := &big.Int{}
	  // summDiffSquareRoot := bigInt.Sqrt(sumDiff)
	  
	  // var stdDevSquared = (new(big.Rat).SetFrac(sumDiff, new(big.Int).SetUint64(uint64(n-1))))
	  // bigInt := &big.Int{}
	  // sqrt := bigInt.Sqrt(stdDevSquared)
	  // _ = sqrt
	  
	  // var Str = `10000000000000000000000000000000000000000000000000000`
	

	  // value, _ := bigInt.SetString(Str, 10)
	  // sqrt := bigInt.Sqrt(value)
	  
	  // stdDevSquareRoot = bigInt.Sqrt(stdDevSquareRoot)
	  
	  // var stdDev,e = new(big.Int), big.NewInt(2)

	  // var stdDev = (sum_of_differences / (len(CFootArray) - 1)) ** 0.5
	  
	  // var diff[n]*big.Int
	  // for i := 0; i < n; i++ {
		// adding the values of
		// array to the variable sum
	//	diff[i] = (math.Pow((nodesFootprint[i]-avg), 2))
	//	sumdiff+=diff[i]
		return nil,nil
	}
	   
func (wp *ZScoreRewardComputation) CalculateCarbonFootprintReward(nbNodes *big.Int, totalFootprint *big.Int, footprint *big.Int, totalCryptoAmount *big.Int) (*big.Int, error) {
	panic("CalculateCarbonFootprintRewardCollection not implemented")
}
