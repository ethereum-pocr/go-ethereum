package cliquepcr

import (
	"fmt"
	"math/big"
	"testing"
)

type TestCase struct {
	nbNodes    int64
	totalCRC   int64
	footprint  int64
	shouldFail bool
	result     *big.Int
}

func toInt(v int64) *int64 {
	x := v
	return &x
}

func TestCalcReward(t *testing.T) {
	testCases := make([]TestCase, 0, 100)
	n1 := new(big.Int)
	n1, ok := n1.SetString("0", 10)
	if !ok {
		fmt.Println("SetString: error")
		return
	}
	testCases = append(testCases, TestCase{10, 5e+8, 1000, false, n1})

	for index, test := range testCases {

		t.Run(fmt.Sprintf("Test case %v", index), func(t *testing.T) {
			nbNodes := big.NewInt(test.nbNodes)
			totalCRC := big.NewInt(test.totalCRC)
			footprint := big.NewInt(test.footprint)

			var rewardComputation RaceRankComputation
			cf := make([]*big.Int, nbNodes.Int64())

			for i := 0; i < int(nbNodes.Int64()); i++ {
				cf[i] = big.NewInt(int64(100000 + i*100000))
			}

			reward, err := rewardComputation.CalculateCarbonFootprintRewardCollection(cf, footprint, totalCRC)
			// reward, err := CalculateCarbonFootprintReward(nbNodes, totalFootprint, footprint)

			t.Logf("Testing calculation nb=%v total=%v footprint=%v  ==> (%v, %e)", nbNodes, totalCRC, footprint, reward, err)
			if err != nil {
				if test.shouldFail {
					return
				}
				t.Errorf("Unexpected error %v", err)
			}
			if test.shouldFail {
				t.Errorf("The calculation should have failed")
			}
			if reward.Cmp(test.result) != 0 {
				t.Errorf("Reward expected %v; but got %v", test.result, reward)
			}
		})

	}
}
