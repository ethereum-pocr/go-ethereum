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
	"math"
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
func TestCalculateGlobalInflationControlFactor1(t *testing.T) {
	var rewardComputation WPRewardComputation
	var factor, _ = rewardComputation.CalculateGlobalInflationControlFactor(big.NewInt(100000))
	a := big.NewRat(1, 1)
	if factor.Cmp(a) != 0 {
		x, _ := a.Float64()
		y, _ := factor.Float64()
		t.Errorf("Expected %20.6f\n got %20.6f\n", x, y)
	}
}
func TestCalculateGlobalInflationControlFactor2(t *testing.T) {
	var rewardComputation RaceRankComputation
	var factor, _ = rewardComputation.CalculateGlobalInflationControlFactor(big.NewInt(100000))
	a := float64(1 / 1.5)
	if factor != a {
		t.Errorf("Expected %20.6f\n got %20.6f\n", a, factor)
	}
}

func TestCalculateGlobalInflationControlFactor3(t *testing.T) {
	var rewardComputation RaceRankComputation
	var factor, _ = rewardComputation.CalculateGlobalInflationControlFactor(big.NewInt(200000))
	a := float64(1 / math.Pow(1.5, 2))
	if factor != a {
		t.Errorf("Expected %20.6f\n got %20.6f\n", a, factor)
	}
}

func TestCalculateCarbonFootprintRewardCollection1(t *testing.T) {
	var rewardComputation RaceRankComputation
	cf := make([]*big.Int, 3)
	cf[0] = big.NewInt(100000)
	cf[1] = big.NewInt(200000)
	cf[2] = big.NewInt(300000)

	// (nodesFootprint []*big.Int, footprint *big.Int, totalCryptoAmount *big.Int)
	var reward, _ = rewardComputation.CalculateCarbonFootprintRewardCollection(cf, big.NewInt(150000), big.NewInt(1000000))
	a := big.NewInt(46822130772748056)
	if reward.Cmp(a) != 0 {

		t.Errorf("Expected %20.6f\n got %20.6f\n", a, reward)
	}
	/*var factor, _ = rewardComputation.CalculateGlobalInflationControlFactor(big.NewInt(200000))
	a := float64(1 / math.Pow(1.5, 2))
	if factor != a {
		t.Errorf("Expected %20.6f\n got %20.6f\n", a, factor)
	}
	*/
}
