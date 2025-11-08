package engine

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
			magicIndex := (blocks * bishopMagics[sq]) >> (64 - bishopMaskCount)
			bishopAttacksTable[sq][magicIndex] = bishopAttacksWithBlockers(sq, blocks)
		}

		// Generate rook attacks for all possible blocks configurations
		for i := range rookBlocksIndices {
			blocks := generateBlockConfiguration(i, rookMask)
			magicIndex := (blocks * rookMagics[sq]) >> (64 - rookMaskCount)
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
// Magic Numbers
// -------------------------------------------------------------------
var rookMagics = [64]Bitboard{
	0x80082080104000,
	0x240004010022000,
	0x900200031004188,
	0x80080084900080,
	0x1200200200281014,
	0x80220011800400,
	0x9400021430028809,
	0x8100120020c08100,
	0x6140800020844001,
	0x820020408d0a00,
	0x4008802002801000,
	0x18800800805000,
	0x1015000801004410,
	0x2002200041008,
	0x605004200490004,
	0x8902000081044426,
	0x80808010204000,
	0x30010a0022048240,
	0x808020001000,
	0x209808008001002,
	0x60901000510c800,
	0x4000180110042040,
	0x820040031100288,
	0x200020009024384,
	0x6840400080008020,
	0x200080400181,
	0x9220600080100080,
	0x200100080800800,
	0x43a4040080080080,
	0x92001200041009,
	0x4206440008305b,
	0x8106018200024314,
	0x4040004480800022,
	0x100420a002880,
	0x861000802004,
	0x410204202000812,
	0x2c800802800400,
	0x4800801600802400,
	0x80014824000210,
	0x21060c0082000053,
	0x40284010828002,
	0x10006000d0004000,
	0x120340820022,
	0x44020c200120028,
	0x1008000400088080,
	0x8001000a04010008,
	0x400010210540018,
	0x200001a054020001,
	0x580004000200440,
	0x200a40008880,
	0xd82002030804200,
	0x8010801000080080,
	0x4010a40081080080,
	0x1090400090100,
	0x2002844810600,
	0x30044038200,
	0x400100c080081021,
	0x810040201202,
	0x2801000a20004015,
	0x2000a20304a42,
	0x1011009028000403,
	0x1000400080289,
	0x80080106009004,
	0x80010124008442,
}

var bishopMagics = [64]Bitboard{
	0x8288101004822082,
	0x94882140b620001,
	0x821020400440001,
	0x1004140080022000,
	0x900410688a020008,
	0x1010c2004200880,
	0x84410141000c0,
	0xa02020082180600,
	0x200108a502c00a0,
	0x204024404068200,
	0x22010010a002000,
	0x8800042404800008,
	0x81420210000000,
	0x880820815140800,
	0x2808040401880841,
	0x2010046062086,
	0x4020912409900100,
	0x8802000410440101,
	0x1001203208010100,
	0x8004002824001080,
	0x4004000080e02204,
	0x2010108020600,
	0x20a843108080222,
	0x6002800040480800,
	0x60889020089100,
	0x4004034420820400,
	0x400220010028200,
	0x80040100f0100408,
	0x2020840000802000,
	0x80281009a806000,
	0x884808012181431,
	0x810810006005206,
	0x1408200842048804,
	0x1002022008228802,
	0x40209002080020,
	0x600802090104,
	0x1840802008a2008,
	0x42a10100020088,
	0xa01002088720c400,
	0x808901020200a200,
	0x14100405091002,
	0x1001218a20021080,
	0xe2905a0082001005,
	0x24200820800,
	0x12404408218100,
	0x10d0102040101,
	0x4220010201800201,
	0x2c1480380810300,
	0x80443004100000,
	0x1040201040002,
	0x1008010063101048,
	0x80840800840c08c0,
	0x124001082020000,
	0x6010a00800c200,
	0x520051006014004,
	0x8110c240040c0,
	0x15a0041640c2000,
	0x20010401034800,
	0x258880825080800,
	0x1510420003840410,
	0x4540000010020885,
	0x442021002101440,
	0x424488008104,
	0x484411101c008080,
}
