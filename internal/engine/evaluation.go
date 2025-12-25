package engine

const (
	// Mobility
	QueenMobilityBonusMg  = 3
	QueenMobilityBonusEg  = 11
	QueenMobilityBase     = 7
	RookMobilityBonusMg   = 10
	RookMobilityBonusEg   = 2
	RookMobilityBase      = 4
	BishopMobilityBonusMg = 10
	BishopMobilityBonusEg = 2
	BishopMobilityBase    = 4
	KnightMobilityBonusMg = 11
	KnightMobilityBonusEg = 0
	KnightMobilityBase    = 2

	// Pawn Structure
	DoubledPawnPenaltyMg  = -1
	DoubledPawnPenaltyEg  = -16
	IsolatedPawnPenaltyMg = -18
	IsolatedPawnPenaltyEg = -7
	BackwardPawnPenaltyMg = -8
	BackwardPawnPenaltyEg = -5

	// Material Adjustment
	BishopPairBonusMg    = 26
	BishopPairBonusEg    = 14
	RookOnOpenFileMg     = 22
	RookOnSemiOpenFileMg = 8

	// King Safety
	KingOnOpenFilePenaltyMg     = -30
	KingOnSemiOpenFilePenaltyMg = -18
)

// PassedPawnsBonusMg contains the bonus for passed pawns on each rank for mg phase
var PassedPawnsBonusMg = [8]int{0, 2, -5, -14, 2, -3, 17, 0}

// PassedPawnsBonusEg contains the bonus for passed pawns on each rank for eg phase
var PassedPawnsBonusEg = [8]int{0, 4, 11, 35, 62, 132, 150, 0}

// Evaluation contains the different evaluation elements of a position
type Evaluation struct {
	mgMaterial      [2]int // White and Black scores
	egMaterial      [2]int
	mgMobility      [2]int
	egMobility      [2]int
	mgPawnStrucutre [2]int
	egPawnStructure [2]int
	phase           int
	// Evaluation tables
	pawnHashTable *PawnHashTable
}

// NewEvaluation returns a new Evaluation
func NewEvaluation(size int) *Evaluation {
	return &Evaluation{
		pawnHashTable: NewPawnHashTable(size),
	}
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

// Evaluate returns the static score of the position
func (pos *Position) Evaluate() int {
	pos.eval.clear()
	blocks := ^pos.EmptySquares()

	enemyPawnsAttacks := [2]Bitboard{
		pawnAttacks(&pos.Bitboards[BlackPawn], Black),
		pawnAttacks(&pos.Bitboards[WhitePawn], White),
	}
	pawns := [2]Bitboard{
		pos.Bitboards[WhitePawn],
		pos.Bitboards[BlackPawn],
	}

	for piece, bb := range pos.Bitboards {
		color := Color(piece / 6)

		for bb > 0 {
			bb := bb.NextBit()
			sq := Bsf(bb)

			switch pieceRole(piece) {
			case King:
				pos.eval.evaluateKing(sq, pawns, color)
			case Queen:
				pos.eval.evaluateQueen(sq, blocks, enemyPawnsAttacks[color], color)
			case Rook:
				pos.eval.evaluateRook(sq, blocks, enemyPawnsAttacks[color], pawns, color)
			case Bishop:
				pos.eval.evaluateBishop(sq, blocks, enemyPawnsAttacks[color], color)
			case Knight:
				pos.eval.evaluateKnight(sq, blocks, enemyPawnsAttacks[color], color)
			case Pawn:
				pos.eval.evaluatePawn(sq, color)
			}
		}
	}

	// Bishop pair bonus
	if pos.Bitboards[WhiteBishop].count() >= 2 {
		pos.eval.mgMaterial[White] += BishopPairBonusMg
		pos.eval.egMaterial[White] += BishopPairBonusEg
	}

	if pos.Bitboards[BlackBishop].count() >= 2 {
		pos.eval.mgMaterial[Black] += BishopPairBonusMg
		pos.eval.egMaterial[Black] += BishopPairBonusEg
	}

	mgSc, egSc, ok := pos.eval.pawnHashTable.probe(pos.PawnHash, pos.Turn)
	if ok {
		pos.eval.mgPawnStrucutre[pos.Turn] = mgSc
		pos.eval.egPawnStructure[pos.Turn] = egSc

	} else {
		pos.eval.evaluatePawnStructure(pos, enemyPawnsAttacks[White], White)
		pos.eval.evaluatePawnStructure(pos, enemyPawnsAttacks[Black], Black)

		mgSc = pos.eval.mgPawnStrucutre[pos.Turn] - pos.eval.mgPawnStrucutre[pos.Turn.Opponent()]
		egSc = pos.eval.egPawnStructure[pos.Turn] - pos.eval.egPawnStructure[pos.Turn.Opponent()]
		pos.eval.pawnHashTable.store(pos.PawnHash, mgSc, egSc, pos.Turn)
	}

	return pos.eval.score(pos.Turn)
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
func (ev *Evaluation) evaluateKing(from int, pawns [2]Bitboard, side Color) {
	piece := pieceColor(King, side)

	// King On open files
	kingFile := from % 8
	for file := kingFile - 1; file >= kingFile+1; file++ {
		if file < 0 || file > 7 {
			continue
		}

		if (pawns[White]|pawns[Black])&Files[file] == 0 {
			if kingFile == file {
				ev.mgMaterial[side] -= KingOnOpenFilePenaltyMg
			} else {
				ev.mgMaterial[side] -= KingOnOpenFilePenaltyMg / 3
			}
		}
		if pawns[side]&Files[file] == 0 && pawns[side.Opponent()]&Files[file] > 0 {
			if kingFile == file {
				ev.mgMaterial[side] -= KingOnSemiOpenFilePenaltyMg
			} else {
				ev.mgMaterial[side] -= KingOnSemiOpenFilePenaltyMg / 3
			}
		}
	}

	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]
}

// evaluateQueen evaluates the score of a queen
func (ev *Evaluation) evaluateQueen(from int, blocks Bitboard, enemyPawnsAttacks Bitboard, side Color) {
	piece := pieceColor(Queen, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	squares := (attacks & ^enemyPawnsAttacks).count()

	ev.mgMobility[side] += (squares - QueenMobilityBase) * QueenMobilityBonusMg
	ev.egMobility[side] += (squares - QueenMobilityBase) * QueenMobilityBonusEg

	ev.phase += 9
}

// evaluateRook evaluates the score of a rook
func (ev *Evaluation) evaluateRook(from int, blocks Bitboard, enemyPawnsAttacks Bitboard, pawns [2]Bitboard, side Color) {
	piece := pieceColor(Rook, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	file := from % 8
	if (pawns[White]|pawns[Black])&Files[file] == 0 {
		ev.mgMaterial[side] += RookOnOpenFileMg
	}

	if pawns[side]&Files[file] == 0 && pawns[side.Opponent()]&Files[file] > 0 {
		ev.mgMaterial[side] += RookOnSemiOpenFileMg
	}

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	squares := (attacks & ^enemyPawnsAttacks).count()

	ev.mgMobility[side] += (squares - RookMobilityBase) * RookMobilityBonusMg
	ev.egMobility[side] += (squares - RookMobilityBase) * RookMobilityBonusEg

	ev.phase += 5
}

// evaluateBishop evaluates the score of a bishop
func (ev *Evaluation) evaluateBishop(from int, blocks Bitboard, enemyPawnsAttacks Bitboard, side Color) {
	piece := pieceColor(Bishop, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	squares := (attacks & ^enemyPawnsAttacks).count()

	ev.mgMobility[side] += (squares - BishopMobilityBase) * BishopMobilityBonusMg
	ev.egMobility[side] += (squares - BishopMobilityBase) * BishopMobilityBonusEg

	ev.phase += 3
}

// evaluateKnight evaluates the score of a knight
func (ev *Evaluation) evaluateKnight(from int, blocks Bitboard, enemyPawnsAttacks Bitboard, side Color) {
	piece := pieceColor(Knight, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
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
func (ev *Evaluation) evaluatePawnStructure(pos *Position, enemyPawnsAttacks Bitboard, side Color) {
	doubledPawns := DoubledPawns(pos, side)
	ev.mgPawnStrucutre[side] += doubledPawns.count() * DoubledPawnPenaltyMg
	ev.egPawnStructure[side] += doubledPawns.count() * DoubledPawnPenaltyEg

	isolatedPawns := IsolatedPawns(pos, side)
	ev.mgPawnStrucutre[side] += isolatedPawns.count() * IsolatedPawnPenaltyMg
	ev.egPawnStructure[side] += isolatedPawns.count() * IsolatedPawnPenaltyEg

	pawns := pos.Bitboards[pieceColor(Pawn, side)]
	backwardPawns := BackwardPawns(pawns, enemyPawnsAttacks, side)
	ev.mgPawnStrucutre[side] += backwardPawns.count() * BackwardPawnPenaltyMg
	ev.egPawnStructure[side] += backwardPawns.count() * BackwardPawnPenaltyEg

	passedPawns := PassedPawns(pawns, pos.Bitboards[pieceColor(Pawn, side.Opponent())], side)
	for passedPawns > 0 {
		fromBB := passedPawns.NextBit()
		sq := Bsf(fromBB)
		rank := sq / 8
		if side == Black {
			rank = 7 - rank
		}

		ev.mgPawnStrucutre[side] += PassedPawnsBonusMg[rank]
		ev.egPawnStructure[side] += PassedPawnsBonusEg[rank]
	}
}

// DoubledPawns returns a bitboard with the files with more than 1 pawn
func DoubledPawns(pos *Position, side Color) Bitboard {
	doubledPawns := Bitboard(0)
	pawns := pos.Bitboards[pieceColor(Pawn, side)]

	for file := range 8 {
		pawnsInFile := pawns & Files[file]
		if pawnsInFile.count() > 1 {
			pawnsInFile.NextBit() // removes one to not double count the penalty
			doubledPawns |= pawnsInFile
		}
	}
	return doubledPawns
}

// IsolatedPawns a bitboard with the isolated pawns for the side
func IsolatedPawns(pos *Position, side Color) Bitboard {
	isolatedPawns := Bitboard(0)
	pawns := pos.Bitboards[pieceColor(Pawn, side)]

	for file := range 8 {
		if isolatedAdjacentFilesMask[file]&pawns == 0 {
			isolatedPawns |= Files[file] & pawns
		}
	}
	return isolatedPawns
}

// BackwardPawns returns a bitboard with the pawns that are backwards
// A backward pawn is a pawn that is not member of own front-attackspans but controlled by a sentry (definition from CPW)
func BackwardPawns(pawns Bitboard, enemyPawnsAttacks Bitboard, side Color) Bitboard {
	stops := pawns << 8
	if side == Black {
		stops = pawns >> 8
	}

	attackFrontSpans := Bitboard(0)
	for pawns > 0 {
		pawn := pawns.NextBit()
		attackFrontSpans |= attacksFrontSpans[side][Bsf(pawn)]
	}

	if side == White {
		return (stops & enemyPawnsAttacks & ^attackFrontSpans) >> 8
	} else {
		return (stops & enemyPawnsAttacks & ^attackFrontSpans) << 8
	}
}

// PassedPawns returns a bitboard with the passed pawns for the side
// A passed pawn is a pawn whose path to promotion is not blocke nor attacked by the enemy pawns
func PassedPawns(alliedPawns Bitboard, enemyPawns Bitboard, side Color) (passedPawns Bitboard) {
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
