package engine

const (
	// Pawn Structure
	DoubledPawnPenaltyMg  = -5
	DoubledPawnPenaltyEg  = -15
	IsolatedPawnPenaltyMg = -10
	IsolatedPawnPenaltyEg = -12
	BackwardPawnPenaltyMg = -9
	BackwardPawnPenaltyEg = -8

	// Material Adjustment
	BishopPairBonusMg    = 22
	BishopPairBonusEg    = 67
	RookOnOpenFileMg     = 37
	RookOnSemiOpenFileMg = 21

	KnightOutpostBonusMg = 35
	KnightOutpostBonusEg = 19
	BishopOutpostBonusMg = 40
	BishopOutpostBonusEg = -3

	KnightAttackWeight   = 19
	BishopAttackWeight   = 15
	RookAttackWeight     = 20
	QueenAttackWeight    = 11
	KingZoneDefenseBonus = 16

	KingOnOpenFilePenalty   = -48
	KingNearOpenFilePenalty = -15

	// Threats
	MinorAttackedByPawnThreatPenalty  = -51
	RookAttackedByPawnThreatPenalty   = -51
	QueenAttackedByPawnThreatPenalty  = -45
	RookAttackedByMinorThreatPenalty  = -38
	QueenAttackedByMinorThreatPenalty = -45

	SafeQueenCheckThreatBonus  = 15
	SafeRookCheckThreatBonus   = 13
	SafeBishopCheckThreatBonus = 17
	SafeKnightCheckThreatBonus = 14

	PinnedQueenThreatPenalty  = -62
	PinnedRookThreatPenalty   = -32
	PinnedBishopThreatPenalty = -33
	PinnedKnightThreatPenalty = -39

	TempoBonus = 24
)

var (
	// Queen Mobility mg/eg contains the bonus for queen mobility
	QueenMobilityMg = [28]int{-21, -18, -33, -54, -39, -22, -20, -18, -16, -14, -11, -7, -4, 0, 1, 2, 2, 2, 1, 3, 11, 23, 36, 47, 37, 75, 37, 23}
	QueenMobilityEg = [28]int{-77, -66, -56, -74, 0, 59, 104, 131, 153, 176, 182, 187, 193, 192, 194, 197, 197, 199, 201, 197, 193, 172, 165, 148, 153, 140, 162, 159}

	// Rook Mobility mg/eg contains the bonus for rook mobility
	RookMobilityMg = [15]int{-42, -31, -8, -1, 3, 5, 6, 7, 10, 14, 17, 18, 22, 26, 29}
	RookMobilityEg = [15]int{-17, 5, 30, 52, 63, 73, 80, 86, 88, 91, 94, 96, 97, 95, 91}

	// Bishop Mobility mg/eg contains the bonus for bishop mobility
	BishopMobilityMg = [14]int{-53, -62, -29, -19, -6, 1, 7, 12, 14, 18, 22, 38, 44, 56}
	BishopMobilityEg = [14]int{-130, -55, -3, 22, 32, 39, 48, 53, 58, 58, 59, 50, 50, 37}

	// KnightMobility mg/eg contains the bonus for knight mobility
	KnightMobilityMg = [9]int{-141, -37, -10, 0, 12, 15, 28, 40, 53}
	KnightMobilityEg = [9]int{-76, -21, 10, 34, 45, 57, 59, 61, 55}

	// PassedPawnsBonus mg/eg contains the bonus for passed pawns
	PassedPawnsBonusMg = [8]int{0, -8, -14, -14, 12, 0, 13, 0}
	PassedPawnsBonusEg = [8]int{0, 9, 13, 40, 67, 139, 117, 0}

	// PawnShieldFrontBonus/PawnShieldSideBonus contains the bonus for pawns on the front and side ofthe king file(s)
	PawnShieldFrontBonus = [4]int{0, 22, 21, 2}
	PawnShieldSideBonus  = [4]int{27, 16, 11, 1}

	// PawnStormFrontPenalty/PawnStormSidePenalty contains the penalty for the enemy pawns on the front and side of king file(s)
	PawnStormFrontPenalty = [4]int{117, -6, -5, 0}
	PawnStormSidePenalty  = [4]int{-4, -23, -26, -5}

	// OutpostsRanks contains the bitboard mask for ranks that are considered outposts
	OutpostsRanks = [2]Bitboard{
		Ranks[3] | Ranks[4] | Ranks[5],
		Ranks[2] | Ranks[3] | Ranks[4],
	}
)

// Evaluation contains the elements for evaluation of a position
type Evaluation struct {
	Eval      EvalVector
	PawnCache PawnHashTable
}

// EvalVector contains the different evaluation elements of a position
type EvalVector struct {
	mgMaterial         [2]int // White and Black scores
	egMaterial         [2]int
	mgMobility         [2]int
	egMobility         [2]int
	mgPawnStrucutre    [2]int
	egPawnStructure    [2]int
	mgKingSafety       [2]int
	threats            [2]int
	kingAttackersCount [2]int
	kingAttacksWeight  [2]int
	phase              int
}

// NewEvaluation returns a new Evaluation
func NewEvaluation(size int) *Evaluation {
	return &Evaluation{
		Eval:      EvalVector{},
		PawnCache: *NewPawnHashTable(size),
	}
}

// Clear clears the evaluation
func (ev *Evaluation) Clear() {
	ev.Eval.clear()
	ev.PawnCache.clear()
}

// clear clears the evaluation vector
func (ev *EvalVector) clear() {
	ev.mgMaterial = [2]int{0, 0}
	ev.egMaterial = [2]int{0, 0}
	ev.mgMobility = [2]int{0, 0}
	ev.egMobility = [2]int{0, 0}
	ev.mgKingSafety = [2]int{0, 0}
	ev.mgPawnStrucutre = [2]int{0, 0}
	ev.egPawnStructure = [2]int{0, 0}
	ev.threats = [2]int{0, 0}
	ev.kingAttackersCount = [2]int{0, 0}
	ev.kingAttacksWeight = [2]int{0, 0}
	ev.phase = 0
}

// Evaluate returns the static score of the position
func (ev *Evaluation) Evaluate(pos *Position) int {
	ev.Eval.clear()

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
	pins := pos.PinnedPieces(White) | pos.PinnedPieces(Black)

	for piece, bb := range pos.Bitboards {
		color := Color(piece / 6)

		for bb > 0 {
			bb := bb.NextBit()
			sq := Bsf(bb)

			switch pieceRole(piece) {
			case King:
				ev.Eval.evaluateKing(sq, pawns, color)
			case Queen:
				ev.Eval.evaluateQueen(sq, blocks, enemyPawnsAttacks[color], pos.KingPosition(color.Opponent()), pins, color)
			case Rook:
				ev.Eval.evaluateRook(sq, blocks, enemyPawnsAttacks[color], pawns, pos.KingPosition(color.Opponent()), pins, color)
			case Bishop:
				ev.Eval.evaluateBishop(sq, blocks, enemyPawnsAttacks[color], pos.KingPosition(color.Opponent()), outpostSquares[color], pins, color, pos)
			case Knight:
				ev.Eval.evaluateKnight(sq, blocks, enemyPawnsAttacks[color], pos.KingPosition(color.Opponent()), outpostSquares[color], pins, color, pos)
			case Pawn:
				ev.Eval.evaluatePawn(sq, color)
			}
		}
	}

	// Bishop pair bonus
	if pos.Bitboards[WhiteBishop].count() >= 2 {
		ev.Eval.mgMaterial[White] += BishopPairBonusMg
		ev.Eval.egMaterial[White] += BishopPairBonusEg
	}

	if pos.Bitboards[BlackBishop].count() >= 2 {
		ev.Eval.mgMaterial[Black] += BishopPairBonusMg
		ev.Eval.egMaterial[Black] += BishopPairBonusEg
	}

	// Safety
	// Apply King Safety Penalties to opponent only if there are at least 2 attackers and one of the pieces is a queen
	if ev.Eval.kingAttackersCount[White] >= 2 && pos.Bitboards[pieceColor(Queen, White)] > 0 {
		zoneDefense := KingZone[Black][Bsf(pos.KingPosition(Black))] & enemyPawnsAttacks[White]
		ev.Eval.mgKingSafety[Black] += -ev.Eval.kingAttacksWeight[White] + KingZoneDefenseBonus*zoneDefense.count()
	}

	if ev.Eval.kingAttackersCount[Black] >= 2 && pos.Bitboards[pieceColor(Queen, Black)] > 0 {
		zoneDefense := KingZone[White][Bsf(pos.KingPosition(White))] & enemyPawnsAttacks[Black]
		ev.Eval.mgKingSafety[White] += -ev.Eval.kingAttacksWeight[Black] + KingZoneDefenseBonus*zoneDefense.count()
	}

	// TempoBonus
	ev.Eval.mgMaterial[pos.Turn] += TempoBonus
	ev.Eval.egMaterial[pos.Turn] += TempoBonus

	mgSc, egSc, ok := ev.PawnCache.probe(pos.PawnHash, pos.Turn)
	if ok {
		ev.Eval.mgPawnStrucutre[pos.Turn] = mgSc
		ev.Eval.egPawnStructure[pos.Turn] = egSc
	} else {
		ev.Eval.evaluatePawnStructure(pos, enemyPawnsAttacks[White], White)
		ev.Eval.evaluatePawnStructure(pos, enemyPawnsAttacks[Black], Black)

		// Store always from White's perspective
		mgScWhite := ev.Eval.mgPawnStrucutre[White] - ev.Eval.mgPawnStrucutre[Black]
		egScWhite := ev.Eval.egPawnStructure[White] - ev.Eval.egPawnStructure[Black]
		ev.PawnCache.store(pos.PawnHash, mgScWhite, egScWhite)
	}

	return ev.Eval.score(pos.Turn)
}

// score returns the score relative to the side
func (ev *EvalVector) score(side Color) int {
	opponent := side.Opponent()

	mg := ev.mgMaterial[side] - ev.mgMaterial[opponent]
	eg := ev.egMaterial[side] - ev.egMaterial[opponent]
	mg += ev.mgMobility[side] - ev.mgMobility[opponent]
	eg += ev.egMobility[side] - ev.egMobility[opponent]
	mg += ev.mgPawnStrucutre[side] - ev.mgPawnStrucutre[opponent]
	eg += ev.egPawnStructure[side] - ev.egPawnStructure[opponent]
	mg += ev.mgKingSafety[side] - ev.mgKingSafety[opponent]
	mg += ev.threats[side] - ev.threats[opponent]
	eg += ev.threats[side] - ev.threats[opponent]

	mgPhase := min(ev.phase, 62)
	egPhase := 62 - mgPhase
	return (mg*mgPhase + eg*egPhase) / 62
}

// evaluateKing evaluates the score of a king
func (ev *EvalVector) evaluateKing(from int, pawns [2]Bitboard, side Color) {
	piece := pieceColor(King, side)

	direction := North
	if side == Black {
		direction = South
	}

	// Pawn Shield / Storm
	kingFile, kingRank := from%8, from/8
	for file := max(0, kingFile-1); file <= min(7, kingFile+1); file++ {
		from := kingRank*8 + file
		frontMask := RayAttacks[direction][from] | bitboardFromIndex(from)

		shielders := pawns[side] & frontMask
		stormers := pawns[side.Opponent()] & frontMask

		shield := NearestFromSide(shielders&Files[file], side)
		storm := NearestFromSide(stormers&Files[file], side.Opponent())
		hasShield := shield != 64 && shield != -1
		hasStorm := storm != 64 && storm != -1
		shieldRank := shield / 8
		stormRank := storm / 8

		shieldDist := abs(kingRank - shieldRank)
		stormDist := abs(kingRank - stormRank)

		if hasShield && shieldDist < 4 {
			if file == kingFile {
				ev.mgKingSafety[side] += PawnShieldFrontBonus[shieldDist]
			} else {
				ev.mgKingSafety[side] += PawnShieldSideBonus[shieldDist]
			}
		}

		// If the pawns are locked (one in front of the other), we skip the storm penalty
		// since the enemy pawn cannot be pushed nor open the file for attacks
		// NOTE: Use -1 due to array indexing. Storms count starts from the 1 rank distance, shield can be in the same rank as the king
		if hasStorm && stormDist > 0 && stormDist < 5 && shieldDist != stormDist-1 {
			if file == kingFile {
				ev.mgKingSafety[side] += PawnStormFrontPenalty[stormDist-1]
			} else {
				ev.mgKingSafety[side] += PawnStormSidePenalty[stormDist-1]
			}
		}

		// Open/SemiOpen files near the king
		if (pawns[side]|pawns[side.Opponent()])&Files[file] == 0 {
			if file == kingFile {
				ev.mgKingSafety[side] += KingOnOpenFilePenalty
			} else {
				ev.mgKingSafety[side] += KingNearOpenFilePenalty
			}
		}
	}

	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]
}

// evaluateQueen evaluates the score of a queen
func (ev *EvalVector) evaluateQueen(from int, blocks Bitboard, enemyPawnsAttacks Bitboard, enemyKing Bitboard, pins Bitboard, side Color) {
	piece := pieceColor(Queen, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	squares := (attacks & ^enemyPawnsAttacks).count()

	enemyKingZone := KingZone[side.Opponent()][Bsf(enemyKing)]
	if attacks&enemyKingZone != 0 {
		ev.kingAttackersCount[side]++
		ev.kingAttacksWeight[side] += QueenAttackWeight * (attacks & enemyKingZone).count()
	}

	ev.mgMobility[side] += QueenMobilityMg[squares]
	ev.egMobility[side] += QueenMobilityEg[squares]

	if enemyPawnsAttacks&fromBB > 0 {
		ev.threats[side] += QueenAttackedByPawnThreatPenalty
	}

	if fromBB&pins > 0 {
		ev.threats[side] += PinnedQueenThreatPenalty
	}

	// Safe checks. Squares not defended by enemy pawns
	// where the queen can move to give check
	safeQueenChecks := Attacks(piece, enemyKing, blocks) & ^enemyPawnsAttacks & attacks
	if safeQueenChecks > 0 {
		ev.threats[side] += SafeQueenCheckThreatBonus * safeQueenChecks.count()
	}

	ev.phase += 9
}

// evaluateRook evaluates the score of a rook
func (ev *EvalVector) evaluateRook(from int, blocks Bitboard, enemyPawnsAttacks Bitboard, pawns [2]Bitboard, enemyKing Bitboard, pins Bitboard, side Color) {
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

	enemyKingZone := KingZone[side.Opponent()][Bsf(enemyKing)]
	if attacks&enemyKingZone != 0 {
		ev.kingAttackersCount[side]++
		ev.kingAttacksWeight[side] += RookAttackWeight * (attacks & enemyKingZone).count()
	}

	if enemyPawnsAttacks&fromBB > 0 {
		ev.threats[side] += RookAttackedByPawnThreatPenalty
	}

	if fromBB&pins > 0 {
		ev.threats[side] += PinnedRookThreatPenalty
	}

	safeRookChecks := Attacks(piece, enemyKing, blocks) & ^enemyPawnsAttacks & attacks
	if safeRookChecks > 0 {
		ev.threats[side] += SafeRookCheckThreatBonus * safeRookChecks.count()
	}

	ev.mgMobility[side] += RookMobilityMg[squares]
	ev.egMobility[side] += RookMobilityEg[squares]

	ev.phase += 5
}

// evaluateBishop evaluates the score of a bishop
func (ev *EvalVector) evaluateBishop(from int, blocks Bitboard, enemyPawnsAttacks Bitboard, enemyKing Bitboard, outpostMask Bitboard, pins Bitboard, side Color, pos *Position) {
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

	enemyKingZone := KingZone[side.Opponent()][Bsf(enemyKing)]
	if attacks&enemyKingZone != 0 {
		ev.kingAttackersCount[side]++
		ev.kingAttacksWeight[side] += BishopAttackWeight * (attacks & enemyKingZone).count()
	}

	if enemyPawnsAttacks&fromBB > 0 {
		ev.threats[side] += MinorAttackedByPawnThreatPenalty
	}
	if attacks&pos.Bitboards[pieceColor(Queen, side.Opponent())] > 0 {
		ev.threats[side.Opponent()] += QueenAttackedByMinorThreatPenalty
	}
	if attacks&pos.Bitboards[pieceColor(Rook, side.Opponent())] > 0 {
		ev.threats[side.Opponent()] += RookAttackedByMinorThreatPenalty
	}

	if fromBB&pins > 0 {
		ev.threats[side] += PinnedBishopThreatPenalty
	}

	safeBishopChecks := Attacks(piece, enemyKing, blocks) & ^enemyPawnsAttacks & attacks
	if safeBishopChecks > 0 {
		ev.threats[side] += SafeBishopCheckThreatBonus * safeBishopChecks.count()
	}

	ev.mgMobility[side] += BishopMobilityMg[squares]
	ev.egMobility[side] += BishopMobilityEg[squares]

	ev.phase += 3
}

// evaluateKnight evaluates the score of a knight
func (ev *EvalVector) evaluateKnight(from int, blocks Bitboard, enemyPawnsAttacks Bitboard, enemyKing Bitboard, outpostMask Bitboard, pins Bitboard, side Color, pos *Position) {
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

	enemyKingZone := KingZone[side.Opponent()][Bsf(enemyKing)]
	if attacks&enemyKingZone != 0 {
		ev.kingAttackersCount[side]++
		ev.kingAttacksWeight[side] += KnightAttackWeight * (attacks & enemyKingZone).count()
	}

	if enemyPawnsAttacks&fromBB > 0 {
		ev.threats[side] += MinorAttackedByPawnThreatPenalty
	}
	if attacks&pos.Bitboards[pieceColor(Queen, side.Opponent())] > 0 {
		ev.threats[side.Opponent()] += QueenAttackedByMinorThreatPenalty
	}
	if attacks&pos.Bitboards[pieceColor(Rook, side.Opponent())] > 0 {
		ev.threats[side.Opponent()] += RookAttackedByMinorThreatPenalty
	}

	if fromBB&pins > 0 {
		ev.threats[side] += PinnedKnightThreatPenalty
	}

	safeKnightChecks := Attacks(piece, enemyKing, blocks) & ^enemyPawnsAttacks & attacks
	if safeKnightChecks > 0 {
		ev.threats[side] += SafeKnightCheckThreatBonus * safeKnightChecks.count()
	}

	ev.mgMobility[side] += KnightMobilityMg[squares]
	ev.egMobility[side] += KnightMobilityEg[squares]

	ev.phase += 3
}

// evaluatePawn evaluates the score of a pawn
func (ev *EvalVector) evaluatePawn(from int, side Color) {
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
func (ev *EvalVector) evaluatePawnStructure(pos *Position, enemyPawnsAttacks Bitboard, side Color) {
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
		frontAndAdjacentSquares := attacksFrontSpans[side][Bsf(pawn)] | RayAttacks[direction][Bsf(pawn)]

		if frontAndAdjacentSquares&enemyPawns == 0 {
			passedPawns |= pawn
		}
	}

	return
}
