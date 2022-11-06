package cliquepcr

import (
	"errors"
	// "math"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/log"
)

// The standard WhitePaper computation
type RaceRankComputation struct {
	rankArray          []*big.Rat
}

func NewRaceRankComputation() IRewardComputation {
	return &RaceRankComputation{
		rankArray: []*big.Rat{big.NewRat(1,1)},
	}
}

func (wp *RaceRankComputation) getRanking(rank int) *big.Rat {
	// calculate the rational value for the given index if it does not already exists
	previous := wp.rankArray[len(wp.rankArray)-1]
	for i := len(wp.rankArray); i <= rank; i++ {
		// multiply the previous one by 0,9
		previous = new(big.Rat).Mul(previous, big.NewRat(9,10))
		wp.rankArray = append(wp.rankArray, previous)
	}
	return wp.rankArray[rank]
}


func (wp *RaceRankComputation) GetAlgorithmId() int {
	return 3
}


// On an annual basis, what is the minimimum amount of CRC tokens that have to be generated each year
// var minRequiredInflation = 0.03
// Inflation control speed at 10^7
var inflationDenominator = new(big.Int).Exp(big.NewInt(10), big.NewInt(7), nil)
// var alpha = 2

func (wp *RaceRankComputation) CalculateRanking(footprint *big.Int, nodesFootprint []*big.Int) (rank *big.Rat, nbNodes int, err error) {
	if footprint.Cmp(zero) <= 0 {
		return nil, 0, errors.New("cannot proceed with zero or negative footprint")
	}
	var NbItemsAbove int
	nbNodes = len(nodesFootprint)
	
	if nbNodes == 0 {
		return nil, 0, errors.New("cannot rank zero node")
	}
	sort.Slice(nodesFootprint, func(a, b int) bool {
		// sort direction low before high.
		return nodesFootprint[a].Cmp(nodesFootprint[b]) < 0
	})

	for i := 0; i < nbNodes; i++ {
		if nodesFootprint[i].Cmp(footprint) == -1 {
			NbItemsAbove++
		}
	}
	rank = wp.getRanking(NbItemsAbove)
	log.Debug("RaceRankComputation.CalculateRanking", "NbItemsAbove", NbItemsAbove)
	log.Debug("RaceRankComputation.CalculateRanking", "rank", rank)
	// log.Debug("RaceRankComputation.CalculateRanking", "rank 9", wp.getRanking(9))
	// log.Debug("RaceRankComputation.CalculateRanking", "rank 99", wp.getRanking(99))

	return rank, nbNodes, nil
}

// Public function for auditing, but used internally only
func (wp *RaceRankComputation) calculateGlobalInflationControlFactor(M *big.Int) (*big.Rat, error) {
	// L = TotalCRC / InflationDenominator
	// D = pow(alpha, L)
	// GlobalInflation  = 1/D


	// If there is no crpto created, return 1
	if M.Cmp(zero) == 0 {
		return big.NewRat(1,1), nil
	}

	L := new(big.Rat).SetFrac(M, new(big.Int).Mul(CTCUnit, inflationDenominator))

	L = L.Mul(L, big.NewRat(7, 10)) // mul by 0,7 to be able to apply the limited devt on alpha = 2
	// resolve the alpha^L in big.Int by using limited development formula
	// 𝛴 (x^k)/k! with 4 levels only
	D := big.NewRat(1,1) // D = 1
	D = D.Add(D, L) // 1 + L

	L2 := new(big.Rat).Mul(L, L) // L^2
	D = D.Add(D, new(big.Rat).Mul(L2, big.NewRat(1,2))) // + L^2 / 2
	L2 = L2.Mul(L2, L) // L^3
	D = D.Add(D, new(big.Rat).Mul(L2, big.NewRat(1,6))) // + L^3 / 6
	L2 = L2.Mul(L2, L) // L^4
	D = D.Add(D, new(big.Rat).Mul(L2, big.NewRat(1,24))) // + L^3 / 24

	return D.Inv(D), nil
}

func (wp *RaceRankComputation) CalculateCarbonFootprintReward(rank *big.Rat, nbNodes int, totalCryptoAmount *big.Int) (*big.Int, error) {
	// In CRC Unit : 0.9^rank
	rewardCRCUnit := new(big.Rat).Mul(rank, new(big.Rat).SetInt(CTCUnit))

	// 0.9^rank x N
	rewardCRCUnit = rewardCRCUnit.Mul(rewardCRCUnit, big.NewRat(int64(nbNodes), 1))

	// 0.9^rank x N * Inflation
	inflationFactor, err := wp.calculateGlobalInflationControlFactor(totalCryptoAmount)
	if err != nil {
		return nil, err
	}
	rewardCRCUnit = rewardCRCUnit.Mul(rewardCRCUnit, inflationFactor)

	u := new(big.Int).Div(rewardCRCUnit.Num(), rewardCRCUnit.Denom())  

	return u, nil
}
