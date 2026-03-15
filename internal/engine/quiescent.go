package engine

var SEEPieceValues = [6]int{10000, 900, 500, 300, 300, 100}

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
// based on Ethereal staticExchangeEvaluation code: https://github.com/AndyGrant/Ethereal/blob/master/src/search.c#L929C5-L929C29
func (pos *Position) see(move *Move) int {
	from, to := move.from(), move.to()
	materialGain := [32]int{}
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

	diagonalAttackers := pos.Bitboards[WhiteBishop] | pos.Bitboards[BlackBishop] |
		pos.Bitboards[WhiteQueen] | pos.Bitboards[BlackQueen]
	orthogonalAttackers := pos.Bitboards[WhiteRook] | pos.Bitboards[BlackRook] |
		pos.Bitboards[WhiteQueen] | pos.Bitboards[BlackQueen]

	blockers := ^pos.EmptySquares()
	attackers := pos.attackersTo(to)
	alreadyAttacked := Bitboard(0)
	materialGain[depth] = SEEPieceValues[targetRole]

	for attackers > 0 {
		depth++
		materialGain[depth] = SEEPieceValues[attackerRole] - materialGain[depth-1]

		// Early termination, if we're already losing
		if max(-materialGain[depth-1], materialGain[depth]) < 0 {
			break
		}

		attackers &= ^fromSq
		blockers &= ^fromSq
		alreadyAttacked |= fromSq

		// Find new attackers(by xrays) when removing the already considered pieces into the exchange
		// Diagonal moves could reveal more bishop like attackers
		if attackerRole == Pawn || attackerRole == Bishop || attackerRole == Queen {
			attackers |= bishopAttacks(to, blockers) & diagonalAttackers & ^alreadyAttacked
		}

		// Orthogonal moves could reveal more rook like attackers
		if attackerRole == Rook || attackerRole == Queen {
			attackers |= rookAttacks(to, blockers) & orthogonalAttackers & ^alreadyAttacked
		}

		side = side.Opponent()
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
// Using the square attacked by algorithm - https://www.chessprogramming.org/Square_Attacked_By#Attacks_to_a_Square
func (pos *Position) attackersTo(to int) (attackers Bitboard) {
	toSq := Bitboard(1 << to)
	blocks := ^pos.EmptySquares()

	knights := pos.Bitboards[WhiteKnight] | pos.Bitboards[BlackKnight]
	bishops := pos.Bitboards[WhiteBishop] | pos.Bitboards[BlackBishop]
	rooks := pos.Bitboards[WhiteRook] | pos.Bitboards[BlackRook]
	queens := pos.Bitboards[WhiteQueen] | pos.Bitboards[BlackQueen]
	kings := pos.Bitboards[WhiteKing] | pos.Bitboards[BlackKing]

	return pawnAttacks(&toSq, White)&pos.Bitboards[BlackPawn] |
		pawnAttacks(&toSq, Black)&pos.Bitboards[WhitePawn] |
		knightAttacksTable[to]&knights |
		bishopAttacks(to, blocks)&bishops |
		rookAttacks(to, blocks)&rooks |
		Attacks(Queen, toSq, blocks)&queens |
		kingAttacksTable[to]&kings
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
