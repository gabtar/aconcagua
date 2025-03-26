package aconcagua

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

// newQueenMoves returns a moves array with the queen moves in chessMove format
func newQueenMoves(from *Bitboard, pos *Position, side Color) (moves []chessMove) {
	toSquares := bishopMoves(from, pos, side) | rookMoves(from, pos, side)
	opponentPieces := pos.Pieces(side.Opponent())

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := quiet

		if toSquare&opponentPieces > 0 {
			flag = capture
		}

		moves = append(moves, encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}

	return
}
