package aconcagua

// -------------
// QUEEN ♕
// -------------

// queenAttacks returns a Bitboard with all the squares a queen is attacking
func queenAttacks(q *Bitboard, pos *Position) (attacks Bitboard) {
	blockers := pos.Pieces(White) | pos.Pieces(Black)
	attacks = rookMagicAttacks(Bsf(*q), blockers) | bishopMagicAttacks(Bsf(*q), blockers)
	return
}

// genQueenMoves generates the queen moves in the move list
func genQueenMoves(from *Bitboard, pos *Position, side Color, ml *moveList) {
	toSquares := bishopMoves(from, pos, side) | rookMoves(from, pos, side)
	opponentPieces := pos.Pieces(side.Opponent())

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := quiet

		if toSquare&opponentPieces > 0 {
			flag = capture
		}

		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
}
