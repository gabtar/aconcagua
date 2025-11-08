package engine

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
	for i := range blockersConfigurations {
		blockersPatterns[i] = generateBlockConfiguration(i, attackMask)

		if isRook {
			attacksPatterns[i] = rooksAttacksWithBlockers(square, blockersPatterns[i])
		} else {
			attacksPatterns[i] = bishopAttacksWithBlockers(square, blockersPatterns[i])
		}
	}

	for range 1000000 {
		magic = Bitboard(rand.Uint64() & rand.Uint64() & rand.Uint64())

		if Bitboard((attackMask*magic)&0xffffffffffffffff) < 6 {
			continue
		}

		used := make([]Bitboard, blockersConfigurations)
		fail := false
		for j := range blockersConfigurations {
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

// GenerateMagicNumbersForRooksAndBishops prints the magic number for each square of the board for a bishop and a rook
func GenerateMagicNumbersForRooksAndBishops() {
	fmt.Println("Rooks magic numbers: ")
	fmt.Println("---------------------")
	for i := range 64 {
		rookMagics[i] = findMagicNumber(i, true)
		fmt.Printf("sq %d: 0x%x,\n", i, rookMagics[i])
	}

	fmt.Println("Bishops magic numbers: ")
	fmt.Println("-----------------------")
	for i := range 64 {
		bishopMagics[i] = findMagicNumber(i, false)
		fmt.Printf("sq %d: 0x%x,\n", i, bishopMagics[i])
	}
}
