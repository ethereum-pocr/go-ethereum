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

func toBigInt(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 10)
	if ok {
		return n
	} else {
		return big.NewInt(0)
	}
}

func TestCalcReward(t *testing.T) {
	testCases := make([]TestCase, 0, 100)
	testCases = append(testCases, TestCase{10, 5000, 1000, false, toBigInt("9996400647922246998")})
	testCases = append(testCases, TestCase{10, 200000, 1000, false, toBigInt("9857031841274683325")})
	testCases = append(testCases, TestCase{10, 200000, 0, true, big.NewInt(int64(10000000000000000))})
	testCases = append(testCases, TestCase{10, 0, 1000, false, toBigInt("10000000000000000000")})
	// cannot have no node
	testCases = append(testCases, TestCase{0, 20000, 1000, true, big.NewInt(int64(0))})
	// cannot have one node
	testCases = append(testCases, TestCase{1, 20000, 1000, false, toBigInt("998561036302515158")})
	// cannot have negative footprint
	testCases = append(testCases, TestCase{10, 200000, -1000, true, big.NewInt(int64(0))})
	testCases = append(testCases, TestCase{10, -200000, 1000, true, big.NewInt(int64(0))})
	testCases = append(testCases, TestCase{10, 200000, 10000000, false, toBigInt("3436934486431687377")})
	testCases = append(testCases, TestCase{10, 200000, 1, false, toBigInt("9857031841274683325")})
	testCases = append(testCases, TestCase{10, 1e+8, 1, false, toBigInt("48007128098380047")})
	testCases = append(testCases, TestCase{5, 1e+6, 200000, false, toBigInt("4187389094740496748")})
	testCases = append(testCases, TestCase{3, 1e+7, 2000, false, toBigInt("1461557073530897394")})
	testCases = append(testCases, TestCase{5, 5e+6, 500, false, toBigInt("3488512022919468174")})
	for index, test := range testCases {

		t.Run(fmt.Sprintf("Test case %v", index), func(t *testing.T) {
			nbNodes := big.NewInt(test.nbNodes)
			totalCRC := big.NewInt(test.totalCRC)
			totalCRC = totalCRC.Mul(totalCRC, CTCUnit)
			footprint := big.NewInt(test.footprint)

			rewardComputation := NewRaceRankComputation()
			cf := make([]*big.Int, nbNodes.Int64())

			for i := 0; i < int(nbNodes.Int64()); i++ {
				cf[i] = big.NewInt(int64(100000 + i*100000))
			}
			rank, nodes, err := rewardComputation.CalculateRanking(footprint, cf)
			var reward *big.Int
			if err == nil {
				reward, err = rewardComputation.CalculateCarbonFootprintReward(rank, nodes, totalCRC)
			}

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
