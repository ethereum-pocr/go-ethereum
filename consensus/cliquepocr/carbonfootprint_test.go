package cliquepocr

import (
	"fmt"
	"math/big"
	"testing"
)

type TestCase struct {
	nbNodes int64
	// total CRC expressed in CRC, not CRCUnit
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
	n1, ok := n1.SetString("10000000000000000000", 10)
	if !ok {
		fmt.Println("SetString: error")
		return
	}
	testCases = append(testCases, TestCase{10, 5000, 1000, false, n1})
	testCases = append(testCases, TestCase{10, 200000, 1000, false, big.NewInt(int64(4444444444444444160))})
	testCases = append(testCases, TestCase{10, 200000, 0, true, big.NewInt(int64(10000000000000000))})
	testCases = append(testCases, TestCase{10, 0, 1000, false, n1})
	// cannot have no node
	testCases = append(testCases, TestCase{0, 20000, 1000, true, big.NewInt(int64(0))})
	// cannot have one node
	testCases = append(testCases, TestCase{1, 20000, 1000, true, big.NewInt(int64(0))})
	// cannot have negative footprint
	testCases = append(testCases, TestCase{10, 200000, -1000, true, big.NewInt(int64(0))})
	testCases = append(testCases, TestCase{10, 200000, 10000000, false, big.NewInt(int64(1549681956000000512))})
	testCases = append(testCases, TestCase{10, 200000, 1, false, big.NewInt(int64(4444444444444444160))})
	testCases = append(testCases, TestCase{10, 1e+8, 1, false, big.NewInt(int64(0))})
	testCases = append(testCases, TestCase{5, 1e+6, 200000, false, big.NewInt(int64(78036884621246752))})
	testCases = append(testCases, TestCase{3, 1e+7, 2000, false, big.NewInt(int64(7))})
	testCases = append(testCases, TestCase{5, 5e+6, 500, false, big.NewInt(int64(7841642727))})
	for index, test := range testCases {

		t.Run(fmt.Sprintf("Test case %v", index), func(t *testing.T) {
			nbNodes := big.NewInt(test.nbNodes)
			totalCRC := big.NewInt(test.totalCRC)
			totalCRC = totalCRC.Mul(totalCRC, CTCUnit)
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
