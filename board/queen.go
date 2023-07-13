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
	piece := WhiteQueen
	if side == Black {
		piece = BlackQueen
	}

	for movesBB > 0 {
		to := movesBB.nextOne()
		// bishop moves type only -> capture or normal
		moveType := NORMAL
		if to&pieces > 0 {
			moveType = CAPTURE
		}
		moves = append(moves, MoveEncode(from, Bsf(to), int(piece), 0, moveType))
	}
	return
}
