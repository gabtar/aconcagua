package engine

const (
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

	KnightOutpostBonusMg = 15
	KnightOutpostBonusEg = 5
	BishopOutpostBonusMg = 10
	BishopOutpostBonusEg = 4

	// King Safety
	KingOnOpenFilePenaltyMg       = -30
	KingOnSemiOpenFilePenaltyMg   = -18
	KingNearOpenFilePenaltyMg     = -10
	KingNearSemiOpenFilePenaltyMg = -6

	KnightAttackWeight = 3
	BishopAttackWeight = 3
	RookAttackWeight   = 5
	QueenAttackWeight  = 9

	TempoBonus = 5
)

var (
	// Queen Mobility mg/eg contains the bonus for queen mobility
	QueenMobilityMg = [28]int{-21, -18, -15, -12, -9, -6, -3, 0, 3, 6, 9, 12, 15, 18, 21, 24, 27, 30, 33, 36, 39, 42, 45, 48, 51, 54, 57, 60}
	QueenMobilityEg = [28]int{-77, -66, -55, -44, -33, -22, -11, 0, 11, 22, 33, 44, 55, 66, 77, 88, 99, 110, 121, 132, 143, 154, 165, 176, 187, 198, 209, 220}

	// Rook Mobility mg/eg contains the bonus for rook mobility
	RookMobilityMg = [15]int{-40, -30, -20, -10, 0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	RookMobilityEg = [15]int{-8, -6, -4, -2, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20}

	// Bishop Mobility mg/eg contains the bonus for bishop mobility
	BishopMobilityMg = [14]int{-40, -30, -20, -10, 0, 10, 20, 30, 40, 50, 60, 70, 80, 90}
	BishopMobilityEg = [14]int{-8, -6, -4, -2, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18}

	// KnightMobility mg/eg contains the bonus for knight mobility
	KnightMobilityMg = [9]int{-22, -11, 0, 11, 22, 33, 44, 55, 66}
	KnightMobilityEg = [9]int{0, 0, 0, 0, 0, 0, 0, 0, 0}

	// PawnShieldBonusMg contains the bonus for pawn shields for mg phase based on rank distance to the king
	PawnShieldBonusMg = [3]int{15, 10, 5}

	// PassedPawnsBonus mg/eg contains the bonus for passed pawns
	PassedPawnsBonusMg = [8]int{0, -5, -9, -11, 15, 11, 11, 0}
	PassedPawnsBonusEg = [8]int{0, 10, 14, 37, 60, 119, 134, 0}

	// OutpostsRanks contains the bitboard mask for ranks that are considered outposts
	OutpostsRanks = [2]Bitboard{
		Ranks[3] | Ranks[4] | Ranks[5],
		Ranks[2] | Ranks[3] | Ranks[4],
	}

	// KingSafetyTable contains the penalties for king safety based on the number wheigted attacks to the king
	KingSafetyTable = [50]int{
		0, 0, 1, 2, 3, 5, 7, 9, 12, 15,
		18, 22, 26, 30, 35, 40, 45, 51, 57, 64,
		71, 79, 88, 97, 107, 118, 130, 142, 155, 168,
		182, 197, 213, 230, 248, 267, 287, 308, 330, 353,
		377, 400, 400, 400, 400, 400, 400, 400, 400, 400,
	}
)

// Evaluation contains the different evaluation elements of a position
type Evaluation struct {
	mgMaterial         [2]int // White and Black scores
	egMaterial         [2]int
	mgMobility         [2]int
	egMobility         [2]int
	mgPawnStrucutre    [2]int
	egPawnStructure    [2]int
	mgKingSafety       [2]int
	kingAttackersCount [2]int
	kingAttackWeight   [2]int
	phase              int
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
	ev.mgKingSafety = [2]int{0, 0}
	ev.mgPawnStrucutre = [2]int{0, 0}
	ev.egPawnStructure = [2]int{0, 0}
	ev.kingAttackersCount = [2]int{0, 0}
	ev.kingAttackWeight = [2]int{0, 0}
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
	outpostSquares := [2]Bitboard{
		OutpostSquares(pawns[White], pawns[Black], White),
		OutpostSquares(pawns[Black], pawns[White], Black),
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
				pos.eval.evaluateQueen(sq, blocks, enemyPawnsAttacks[color], pos.KingPosition(color.Opponent()), color)
			case Rook:
				pos.eval.evaluateRook(sq, blocks, enemyPawnsAttacks[color], pawns, pos.KingPosition(color.Opponent()), color)
			case Bishop:
				pos.eval.evaluateBishop(sq, blocks, enemyPawnsAttacks[color], pos.KingPosition(color.Opponent()), outpostSquares[color], color)
			case Knight:
				pos.eval.evaluateKnight(sq, blocks, enemyPawnsAttacks[color], pos.KingPosition(color.Opponent()), outpostSquares[color], color)
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

	// TempoBonus
	pos.eval.mgMaterial[pos.Turn] += TempoBonus
	pos.eval.egMaterial[pos.Turn] += TempoBonus

	// Pawn Structure evaluation
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

	// Apply King Safety Penalties to opponent only if there are at least 2 attackers
	if ev.kingAttackersCount[side] >= 2 {
		weight := min(ev.kingAttackWeight[side], 49)
		ev.mgKingSafety[side] += KingSafetyTable[weight]
	}

	if ev.kingAttackersCount[opponent] >= 2 {
		weight := min(ev.kingAttackWeight[opponent], 49)
		ev.mgKingSafety[opponent] += KingSafetyTable[weight]
	}
	mg += ev.mgKingSafety[side] - ev.mgKingSafety[opponent]

	mgPhase := min(ev.phase, 62)
	egPhase := 62 - mgPhase
	return (mg*mgPhase + eg*egPhase) / 62
}

// evaluateKing evaluates the score of a king
func (ev *Evaluation) evaluateKing(from int, pawns [2]Bitboard, side Color) {
	piece := pieceColor(King, side)

	kingFile := from % 8
	kingRank := from / 8
	for file := kingFile - 1; file <= kingFile+1; file++ {
		if file < 0 || file > 7 {
			continue
		}

		// Evaluate pawn Shield
		pawnsInFile := pawns[side] & Files[file]
		if pawnsInFile > 0 {
			rankIncrement := 1
			if side == Black {
				rankIncrement = -1
			}
			for i := range 3 {
				rank := (kingRank + rankIncrement) + i*rankIncrement
				if rank < 0 || rank > 7 {
					break
				}

				if pawnsInFile&Ranks[rank] > 0 {
					ev.mgKingSafety[side] += PawnShieldBonusMg[i]
					break // only count the first pawn
				}
			}
		}

		// Evaluate King on Open/Semi Open File
		if (pawns[White]|pawns[Black])&Files[file] == 0 {
			if kingFile == file {
				ev.mgKingSafety[side] += KingOnOpenFilePenaltyMg
			} else {
				ev.mgKingSafety[side] += KingNearOpenFilePenaltyMg
			}
		} else if pawns[side]&Files[file] == 0 && pawns[side.Opponent()]&Files[file] > 0 {
			if kingFile == file {
				ev.mgKingSafety[side] += KingOnSemiOpenFilePenaltyMg
			} else {
				ev.mgKingSafety[side] += KingNearSemiOpenFilePenaltyMg
			}
		}
	}

	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]
}

// evaluateQueen evaluates the score of a queen
func (ev *Evaluation) evaluateQueen(from int, blocks Bitboard, enemyPawnsAttacks Bitboard, enemyKing Bitboard, side Color) {
	piece := pieceColor(Queen, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	squares := (attacks & ^enemyPawnsAttacks).count()

	enemyKingZone := KingZone[Bsf(enemyKing)]
	if attacks&enemyKingZone != 0 {
		ev.kingAttackersCount[side]++
		ev.kingAttackWeight[side] += QueenAttackWeight * (attacks & enemyKingZone).count()
	}

	ev.mgMobility[side] += QueenMobilityMg[squares]
	ev.egMobility[side] += QueenMobilityEg[squares]

	ev.phase += 9
}

// evaluateRook evaluates the score of a rook
func (ev *Evaluation) evaluateRook(from int, blocks Bitboard, enemyPawnsAttacks Bitboard, pawns [2]Bitboard, enemyKing Bitboard, side Color) {
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

	enemyKingZone := KingZone[Bsf(enemyKing)]
	if attacks&enemyKingZone != 0 {
		ev.kingAttackersCount[side]++
		ev.kingAttackWeight[side] += RookAttackWeight * (attacks & enemyKingZone).count()
	}

	ev.mgMobility[side] += RookMobilityMg[squares]
	ev.egMobility[side] += RookMobilityEg[squares]

	ev.phase += 5
}

// evaluateBishop evaluates the score of a bishop
func (ev *Evaluation) evaluateBishop(from int, blocks Bitboard, enemyPawnsAttacks Bitboard, enemyKing Bitboard, outpostMask Bitboard, side Color) {
	piece := pieceColor(Bishop, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	if outpostMask&bitboardFromIndex(from) > 0 {
		ev.mgMaterial[side] += BishopOutpostBonusMg
		ev.egMaterial[side] += BishopOutpostBonusEg
	}

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	squares := (attacks & ^enemyPawnsAttacks).count()

	enemyKingZone := KingZone[Bsf(enemyKing)]
	if attacks&enemyKingZone != 0 {
		ev.kingAttackersCount[side]++
		ev.kingAttackWeight[side] += BishopAttackWeight * (attacks & enemyKingZone).count()
	}

	ev.mgMobility[side] += BishopMobilityMg[squares]
	ev.egMobility[side] += BishopMobilityEg[squares]

	ev.phase += 3
}

// evaluateKnight evaluates the score of a knight
func (ev *Evaluation) evaluateKnight(from int, blocks Bitboard, enemyPawnsAttacks Bitboard, enemyKing Bitboard, outpostMask Bitboard, side Color) {
	piece := pieceColor(Knight, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	if outpostMask&bitboardFromIndex(from) > 0 {
		ev.mgMaterial[side] += KnightOutpostBonusMg
		ev.egMaterial[side] += KnightOutpostBonusEg
	}

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	squares := (attacks & ^enemyPawnsAttacks).count()

	enemyKingZone := KingZone[Bsf(enemyKing)]
	if attacks&enemyKingZone != 0 {
		ev.kingAttackersCount[side]++
		ev.kingAttackWeight[side] += KnightAttackWeight * (attacks & enemyKingZone).count()
	}

	ev.mgMobility[side] += KnightMobilityMg[squares]
	ev.egMobility[side] += KnightMobilityEg[squares]

	ev.phase += 3
}

// evaluatePawn evaluates the score of a pawn
func (ev *Evaluation) evaluatePawn(from int, side Color) {
	piece := pieceColor(Pawn, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]
}

// OutpostSquares returns a bitboard of outpost squares for the given side
// An outpost square is:
// - In enemy territory (rank 4-6 for white, 3-5 for black)
// - Cannot be attacked by enemy pawns
// - Protected by own pawn(s)
func OutpostSquares(alliedPawns Bitboard, enemyPawns Bitboard, side Color) Bitboard {
	outpostRanks := OutpostsRanks[side]

	enemyAttacksFrontSpans := Bitboard(0)
	for enemyPawns > 0 {
		pawn := enemyPawns.NextBit()
		enemyAttacksFrontSpans |= attacksFrontSpans[side.Opponent()][Bsf(pawn)]
	}
	protectedByPawns := pawnAttacks(&alliedPawns, side)

	return ^enemyAttacksFrontSpans & protectedByPawns & outpostRanks
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
