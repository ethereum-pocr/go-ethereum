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

package cliquepocr

import (
	"math"
	"math/big"
	"testing"
)

func TestCalculateGlobalInflationControlFactor(t *testing.T) {
	rewardComputation := NewRaceRankComputation()
	n1 := new(big.Int)
	n1, ok := n1.SetString("20000000000000000000000000", 10)
	if !ok {
		t.Errorf("SetString: error")
		return
	}
	var factor, _ = rewardComputation.CalculateGlobalInflationControlFactor(n1)
	a := "0.240754"
	b := factor.FloatString(6)
	if b != a {
		t.Errorf("Expected %s\n got %s\n", a, b)
	}
}



func TestCalculateRanking1(t *testing.T) {
	rewardComputation := NewRaceRankComputation()
	cf := make([]*big.Int, 3)
	cf[0] = big.NewInt(100000)
	cf[1] = big.NewInt(200000)
	cf[2] = big.NewInt(300000)

	var rank, nodes, _ = rewardComputation.CalculateRanking(big.NewInt(150000), cf)
	n1 := math.Pow(0.9, 1)
	n2, _ := rank.Float64()
	if n1 != n2 {
		t.Errorf("Expected %20.6f\n got %20.6f\n", n1, n2)
	}
	if nodes != 3 {
		t.Errorf("Expected %v\n got %v\n", 3, nodes)
	}
}

// func TestCalculateCarbonFootprintReward1(t *testing.T) {
// 	var rewardComputation RaceRankComputation
// 	cf := make([]*big.Int, 3)
// 	cf[0] = big.NewInt(100000)
// 	cf[1] = big.NewInt(200000)
// 	cf[2] = big.NewInt(300000)

// 	// (nodesFootprint []*big.Int, footprint *big.Int, totalCryptoAmount *big.Int)
// 	var reward, _ = rewardComputation.CalculateCarbonFootprintReward(cf, big.NewInt(150000), big.NewInt(1000000))
// 	n1 := new(big.Int)
// 	n1, ok := n1.SetString("2700000000000000000", 10)
// 	if !ok {
// 		t.Errorf("SetString: error")
// 		return
// 	}
// 	if reward.Cmp(n1) != 0 {

// 		t.Errorf("Expected %20.6f\n got %20.6f\n", n1, reward)
// 	}
// }
