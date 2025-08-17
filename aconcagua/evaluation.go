package aconcagua

const (
	QueenMobilityBonusMg  = 3
	QueenMobilityBonusEg  = 1
	QueenMobilityBase     = 7
	RookMobilityBonusMg   = 2
	RookMobilityBonusEg   = 2
	RookMobilityBase      = 4
	BishopMobilityBonusMg = 2
	BishopMobilityBonusEg = 3
	BishopMobilityBase    = 4
	KnightMobilityBonusMg = 1
	KnightMobilityBonusEg = 2
	KnightMobilityBase    = 2

	// NOTE: 500 Games 12+0.1s TC -> 17.39 +/- 26.04. Passed pawns disabled...
	// Try w/ diferent values and play matchs against v3.1.0
	// nps are a bit lower than v3.1.0. Maybe later use a Pawn Hash Table to improve performance
	DoubledPawnPenaltyMg  = -8
	DoubledPawnPenaltyEg  = -12
	IsolatedPawnPenaltyMg = -5
	IsolatedPawnPenaltyEg = -15
	BackwardPawnPenaltyMg = -3
	BackwardPawnPenaltyEg = -8
)

// Evaluation contains the diferent evaluation elements of a position
type Evaluation struct {
	mgMaterial      [2]int // White and Black scores
	egMaterial      [2]int
	mgMobility      [2]int
	egMobility      [2]int
	mgPawnStrucutre [2]int
	egPawnStructure [2]int
	phase           int
}

// NewEvaluation returns a new Evaluation
func NewEvaluation() *Evaluation {
	return &Evaluation{}
}

// clear clears the evaluation
func (ev *Evaluation) clear() {
	ev.mgMaterial = [2]int{0, 0}
	ev.egMaterial = [2]int{0, 0}
	ev.mgMobility = [2]int{0, 0}
	ev.egMobility = [2]int{0, 0}
	ev.mgPawnStrucutre = [2]int{0, 0}
	ev.egPawnStructure = [2]int{0, 0}
	ev.phase = 0
}

// score returns the score relative to the side
func (ev *Evaluation) score(side Color) int {
	opponent := side.Opponent()

	mg := ev.mgMaterial[side] - ev.mgMaterial[opponent]
	eg := ev.egMaterial[side] - ev.egMaterial[opponent]
	mg += ev.mgMobility[side] - ev.mgMobility[opponent]
	eg += ev.egMobility[side] - ev.egMobility[opponent]
	mg += ev.mgPawnStrucutre[side] - ev.mgPawnStrucutre[opponent]
	eg += ev.egPawnStructure[side] - ev.egPawnStructure[opponent]

	mgPhase := min(ev.phase, 62)
	egPhase := 62 - mgPhase

	return (mg*mgPhase + eg*egPhase) / 62
}

// evaluateKing evaluates the score of a king
func (ev *Evaluation) evaluateKing(from int, side Color) {
	piece := pieceColor(King, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]
}

// evaluateQueen evaluates the score of a queen
func (ev *Evaluation) evaluateQueen(from int, blocks Bitboard, enemyPawns Bitboard, side Color) {
	piece := pieceColor(Queen, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	enemyPawnsAttacks := pawnAttacks(&enemyPawns, side.Opponent())
	squares := (attacks & ^enemyPawnsAttacks).count()

	ev.mgMobility[side] += (squares - QueenMobilityBase) * QueenMobilityBonusMg
	ev.egMobility[side] += (squares - QueenMobilityBase) * QueenMobilityBonusEg

	ev.phase += 9
}

// evaluateRook evaluates the score of a rook
func (ev *Evaluation) evaluateRook(from int, blocks Bitboard, enemyPawns Bitboard, side Color) {
	piece := pieceColor(Rook, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	enemyPawnsAttacks := pawnAttacks(&enemyPawns, side.Opponent())
	squares := (attacks & ^enemyPawnsAttacks).count()

	ev.mgMobility[side] += (squares - RookMobilityBase) * RookMobilityBonusMg
	ev.egMobility[side] += (squares - RookMobilityBase) * RookMobilityBonusEg

	ev.phase += 5
}

// evaluateBishop evaluates the score of a bishop
func (ev *Evaluation) evaluateBishop(from int, blocks Bitboard, enemyPawns Bitboard, side Color) {
	piece := pieceColor(Bishop, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	enemyPawnsAttacks := pawnAttacks(&enemyPawns, side.Opponent())
	squares := (attacks & ^enemyPawnsAttacks).count()

	ev.mgMobility[side] += (squares - BishopMobilityBase) * BishopMobilityBonusMg
	ev.egMobility[side] += (squares - BishopMobilityBase) * BishopMobilityBonusEg

	ev.phase += 3
}

// evaluateKnight evaluates the score of a knight
func (ev *Evaluation) evaluateKnight(from int, blocks Bitboard, enemyPawns Bitboard, side Color) {
	piece := pieceColor(Knight, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	enemyPawnsAttacks := pawnAttacks(&enemyPawns, side.Opponent())
	squares := (attacks & ^enemyPawnsAttacks).count()

	ev.mgMobility[side] += (squares - KnightMobilityBase) * KnightMobilityBonusMg
	ev.egMobility[side] += (squares - KnightMobilityBase) * KnightMobilityBonusEg

	ev.phase += 3
}

// evaluatePawn evaluates the score of a pawn
func (ev *Evaluation) evaluatePawn(from int, side Color) {
	piece := pieceColor(Pawn, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]
}

// evaluatePawnStructure evaluates the pawn structure for the side in the position passed
func (ev *Evaluation) evaluatePawnStructure(pos *Position, side Color) {
	doubledPawns := doubledPawns(pos, side)
	ev.mgPawnStrucutre[side] += doubledPawns.count() * DoubledPawnPenaltyMg
	ev.egPawnStructure[side] += doubledPawns.count() * DoubledPawnPenaltyEg

	isolatedPawns := isolatedPawns(pos, side)
	ev.mgPawnStrucutre[side] += isolatedPawns.count() * IsolatedPawnPenaltyMg
	ev.egPawnStructure[side] += isolatedPawns.count() * IsolatedPawnPenaltyEg

	backwardPawns := backwardPawns(pos, side)
	ev.mgPawnStrucutre[side] += backwardPawns.count() * BackwardPawnPenaltyMg
	ev.egPawnStructure[side] += backwardPawns.count() * BackwardPawnPenaltyEg

	// passedPawns := passedPawns(pos, side)
	// passedPawnBonus := [8]int{0, 0, 10, 20, 30, 60, 100, 0}
	// for passedPawns > 0 {
	// 	fromBB := passedPawns.NextBit()
	// 	sq := Bsf(fromBB)
	// 	rank := sq / 8
	// 	if side == Black {
	// 		rank = 7 - rank
	// 	}
	//
	// 	ev.mgPawnStrucutre[side] += passedPawnBonus[rank]
	// 	ev.egPawnStructure[side] += passedPawnBonus[rank]
	// }
}

// doubledPawns returns a bitboard with the files with more than 1 pawn
func doubledPawns(pos *Position, side Color) Bitboard {
	doubledPawns := Bitboard(0)
	pawns := pos.Bitboards[pieceColor(Pawn, side)]

	for file := range 8 {
		pawnsInFile := pawns & files[file]
		if pawnsInFile.count() > 1 {
			pawnsInFile.NextBit() // removes one to not double count the penalty
			doubledPawns |= pawnsInFile
		}
	}
	return doubledPawns
}

// isolatedPawns a bitboard with the isolated pawns for the side
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

// backwardPawns returns a bitboard with the pawns that are backwards
// A backward pawn is a pawn that is not member of own front-attackspans but controlled by a sentry (definition from CPW)
func backwardPawns(pos *Position, side Color) Bitboard {
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

func (pos *Position) Evaluate() int {
	pos.evaluation.clear()
	blocks := ^pos.EmptySquares()

	for piece, bb := range pos.Bitboards {
		color := Color(piece / 6)

		for bb > 0 {
			bb := bb.NextBit()
			sq := Bsf(bb)

			switch pieceRole(piece) {
			case King:
				pos.evaluation.evaluateKing(sq, color)
			case Queen:
				pos.evaluation.evaluateQueen(sq, blocks, pos.Bitboards[pieceColor(Pawn, color.Opponent())], color)
			case Rook:
				pos.evaluation.evaluateRook(sq, blocks, pos.Bitboards[pieceColor(Pawn, color.Opponent())], color)
			case Bishop:
				pos.evaluation.evaluateBishop(sq, blocks, pos.Bitboards[pieceColor(Pawn, color.Opponent())], color)
			case Knight:
				pos.evaluation.evaluateKnight(sq, blocks, pos.Bitboards[pieceColor(Pawn, color.Opponent())], color)
			case Pawn:
				pos.evaluation.evaluatePawn(sq, color)
			}
		}
	}

	pos.evaluation.evaluatePawnStructure(pos, White)
	pos.evaluation.evaluatePawnStructure(pos, Black)

	return pos.evaluation.score(pos.Turn)
}
