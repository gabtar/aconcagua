package aconcagua

// -------------
// QUEEN ♕
// -------------

// queenAttacks returns a Bitboard with all the squares a queen is attacking
func queenAttacks(q *Bitboard, blocks Bitboard) (attacks Bitboard) {
	attacks = rookMagicAttacks(Bsf(*q), blocks) | bishopMagicAttacks(Bsf(*q), blocks)
	return
}

// genQueenMoves generates the queen moves in the move list
func genQueenMoves(from *Bitboard, ml *moveList, pd *PositionData) {
	toSquares := bishopMoves(from, pd) | rookMoves(from, pd)

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := quiet

		if toSquare&pd.enemies > 0 {
			flag = capture
		}

		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
}
