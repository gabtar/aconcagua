package board

// -------------
// KNIGHT ♘
// -------------

// getKnightMoves returns a move slice of all posible moves of the knight passed
func getKnightMoves(b *Bitboard, pos *Position, side Color) (moves []Move) {
	movesBB := knightMoves(b, pos, side)
	pieces := ^pos.EmptySquares()
	from := Bsf(*b)
	piece := pieceOfColor[Knight][side]

	for movesBB > 0 {
		to := movesBB.nextOne()
		move := newMove().
			setFromSq(from).
			setToSq(Bsf(to)).
			setPiece(piece).
			setMoveType(NORMAL).
			setEpTargetBefore(pos.enPassantTarget).
			setRule50Before(pos.halfmoveClock).
			setCastleRightsBefore(pos.castlingRights)

		if to&pieces > 0 {
			capturedPiece, _ := pos.PieceAt(squareReference[Bsf(to)])
			move.setMoveType(CAPTURE).setCapturedPiece(capturedPiece)
		}
		moves = append(moves, *move)
	}
	return
}

// knightMoves returns a bitboard with the legal moves of the knight from the bitboard passed
func knightMoves(k *Bitboard, pos *Position, side Color) (moves Bitboard) {
	// If the knight is pinned it can move at all
	if isPinned(*k, side, pos) {
		return Bitboard(0)
	}
	moves = knightAttacks(k, pos) & ^pos.Pieces(side) &
		checkRestrictedMoves(*k, side, pos)
	return
}

// knightAttacks returns a bitboard with the attacks of a knight from the bitboard passed
// TODO: use precomputed hash/array
func knightAttacks(k *Bitboard, pos *Position) (attacks Bitboard) {
	notInHFile := *k & ^(*k & files[7])
	notInAFile := *k & ^(*k & files[0])
	notInABFiles := *k & ^(*k & (files[0] | files[1]))
	notInGHFiles := *k & ^(*k & (files[7] | files[6]))

	attacks = notInAFile<<15 | notInHFile<<17 | notInGHFiles<<10 |
		notInABFiles<<6 | notInHFile>>15 | notInAFile>>17 |
		notInABFiles>>10 | notInGHFiles>>6
	return
}
