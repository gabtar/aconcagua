package engine

// Quiescent is an evaluation function that takes into account some dynamic possibilities
func Quiescent(pos *Position, s *Search, alpha int, beta int) int {
	s.nodes++
	if s.TimeControl.stop {
		return 0
	}

	score := pos.Evaluate()

	if score >= beta {
		return beta
	}

	if score > alpha {
		alpha = score
	}

	ml := NewMoveList()
	pd := pos.generatePositionData()
	pos.generateCaptures(ml, &pd)

	for i := range ml.length {
		see := pos.see(ml.moves[i].from(), ml.moves[i].to())
		if see < 0 {
			continue
		}

		pos.MakeMove(&ml.moves[i])
		score = -Quiescent(pos, s, -beta, -alpha)
		pos.UnmakeMove(&ml.moves[i])
		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha
}

// see implements an static exchange evaluation on the square passed
func (pos *Position) see(from int, to int) int {
	materialGain := [32]int{}
	pieceValue := [6]int{10000, 900, 500, 300, 300, 100}
	depth := 0
	side := pos.Turn
	fromSq := bitboardFromIndex(from)

	targetPiece := pos.PieceAt(to)
	if targetPiece == NoPiece { // should be an ep capture
		targetPiece = pieceColor(Pawn, side.Opponent())
	}
	targetRole := pieceRole(targetPiece)
	attackerRole := pieceRole(pos.PieceAt(from))

	blockers := ^pos.EmptySquares()
	attackers := pos.attackers(to, side.Opponent(), blockers)
	alreadyAttacked := Bitboard(0)
	materialGain[depth] = pieceValue[targetRole]

	for attackers > 0 {
		depth++
		materialGain[depth] = pieceValue[attackerRole] - materialGain[depth-1]
		attackers &= ^fromSq
		blockers &= ^fromSq
		alreadyAttacked |= fromSq

		// Find new attackers(by xrays) when removing the already considered pieces into the exchange
		side = side.Opponent()
		attackers = pos.attackers(to, side, blockers) & ^alreadyAttacked
		fromSq, attackerRole = pos.getLeastValuableAttacker(attackers, side)
		if attackerRole == NoPiece {
			break
		}
	}

	// Negamax the material gain to get the final static exchange evaluation
	for depth = depth - 1; depth > 0; depth-- { // start with depth -1 because we use the speculative material store for capture at the end of the tactical sequence
		materialGain[depth-1] = -max(-materialGain[depth-1], materialGain[depth])
	}

	return materialGain[0]
}

// attackers returns a bitboard with all the attackers of the square passed
func (pos *Position) attackers(to int, side Color, blocks Bitboard) (attackers Bitboard) {
	toSq := Bitboard(1 << to)

	for p, bitboard := range pos.getBitboards(side) {
		piece := pieceColor(p, side)
		from := bitboard.NextBit()
		for from > 0 {
			attackedSquares := Attacks(piece, from, blocks)
			if attackedSquares&toSq > 0 {
				attackers |= from
			}
			from = bitboard.NextBit()
		}
	}
	return
}

// getLeastValuableAttacker returns the least valuable attacker from the attackers bitboard
func (pos *Position) getLeastValuableAttacker(attackers Bitboard, side Color) (Bitboard, int) {
	bitboards := pos.getBitboards(side)
	for piece := Pawn; piece >= King; piece-- {
		attackingPieces := bitboards[piece] & attackers
		if attackingPieces > 0 {
			return attackingPieces.NextBit(), piece
		}
	}
	return Bitboard(0), NoPiece
}
