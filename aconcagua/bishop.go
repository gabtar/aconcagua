package aconcagua

// -------------
// BISHOP ♗
// -------------

// bishopMoves returns a bitboard with the legal moves of the bishop from the bitboard passed
func bishopMoves(b *Bitboard, pos *Position, side Color) (moves Bitboard) {
	return bishopAttacks(b, pos) & ^pos.Pieces(side) &
		pinRestrictedDirection(b, side, pos) &
		checkRestrictedMoves(*b, side, pos)
}

// bishopAttacks returns a bitboard with the attacks of a bishop from the bitboard passed
func bishopAttacks(b *Bitboard, pos *Position) (attacks Bitboard) {
	square := Bsf(*b)

	for _, direction := range []uint64{NORTHEAST, SOUTHEAST, SOUTHWEST, NORTHWEST} {
		attacks |= raysAttacks[direction][square]
		nearestBlocker := nearestPieceInDirection(b, pos, direction)

		if nearestBlocker > 0 {
			attacks &= ^raysAttacks[direction][Bsf(nearestBlocker)]
		}
	}
	return
}

// genBishopMoves generates the bishop moves in the move list
func genBishopMoves(from *Bitboard, pos *Position, side Color, ml *moveList) {
	toSquares := bishopMoves(from, pos, side)
	opponentPieces := pos.Pieces(side.Opponent())

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := quiet

		if toSquare&opponentPieces > 0 {
			flag = capture
		}

		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}

	return
}
