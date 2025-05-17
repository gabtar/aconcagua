package aconcagua

// -------------
// ROOK ♖
// -------------

// rookMoves returns a bitboard with the legal moves of the rook from the bitboard passed
func rookMoves(r *Bitboard, pos *Position, side Color) (moves Bitboard) {
	return rookMagicAttacks(Bsf(*r), pos.Pieces(White)|pos.Pieces(Black)) &
		^pos.Pieces(side) &
		pinRestrictedDirection(r, side, pos) &
		checkRestrictedMoves(*r, side, pos)
}

// rookMagicAttacks returns a bitboard with the attack mask of a rook from the square passed taking into account the blockers
func rookMagicAttacks(square int, blocks Bitboard) Bitboard {
	blocks &= rooksMaskTable[square]
	magicIndex := (blocks * rookMagics[square]) >> (64 - rooksMaskTable[square].count())
	return rookAttacksTable[square][magicIndex]
}

// genRookMoves generates the rook moves in the move list
func genRookMoves(from *Bitboard, pos *Position, side Color, ml *moveList) {
	toSquares := rookMoves(from, pos, side)
	opponentPieces := pos.Pieces(side.Opponent())

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := quiet

		if toSquare&opponentPieces > 0 {
			flag = capture
		}

		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
}
