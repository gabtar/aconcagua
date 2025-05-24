package aconcagua

// -------------
// ROOK ♖
// -------------

// rookMoves returns a bitboard with the legal moves of the rook from the bitboard passed
func rookMoves(r *Bitboard, pd *PositionData) (moves Bitboard) {
	return rookMagicAttacks(Bsf(*r), pd.allies|pd.enemies) &
		^pd.allies &
		pd.checkRestrictedSquares &
		pinRestrictedSquares(*r, pd.kingPosition, pd.pinnedPieces)
}

// rookMagicAttacks returns a bitboard with the attack mask of a rook from the square passed taking into account the blockers
func rookMagicAttacks(square int, blocks Bitboard) Bitboard {
	blocks &= rooksMaskTable[square]
	magicIndex := (blocks * rookMagics[square]) >> (64 - rooksMaskTable[square].count())
	return rookAttacksTable[square][magicIndex]
}

// genRookMoves generates the rook moves in the move list
func genRookMoves(from *Bitboard, ml *moveList, pd *PositionData) {
	toSquares := rookMoves(from, pd)

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := quiet

		if toSquare&pd.enemies > 0 {
			flag = capture
		}

		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
}
