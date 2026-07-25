package engine

import (
	"fmt"
	"math/rand/v2"
)

// Fixed shifts for rooks and bishops
const (
	BishopMagicShift = 55
	RookMagicShift   = 52
)

var rooksMaskTable [64]Bitboard
var bishopMaskTable [64]Bitboard
var bishopAttacksTable [64][512]Bitboard
var rookAttacksTable [64][4096]Bitboard

// init initializes the tables for use w/ magic numbers
func init() {
	// initialize rooksAttacksMask and bishopsAttacksMask
	for sq := range 64 {
		rooksMaskTable[sq] = rookMask(sq)
		bishopMaskTable[sq] = bishopMask(sq)
	}

	// initialize rookAttacksTable and bishopAttacksTable
	for sq := range 64 {
		bishopMask := bishopMaskTable[sq]
		rookMask := rooksMaskTable[sq]
		bishopMaskCount := bishopMask.count()
		rookMaskCount := rookMask.count()
		bishopBlocksIndices := 1 << bishopMaskCount
		rookBlocksIndices := 1 << rookMaskCount

		// Generate bishop attacks for all possible blocks configurations
		for i := range bishopBlocksIndices {
			blocks := generateBlockConfiguration(i, bishopMask)
			magicIndex := (blocks * bishopMagics[sq]) >> BishopMagicShift
			bishopAttacksTable[sq][magicIndex] = bishopAttacksWithBlockers(sq, blocks)
		}

		// Generate rook attacks for all possible blocks configurations
		for i := range rookBlocksIndices {
			blocks := generateBlockConfiguration(i, rookMask)
			magicIndex := (blocks * rookMagics[sq]) >> RookMagicShift
			rookAttacksTable[sq][magicIndex] = rooksAttacksWithBlockers(sq, blocks)
		}
	}
}

// generateBlockConfiguration creates an blocks bitboard based on an index and mask
func generateBlockConfiguration(index int, mask Bitboard) Bitboard {
	var blocks Bitboard
	bitCount := mask.count()

	for i := range bitCount {
		bitPos := Bsf(mask)
		mask &= mask - 1

		// If the corresponding bit in the index is set, set the bit in the blocks config
		if (index & (1 << i)) != 0 {
			blocks |= (1 << bitPos)
		}
	}

	return blocks
}

// rookMask returns a Bitboard with all the squares a rook is attacking
// rookMask does not take into account the edges (no outer squares)
func rookMask(square int) (attacks Bitboard) {
	rank, file := square/8, square%8
	for r := rank + 1; r <= 6; r++ {
		attacks |= 1 << (r*8 + file)
	}
	for r := rank - 1; r >= 1; r-- {
		attacks |= 1 << (r*8 + file)
	}
	for f := file + 1; f <= 6; f++ {
		attacks |= 1 << (rank*8 + f)
	}
	for f := file - 1; f >= 1; f-- {
		attacks |= 1 << (rank*8 + f)
	}
	return
}

// bishopMask returns a Bitboard with all the squares a bishop is attacking
// bishopMask does not take into account the edges (no outer squares)
func bishopMask(square int) (attacks Bitboard) {
	rank, file := square/8, square%8
	for r, f := rank+1, file+1; r <= 6 && f <= 6; r, f = r+1, f+1 {
		attacks |= 1 << (r*8 + f)
	}
	for r, f := rank+1, file-1; r <= 6 && f >= 1; r, f = r+1, f-1 {
		attacks |= 1 << (r*8 + f)
	}
	for r, f := rank-1, file+1; r >= 1 && f <= 6; r, f = r-1, f+1 {
		attacks |= 1 << (r*8 + f)
	}
	for r, f := rank-1, file-1; r >= 1 && f >= 1; r, f = r-1, f-1 {
		attacks |= 1 << (r*8 + f)
	}
	return
}

// rooksAttacksWithBlockers returns a bitboard with the squares attacks from the passed square
func rooksAttacksWithBlockers(sq int, blockers Bitboard) (attacks Bitboard) {
	rank, file := sq/8, sq%8
	for r := rank + 1; r <= 7; r++ {
		square := r*8 + file
		attacks |= 1 << square
		if blockers&(1<<square) > 0 {
			break
		}
	}
	for r := rank - 1; r >= 0; r-- {
		square := r*8 + file
		attacks |= 1 << square
		if blockers&(1<<square) > 0 {
			break
		}
	}
	for f := file + 1; f <= 7; f++ {
		square := rank*8 + f
		attacks |= 1 << square
		if blockers&(1<<square) > 0 {
			break
		}
	}
	for f := file - 1; f >= 0; f-- {
		square := rank*8 + f
		attacks |= 1 << square
		if blockers&(1<<square) > 0 {
			break
		}
	}
	return
}

// bishopAttacksWithBlockers returns a bitboard with the squares attacks from the passed square
func bishopAttacksWithBlockers(sq int, blockers Bitboard) (attacks Bitboard) {
	rank, file := sq/8, sq%8
	for r, f := rank+1, file+1; r <= 7 && f <= 7; r, f = r+1, f+1 {
		square := r*8 + f
		attacks |= 1 << square
		if blockers&(1<<square) > 0 {
			break
		}
	}
	for r, f := rank+1, file-1; r <= 7 && f >= 0; r, f = r+1, f-1 {
		square := r*8 + f
		attacks |= 1 << square
		if blockers&(1<<square) > 0 {
			break
		}
	}
	for r, f := rank-1, file+1; r >= 0 && f <= 7; r, f = r-1, f+1 {
		square := r*8 + f
		attacks |= 1 << square
		if blockers&(1<<square) > 0 {
			break
		}
	}
	for r, f := rank-1, file-1; r >= 0 && f >= 0; r, f = r-1, f-1 {
		square := r*8 + f
		attacks |= 1 << square
		if blockers&(1<<square) > 0 {
			break
		}
	}
	return
}

// -------------------------------------------------------------------
// Precalculated magic numbers with fixed shift
// These tables can be obtained by using the findMagicNumber function
// for each specific square
// -------------------------------------------------------------------
var rookMagics = [64]Bitboard{
	0x80004001801024, 0x40022018005040, 0x820009408001000, 0x2000a0020900c08, 0x2080004000a2408, 0x80420410140500, 0x9080010002000880, 0x5480024060800900,
	0x400040102000, 0xa001040444022410, 0x804004a5004, 0x82104020020100, 0x220180001000218, 0xa05000468010006, 0xa01080804600600, 0x22101000240804,
	0x1b408080004000, 0x8400081800900110, 0x920042002000, 0x420041008000210, 0x4800070008001402, 0x20010000844088, 0x8408024020018601, 0x8200688009040,
	0x18001002100400, 0x2108410010c000, 0x2000120010010480, 0x81201201001000, 0x80200213050001, 0x4282084484002040, 0x28000080a0004201, 0x2040089002,
	0x4110416008400010, 0x104080210602000, 0x11408120011c0004, 0x2100100081008, 0x48b1001015000800, 0xc0210200204402, 0x11800884200, 0x400028c0080a80,
	0x160011001ac2000, 0x40012000a2002, 0xc201000400222000, 0x2102441000800, 0x204540100031800, 0x80060a110020004, 0x90000100804c2002, 0x800200a1004,
	0x4000200010004240, 0x480880022085108, 0x200810840880, 0x10280a00d0011110, 0xd012800406204402, 0xc942000908009020, 0x8001010008008440, 0x100810402008,
	0x1010401100800065, 0x10a0140a00812, 0x2411012000181005, 0x12490082040c2, 0x300024800042b, 0x92000804009062, 0x20008034004203, 0x208001024080a402,
}

var bishopMagics = [64]Bitboard{
	0x56044080250080, 0x1048a0010042141, 0x2f8002a640040, 0x80200c2082208000, 0x200100404080812, 0x90000a0222000904, 0x8001002910005006, 0x110410080304400,
	0x2400018102020010, 0x100002c2520082a4, 0x5878100100408804, 0x800201200070000, 0x11040000001, 0x8082221011425008, 0x100020069200400, 0x80002018406c448,
	0x20200838102081, 0x1210044012020, 0x2a2005200c0005, 0x5042a00260060090, 0x12101080200200, 0x21020008222480c0, 0x8200304a36448290, 0x400020110a48001d,
	0x40d0100001040083, 0x204001122308, 0x201008110400a00, 0x21040004440080, 0x4040848014002000, 0x8086e4040a008063, 0x8010b20000401084, 0x8406018080a08490,
	0x600804041081000, 0x1208040146040284, 0x82040101020a24, 0xa0400a00082200, 0x9010008200802200, 0x420020401c0801, 0x20282870800a0020, 0x8004000a80802400,
	0x4184714202060042, 0x22401082040400, 0x10470d0000410, 0x10244a0404000805, 0x200ea0246c04400, 0xc004040220e20220, 0x1908904001088, 0x802420600020,
	0x2001001004310, 0x6002004603034010, 0x2a1501300404000a, 0x600b2002001, 0x8006a804510000, 0x420e16052000, 0x12042040a10050, 0x18010024002502,
	0x16001891088440, 0x20002018250c040, 0x4000905120131008, 0x118024008080840, 0x40c2200e2900c0, 0x82014c10102024, 0x242000c20080, 0x1800898049086820,
}

// find_magics.go is an utility to find the magic numbers for the bitboards
// based on the original code of Tord Romstad's proposal to find magics:
// Just trying out random numbers with a low number of nonzero bits until you find a number which works
// is by far the fastest and easiest way to generate the magic numbers, in my experience. On my Core Duo 2.8 GHz,
// it takes less than a second to find magic numbers for rooks and bishops for all squares (and I have made no
// 	attempt to optimize the code, it should be easy to make it much faster).

// findMagicNumber finds the magic number for the given square for bishops or rooks
func findMagicNumber(square int, isRook int) (magic Bitboard) {
	magicShift := [2]int{BishopMagicShift, RookMagicShift}
	tableSize := [2]int{512, 4096}
	attackMask := bishopMask(square)
	if isRook == 1 {
		attackMask = rookMask(square)
	}

	maskTotalBits := attackMask.count()
	blockersConfigurations := 1 << maskTotalBits

	attacksPatterns := make([]Bitboard, blockersConfigurations)
	blockersPatterns := make([]Bitboard, blockersConfigurations)
	for i := range blockersConfigurations {
		blockersPatterns[i] = generateBlockConfiguration(i, attackMask)

		if isRook == 1 {
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

		used := make([]Bitboard, tableSize[isRook])
		fail := false
		for j := range blockersConfigurations {
			magicIndex := (magic * blockersPatterns[j]) >> magicShift[isRook]
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
		rookMagics[i] = findMagicNumber(i, 1)
		fmt.Printf("sq %d: 0x%x,\n", i, rookMagics[i])
	}

	fmt.Println("Bishops magic numbers: ")
	fmt.Println("-----------------------")
	for i := range 64 {
		bishopMagics[i] = findMagicNumber(i, 0)
		fmt.Printf("sq %d: 0x%x,\n", i, bishopMagics[i])
	}
}
