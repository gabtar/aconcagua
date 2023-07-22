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
