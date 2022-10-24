// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package cliquepcr

import (
	"math/big"
	"testing"
)

func TestCarbonFootprintReward(t *testing.T) {
	var rewardComputation WPRewardComputation
	var reward, _ = rewardComputation.CalculateCarbonFootprintReward(big.NewInt(3), big.NewInt(600), big.NewInt(10), big.NewInt(8e+7))
	a := big.NewInt(2000000000000000000)
	if reward.Uint64() != a.Uint64() {
		t.Errorf("Expected %d got %d", a, reward)
	}
}

func TestCalculateAcceptNewSealersRewardWP(t *testing.T) {
	var rewardComputation WPRewardComputation
	var reward, _ = rewardComputation.CalculateAcceptNewSealersReward(big.NewInt(10))
	a := big.NewInt(3000000000000000000)
	if reward.Uint64() != a.Uint64() {
		t.Errorf("Expected %d got %d", a, reward)
	}
}

func TestCalculateAcceptNewSealersRewardZSCORE(t *testing.T) {
	var rewardComputation ZScoreRewardComputation
	var reward, _ = rewardComputation.CalculateAcceptNewSealersReward(big.NewInt(10))
	a := big.NewInt(3000000000000000000)
	if reward.Uint64() != a.Uint64() {
		t.Errorf("Expected %d got %d", a, reward)
	}
}

func TestCalculateAcceptNewSealersRewardPERCENTILE(t *testing.T) {
	var rewardComputation PercentileRankRewardComputation
	var reward, _ = rewardComputation.CalculateAcceptNewSealersReward(big.NewInt(10))
	a := big.NewInt(3000000000000000000)
	if reward.Uint64() != a.Uint64() {
		t.Errorf("Expected %d got %d", a, reward)
	}
}
func TestCalculateGlobalInflationControlFactor(t *testing.T) {
	var rewardComputation WPRewardComputation
	var factor, _ = rewardComputation.CalculateGlobalInflationControlFactor(big.NewInt(1000000))
	a := big.NewRat(1, 0)
	if factor.Cmp(a) != 0 {
		x, _ := a.Float64()
		y, _ := factor.Float64()
		t.Errorf("Expected %20.6f\n got %20.6f\n", x, y)
	}
}
