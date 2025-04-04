package aconcagua

// -------------
// ROOK ♖
// -------------

// rookMoves returns a bitboard with the legal moves of the rook from the bitboard passed
func rookMoves(r *Bitboard, pos *Position, side Color) (moves Bitboard) {
	return rookAttacks(r, pos) & ^pos.Pieces(side) &
		pinRestrictedDirection(r, side, pos) &
		checkRestrictedMoves(*r, side, pos)
}

// rookAttacks retuns all squares a rook attacks from the passed square
func rookAttacks(r *Bitboard, pos *Position) (attacks Bitboard) {
	square := Bsf(*r)

	for _, direction := range []uint64{NORTH, SOUTH, WEST, EAST} {
		attacks |= raysAttacks[direction][square]
		nearestBlocker := nearestPieceInDirection(r, pos, direction)

		if nearestBlocker > 0 {
			attacks &= ^raysAttacks[direction][Bsf(nearestBlocker)]
		}
	}
	return
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
