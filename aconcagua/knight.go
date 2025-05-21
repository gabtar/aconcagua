package aconcagua

// -------------
// KNIGHT ♘
// -------------

// knightMoves returns a bitboard with the legal moves of the knight from the bitboard passed
func knightMoves(k *Bitboard, pos *Position, side Color) (moves Bitboard) {
	// If the knight is pinned it can move at all
	if isPinned(k, side, pos) {
		return Bitboard(0)
	}
	moves = knightAttacksTable[Bsf(*k)] & ^pos.Pieces(side) &
		checkRestrictedMoves(side, pos)
	return
}

// genKnightMoves generates the knight moves in the move list
func genKnightMoves(from *Bitboard, pos *Position, side Color, ml *moveList) {
	toSquares := knightMoves(from, pos, side)
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
