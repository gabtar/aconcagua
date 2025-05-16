package aconcagua

import (
	"fmt"
	"math/rand/v2"
)

// find_magics.go is an utility to find the magic numbers for the bitboards
// based on the original code of Tord Romstad's proposal to find magics:
// Just trying out random numbers with a low number of nonzero bits until you find a number which works
// is by far the fastest and easiest way to generate the magic numbers, in my experience. On my Core Duo 2.8 GHz,
// it takes less than a second to find magic numbers for rooks and bishops for all squares (and I have made no
// 	attempt to optimize the code, it should be easy to make it much faster).

// findMagicNumber finds the magic number for the given square for bishops or rooks
func findMagicNumber(square int, isRook bool) (magic Bitboard) {
	attackMask := bishopMask(square)
	if isRook {
		attackMask = rookMask(square)
	}

	maskTotalBits := attackMask.count()
	blockersConfigurations := 1 << maskTotalBits

	attacksPatterns := make([]Bitboard, blockersConfigurations)
	blockersPatterns := make([]Bitboard, blockersConfigurations)
	for i := 0; i < blockersConfigurations; i++ {
		blockersPatterns[i] = generateBlockConfiguration(i, attackMask)

		if isRook {
			attacksPatterns[i] = rooksAttacksWithBlockers(square, blockersPatterns[i])
		} else {
			attacksPatterns[i] = bishopAttacksWithBlockers(square, blockersPatterns[i])
		}
	}

	for i := 0; i < 1000000; i++ {
		magic = Bitboard(rand.Uint64() & rand.Uint64() & rand.Uint64())

		if Bitboard((attackMask*magic)&0xffffffffffffffff) < 6 {
			continue
		}

		used := make([]Bitboard, blockersConfigurations)
		fail := false
		for j := 0; j < blockersConfigurations; j++ {
			magicIndex := (magic * blockersPatterns[j]) >> (64 - maskTotalBits)
			if used[magicIndex] == 0 {
				used[magicIndex] = attacksPatterns[j]
			} else if used[magicIndex] != attacksPatterns[j] {
				fail = true
			}
		}
		if !fail {
			break
		}
	}

	return
}

func GenerateMagicNumbersForBishopsAndRooks() {
	fmt.Println("Rooks magic numbers: ")
	fmt.Println("---------------------")
	for i := 0; i < 64; i++ {
		rookMagics[i] = findMagicNumber(i, true)
		fmt.Printf("sq %d: 0x%x,\n", i, rookMagics[i])
	}

	fmt.Println("Bishops magic numbers: ")
	fmt.Println("-----------------------")
	for i := 0; i < 64; i++ {
		bishopMagics[i] = findMagicNumber(i, false)
		fmt.Printf("sq %d: 0x%x,\n", i, bishopMagics[i])
	}
}
