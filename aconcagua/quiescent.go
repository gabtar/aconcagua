package aconcagua

// Quiescent is an evaluation function that takes into account some dynamic possibilities
func Quiescent(pos *Position, s *Search, alpha int, beta int) int {
	if s.timeControl.stop {
		return 0
	}

	score := pos.Evaluate(&s.PawnTable)

	if score >= beta {
		return beta
	}

	if score > alpha {
		alpha = score
	}

	ml := NewMoveList(40)
	pos.generateCaptures(&ml)

	for i := range len(ml) {
		see := pos.see(ml[i].from(), ml[i].to())
		if see < 0 {
			continue
		}

		pos.MakeMove(&ml[i])
		score = -Quiescent(pos, s, -beta, -alpha)
		pos.UnmakeMove(&ml[i])
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

	targetPiece := pieceRole(pos.PieceAt(squareReference[to]))
	attackerPiece := pieceRole(pos.PieceAt(squareReference[from]))

	blockers := ^pos.EmptySquares()
	attackers := pos.attackers(to, side, blockers) | pos.attackers(to, side.Opponent(), blockers)
	alreadyAttacked := Bitboard(0)
	materialGain[depth] = pieceValue[targetPiece]

	for attackers > 0 {
		depth++
		materialGain[depth] = pieceValue[attackerPiece] - materialGain[depth-1]
		attackers &= ^fromSq
		blockers &= ^fromSq
		alreadyAttacked |= fromSq

		// Find new attackers(by xrays) when removing the already considered pieces into the exchange
		attackers = (pos.attackers(to, side, blockers) | pos.attackers(to, side.Opponent(), blockers)) & ^alreadyAttacked

		side = side.Opponent()
		fromSq, attackerPiece = pos.getLeastValuableAttacker(attackers, side)
		if attackerPiece == NoPiece {
			break
		}
	}

	// Negamax the material gain to get the final static exchange evaluation
	for depth = depth - 1; depth > 0; depth-- { // start with depth -1 beacuse the speculative material store for capture at the end of the tactical sequence
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
	for piece := 5; piece >= 0; piece-- {
		if bitboards[piece]&attackers > 0 {
			attacker := bitboards[piece] & attackers
			return attacker.NextBit(), piece
		}
	}
	return Bitboard(0), NoPiece
}
