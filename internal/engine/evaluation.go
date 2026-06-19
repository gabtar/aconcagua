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
	EvalData  EvalData
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

// EvalData contains positional data about the current position
type EvalData struct {
	kings           [2]Bitboard
	attackedByPawns [2]Bitboard
	pawns           [2]Bitboard
	outposts        [2]Bitboard
	blocks          Bitboard
	pinned          Bitboard
}

// NewEvaluation returns a new Evaluation
func NewEvaluation(size int) *Evaluation {
	return &Evaluation{
		Eval:      EvalVector{},
		EvalData:  EvalData{},
		PawnCache: *NewPawnHashTable(size),
	}
}

// Clear clears the evaluation
func (ev *Evaluation) Clear() {
	ev.Eval.clear()
	ev.EvalData.clear()
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

// clear clears the EvalData
func (ed *EvalData) clear() {
	ed.kings = [2]Bitboard{}
	ed.attackedByPawns = [2]Bitboard{}
	ed.pawns = [2]Bitboard{}
	ed.outposts = [2]Bitboard{}
	ed.blocks = 0
	ed.pinned = 0
}

// init initializes the evaluation data
func (ed *EvalData) init(pos *Position) {
	ed.kings = [2]Bitboard{
		pos.KingPosition(White),
		pos.KingPosition(Black),
	}
	ed.attackedByPawns = [2]Bitboard{
		pawnAttacks(&pos.Bitboards[WhitePawn], White),
		pawnAttacks(&pos.Bitboards[BlackPawn], Black),
	}
	ed.pawns = [2]Bitboard{
		pos.Bitboards[WhitePawn],
		pos.Bitboards[BlackPawn],
	}
	ed.outposts = [2]Bitboard{
		OutpostSquares(ed.pawns[White], ed.pawns[Black], White),
		OutpostSquares(ed.pawns[Black], ed.pawns[White], Black),
	}
	ed.blocks = ^pos.EmptySquares()
	ed.pinned = pos.PinnedPieces(White) | pos.PinnedPieces(Black)
}

// Evaluate returns the static score of the position
func (ev *Evaluation) Evaluate(pos *Position) int {
	ev.Eval.clear()
	ev.EvalData.init(pos)

	for piece, bb := range pos.Bitboards {
		color := Color(piece / 6)

		for bb > 0 {
			bb := bb.NextBit()
			sq := Bsf(bb)

			switch pieceRole(piece) {
			case King:
				ev.evaluateKing(sq, color)
			case Queen:
				ev.evaluateQueen(sq, color)
			case Rook:
				ev.evaluateRook(sq, color)
			case Bishop:
				ev.evaluateBishop(sq, color, pos)
			case Knight:
				ev.evaluateKnight(sq, color, pos)
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
		zoneDefense := KingZone[Black][Bsf(pos.KingPosition(Black))] & ev.EvalData.attackedByPawns[Black]
		ev.Eval.mgKingSafety[Black] += -ev.Eval.kingAttacksWeight[White] + KingZoneDefenseBonus*zoneDefense.count()
	}

	if ev.Eval.kingAttackersCount[Black] >= 2 && pos.Bitboards[pieceColor(Queen, Black)] > 0 {
		zoneDefense := KingZone[White][Bsf(pos.KingPosition(White))] & ev.EvalData.attackedByPawns[White]
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
		ev.Eval.evaluatePawnStructure(pos, ev.EvalData.attackedByPawns[Black], White)
		ev.Eval.evaluatePawnStructure(pos, ev.EvalData.attackedByPawns[White], Black)

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
func (ev *Evaluation) evaluateKing(from int, side Color) {
	piece := pieceColor(King, side)
	direction := [2]int{North, South}

	// Pawn Shield / Storm
	kingFile, kingRank := from%8, from/8
	for file := max(0, kingFile-1); file <= min(7, kingFile+1); file++ {
		from := kingRank*8 + file
		frontMask := RayAttacks[direction[side]][from] | bitboardFromIndex(from)

		shielders := ev.EvalData.pawns[side] & frontMask
		stormers := ev.EvalData.pawns[side.Opponent()] & frontMask

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
				ev.Eval.mgKingSafety[side] += PawnShieldFrontBonus[shieldDist]
			} else {
				ev.Eval.mgKingSafety[side] += PawnShieldSideBonus[shieldDist]
			}
		}

		// If the pawns are locked (one in front of the other), we skip the storm penalty
		// since the enemy pawn cannot be pushed nor open the file for attacks
		// NOTE: Use -1 due to array indexing. Storms count starts from the 1 rank distance, shield can be in the same rank as the king
		if hasStorm && stormDist > 0 && stormDist < 5 && shieldDist != stormDist-1 {
			if file == kingFile {
				ev.Eval.mgKingSafety[side] += PawnStormFrontPenalty[stormDist-1]
			} else {
				ev.Eval.mgKingSafety[side] += PawnStormSidePenalty[stormDist-1]
			}
		}

		// Open/SemiOpen files near the king
		if (ev.EvalData.pawns[side]|ev.EvalData.pawns[side.Opponent()])&Files[file] == 0 {
			if file == kingFile {
				ev.Eval.mgKingSafety[side] += KingOnOpenFilePenalty
			} else {
				ev.Eval.mgKingSafety[side] += KingNearOpenFilePenalty
			}
		}
	}

	ev.Eval.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.Eval.egMaterial[side] += endgamePiecesScore[piece][from]
}

// evaluateQueen evaluates the score of a queen
func (ev *Evaluation) evaluateQueen(from int, side Color) {
	piece := pieceColor(Queen, side)
	opponent := side.Opponent()
	ev.Eval.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.Eval.egMaterial[side] += endgamePiecesScore[piece][from]

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, ev.EvalData.blocks)
	squares := (attacks & ^ev.EvalData.attackedByPawns[opponent]).count()

	enemyKingZone := KingZone[opponent][Bsf(ev.EvalData.kings[opponent])]
	if attacks&enemyKingZone != 0 {
		ev.Eval.kingAttackersCount[side]++
		ev.Eval.kingAttacksWeight[side] += QueenAttackWeight * (attacks & enemyKingZone).count()
	}

	ev.Eval.mgMobility[side] += QueenMobilityMg[squares]
	ev.Eval.egMobility[side] += QueenMobilityEg[squares]

	if ev.EvalData.attackedByPawns[opponent]&fromBB > 0 {
		ev.Eval.threats[side] += QueenAttackedByPawnThreatPenalty
	}

	if fromBB&ev.EvalData.pinned > 0 {
		ev.Eval.threats[side] += PinnedQueenThreatPenalty
	}

	// Safe checks. Squares not defended by enemy pawns
	// where the queen can move to give check
	safeQueenChecks := Attacks(piece, ev.EvalData.kings[opponent], ev.EvalData.blocks) & ^ev.EvalData.attackedByPawns[opponent] & attacks
	if safeQueenChecks > 0 {
		ev.Eval.threats[side] += SafeQueenCheckThreatBonus * safeQueenChecks.count()
	}

	ev.Eval.phase += 9
}

// evaluateRook evaluates the score of a rook
func (ev *Evaluation) evaluateRook(from int, side Color) {
	piece := pieceColor(Rook, side)
	opponent := side.Opponent()
	ev.Eval.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.Eval.egMaterial[side] += endgamePiecesScore[piece][from]

	file := from % 8
	if (ev.EvalData.pawns[White]|ev.EvalData.pawns[Black])&Files[file] == 0 {
		ev.Eval.mgMaterial[side] += RookOnOpenFileMg
	}

	if ev.EvalData.pawns[side]&Files[file] == 0 && ev.EvalData.pawns[opponent]&Files[file] > 0 {
		ev.Eval.mgMaterial[side] += RookOnSemiOpenFileMg
	}

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, ev.EvalData.blocks)
	squares := (attacks & ^ev.EvalData.attackedByPawns[opponent]).count()

	enemyKingZone := KingZone[opponent][Bsf(ev.EvalData.kings[opponent])]
	if attacks&enemyKingZone != 0 {
		ev.Eval.kingAttackersCount[side]++
		ev.Eval.kingAttacksWeight[side] += RookAttackWeight * (attacks & enemyKingZone).count()
	}

	if ev.EvalData.attackedByPawns[opponent]&fromBB > 0 {
		ev.Eval.threats[side] += RookAttackedByPawnThreatPenalty
	}

	if fromBB&ev.EvalData.pinned > 0 {
		ev.Eval.threats[side] += PinnedRookThreatPenalty
	}

	safeRookChecks := Attacks(piece, ev.EvalData.kings[opponent], ev.EvalData.blocks) & ^ev.EvalData.attackedByPawns[opponent] & attacks
	if safeRookChecks > 0 {
		ev.Eval.threats[side] += SafeRookCheckThreatBonus * safeRookChecks.count()
	}

	ev.Eval.mgMobility[side] += RookMobilityMg[squares]
	ev.Eval.egMobility[side] += RookMobilityEg[squares]

	ev.Eval.phase += 5
}

// evaluateBishop evaluates the score of a bishop
func (ev *Evaluation) evaluateBishop(from int, side Color, pos *Position) {
	piece := pieceColor(Bishop, side)
	opponent := side.Opponent()
	ev.Eval.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.Eval.egMaterial[side] += endgamePiecesScore[piece][from]

	if ev.EvalData.outposts[side]&bitboardFromIndex(from) > 0 {
		ev.Eval.mgMaterial[side] += BishopOutpostBonusMg
		ev.Eval.egMaterial[side] += BishopOutpostBonusEg
	}

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, ev.EvalData.blocks)
	squares := (attacks & ^ev.EvalData.attackedByPawns[opponent]).count()

	enemyKingZone := KingZone[opponent][Bsf(ev.EvalData.kings[opponent])]
	if attacks&enemyKingZone != 0 {
		ev.Eval.kingAttackersCount[side]++
		ev.Eval.kingAttacksWeight[side] += BishopAttackWeight * (attacks & enemyKingZone).count()
	}

	if ev.EvalData.attackedByPawns[opponent]&fromBB > 0 {
		ev.Eval.threats[side] += MinorAttackedByPawnThreatPenalty
	}
	if attacks&pos.Bitboards[pieceColor(Queen, side.Opponent())] > 0 {
		ev.Eval.threats[side.Opponent()] += QueenAttackedByMinorThreatPenalty
	}
	if attacks&pos.Bitboards[pieceColor(Rook, side.Opponent())] > 0 {
		ev.Eval.threats[side.Opponent()] += RookAttackedByMinorThreatPenalty
	}

	if fromBB&ev.EvalData.pinned > 0 {
		ev.Eval.threats[side] += PinnedBishopThreatPenalty
	}

	safeBishopChecks := Attacks(piece, ev.EvalData.kings[opponent], ev.EvalData.blocks) & ^ev.EvalData.attackedByPawns[opponent] & attacks
	if safeBishopChecks > 0 {
		ev.Eval.threats[side] += SafeBishopCheckThreatBonus * safeBishopChecks.count()
	}

	ev.Eval.mgMobility[side] += BishopMobilityMg[squares]
	ev.Eval.egMobility[side] += BishopMobilityEg[squares]

	ev.Eval.phase += 3
}

// evaluateKnight evaluates the score of a knight
func (ev *Evaluation) evaluateKnight(from int, side Color, pos *Position) {
	piece := pieceColor(Knight, side)
	opponent := side.Opponent()
	ev.Eval.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.Eval.egMaterial[side] += endgamePiecesScore[piece][from]

	if ev.EvalData.outposts[side]&bitboardFromIndex(from) > 0 {
		ev.Eval.mgMaterial[side] += KnightOutpostBonusMg
		ev.Eval.egMaterial[side] += KnightOutpostBonusEg
	}

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, ev.EvalData.blocks)
	squares := (attacks & ^ev.EvalData.attackedByPawns[opponent]).count()

	enemyKingZone := KingZone[opponent][Bsf(ev.EvalData.kings[opponent])]
	if attacks&enemyKingZone != 0 {
		ev.Eval.kingAttackersCount[side]++
		ev.Eval.kingAttacksWeight[side] += KnightAttackWeight * (attacks & enemyKingZone).count()
	}

	if ev.EvalData.attackedByPawns[opponent]&fromBB > 0 {
		ev.Eval.threats[side] += MinorAttackedByPawnThreatPenalty
	}
	if attacks&pos.Bitboards[pieceColor(Queen, opponent)] > 0 {
		ev.Eval.threats[side.Opponent()] += QueenAttackedByMinorThreatPenalty
	}
	if attacks&pos.Bitboards[pieceColor(Rook, opponent)] > 0 {
		ev.Eval.threats[side.Opponent()] += RookAttackedByMinorThreatPenalty
	}

	if fromBB&ev.EvalData.pinned > 0 {
		ev.Eval.threats[side] += PinnedKnightThreatPenalty
	}

	safeKnightChecks := Attacks(piece, ev.EvalData.kings[opponent], ev.EvalData.blocks) & ^ev.EvalData.attackedByPawns[opponent] & attacks
	if safeKnightChecks > 0 {
		ev.Eval.threats[side] += SafeKnightCheckThreatBonus * safeKnightChecks.count()
	}

	ev.Eval.mgMobility[side] += KnightMobilityMg[squares]
	ev.Eval.egMobility[side] += KnightMobilityEg[squares]

	ev.Eval.phase += 3
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
