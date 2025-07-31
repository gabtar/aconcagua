package aconcagua

const (
	BackwardPawnPenaltyMg = -3
	BackwardPawnPenaltyEg = -8
	DoubledPawnPenaltyMg  = -8
	DoubledPawnPenaltyEg  = -12
	IsolatedPawnPenaltyMg = -10
	IsolatedPawnPenaltyEg = -15
)

// Evaluation is a vector containing the diferent evaluations of the position
type Evaluation struct {
	mgMaterial [2]int // PSQT + material weight [white, black]
	egMaterial [2]int
	mgMobility [2]int
	egMobility [2]int
	// TODO: use a pawn hash table to eval pawn structure/so it can be cached?
	// mgPawnStrucutre [2]int
	// egPawnStructure [2]int
	phase int
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

	// TODO: use a separate pawn structure evaluation
	// Total Backwards pawns for each side, eg:
	// ev.pawnStructure = evaluatePawnStructure(pos, White)
	// ev.pawnStructure = evaluatePawnStructure(pos, Black)

	// Doubled Pawns
	// ev.mgMaterial[White] += doubledPawns(pos, White).count() * DoubledPawnPenaltyMg
	// ev.egMaterial[Black] += doubledPawns(pos, Black).count() * DoubledPawnPenaltyEg

	// Isolated Pawns
	ev.mgMaterial[White] += isolatedPawns(pos, White).count() * IsolatedPawnPenaltyMg
	ev.egMaterial[Black] += isolatedPawns(pos, Black).count() * IsolatedPawnPenaltyEg

	// Backwards Pawns
	// whiteBackwardPawns := backwardsPawns(pos, White)
	// blackBackwardPawns := backwardsPawns(pos, Black)
	// ev.mgMaterial[White] += whiteBackwardPawns.count() * BackwardPawnPenaltyMg
	// ev.egMaterial[Black] += blackBackwardPawns.count() * BackwardPawnPenaltyEg

	// Passed Pawns
	// whitePassedPawns := passedPawns(pos, White)
	// blackPassedPawns := passedPawns(pos, Black)
	//
	// // Passed pawn bonus depending on the ranks to go to promotion
	// passedPawnBonus := [8]int{0, 0, 10, 20, 30, 60, 100, 0}
	// for whitePassedPawns > 0 {
	// 	fromBB := whitePassedPawns.NextBit()
	// 	sq := Bsf(fromBB)
	// 	ev.mgMaterial[White] += passedPawnBonus[sq/8]
	// 	ev.egMaterial[White] += passedPawnBonus[sq/8]
	// }
	//
	// for blackPassedPawns > 0 {
	// 	fromBB := blackPassedPawns.NextBit()
	// 	sq := Bsf(fromBB)
	// 	ev.mgMaterial[Black] += passedPawnBonus[7-sq/8]
	// 	ev.egMaterial[Black] += passedPawnBonus[7-sq/8]
	// }

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
func (ev *Evaluation) evaluateKing(pos *Position, king *Bitboard, side Color) {
	sq := Bsf(*king)

	ev.mgMaterial[side] += middlegamePiecesScore[pieceColor(King, side)][sq]
	ev.egMaterial[side] += endgamePiecesScore[pieceColor(King, side)][sq]

	return
}

// evaluateQueen returns the middlegame and endgame score of the queen in the position
func (ev *Evaluation) evaluateQueen(pos *Position, queen *Bitboard, side Color) {
	piece := pieceColor(Queen, side)
	sq := Bsf(*queen)

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	// attacksCount := Attacks(piece, *queen, ^pos.EmptySquares()).count()
	// ev.mgMobility[side] += (attacksCount - 7) * 5
	// ev.egMobility[side] += (attacksCount - 7) * 3

	ev.phase += 4

	return
}

// evaluateRook returns the middlegame and endgame score of the rook in the position
func (ev *Evaluation) evaluateRook(pos *Position, rook *Bitboard, side Color) {
	piece := pieceColor(Rook, side)
	sq := Bsf(*rook)

	// TODO: open files bonus???
	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	// attacksCount := Attacks(piece, *rook, ^pos.EmptySquares()).count()
	// ev.mgMobility[side] += (attacksCount - 5) * 3
	// ev.egMobility[side] += (attacksCount - 5) * 3

	ev.phase += 2

	return
}

// evaluateBishop returns the middlegame and endgame score of the bishop in the position
func (ev *Evaluation) evaluateBishop(pos *Position, bishop *Bitboard, side Color) {
	piece := pieceColor(Bishop, side)
	sq := Bsf(*bishop)

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	// attacksCount := Attacks(piece, *bishop, ^pos.EmptySquares()).count()
	// ev.mgMobility[side] += (attacksCount - 5) * 3
	// ev.egMobility[side] += (attacksCount - 5) * 4

	ev.phase += 1

	return
}

// evaluateKnight returns the middlegame and endgame score of the knight in the position
func (ev *Evaluation) evaluateKnight(pos *Position, knight *Bitboard, side Color) {
	piece := pieceColor(Knight, side)
	sq := Bsf(*knight)

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	// TODO: extract to function?
	// attacksCount := Attacks(piece, *knight, ^pos.EmptySquares()).count()
	// ev.mgMobility[side] += (attacksCount - 3) * 3
	// ev.egMobility[side] += (attacksCount - 3) * 4

	ev.phase += 1

	return
}

// evaluatePawn returns the middlegame and endgame score of the pawn in the position
func (ev *Evaluation) evaluatePawn(pos *Position, pawn *Bitboard, side Color) {
	piece := pieceColor(Pawn, side)
	sq := Bsf(*pawn)

	// if isIsolated(pawn, pos, side) {
	// 	ev.mgMaterial[side] -= 5
	// 	ev.egMaterial[side] -= 15
	// }

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	return
}

// doubledPawns returns a bitboard with the files with more than 1 pawn
func doubledPawns(pos *Position, side Color) Bitboard {
	doubledPawns := Bitboard(0)
	pawns := pos.Bitboards[pieceColor(Pawn, side)]

	for file := range 8 {
		pawnsInFile := pawns & files[file]
		if pawnsInFile.count() > 1 {
			pawnsInFile.NextBit() // remove one to not double count
			doubledPawns |= pawnsInFile
		}
	}
	return doubledPawns
}

// isolatedPawns with the isolated pawns for the side
func isolatedPawns(pos *Position, side Color) Bitboard {
	isolatedPawns := Bitboard(0)
	pawns := pos.Bitboards[pieceColor(Pawn, side)]

	for file := range 8 {
		if IsolatedAdjacentFilesMask[file]&pawns == 0 {
			isolatedPawns |= files[file] & pawns
		}
	}
	return isolatedPawns
}

// IsolatedAdjacentFilesMask contains the adjacent files of a pawn to test if it is isolated
var IsolatedAdjacentFilesMask = [8]Bitboard{
	files[1],
	files[0] | files[2],
	files[1] | files[3],
	files[2] | files[4],
	files[3] | files[5],
	files[4] | files[6],
	files[5] | files[7],
	files[6],
}

// attacksFrontSpans is a precalculated table containing the front attack spans for each square
// front attack spans includes the attacked squares itself, thus it is like a fill of attacked squares in the appropriate direction
// front attack span for pawn on d4
// . . 1 . 1 . . .
// . . 1 . 1 . . .
// . . 1 . 1 . . .
// . . 1 . 1 . . .
// . . . w . . . .
// . . . . . . . .
// . . . . . . . .
// . . . . . . . .
var attacksFrontSpans [2][64]Bitboard

func generateAttacksFrontSpans() (attacksFrontSpans [2][64]Bitboard) {

	for sq := range 64 {
		file, rank := sq%8, sq/8
		eastFront, westFront := rank*8+file+1, rank*8+file-1

		if file < 7 {
			attacksFrontSpans[White][sq] |= rayAttacks[North][eastFront]
			attacksFrontSpans[Black][sq] |= rayAttacks[South][eastFront]
		}
		if file > 0 {
			attacksFrontSpans[White][sq] |= rayAttacks[North][westFront]
			attacksFrontSpans[Black][sq] |= rayAttacks[South][westFront]
		}

	}

	return
}

// backwardsPawns returns a bitboard with the pawns that are backwards
// A backward pawn is a pawn that is not member of own front-attackspans but controlled by a sentry (definition from CPW)
func backwardsPawns(pos *Position, side Color) Bitboard {
	pawns := pos.Bitboards[pieceColor(Pawn, side)]
	stops := pawns << 8
	if side == Black {
		stops = pawns >> 8
	}

	attackFrontSpans := Bitboard(0)
	for pawns > 0 {
		pawn := pawns.NextBit()
		attackFrontSpans |= attacksFrontSpans[side][Bsf(pawn)]
	}

	enemyPawns := pos.Bitboards[pieceColor(Pawn, side.Opponent())]
	enemyPawnsAttacks := pawnAttacks(&enemyPawns, side.Opponent())

	if side == White {
		return (stops & enemyPawnsAttacks & ^attackFrontSpans) >> 8
	} else {
		return (stops & enemyPawnsAttacks & ^attackFrontSpans) << 8
	}
}

// passedPawns returns a bitboard with the passed pawns for the side
// A passed pawn is a pawn whose path to promotion is not blocke nor attacked by the enemy pawns
func passedPawns(pos *Position, side Color) (passedPawns Bitboard) {
	alliedPawns := pos.Bitboards[pieceColor(Pawn, side)]
	enemyPawns := pos.Bitboards[pieceColor(Pawn, side.Opponent())]
	direction := North
	if side == Black {
		direction = South
	}

	for alliedPawns > 0 {
		pawn := alliedPawns.NextBit()
		frontAndAdjacentSquares := attacksFrontSpans[side][Bsf(pawn)] | rayAttacks[direction][Bsf(pawn)]

		if frontAndAdjacentSquares&enemyPawns == 0 {
			passedPawns |= pawn
		}
	}

	return
}
