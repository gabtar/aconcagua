package aconcagua

// -------------
// KNIGHT ♘
// -------------

// knightMoves returns a bitboard with the legal moves of the knight from the bitboard passed
func knightMoves(k *Bitboard, pd *PositionData) (moves Bitboard) {
	// If the knight is pinned it can move at all
	if *k&pd.pinnedPieces > 0 {
		return Bitboard(0)
	}
	return knightAttacksTable[Bsf(*k)] & ^pd.allies & pd.checkRestrictedSquares
}

// genKnightMoves generates the knight moves in the move list
func genKnightMoves(from *Bitboard, pos *Position, side Color, ml *moveList, pd *PositionData) {
	toSquares := knightMoves(from, pd)

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := quiet

		if toSquare&pd.enemies > 0 {
			flag = capture
		}

		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
}
