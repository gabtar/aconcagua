package aconcagua

// quiescent is an evaluation function that takes into account some dynamic possibilities
func quiescent(pos *Position, s *Search, alpha int, beta int) int {
	if s.timeControl.stop {
		return 0
	}

	score := Eval(pos)

	if score >= beta {
		return beta
	}

	if score > alpha {
		alpha = score
	}

	ml := pos.LegalMoves()
	ml.capturesOnly()

	for i := 0; i < ml.length; i++ {
		pos.MakeMove(&ml.moves[i])
		score = -quiescent(pos, s, -beta, -alpha)
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
	side := pos.Turn.Opponent()
	fromSq := Bitboard(1 << from)

	targetPiece := pos.PieceAt(squareReference[to])
	if targetPiece > 5 {
		targetPiece = targetPiece - 6
	}

	attackerPiece := NoPiece
	blockers := ^pos.EmptySquares()
	attackers := pos.attackers(to, side, blockers) | pos.attackers(to, side.Opponent(), blockers)
	alreadyAttacked := Bitboard(0)
	materialGain[depth] = pieceValue[targetPiece]

	for attackers > 0 {
		depth++
		side = side.Opponent()
		fromSq, attackerPiece = pos.getLeastValuableAttacker(attackers, side)
		if attackerPiece == NoPiece {
			break
		}
		materialGain[depth] = pieceValue[attackerPiece] - materialGain[depth-1]

		attackers &= ^fromSq
		blockers &= ^fromSq
		alreadyAttacked |= fromSq

		// Find new attackers(by xrays) when removing the already considered pieces into the exchange
		attackers = (pos.attackers(to, side, blockers) | pos.attackers(to, side.Opponent(), blockers)) & ^alreadyAttacked
	}

	// Negamax the material gain to get the final static exchange evaluation
	for depth = depth - 1; depth > 0; depth-- { // start with depth -1 beacuse the speculative material store for capture at the end of the tactical sequence
		materialGain[depth-1] = -max(-materialGain[depth-1], materialGain[depth])
	}

	return materialGain[0]
}

// max returns the maximum of a and b
func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// attackers returns a bitboard with all the attackers of the square passed
func (pos *Position) attackers(to int, side Color, blocks Bitboard) (attackers Bitboard) {
	pieceStart := startingPieceNumber(side)
	toSq := Bitboard(1 << to)

	for piece, bitboard := range pos.getBitboards(side) {
		from := bitboard.NextBit()
		for from > 0 {
			attackedSquares := Attacks(piece+pieceStart, from, blocks)
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
	for piece := 5; piece >= 0; piece-- {
		if bitboards[piece]&attackers > 0 {
			attacker := bitboards[piece] & attackers
			return attacker.NextBit(), piece
		}
	}
	return Bitboard(0), NoPiece
}
