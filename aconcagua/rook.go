package aconcagua

// -------------
// ROOK ♖
// -------------

// getRookMoves returns a move slice with all the legal moves of a rook from the bitboard passed
func getRookMoves(r *Bitboard, pos *Position, side Color) (moves []Move) {
	movesBB := rookMoves(r, pos, side)
	pieces := ^pos.EmptySquares()
	from := Bsf(*r)
	piece := pieceOfColor[Rook][side]

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
