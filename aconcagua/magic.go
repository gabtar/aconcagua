package aconcagua

var rooksMaskTable [64]Bitboard
var bishopMaskTable [64]Bitboard
var bishopAttacksTable [64][512]Bitboard
var rookAttacksTable [64][4096]Bitboard

// init initializes the tables for use w/ magic numbers
func init() {
	// initialize rooksAttacksMask and bishopsAttacksMask
	for sq := 0; sq < 64; sq++ {
		rooksMaskTable[sq] = rookMask(sq)
		bishopMaskTable[sq] = bishopMask(sq)
	}

	// initialize rookAttacksTable and bishopAttacksTable
	for sq := 0; sq < 64; sq++ {
		bishopMask := bishopMaskTable[sq]
		rookMask := rooksMaskTable[sq]
		bishopMaskCount := bishopMask.count()
		rookMaskCount := rookMask.count()
		bishopBlocksIndices := 1 << bishopMaskCount
		rookBlocksIndices := 1 << rookMaskCount

		// Generate bishop attacks for all possible blocks configurations
		for i := 0; i < bishopBlocksIndices; i++ {
			blocks := generateBlockConfiguration(i, bishopMask)
			magicIndex := (blocks * bishopMagics[sq]) >> (64 - bishopMaskCount)
			bishopAttacksTable[sq][magicIndex] = bishopAttacksWithBlockers(sq, blocks)
		}

		// Generate rook attacks for all possible blocks configurations
		for i := 0; i < rookBlocksIndices; i++ {
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

	for i := 0; i < bitCount; i++ {
		bitPos := Bsf(mask)
		mask &= mask - 1

		// If the corresponding bit in the index is set, set the bit in the blocks config
		if (index & (1 << i)) != 0 {
			blocks |= (1 << bitPos)
		}
	}

	return blocks
}

// bishop returns a bitboard with bishop attacks from the given square with the given blocks
func bishop(square int, blocks Bitboard) Bitboard {
	blocks &= bishopMaskTable[square]
	magicIndex := (blocks * bishopMagics[square]) >> (64 - bishopMaskTable[square].count())
	return bishopAttacksTable[square][magicIndex]
}

// rook returns a bitboard with rook attacks from the given square with the given blocks
func rook(square int, blocks Bitboard) Bitboard {
	blocks &= rooksMaskTable[square]
	magicIndex := (blocks * rookMagics[square]) >> (64 - rooksMaskTable[square].count())
	return rookAttacksTable[square][magicIndex]
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
// TODO: I still need to check out these numbers found w/ my findMagic implementation
// in simple tests are working ok. When i redesign the move generation i will check it out with perft tests
var rookMagics = [64]Bitboard{
	0x1080002084400010,
	0x4040002000100040,
	0x28800d8060001000,
	0x180100008008680,
	0x80040180180002,
	0x5200020003900408,
	0x80028001004200,
	0x800b000020c180,
	0x802040088001,
	0x8040804000842000,
	0x4141002001004094,
	0x1400801000180082,
	0x8018009c0081,
	0x10c2000806001004,
	0x404001002041801,
	0x120801080004300,
	0x8400208000844002,
	0x1002828020144000,
	0x410011006000,
	0x8818008001001,
	0x1808044004800,
	0x108808014000200,
	0x2880040088013002,
	0x8010020008440887,
	0x180208880004000,
	0x8228500040002008,
	0x600080801000,
	0x1089002100181000,
	0x2008600082010,
	0x620240080800200,
	0xeb000100020004,
	0x304100010000438a,
	0x40012041800082,
	0x5010006000400044,
	0x5202004822001080,
	0x810280080801004,
	0x1242040080800800,
	0x111000249000400,
	0x1b00100604004801,
	0x200204540a000991,
	0x400080028020,
	0xcbe4500020004001,
	0x1004020010010,
	0x500018008080,
	0x808008500110008,
	0x2020004008080,
	0x800002d008040001,
	0x2000024281020014,
	0x40008008a180,
	0x202400300228100,
	0x82000d0008080,
	0x12420008a03200,
	0x1802100801004500,
	0x40080aa0080,
	0x4000082201300400,
	0xa240100428a00,
	0x6080810048213202,
	0x10010242007187a2,
	0x1001088422001,
	0xa000820400432,
	0x4007000800063015,
	0x4222006108041006,
	0x6401a1c830020104,
	0x4024050080a242,
}
var bishopMagics = [64]Bitboard{
	0x80000000000801,
	0x6c00600814404000,
	0x40000e4202002,
	0x1001200040080002,
	0x2842006c00808080,
	0x18020400440c04,
	0x4c00000092000,
	0x3000322108201802,
	0x2000020080c0000,
	0x804040428814008,
	0xb004c80438000000,
	0x44c6440000202401,
	0x2000440080040,
	0x10022904000a2805,
	0x1048420008200900,
	0xa142000100600000,
	0x4220008007000c0,
	0x10208200001a0,
	0x1070001400a240,
	0x405010001108500e,
	0x1140808010010,
	0x2000101,
	0x1009020800045000,
	0x10120a0010090001,
	0xa004032000002,
	0x4a00010884082e00,
	0x1002110000a00,
	0x9a021000602084,
	0x8000008010002001,
	0x5000005411010082,
	0x820020824080000,
	0x20d018008040040,
	0x1b80,
	0x80102402208000,
	0x220100120200008,
	0x409000830208809,
	0x100000006b8200a8,
	0x80089c010,
	0x40400022002000,
	0x4181080048180000,
	0x1400201100204002,
	0x48040000a004000,
	0x104008800040230,
	0x80182142086020,
	0x1000017400800,
	0x1110080004000209,
	0x200a0001000009,
	0x400a060084800408,
	0x8004002085010022,
	0x20e000e000020000,
	0x1480000200002000,
	0x50000660041c0,
	0x8304025002002a9,
	0x8801004000880,
	0x41080901060a8800,
	0x80000004420000,
	0x8000000082500800,
	0x200404000020080c,
	0x20092400d004440,
	0x54210014140a0b00,
	0x8020000042000204,
	0x4000082801204444,
	0x2502010200490,
	0x811004000000002,
}
