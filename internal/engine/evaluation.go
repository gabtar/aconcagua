package engine

const (
	// Pawn Structure
	DoubledPawnPenaltyMg  = 2
	DoubledPawnPenaltyEg  = -8
	IsolatedPawnPenaltyMg = -1
	IsolatedPawnPenaltyEg = -4
	BackwardPawnPenaltyMg = -2
	BackwardPawnPenaltyEg = -6
	DefendedPawnBonusMg   = 10
	DefendedPawnBonusEg   = 7
	ConnectedPawnBonusMg  = 7
	ConnectedPawnBonusEg  = 4

	// Material Adjustment
	BishopPairBonusMg    = 24
	BishopPairBonusEg    = 64
	RookOnOpenFileMg     = 37
	RookOnSemiOpenFileMg = 21

	KnightOutpostBonusMg = 36
	KnightOutpostBonusEg = 19
	BishopOutpostBonusMg = 41
	BishopOutpostBonusEg = -2

	KnightAttackWeight   = 20
	BishopAttackWeight   = 16
	RookAttackWeight     = 21
	QueenAttackWeight    = 12
	KingZoneDefenseBonus = 17

	KingOnOpenFilePenalty   = -49
	KingNearOpenFilePenalty = -16

	// Threats
	MinorAttackedByPawnThreatPenalty  = -55
	RookAttackedByPawnThreatPenalty   = -52
	QueenAttackedByPawnThreatPenalty  = -50
	RookAttackedByMinorThreatPenalty  = -38
	QueenAttackedByMinorThreatPenalty = -45

	SafeQueenCheckThreatBonus  = 15
	SafeRookCheckThreatBonus   = 12
	SafeBishopCheckThreatBonus = 17
	SafeKnightCheckThreatBonus = 14

	PinnedQueenThreatPenaltyMg  = -77
	PinnedRookThreatPenaltyMg   = -59
	PinnedBishopThreatPenaltyMg = -5
	PinnedKnightThreatPenaltyMg = -30

	PinnedQueenThreatPenaltyEg  = -48
	PinnedRookThreatPenaltyEg   = -7
	PinnedBishopThreatPenaltyEg = -70
	PinnedKnightThreatPenaltyEg = -57

	TempoBonus = 25
)

var (
	// Queen Mobility mg/eg contains the bonus for queen mobility
	QueenMobilityMg = [28]int{-21, -18, -38, -50, -45, -27, -23, -19, -17, -15, -12, -8, -5, -1, 0, 0, 2, 1, 2, 5, 13, 27, 42, 64, 51, 91, 25, 6}
	QueenMobilityEg = [28]int{-77, -66, -56, -74, 26, 90, 124, 139, 158, 179, 184, 191, 197, 196, 197, 201, 199, 202, 202, 197, 193, 173, 162, 138, 141, 112, 133, 127}

	// Rook Mobility mg/eg contains the bonus for rook mobility
	RookMobilityMg = [15]int{-39, -28, -7, -1, 2, 4, 6, 7, 10, 15, 18, 20, 24, 30, 38}
	RookMobilityEg = [15]int{-17, 11, 24, 50, 64, 73, 81, 86, 88, 90, 93, 95, 96, 94, 88}

	// Bishop Mobility mg/eg contains the bonus for bishop mobility
	BishopMobilityMg = [14]int{-36, -55, -28, -16, -5, 2, 8, 14, 15, 21, 24, 43, 43, 59}
	BishopMobilityEg = [14]int{-139, -53, 0, 22, 32, 38, 47, 51, 56, 56, 56, 47, 50, 36}

	// KnightMobility mg/eg contains the bonus for knight mobility
	KnightMobilityMg = [9]int{-140, -31, -9, 3, 15, 18, 30, 41, 56}
	KnightMobilityEg = [9]int{-80, -26, 9, 32, 43, 55, 56, 58, 51}

	// PassedPawnsBonus mg/eg contains the bonus for passed pawns
	PassedPawnsBonusMg = [8]int{0, 0, -6, -8, 16, 0, 14, 0}
	PassedPawnsBonusEg = [8]int{0, 10, 15, 42, 70, 141, 123, 0}

	// PawnShieldFrontBonus/PawnShieldSideBonus contains the bonus for pawns on the front and side ofthe king file(s)
	PawnShieldFrontBonus = [4]int{0, 21, 16, 2}
	PawnShieldSideBonus  = [4]int{21, 15, 7, 0}

	// PawnStormFrontPenalty/PawnStormSidePenalty contains the penalty for the enemy pawns on the front and side of king file(s)
	PawnStormFrontPenalty = [4]int{137, -9, -6, -2}
	PawnStormSidePenalty  = [4]int{-13, -22, -27, -6}

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
	mgThreats          [2]int
	egThreats          [2]int
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
	ev.mgThreats = [2]int{0, 0}
	ev.egThreats = [2]int{0, 0}
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
		pawnAttacks(&pos.Pieces[WhitePawn], White),
		pawnAttacks(&pos.Pieces[BlackPawn], Black),
	}
	ed.pawns = [2]Bitboard{
		pos.Pieces[WhitePawn],
		pos.Pieces[BlackPawn],
	}
	ed.outposts = [2]Bitboard{
		OutpostSquares(ed.pawns[White], ed.pawns[Black], White),
		OutpostSquares(ed.pawns[Black], ed.pawns[White], Black),
	}
	ed.blocks = pos.Sides[All]
	ed.pinned = pos.PinnedPieces(White) | pos.PinnedPieces(Black)
}

// Evaluate returns the static score of the position
func (ev *Evaluation) Evaluate(pos *Position) int {
	ev.Eval.clear()
	ev.EvalData.init(pos)

	ev.evaluatePawns(pos)

	for piece, bb := range pos.Pieces {
		color := Color(piece / 6)

		// Skip pawns
		if pieceRole(piece) == Pawn {
			continue
		}

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
			}
		}
	}

	// Bishop pair bonus
	if pos.Pieces[WhiteBishop].count() >= 2 {
		ev.Eval.mgMaterial[White] += BishopPairBonusMg
		ev.Eval.egMaterial[White] += BishopPairBonusEg
	}

	if pos.Pieces[BlackBishop].count() >= 2 {
		ev.Eval.mgMaterial[Black] += BishopPairBonusMg
		ev.Eval.egMaterial[Black] += BishopPairBonusEg
	}

	// Safety
	// Apply King Safety Penalties to opponent only if there are at least 2 attackers and one of the pieces is a queen
	if ev.Eval.kingAttackersCount[White] >= 2 && pos.Pieces[pieceColor(Queen, White)] > 0 {
		zoneDefense := KingZone[Black][Bsf(pos.KingPosition(Black))] & ev.EvalData.attackedByPawns[Black]
		ev.Eval.mgKingSafety[Black] += -ev.Eval.kingAttacksWeight[White] + KingZoneDefenseBonus*zoneDefense.count()
	}

	if ev.Eval.kingAttackersCount[Black] >= 2 && pos.Pieces[pieceColor(Queen, Black)] > 0 {
		zoneDefense := KingZone[White][Bsf(pos.KingPosition(White))] & ev.EvalData.attackedByPawns[White]
		ev.Eval.mgKingSafety[White] += -ev.Eval.kingAttacksWeight[Black] + KingZoneDefenseBonus*zoneDefense.count()
	}

	// TempoBonus
	ev.Eval.mgMaterial[pos.Turn] += TempoBonus
	ev.Eval.egMaterial[pos.Turn] += TempoBonus

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
	mg += ev.mgThreats[side] - ev.mgThreats[opponent]
	eg += ev.egThreats[side] - ev.egThreats[opponent]

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
		ev.Eval.mgThreats[side] += QueenAttackedByPawnThreatPenalty
		ev.Eval.egThreats[side] += QueenAttackedByPawnThreatPenalty
	}

	if fromBB&ev.EvalData.pinned > 0 {
		ev.Eval.mgThreats[side] += PinnedQueenThreatPenaltyMg
		ev.Eval.egThreats[side] += PinnedQueenThreatPenaltyEg
	}

	// Safe checks. Squares not defended by enemy pawns
	// where the queen can move to give check
	safeQueenChecks := Attacks(piece, ev.EvalData.kings[opponent], ev.EvalData.blocks) & ^ev.EvalData.attackedByPawns[opponent] & attacks
	if safeQueenChecks > 0 {
		ev.Eval.mgThreats[side] += SafeQueenCheckThreatBonus * safeQueenChecks.count()
		ev.Eval.egThreats[side] += SafeQueenCheckThreatBonus * safeQueenChecks.count()
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
		ev.Eval.mgThreats[side] += RookAttackedByPawnThreatPenalty
		ev.Eval.egThreats[side] += RookAttackedByPawnThreatPenalty
	}

	if fromBB&ev.EvalData.pinned > 0 {
		ev.Eval.mgThreats[side] += PinnedRookThreatPenaltyMg
		ev.Eval.egThreats[side] += PinnedRookThreatPenaltyEg
	}

	safeRookChecks := Attacks(piece, ev.EvalData.kings[opponent], ev.EvalData.blocks) & ^ev.EvalData.attackedByPawns[opponent] & attacks
	if safeRookChecks > 0 {
		ev.Eval.mgThreats[side] += SafeRookCheckThreatBonus * safeRookChecks.count()
		ev.Eval.egThreats[side] += SafeRookCheckThreatBonus * safeRookChecks.count()
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
		ev.Eval.mgThreats[side] += MinorAttackedByPawnThreatPenalty
		ev.Eval.egThreats[side] += MinorAttackedByPawnThreatPenalty
	}
	if attacks&pos.Pieces[pieceColor(Queen, side.Opponent())] > 0 {
		ev.Eval.mgThreats[side.Opponent()] += QueenAttackedByMinorThreatPenalty
		ev.Eval.egThreats[side.Opponent()] += QueenAttackedByMinorThreatPenalty
	}
	if attacks&pos.Pieces[pieceColor(Rook, side.Opponent())] > 0 {
		ev.Eval.mgThreats[side.Opponent()] += RookAttackedByMinorThreatPenalty
		ev.Eval.egThreats[side.Opponent()] += RookAttackedByMinorThreatPenalty
	}

	if fromBB&ev.EvalData.pinned > 0 {
		ev.Eval.mgThreats[side] += PinnedBishopThreatPenaltyMg
		ev.Eval.egThreats[side] += PinnedBishopThreatPenaltyEg
	}

	safeBishopChecks := Attacks(piece, ev.EvalData.kings[opponent], ev.EvalData.blocks) & ^ev.EvalData.attackedByPawns[opponent] & attacks
	if safeBishopChecks > 0 {
		ev.Eval.mgThreats[side] += SafeBishopCheckThreatBonus * safeBishopChecks.count()
		ev.Eval.egThreats[side] += SafeBishopCheckThreatBonus * safeBishopChecks.count()
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
		ev.Eval.mgThreats[side] += MinorAttackedByPawnThreatPenalty
		ev.Eval.egThreats[side] += MinorAttackedByPawnThreatPenalty
	}
	if attacks&pos.Pieces[pieceColor(Queen, side.Opponent())] > 0 {
		ev.Eval.mgThreats[side.Opponent()] += QueenAttackedByMinorThreatPenalty
		ev.Eval.egThreats[side.Opponent()] += QueenAttackedByMinorThreatPenalty
	}
	if attacks&pos.Pieces[pieceColor(Rook, side.Opponent())] > 0 {
		ev.Eval.mgThreats[side.Opponent()] += RookAttackedByMinorThreatPenalty
		ev.Eval.egThreats[side.Opponent()] += RookAttackedByMinorThreatPenalty
	}

	if fromBB&ev.EvalData.pinned > 0 {
		ev.Eval.mgThreats[side] += PinnedKnightThreatPenaltyMg
		ev.Eval.egThreats[side] += PinnedKnightThreatPenaltyEg
	}

	safeKnightChecks := Attacks(piece, ev.EvalData.kings[opponent], ev.EvalData.blocks) & ^ev.EvalData.attackedByPawns[opponent] & attacks
	if safeKnightChecks > 0 {
		ev.Eval.mgThreats[side] += SafeKnightCheckThreatBonus * safeKnightChecks.count()
		ev.Eval.egThreats[side] += SafeKnightCheckThreatBonus * safeKnightChecks.count()
	}

	ev.Eval.mgMobility[side] += KnightMobilityMg[squares]
	ev.Eval.egMobility[side] += KnightMobilityEg[squares]

	ev.Eval.phase += 3
}

// evaluatePawn evaluates the score of the pawn structure of the position
func (ev *Evaluation) evaluatePawns(pos *Position) {

	// Check Pawn Cache
	mgSc, egSc, hasPawnCache := ev.PawnCache.probe(pos.PawnHash, pos.Turn)
	if hasPawnCache {
		ev.Eval.mgPawnStrucutre[pos.Turn] = mgSc
		ev.Eval.egPawnStructure[pos.Turn] = egSc
		return
	}

	for side := Color(White); side <= Black; side++ {
		piece := pieceColor(Pawn, side)
		pawns := pos.Pieces[piece]
		opponent := side.Opponent()
		sidePawns := pawns

		backwardPawns := BackwardPawns(ev.EvalData.pawns[side], ev.EvalData.attackedByPawns[opponent], side)
		passedPawns := PassedPawns(ev.EvalData.pawns[side], ev.EvalData.pawns[opponent], side)

		for sidePawns > 0 {
			fromBB := sidePawns.NextBit()
			from := Bsf(fromBB)
			file := from % 8

			ev.Eval.mgPawnStrucutre[side] += middlegamePiecesScore[piece][from]
			ev.Eval.egPawnStructure[side] += endgamePiecesScore[piece][from]

			// Doubled. A pawn is doubled when another pawn is in the same file
			pawnsInFile := pawns & Files[file]
			if pawnsInFile.count() > 1 {
				ev.Eval.mgPawnStrucutre[side] += DoubledPawnPenaltyMg
				ev.Eval.egPawnStructure[side] += DoubledPawnPenaltyEg
			}

			// Isolated. A pawn is isolated when the adjacent files have no allied pawns
			if IsolatedAdjacentFilesMask[file]&pawns == 0 {
				ev.Eval.mgPawnStrucutre[side] += IsolatedPawnPenaltyMg
				ev.Eval.egPawnStructure[side] += IsolatedPawnPenaltyEg
			}

			// Backward. A pawn that cannot be safely advanced, because it will be captured by enemy pawns
			backward := backwardPawns&fromBB > 0
			if backward {
				ev.Eval.mgPawnStrucutre[side] += BackwardPawnPenaltyMg
				ev.Eval.egPawnStructure[side] += BackwardPawnPenaltyEg
			}

			// Passed. A pawn whose path to promotion is not blocked nor attacked by enemy pawns
			if passedPawns&fromBB > 0 {
				rank := from / 8
				if side == Black {
					rank = 7 - rank
				}
				ev.Eval.mgPawnStrucutre[side] += PassedPawnsBonusMg[rank]
				ev.Eval.egPawnStructure[side] += PassedPawnsBonusEg[rank]
			}

			// Defended pawn. A pawn that is defended by the same side pawns
			defenders := (pawnAttacks(&fromBB, opponent) & pawns).count()
			if defenders > 0 {
				ev.Eval.mgPawnStrucutre[side] += DefendedPawnBonusMg * defenders
				ev.Eval.egPawnStructure[side] += DefendedPawnBonusEg * defenders
			}

			// Connected Pawn. A pawn that has allies at adjacent files, and its not backward
			connected := (IsolatedAdjacentFilesMask[file] & pawns).count()
			if !backward && connected > 0 {
				ev.Eval.mgPawnStrucutre[side] += ConnectedPawnBonusMg * connected
				ev.Eval.egPawnStructure[side] += ConnectedPawnBonusEg * connected
			}
		}

		// Store pawn score on cache, always from White's perspective
		mgScWhite := ev.Eval.mgPawnStrucutre[White] - ev.Eval.mgPawnStrucutre[Black]
		egScWhite := ev.Eval.egPawnStructure[White] - ev.Eval.egPawnStructure[Black]
		ev.PawnCache.store(pos.PawnHash, mgScWhite, egScWhite)
	}

}

// OutpostSquares returns a bitboard of outpost squares for the given side
// An outpost square is:
// - In enemy territory (rank 4-6 for white, 3-5 for black)
// - Cannot be attacked by enemy pawns
// - Protected by own pawn(s)
func OutpostSquares(alliedPawns Bitboard, enemyPawns Bitboard, side Color) Bitboard {
	outpostRanks := OutpostsRanks[side]

	frontSpans := Bitboard(0)
	if side == White {
		frontSpans = ((fillDown(enemyPawns)&notAFile)>>1 | (fillDown(enemyPawns)&notHFile)<<1) >> 8
	} else {
		frontSpans = ((fillUp(enemyPawns)&notAFile)>>1 | (fillUp(enemyPawns)&notHFile)<<1) << 8
	}

	protectedByPawns := pawnAttacks(&alliedPawns, side)

	return ^frontSpans & protectedByPawns & outpostRanks
}

// BackwardPawns returns a bitboard with the pawns that are backwards
// A backward pawn is a pawn that is not member of own front-attackspans but controlled by a sentry (definition from CPW)
func BackwardPawns(pawns Bitboard, enemyPawnsAttacks Bitboard, side Color) Bitboard {
	if side == White {
		stops := pawns << 8
		frontSpans := (fillUp(stops)&notAFile)>>1 | (fillUp(stops)&notHFile)<<1
		return (stops & enemyPawnsAttacks & ^frontSpans) >> 8
	} else {
		stops := pawns >> 8
		frontSpans := (fillDown(stops)&notAFile)>>1 | (fillDown(stops)&notHFile)<<1
		return (stops & enemyPawnsAttacks & ^frontSpans) << 8
	}
}

// PassedPawns returns a bitboard with the passed pawns for the side
// A passed pawn is a pawn whose path to promotion is not blocked nor attacked by the enemy pawns
func PassedPawns(alliedPawns Bitboard, enemyPawns Bitboard, side Color) (passedPawns Bitboard) {
	fileSpan, adjacentSpans := Bitboard(0), Bitboard(0)
	if side == White {
		fileSpan = fillDown(enemyPawns)
		adjacentSpans = fillDown(enemyPawns >> 8)
	} else {
		fileSpan = fillUp(enemyPawns)
		adjacentSpans = fillUp(enemyPawns << 8)
	}

	blockedOrAttacked := fileSpan | (adjacentSpans&notAFile)>>1 | (adjacentSpans&notHFile)<<1
	return alliedPawns &^ blockedOrAttacked
}
