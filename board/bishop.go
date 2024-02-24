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

	for movesBB > 0 {
		to := movesBB.NextBit()
		move := newMove().
			setFromSq(from).
			setToSq(Bsf(to)).
			setPiece(piece).
			setMoveType(Normal).
			setEpTargetBefore(pos.enPassantTarget).
			setRule50Before(pos.halfmoveClock).
			setCastleRightsBefore(pos.castlingRights)

		if to&pieces > 0 {
			capturedPiece, _ := pos.PieceAt(squareReference[Bsf(to)])
			move.setMoveType(Capture).setCapturedPiece(capturedPiece)
		}
		moves = append(moves, *move)
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
		nearestBlocker = bitboardFromIndex(Bsf(blockersInDirection))
	case SOUTH, WEST, SOUTHEAST, SOUTHWEST:
		nearestBlocker = bitboardFromIndex(63 - Bsr(blockersInDirection))
	}
	return
}
