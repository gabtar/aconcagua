package engine

// queenAttacks returns a Bitboard with all the squares a queen is attacking
func queenAttacks(q *Bitboard, blocks Bitboard) (attacks Bitboard) {
	attacks = rookAttacks(Bsf(*q), blocks) | bishopAttacks(Bsf(*q), blocks)
	return
}

// rookAttacks returns a bitboard with the attack mask of a rook from the square passed taking into account the blockers
func rookAttacks(square int, blocks Bitboard) Bitboard {
	blocks &= rooksMaskTable[square]
	magicIndex := (blocks * rookMagics[square]) >> RookMagicShift
	return rookAttacksTable[square][magicIndex]
}

// bishopAttacks returns a bitboard with the attack mask of a bishop from the square passed taking into account the blockers
func bishopAttacks(square int, blocks Bitboard) Bitboard {
	blocks &= bishopMaskTable[square]
	magicIndex := (blocks * bishopMagics[square]) >> BishopMagicShift
	return bishopAttacksTable[square][magicIndex]
}

// pawnAttacks returns a bitboard with the squares the pawn attacks from the position passed
func pawnAttacks(p *Bitboard, side Color) (attacks Bitboard) {
	if side == White {
		attacks = (*p&notAFile)<<7 | (*p&notHFile)<<9
	} else {
		attacks = (*p&notAFile)>>9 | (*p&notHFile)>>7
	}
	return
}

// Attacks returns a bitboard with all the squares the piece passed attacks
func Attacks(piece int, from Bitboard, blocks Bitboard) (attacks Bitboard) {
	switch piece {
	case WhiteKing, BlackKing:
		attacks |= kingAttacksTable[Bsf(from)]
	case WhiteQueen, BlackQueen:
		attacks |= queenAttacks(&from, blocks)
	case WhiteRook, BlackRook:
		attacks |= rookAttacks(Bsf(from), blocks)
	case WhiteBishop, BlackBishop:
		attacks |= bishopAttacks(Bsf(from), blocks)
	case WhiteKnight, BlackKnight:
		attacks |= knightAttacksTable[Bsf(from)]
	case WhitePawn:
		attacks |= pawnAttacks(&from, White)
	case BlackPawn:
		attacks |= pawnAttacks(&from, Black)
	}
	return
}
