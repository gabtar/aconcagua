package aconcagua

// TODO: New Evaluation Tests (Handcraft/'Intuitive' Tunning...) - 50 games 60s+1s TC (Blitz_Testing_4moves.epd)
// ------------------------------------------------------------------------------------------------------------------
// Feature                      |      Implemented            |   Elo improvement (from base PSQT eval - not tunned)
// ------------------------------------------------------------------------------------------------------------------
// 0. New PSQT														✔																	0.0       ( -219.87 vs Aconcagua-v3.0.0)
// 1. Mobility v1                         ✔																 -20.87     (  -99.95 vs Aconcagua-v3.0.0)
// 1. Mobility v2                         ✔																 +200.24    ( -34.68 vs Aconcagua-v3.0.0)
// 2. Pawn structure analysis
//		2.1 Doubled Pawns
//		2.2 Isolated Pawn
//		2.3 Backward Pawns
//		2.4 Passed Pawns
// 3. King safety
//		3.1 Pawn shield
//		3.2 Pawn storm
//    3.3 Open file
//    3.4 King attackers
// 4. Center control
// 5. Open files
// 6. Bishop pair bonus

// Evaluation is a vector containing the diferent evaluations of the position
type Evaluation struct {
	mgMaterial [2]int // PSQT + material weight [white, black]
	egMaterial [2]int
	mgMobility [2]int
	egMobility [2]int
	phase      int
}

// Evaluate returns the evaluation of the position
func Evaluate(pos *Position) int {
	ev := Evaluation{}

	for p, bb := range pos.Bitboards {
		color := p / 6

		for bb > 0 {
			fromBB := bb.NextBit()
			switch pieceRole(p) {
			case King:
				ev.evaluateKing(pos, &fromBB, Color(color))
			case Queen:
				ev.evaluateQueen(pos, &fromBB, Color(color))
			case Rook:
				ev.evaluateRook(pos, &fromBB, Color(color))
			case Bishop:
				ev.evaluateBishop(pos, &fromBB, Color(color))
			case Knight:
				ev.evaluateKnight(pos, &fromBB, Color(color))
			case Pawn:
				ev.evaluatePawn(pos, &fromBB, Color(color))
			}
		}
	}

	return ev.score(pos.Turn)
}

// score returns the score relative to the side passed
func (ev *Evaluation) score(side Color) int {
	opponent := side.Opponent()
	mgPhase := min(ev.phase, 24)
	egPhase := 24 - mgPhase

	mg := ev.mgMaterial[side] - ev.mgMaterial[opponent] + ev.mgMobility[side] - ev.mgMobility[opponent]
	eg := ev.egMaterial[side] - ev.egMaterial[opponent] + ev.egMobility[side] - ev.egMobility[opponent]

	return (mg*mgPhase + eg*egPhase) / 24
}

// evaluateKing returns the middlegame and endgame score of the king in the position
func (ev *Evaluation) evaluateKing(pos *Position, king *Bitboard, side Color) (mgScore int, egScore int) {
	sq := Bsf(*king)

	ev.mgMaterial[side] += middlegamePiecesScore[pieceColor(King, side)][sq]
	ev.egMaterial[side] += endgamePiecesScore[pieceColor(King, side)][sq]
	return
}

// evaluateQueen returns the middlegame and endgame score of the queen in the position
func (ev *Evaluation) evaluateQueen(pos *Position, queen *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Queen, side)
	sq := Bsf(*queen)

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	attacksCount := Attacks(piece, *queen, ^pos.EmptySquares()).count()
	ev.mgMobility[side] += (attacksCount - 7) * 5
	ev.egMobility[side] += (attacksCount - 7) * 3

	ev.phase += 4

	return
}

// evaluateRook returns the middlegame and endgame score of the rook in the position
func (ev *Evaluation) evaluateRook(pos *Position, rook *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Rook, side)
	sq := Bsf(*rook)

	// TODO: open files bonus???
	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	attacksCount := Attacks(piece, *rook, ^pos.EmptySquares()).count()
	ev.mgMobility[side] += (attacksCount - 5) * 3
	ev.egMobility[side] += (attacksCount - 5) * 3

	ev.phase += 2

	return
}

// evaluateBishop returns the middlegame and endgame score of the bishop in the position
func (ev *Evaluation) evaluateBishop(pos *Position, bishop *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Bishop, side)
	sq := Bsf(*bishop)

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	attacksCount := Attacks(piece, *bishop, ^pos.EmptySquares()).count()
	ev.mgMobility[side] += (attacksCount - 5) * 3
	ev.egMobility[side] += (attacksCount - 5) * 4

	ev.phase += 1

	return
}

// evaluateKnight returns the middlegame and endgame score of the knight in the position
func (ev *Evaluation) evaluateKnight(pos *Position, knight *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Knight, side)
	sq := Bsf(*knight)

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	// TODO: extract to function?
	attacksCount := Attacks(piece, *knight, ^pos.EmptySquares()).count()
	ev.mgMobility[side] += (attacksCount - 3) * 3
	ev.egMobility[side] += (attacksCount - 3) * 4

	ev.phase += 1

	return
}

// evaluatePawn returns the middlegame and endgame score of the pawn in the position
func (ev *Evaluation) evaluatePawn(pos *Position, pawn *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Pawn, side)
	sq := Bsf(*pawn)

	// if isDoubled(pawn, pos, side) {
	// 	ev.mgMaterial[side] -= 5
	// 	ev.egMaterial[side] -= 10
	// }
	//
	// if isIsolated(pawn, pos, side) {
	// 	ev.mgMaterial[side] -= 10
	// 	ev.egMaterial[side] -= 10
	// }
	//
	// if isBackward(pawn, pos, side) {
	// 	ev.mgMaterial[side] -= 5
	// 	ev.egMaterial[side] -= 10
	// }
	//
	// if isPassed(pawn, pos, side) {
	// 	ev.mgMaterial[side] += 20
	// 	ev.egMaterial[side] += 30
	// }

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	return
}

// isDoubled returns true if the pawn is doubled
func isDoubled(pawn *Bitboard, pos *Position, side Color) bool {
	file := Bsf(*pawn) % 8
	filePawns := files[file] & pos.Bitboards[pieceColor(Pawn, side)]

	if filePawns.count() > 1 {
		return true
	}
	return false
}

// isIsolated returns true if the pawn is isolated
func isIsolated(pawn *Bitboard, pos *Position, side Color) bool {
	file := Bsf(*pawn) % 8
	adjacentFiles := Bitboard(0)

	if file < 7 {
		adjacentFiles |= files[file+1]
	}
	if file > 0 {
		adjacentFiles |= files[file-1]
	}

	if adjacentFiles&pos.Bitboards[pieceColor(Pawn, side)] == 0 {
		return true
	}
	return false
}

// isBackward returns true if the pawn is backward
func isBackward(pawn *Bitboard, pos *Position, side Color) bool {
	// From wikipedia
	// In chess, a backward pawn is a pawn that is behind all pawns of the same color on the adjacent files and cannot be safely advanced
	file := Bsf(*pawn) % 8
	rank := Bsf(*pawn) / 8
	backwardDirection := South
	backwardLeftSq, backwardRightSq := rank*8+file-1, rank*8+file+1
	upSq := *pawn << 8
	if side == Black {
		backwardLeftSq, backwardRightSq = rank*8+file-1, rank*8+file+1
		upSq = *pawn >> 8
		backwardDirection = North
	}

	alliedPawns := pos.Bitboards[pieceColor(Pawn, side)]
	adjacentBackward := Bitboard(0)
	if file < 7 {
		adjacentBackward |= rayAttacks[backwardDirection][backwardRightSq]
	}
	if file > 0 {
		adjacentBackward |= rayAttacks[backwardDirection][backwardLeftSq]
	}

	enemyPawns := pos.Bitboards[pieceColor(Pawn, side.Opponent())]
	enemyPawnsAttacks := Attacks(pieceColor(Pawn, side.Opponent()), enemyPawns, ^pos.EmptySquares())

	if adjacentBackward&alliedPawns == 0 && enemyPawnsAttacks&upSq > 0 {
		return true
	}

	return false
}

// isPassed returns true if the pawn is passed
func isPassed(pawn *Bitboard, pos *Position, side Color) bool {
	file := Bsf(*pawn) % 8
	if file == 0 || file == 7 {
		return false
	}

	direction := North
	upLeftSq, upRightSq := *pawn<<9, *pawn<<7
	if side == Black {
		direction = South
		upLeftSq, upRightSq = *pawn>>7, *pawn>>9
	}

	adjacentUp := Bitboard(0)
	adjacentUp |= rayAttacks[direction][Bsf(*pawn)]
	if file < 7 {
		adjacentUp |= rayAttacks[direction][Bsf(upRightSq)]
	}
	if file > 0 {
		adjacentUp |= rayAttacks[direction][Bsf(upLeftSq)]
	}

	enemyPawns := pos.Bitboards[pieceColor(Pawn, side.Opponent())]
	enemyPawnsAttacks := Attacks(pieceColor(Pawn, side.Opponent()), enemyPawns, ^pos.EmptySquares())

	if adjacentUp&enemyPawns == 0 && enemyPawnsAttacks&adjacentUp == 0 {
		return true
	}
	return false
}
