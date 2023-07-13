package board

// -------------
// ROOK ♖
// -------------

// getRookMoves returns a move slice with all the legal moves of a rook from the bitboard passed
func getRookMoves(r *Bitboard, pos *Position, side Color) (moves []Move) {
	movesBB := rookMoves(r, pos, side)
	pieces := ^pos.EmptySquares()
	from := Bsf(*r)
	piece := WhiteRook
	if side == Black {
		piece = BlackRook
	}

	for movesBB > 0 {
		to := movesBB.nextOne()
		// rook moves type only -> capture or normal
		moveType := NORMAL
		if to&pieces > 0 {
			moveType = CAPTURE
		}
		moves = append(moves, MoveEncode(from, Bsf(to), int(piece), 0, moveType))
	}
	return
}

// rookMoves returns a bitboard with the legal moves of the rook from the bitboard passed
func rookMoves(r *Bitboard, pos *Position, side Color) (moves Bitboard) {
	return rookAttacks(r, pos) & ^pos.Pieces(side) &
		pinRestrictedDirection(*r, side, pos) &
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
