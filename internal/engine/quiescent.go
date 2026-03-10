package engine

// Quiescent is an evaluation function that takes into account some dynamic possibilities
func Quiescent(pos *Position, s *Search, alpha int, beta int, ply int) int {
	s.nodes++
	if s.TimeControl.stop {
		return 0
	}

	s.seldepth = max(s.seldepth, uint8(ply))
	// If the position is a draw avoid redundant search
	if pos.isDraw() {
		return 0
	}

	// Transposition Table probe
	ttScore, ttEval, ttMove, ttHit := s.TranspositionTable.probe(pos.Hash, 0, ply, alpha, beta)
	if ttHit {
		return ttScore
	}

	staticEval := s.evaluate(pos, ttMove, ttEval)

	if staticEval >= beta {
		return beta
	}

	if staticEval > alpha {
		alpha = staticEval
	}

	ml := NewMoveList()
	pd := pos.generatePositionData()
	pos.generateCaptures(ml, &pd)
	genQueenPromotions(pos, pos.Turn, ml, &pd)

	flag := FlagAlpha
	newScore := MinInt
	bestMove := NoMove

	for i := range ml.length {
		see := pos.see(&ml.moves[i])
		if see < 0 {
			continue
		}

		pos.MakeMove(&ml.moves[i])
		newScore = -Quiescent(pos, s, -beta, -alpha, ply+1)
		pos.UnmakeMove(&ml.moves[i])

		if newScore >= beta {
			s.TranspositionTable.store(pos.Hash, 0, ply, FlagBeta, beta, staticEval, ml.moves[i])
			return beta
		}
		if newScore > alpha {
			flag = FlagExact
			bestMove = ml.moves[i]
			alpha = newScore
		}
	}

	s.TranspositionTable.store(pos.Hash, 0, ply, flag, alpha, staticEval, bestMove)
	return alpha
}

// genQueenPromotions generates the queen promotions in the move list
func genQueenPromotions(pos *Position, side Color, ml *MoveList, pd *PositionData) {
	pawns := pos.Bitboards[pieceColor(Pawn, side)]
	promoFromRank := [2]Bitboard{Ranks[6], Ranks[1]}
	posiblesPromotions := pawns & promoFromRank[side] & ^pd.enemies

	for posiblesPromotions > 0 {
		from := posiblesPromotions.NextBit()
		targets := pawnMoves(&from, pd, side)
		to := from << 8
		if side == Black {
			to = from >> 8
		}
		if targets&to > 0 {
			ml.add(*encodeMove(uint16(Bsf(from)), uint16(Bsf(to)), queenPromotion))
		}
	}
}

// see implements an static exchange evaluation on the square passed
func (pos *Position) see(move *Move) int {
	from := move.from()
	to := move.to()
	materialGain := [32]int{}
	pieceValue := [6]int{10000, 900, 500, 300, 300, 100}
	depth := 0
	side := pos.Turn
	fromSq := bitboardFromIndex(from)

	targetPiece := pos.PieceAt(to)
	if targetPiece == NoPiece { // should be an ep capture
		targetPiece = pieceColor(Pawn, side.Opponent())
	}
	// Promotions
	if move.flag() >= knightPromotion {
		targetPiece = pieceColor(Queen, side.Opponent())
	}

	targetRole := pieceRole(targetPiece)
	attackerRole := pieceRole(pos.PieceAt(from))

	blockers := ^pos.EmptySquares()
	attackers := pos.attackersTo(to, side.Opponent(), blockers)
	alreadyAttacked := Bitboard(0)
	materialGain[depth] = pieceValue[targetRole]

	for attackers > 0 {
		depth++
		materialGain[depth] = pieceValue[attackerRole] - materialGain[depth-1]

		// Early termination, if we're already losing
		if max(-materialGain[depth-1], materialGain[depth]) < 0 {
			break
		}

		attackers &= ^fromSq
		blockers &= ^fromSq
		alreadyAttacked |= fromSq

		// Find new attackers(by xrays) when removing the already considered pieces into the exchange
		side = side.Opponent()
		attackers = pos.attackersTo(to, side, blockers) & ^alreadyAttacked
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

// attackersTo returns a bitboard with all the attackersTo of the square passed
func (pos *Position) attackersTo(to int, side Color, blocks Bitboard) (attackers Bitboard) {
	toSq := Bitboard(1 << to)

	// Using the square attacked by algorithm - https://www.chessprogramming.org/Square_Attacked_By#Attacks_to_a_Square
	pawnAttacks := pawnAttacks(&toSq, side.Opponent()) & pos.Bitboards[pieceColor(Pawn, side)]
	knightAttacks := knightAttacksTable[to] & pos.Bitboards[pieceColor(Knight, side)]
	bishopAttacks := bishopAttacks(to, blocks) & pos.Bitboards[pieceColor(Bishop, side)]
	rookAttacks := rookAttacks(to, blocks) & pos.Bitboards[pieceColor(Rook, side)]
	queenAttacks := Attacks(Queen, toSq, blocks) & pos.Bitboards[pieceColor(Queen, side)]
	kingAttacks := kingAttacksTable[to] & pos.Bitboards[pieceColor(King, side)]

	return pawnAttacks | knightAttacks | bishopAttacks | rookAttacks | queenAttacks | kingAttacks
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
