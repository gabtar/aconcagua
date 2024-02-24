package board

// -------------
// QUEEN ♕
// -------------

// queenAttacks returns a Bitboard with all the squares a queen is attacking
func queenAttacks(q *Bitboard, pos *Position) (attacks Bitboard) {
	attacks = rookAttacks(q, pos) | bishopAttacks(q, pos)
	return
}

// getQueenMoves returns a move slice with all the legal moves of a queen from the bitboard passed
func getQueenMoves(q *Bitboard, pos *Position, side Color) (moves []Move) {
	movesBB := bishopMoves(q, pos, side) | rookMoves(q, pos, side)
	pieces := ^pos.EmptySquares()
	from := Bsf(*q)
	piece := pieceOfColor[Queen][side]

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
