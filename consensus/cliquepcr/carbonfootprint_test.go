package cliquepcr

import (
	"fmt"
	"math/big"
	"testing"
)

type TestCase struct {
	nbNodes        int64
	totalFootprint int64
	footprint      int64
	calc           bool
	shouldFail     bool
	result         int64
}

func toInt(v int64) *int64 {
	x := v
	return &x
}

func TestCalcReward(t *testing.T) {
	testCases := make([]TestCase, 0, 100)
	testCases = append(testCases, TestCase{10, 200000, 1000, true, false, 0})
	testCases = append(testCases, TestCase{10, 200000, 0, true, true, 0})
	testCases = append(testCases, TestCase{10, 0, 1000, true, true, 0})
	testCases = append(testCases, TestCase{0, 20000, 1000, false, true, 0})
	testCases = append(testCases, TestCase{10, 200000, -1000, false, true, 0})
	testCases = append(testCases, TestCase{10, 200000, 10000000, true, false, 0})
	testCases = append(testCases, TestCase{10, 200000, 1, true, false, 0})
	testCases = append(testCases, TestCase{10, 1e+8, 1, true, false, 1})

	for index, test := range testCases {

		t.Run(fmt.Sprintf("Test case %v", index), func(t *testing.T) {
			nbNodes := big.NewInt(test.nbNodes)
			totalFootprint := big.NewInt(test.totalFootprint)
			footprint := big.NewInt(test.footprint)
			if test.calc {
				avg := float64(totalFootprint.Uint64() / nbNodes.Uint64())
				ratio := float64(float64(footprint.Uint64()) / avg)
				test.result = int64(1e+18/(ratio+0.2) - 5e+17)
				if test.result < 0 {
					test.result = 0
				}
				if test.result > 2e+18 {
					test.result = 2e+18
				}
			}

			reward, err := CalculateCarbonFootprintReward(nbNodes, totalFootprint, footprint)
			t.Logf("Testing calculation nb=%v total=%v footprint=%v  ==> (%v, %e)", nbNodes, totalFootprint, footprint, reward, err)
			if err != nil {
				if test.shouldFail {
					return
				}
				t.Errorf("Unexpected error %v", err)
			}
			if test.shouldFail {
				t.Errorf("The calculation should have failed")
			}
			if reward.Int64() != test.result {
				t.Errorf("Reward expected %v; but got %v", test.result, reward)
			}
		})

	}
}
