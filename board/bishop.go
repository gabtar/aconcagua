package board

// -------------
// BISHOP ♗
// -------------

// getBishopMoves returns a move slice with all the legal moves of a bishop from the bitboard passed
func getBishopMoves(b *Bitboard, pos *Position, side Color) (moves []Move) {
	movesBB := bishopMoves(b, pos, side)
	pieces := ^pos.EmptySquares()
	from := Bsf(*b)
	piece := pieceOfColor[Bishop][side]
	oldEpTarget := 0
	if pos.enPassantTarget > 0 {
		oldEpTarget = Bsf(pos.enPassantTarget)
	}

	for movesBB > 0 {
		to := movesBB.nextOne()
		// bishop moves type only -> capture or normal
		moveType := NORMAL
		capturedPiece := Piece(0)
		if to&pieces > 0 {
			moveType = CAPTURE
			capturedPiece, _ = pos.PieceAt(to.ToStringSlice()[0])
		}
		moves = append(moves, MoveEncode(from, Bsf(to), int(piece), 0, moveType, int(capturedPiece), oldEpTarget))
	}
	return
}

// bishopMoves returns a bitboard with the legal moves of the bishop from the bitboard passed
func bishopMoves(b *Bitboard, pos *Position, side Color) (moves Bitboard) {
	return bishopAttacks(b, pos) & ^pos.Pieces(side) &
		pinRestrictedDirection(*b, side, pos) &
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

// TODO: extract to position/piece file
// nearestPieceInDirection returns a bitboard with the nearest piece in the direction passed
func nearestPieceInDirection(b *Bitboard, pos *Position, dir uint64) (nearestBlocker Bitboard) {
	blockers := ^pos.EmptySquares()
	blockersInDirection := blockers & raysAttacks[dir][Bsf(*b)]

	switch dir {
	case NORTH, EAST, NORTHEAST, NORTHWEST:
		nearestBlocker = BitboardFromIndex(Bsf(blockersInDirection))
	case SOUTH, WEST, SOUTHEAST, SOUTHWEST:
		nearestBlocker = BitboardFromIndex(63 - Bsr(blockersInDirection))
	}
	return
}
