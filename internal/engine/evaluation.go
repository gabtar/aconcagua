package engine

const (
	MaxPhaseValue = 24
	MgScoreMask   = 0xffff

	// Endgame Scaling Factors
	ScaleNormal      = 128
	ScalePureOCB     = 48
	ScaleOCBPawnStep = 8
	ScaleOCBKnights  = 80
	ScaleOCBMajors   = 100
	ScaleNoPawns     = 8
)

var (
	// Mobility arrays based on the number of squares a piece can attack
	QueenMobility = [28]Score{
		S(-21, -77), S(-18, -66), S(-41, -56), S(-44, -75), S(-45, 30), S(-23, 77),
		S(-19, 117), S(-15, 136), S(-12, 159), S(-10, 182), S(-7, 188), S(-3, 195),
		S(0, 203), S(3, 203), S(4, 205), S(5, 211), S(6, 210), S(5, 216),
		S(7, 215), S(9, 212), S(17, 206), S(30, 183), S(45, 169), S(68, 138),
		S(64, 139), S(82, 117), S(32, 134), S(9, 128),
	}
	RookMobility = [15]Score{
		S(-23, -16), S(-21, 0), S(-10, 38), S(-5, 64), S(-1, 77), S(1, 86),
		S(2, 93), S(4, 100), S(7, 101), S(11, 104), S(13, 109), S(14, 113),
		S(16, 117), S(20, 116), S(30, 109),
	}
	BishopMobility = [14]Score{
		S(-37, -130), S(-51, -47), S(-24, 8), S(-14, 32), S(-3, 43), S(3, 50),
		S(9, 60), S(14, 65), S(15, 71), S(20, 72), S(24, 73), S(38, 67),
		S(42, 71), S(47, 63),
	}
	KnightMobility = [9]Score{
		S(-124, -130), S(-33, -25), S(-12, 15), S(0, 38), S(11, 49), S(14, 61),
		S(24, 63), S(35, 65), S(49, 60),
	}

	// Material Adjustment
	BishopPairBonus         = S(25, 63)
	RookOnOpenFileBonus     = S(31, 13)
	RookOnSemiOpenFileBonus = S(12, 23)
	RookOnSeventhRankBonus  = S(-6, 31)
	QueenOnSeventhRankBonus = S(-8, 24)
	KnightOutpostBonus      = S(36, 19)
	BishopOutpostBonus      = S(39, 0)

	// Pawn Structure
	DoubledPawnPenalty        = S(3, -12)
	IsolatedPawnPenalty       = S(-2, -2)
	BackwardPawnPenalty       = S(-3, -1)
	DefendedPawnBonus         = S(9, 9)
	ConnectedPawnBonus        = S(6, 6)
	PassedPawnsBonus          = [8]Score{S(0, 0), S(-3, -7), S(-9, 4), S(-13, 44), S(9, 87), S(18, 164), S(7, 116), S(0, 0)}
	CandidatePassedPawnsBonus = [2][8]Score{
		{S(0, 0), S(-41, 17), S(-16, 10), S(6, 41), S(31, 64), S(60, 91), S(0, 65), S(0, 0)},
		{S(0, 0), S(-20, 5), S(-4, 23), S(5, 42), S(41, 76), S(69, 153), S(0, 97), S(0, 0)},
	}

	// King Safety
	PawnShield = [2][8]Score{
		{S(31, -13), S(27, -20), S(19, -13), S(14, -24), S(14, -30), S(6, -27), S(4, 0), S(0, 0)},
		{S(0, 0), S(35, -21), S(32, -17), S(17, -28), S(14, -36), S(-16, -12), S(6, -10), S(0, 0)},
	}
	PawnStorm = [2][2][8]Score{
		{
			{S(6, 31), S(-68, -54), S(-22, -2), S(-17, -4), S(1, -5), S(8, -10), S(9, -10), S(0, 0)},
			{S(26, 58), S(-9, -34), S(-4, -13), S(6, -15), S(5, 0), S(10, -8), S(35, 6), S(0, 0)},
		},
		{
			{S(0, 0), S(67, 24), S(-6, -1), S(2, -1), S(9, -4), S(15, -7), S(15, -5), S(0, 0)},
			{S(0, 0), S(41, 65), S(-11, -21), S(14, -4), S(21, -5), S(20, 0), S(61, -17), S(0, 0)},
		},
	}
	KingOnOpenFiles = [2]Score{S(1, -18), S(-26, -14)}

	KnightAttackWeight   = S(20, -2)
	BishopAttackWeight   = S(18, -6)
	RookAttackWeight     = S(18, -16)
	QueenAttackWeight    = S(1, 0)
	KingZoneDefenseBonus = S(9, 5)
	SafeQueenCheck       = S(19, 11)
	SafeRookCheck        = S(40, -1)
	SafeBishopCheck      = S(17, 17)
	SafeKnightCheck      = S(15, 15)

	// Threats
	MinorAttackedByPawnThreat  = S(-52, -21)
	MajorAttackedByPawnThreat  = S(-60, 9)
	RookAttackedByMinorThreat  = S(-52, -5)
	QueenAttackedByMinorThreat = S(-44, 1)
	HangingPawnThreat          = S(0, -19)

	PinnedPieceThreat = [4]Score{S(-67, -60), S(-42, -25), S(-9, -73), S(-28, -64)}

	// Passed Pawns
	PassedKingDistance      = S(3, -8)
	PassedEnemyKingDistance = S(-1, 7)
	PassedPawnProtected     = S(1, -2)

	Tempo = 23

	// OutpostsRanks contains the bitboard mask for ranks that are considered outposts
	OutpostsRanks = [2]Bitboard{
		Ranks[3] | Ranks[4] | Ranks[5],
		Ranks[2] | Ranks[3] | Ranks[4],
	}
)

// Score encodes the values for middlegame and endgame score of an evaluation feature
type Score int32

// NewScore creates an Score from mg and eg values
func NewScore(mg int, eg int) Score {
	return Score(int32(mg) + int32(uint16(eg))<<16)
}

// S is an alias/shorthand for NewScore()
var S = NewScore

// Get returns the mg and eg values of the score
func (sc Score) Get() (mg int, eg int) {
	mg = int(int16(sc & MgScoreMask))
	// NOTE: as used in Ehtereal/Berserk engine the 0x8000 compensates the shift for eg,
	// otherwise we will have overflow and get wrong values
	eg = int(int16((sc + 0x8000) >> 16))
	return
}

// Evaluation contains the elements for evaluation of a position
type Evaluation struct {
	EvalData  EvalData
	PawnCache PawnHashTable
}

// EvalData contains positional data about the current position
type EvalData struct {
	kings               [2]Bitboard
	pawns               [2]Bitboard
	attackedBy          [2][6]Bitboard
	pinnedPieces        Bitboard
	outposts            [2]Bitboard
	kingAttackersCount  [2]int
	kingAttackersWeight [2]Score
}

// NewEvaluation returns a new Evaluation
func NewEvaluation(size int) *Evaluation {
	return &Evaluation{
		EvalData:  EvalData{},
		PawnCache: *NewPawnHashTable(size),
	}
}

// Clear clears the evaluation
func (ev *Evaluation) Clear() {
	ev.EvalData.clear()
	ev.PawnCache.clear()
}

// clear clears the EvalData
func (ed *EvalData) clear() {
	ed.kings = [2]Bitboard{0, 0}
	ed.pawns = [2]Bitboard{0, 0}
	ed.attackedBy = [2][6]Bitboard{}
	ed.pinnedPieces = Bitboard(0)
	ed.outposts = [2]Bitboard{0, 0}
	ed.kingAttackersCount = [2]int{0, 0}
	ed.kingAttackersWeight = [2]Score{0, 0}
}

// init initializes the evaluation data
func (ed *EvalData) init(pos *Position) {
	ed.clear()
	ed.kings[White] = pos.Pieces[WhiteKing]
	ed.kings[Black] = pos.Pieces[BlackKing]
	ed.pawns[White] = pos.Pieces[WhitePawn]
	ed.pawns[Black] = pos.Pieces[BlackPawn]
	ed.attackedBy[White][Pawn] = pawnAttacks(&pos.Pieces[WhitePawn], White)
	ed.attackedBy[Black][Pawn] = pawnAttacks(&pos.Pieces[BlackPawn], Black)
	ed.pinnedPieces = pos.PinnedPieces(White) | pos.PinnedPieces(Black)
	ed.outposts = [2]Bitboard{
		OutpostSquares(ed.pawns[White], ed.pawns[Black], White),
		OutpostSquares(ed.pawns[Black], ed.pawns[White], Black),
	}
}

// GetEvalPhase returns the mg phase value of the position depending on the remaining material on it
func GetEvalPhase(pos *Position) int {
	return (pos.Pieces[WhiteQueen]|pos.Pieces[BlackQueen]).Count()*4 +
		(pos.Pieces[WhiteRook]|pos.Pieces[BlackRook]).Count()*2 +
		(pos.Pieces[WhiteBishop] | pos.Pieces[BlackBishop]).Count() +
		(pos.Pieces[WhiteKnight] | pos.Pieces[BlackKnight]).Count()
}

// Evaluate returns the static score of the position
func (ev *Evaluation) Evaluate(pos *Position) (score int) {
	ev.EvalData.init(pos)
	sc := S(0, 0)

	sc += ev.evaluatePawns(pos)
	sc += ev.evaluateQueens(pos, White) - ev.evaluateQueens(pos, Black)
	sc += ev.evaluateRooks(pos, White) - ev.evaluateRooks(pos, Black)
	sc += ev.evaluateBishops(pos, White) - ev.evaluateBishops(pos, Black)
	sc += ev.evaluateKnights(pos, White) - ev.evaluateKnights(pos, Black)
	sc += ev.evaluateKings(pos, White) - ev.evaluateKings(pos, Black)
	sc += ev.evaluatePassed(pos)
	sc += ev.evaluateThreats(pos)

	phase := GetEvalPhase(pos)
	mgPhase := min(phase, MaxPhaseValue)
	egPhase := MaxPhaseValue - mgPhase
	mg, eg := sc.Get()

	scale := ScaleFactor(pos)
	score = (mg*mgPhase + eg*egPhase*scale/ScaleNormal) / MaxPhaseValue
	if pos.Turn == Black {
		score = -score
	}
	score += Tempo
	return
}

// evaluateKings returns the score of the Kings in the position for the side passed
func (ev *Evaluation) evaluateKings(pos *Position, side Color) (sc Score) {
	opponent := side.Opponent()
	king := pos.Pieces[PieceOf(King, side)]
	from := Bsf(king)
	sc += PieceSquaresScores[King][relativePsqtSquare(from, side)]
	ev.EvalData.attackedBy[side][King] = kingAttacksTable[from]

	kingFile, kingRank := from%8, from/8

	// Squares in front of the king on the king file and the two adjacent files
	frontMask := KingFrontMask[side][from]

	for file := max(0, kingFile-1); file <= min(7, kingFile+1); file++ {
		shielders := pos.Pieces[PieceOf(Pawn, side)] & Files[file] & frontMask
		stormers := pos.Pieces[PieceOf(Pawn, opponent)] & Files[file] & frontMask

		sameFile := 0
		if file == kingFile {
			sameFile = 1
		}

		// Shield. Nearest allied pawn in front of the king
		shieldRank := 8
		if shielders > 0 {
			shieldRank = NearestFromSide(shielders, side) / 8
			sc += PawnShield[sameFile][abs(kingRank-shieldRank)]
		}

		// Storm. Most advanced enemy pawn ahead of the king
		if stormers > 0 {
			stormRank := NearestFromSide(stormers, opponent) / 8
			blocked := 0
			if shieldRank != 8 && abs(shieldRank-stormRank) == 1 {
				blocked = 1
			}
			sc += PawnStorm[sameFile][blocked][abs(kingRank-stormRank)]
		}

		// King on open/near open files
		openFile := (shielders | stormers) == 0
		if openFile {
			sc += KingOnOpenFiles[sameFile]
		}
	}

	// Safety. Only apply penalties when we have enough attackers(1 piece and at least a queen or more than 2 pieces)
	enemyQueens := pos.Pieces[PieceOf(Queen, opponent)].Count()
	if ev.EvalData.kingAttackersCount[opponent] > (1 - enemyQueens) {

		// Possible squares where an enemy piece can move to give check to our king
		defendedByPawns := ev.EvalData.attackedBy[side][Pawn]
		knightChecksThreats := knightAttacksTable[from]
		bishopChecksThreats := bishopAttacks(from, pos.Sides[All])
		rookChecksThreats := rookAttacks(from, pos.Sides[All])
		queenChecksThreats := bishopChecksThreats | rookChecksThreats

		defendedSquares := (KingZone[side][from] & defendedByPawns).Count()
		knightChecks := (knightChecksThreats & ^defendedByPawns & ev.EvalData.attackedBy[opponent][Knight]).Count()
		bishopChecks := (bishopChecksThreats & ^defendedByPawns & ev.EvalData.attackedBy[opponent][Bishop]).Count()
		rookChecks := (rookChecksThreats & ^defendedByPawns & ev.EvalData.attackedBy[opponent][Rook]).Count()
		queenChecks := (queenChecksThreats & ^defendedByPawns & ev.EvalData.attackedBy[opponent][Queen]).Count()

		sc += Score(defendedSquares)*KingZoneDefenseBonus - ev.EvalData.kingAttackersWeight[opponent] -
			Score(knightChecks)*SafeKnightCheck - Score(bishopChecks)*SafeBishopCheck -
			Score(rookChecks)*SafeRookCheck - Score(queenChecks)*SafeQueenCheck
	}

	return sc
}

// evaluateQueens returns the score of the Queens in the position for the side passed
func (ev *Evaluation) evaluateQueens(pos *Position, side Color) (sc Score) {
	opponent := side.Opponent()
	queens := pos.Pieces[PieceOf(Queen, side)]
	for queens > 0 {
		nextQueen := queens.NextBit()
		from := Bsf(nextQueen)
		sc += PieceSquaresScores[Queen][relativePsqtSquare(from, side)]

		attacks := queenAttacks(&nextQueen, pos.Sides[All])
		ev.EvalData.attackedBy[side][Queen] |= attacks
		safeSquares := (attacks & ^ev.EvalData.attackedBy[opponent][Pawn]).Count()
		sc += QueenMobility[safeSquares]

		enemyKingZone := KingZone[opponent][Bsf(ev.EvalData.kings[opponent])]
		if attacks&enemyKingZone != 0 {
			ev.EvalData.kingAttackersCount[side]++
			ev.EvalData.kingAttackersWeight[side] += QueenAttackWeight
		}

		kingRank := RankRelativeToSide(Bsf(ev.EvalData.kings[opponent]), side)
		queenRank := RankRelativeToSide(from, side)
		if queenRank == 6 && kingRank == 7 {
			sc += QueenOnSeventhRankBonus
		}

		if ev.EvalData.pinnedPieces&nextQueen > 0 {
			sc += PinnedPieceThreat[Queen-1]
		}
	}
	return sc
}

// evaluateRooks returns the socre of the Rooks in the position for the side passed
func (ev *Evaluation) evaluateRooks(pos *Position, side Color) (sc Score) {
	opponent := side.Opponent()
	rooks := pos.Pieces[PieceOf(Rook, side)]
	for rooks > 0 {
		nextRook := rooks.NextBit()
		from := Bsf(nextRook)
		file := from % 8
		sc += PieceSquaresScores[Rook][relativePsqtSquare(from, side)]

		attacks := rookAttacks(from, pos.Sides[All])
		ev.EvalData.attackedBy[side][Rook] |= attacks
		safeSquares := (attacks & ^ev.EvalData.attackedBy[opponent][Pawn]).Count()
		sc += RookMobility[safeSquares]

		enemyKingZone := KingZone[opponent][Bsf(ev.EvalData.kings[opponent])]
		if attacks&enemyKingZone != 0 {
			ev.EvalData.kingAttackersCount[side]++
			ev.EvalData.kingAttackersWeight[side] += RookAttackWeight
		}

		// Open files
		if (ev.EvalData.pawns[side]|ev.EvalData.pawns[opponent])&Files[file] == 0 {
			sc += RookOnOpenFileBonus
		} else if ev.EvalData.pawns[side]&Files[file] == 0 && ev.EvalData.pawns[opponent]&Files[file] > 0 {
			sc += RookOnSemiOpenFileBonus
		}

		// Rook on 7th
		kingRank := RankRelativeToSide(Bsf(ev.EvalData.kings[opponent]), side)
		rookRank := RankRelativeToSide(from, side)
		if kingRank == 7 && rookRank == 6 {
			sc += RookOnSeventhRankBonus
		}

		if ev.EvalData.pinnedPieces&nextRook > 0 {
			sc += PinnedPieceThreat[Rook-1]
		}
	}
	return sc
}

// evaluateBishops returns the score of the Bishops in the position for the side passed
func (ev *Evaluation) evaluateBishops(pos *Position, side Color) (sc Score) {
	opponent := side.Opponent()
	bishops := pos.Pieces[PieceOf(Bishop, side)]

	// Bishop pair bonus
	if bishops.Count() >= 2 {
		sc += BishopPairBonus
	}

	for bishops > 0 {
		nextBishop := bishops.NextBit()
		from := Bsf(nextBishop)
		sc += PieceSquaresScores[Bishop][relativePsqtSquare(from, side)]

		attacks := bishopAttacks(from, pos.Sides[All])
		ev.EvalData.attackedBy[side][Bishop] |= attacks
		safeSquares := (attacks & ^ev.EvalData.attackedBy[opponent][Pawn]).Count()
		sc += BishopMobility[safeSquares]

		enemyKingZone := KingZone[opponent][Bsf(ev.EvalData.kings[opponent])]
		if attacks&enemyKingZone != 0 {
			ev.EvalData.kingAttackersCount[side]++
			ev.EvalData.kingAttackersWeight[side] += BishopAttackWeight
		}

		if nextBishop&ev.EvalData.outposts[side] > 0 {
			sc += BishopOutpostBonus
		}

		if ev.EvalData.pinnedPieces&nextBishop > 0 {
			sc += PinnedPieceThreat[Bishop-1]
		}
	}
	return sc
}

// evaluateKnights returns the score of the Knights in the position for the side passed
func (ev *Evaluation) evaluateKnights(pos *Position, side Color) (sc Score) {
	opponent := side.Opponent()
	knights := pos.Pieces[PieceOf(Knight, side)]

	for knights > 0 {
		nextKnight := knights.NextBit()
		from := Bsf(nextKnight)
		sc += PieceSquaresScores[Knight][relativePsqtSquare(from, side)]

		attacks := knightAttacksTable[from]
		ev.EvalData.attackedBy[side][Knight] |= attacks
		safeSquares := (attacks & ^ev.EvalData.attackedBy[opponent][Pawn]).Count()
		sc += KnightMobility[safeSquares]

		enemyKingZone := KingZone[opponent][Bsf(ev.EvalData.kings[opponent])]
		if attacks&enemyKingZone != 0 {
			ev.EvalData.kingAttackersCount[side]++
			ev.EvalData.kingAttackersWeight[side] += KnightAttackWeight
		}

		if nextKnight&ev.EvalData.outposts[side] > 0 {
			sc += KnightOutpostBonus
		}

		if ev.EvalData.pinnedPieces&nextKnight > 0 {
			sc += PinnedPieceThreat[Knight-1]
		}
	}
	return sc
}

// evaluatePawns returns the score of the Pawns in the position for the side passed
func (ev *Evaluation) evaluatePawns(pos *Position) (sc Score) {
	// Check Pawn Cache
	mgSc, egSc, hasPawnCache := ev.PawnCache.probe(pos.PawnHash)
	if hasPawnCache {
		sc = S(mgSc, egSc)
		return
	}

	sideModifier := [2]Score{Score(1), Score(-1)}
	backwardsPawns := BackwardPawns(ev.EvalData.pawns[White], ev.EvalData.attackedBy[Black][Pawn], White) |
		BackwardPawns(ev.EvalData.pawns[Black], ev.EvalData.attackedBy[White][Pawn], Black)
	passedPawns := PassedPawns(ev.EvalData.pawns[White], ev.EvalData.pawns[Black], White) |
		PassedPawns(ev.EvalData.pawns[Black], ev.EvalData.pawns[White], Black)

	for side := Color(White); side <= Black; side++ {
		opponent := side.Opponent()
		alliedPawns := pos.Pieces[PieceOf(Pawn, side)]
		enemyPawns := pos.Pieces[PieceOf(Pawn, opponent)]
		pawns := alliedPawns

		for pawns > 0 {
			nextPawn := pawns.NextBit()
			from := Bsf(nextPawn)
			file := from % 8
			sc += sideModifier[side] * PieceSquaresScores[Pawn][relativePsqtSquare(from, side)]

			// Doubled. A pawn is doubled when another pawn is in the same file
			pawnsInFile := alliedPawns & Files[file]
			if pawnsInFile.Count() > 1 {
				sc += sideModifier[side] * DoubledPawnPenalty
			}

			// Isolated. A pawn is isolated when the adjacent files have no allied pawns
			if IsolatedAdjacentFilesMask[file]&alliedPawns == 0 {
				sc += sideModifier[side] * IsolatedPawnPenalty
			}

			// Backward. A pawn that cannot be safely advanced, because it will be captured by enemy pawns
			backward := backwardsPawns&nextPawn > 0
			if backward {
				sc += sideModifier[side] * BackwardPawnPenalty
			}

			// Passed. A pawn whose path to promotion is not blocked nor attacked by enemy pawns
			if passedPawns&nextPawn > 0 {
				rank := RankRelativeToSide(from, side)
				sc += sideModifier[side] * PassedPawnsBonus[rank]
			} else { // Check possible candidate passed pawn
				candidateFlag, rank := CandidatePassedPawn(nextPawn, alliedPawns, enemyPawns, side)
				if candidateFlag >= 0 {
					sc += sideModifier[side] * CandidatePassedPawnsBonus[candidateFlag][rank]
				}
			}

			// Defended pawn. A pawn that is defended by the same side pawns
			defenders := (pawnAttacks(&nextPawn, opponent) & alliedPawns).Count()
			if defenders > 0 {
				sc += sideModifier[side] * DefendedPawnBonus * Score(defenders)
			}

			// Connected Pawn. A pawn that has allies at adjacent files, and its not backward
			connected := (IsolatedAdjacentFilesMask[file] & alliedPawns).Count()
			if !backward && connected > 0 {
				sc += sideModifier[side] * ConnectedPawnBonus * Score(connected)
			}
		}
	}

	mgSc, egSc = sc.Get()
	ev.PawnCache.store(pos.PawnHash, mgSc, egSc)
	return sc
}

// evaluatePassed returns the score of the passed pawns of the postion
func (ev *Evaluation) evaluatePassed(pos *Position) (sc Score) {
	sideModifier := [2]Score{Score(1), Score(-1)}
	for side := Color(White); side <= Black; side++ {
		opponent := side.Opponent()
		alliedPawns := pos.Pieces[PieceOf(Pawn, side)]
		enemyPawns := pos.Pieces[PieceOf(Pawn, opponent)]
		passed := PassedPawns(alliedPawns, enemyPawns, side)
		protected := pawnAttacks(&alliedPawns, opponent)

		for passed > 0 {
			nextPassed := passed.NextBit()
			from := Bsf(nextPassed)
			alliedKingSq := Bsf(ev.EvalData.kings[side])
			enemyKingSq := Bsf(ev.EvalData.kings[opponent])

			sc += sideModifier[side] * Score(ManhattanDistance(from, alliedKingSq)) * PassedKingDistance
			sc += sideModifier[side] * Score(ManhattanDistance(from, enemyKingSq)) * PassedEnemyKingDistance
			if protected&alliedPawns > 0 {
				sc += sideModifier[side] * PassedPawnProtected
			}
		}
	}
	return
}

// ManhattanDistance between two squares is determined by the minimal number of orthogonal King moves between these squares on
// an empty board
func ManhattanDistance(from int, to int) int {
	fromRank, fromFile := from/8, from%8
	toRank, toFile := to/8, to%8

	return abs(toRank-fromRank) + abs(toFile-fromFile)
}

// evaluateThreats evaluates the threats in the position
func (ev *Evaluation) evaluateThreats(pos *Position) (sc Score) {
	sideModifier := [2]Score{Score(1), Score(-1)}

	for side := Color(White); side <= Black; side++ {
		opponent := side.Opponent()
		attackedByMinors := ev.EvalData.attackedBy[opponent][Knight] | ev.EvalData.attackedBy[opponent][Bishop]
		attackedByMajors := ev.EvalData.attackedBy[opponent][Rook] | ev.EvalData.attackedBy[opponent][Queen]
		attackedByPawns := ev.EvalData.attackedBy[opponent][Pawn]
		attackedByEnemy := attackedByMinors | attackedByMajors | attackedByPawns
		defendedByAllied := ev.EvalData.attackedBy[side][Pawn] | ev.EvalData.attackedBy[side][Knight] |
			ev.EvalData.attackedBy[side][Bishop] | ev.EvalData.attackedBy[side][Rook] |
			ev.EvalData.attackedBy[side][Queen] | ev.EvalData.attackedBy[side][King]

		minors := pos.Pieces[PieceOf(Knight, side)] | pos.Pieces[PieceOf(Bishop, side)]
		rooks := pos.Pieces[PieceOf(Rook, side)]
		queens := pos.Pieces[PieceOf(Queen, side)]

		miniorsAttackedByPawns := (minors & attackedByPawns).Count()
		if miniorsAttackedByPawns > 0 {
			sc += Score(miniorsAttackedByPawns) * sideModifier[side] * MinorAttackedByPawnThreat
		}

		majorsAttackedByPawns := ((rooks | queens) & attackedByPawns).Count()
		if majorsAttackedByPawns > 0 {
			sc += Score(majorsAttackedByPawns) * sideModifier[side] * MajorAttackedByPawnThreat
		}

		rooksAttackedByMinors := (rooks & attackedByMinors).Count()
		if rooksAttackedByMinors > 0 {
			sc += Score(rooksAttackedByMinors) * sideModifier[side] * RookAttackedByMinorThreat
		}

		queensAttackedByMinors := (queens & attackedByMinors).Count()
		if queensAttackedByMinors > 0 {
			sc += Score(queensAttackedByMinors) * sideModifier[side] * QueenAttackedByMinorThreat
		}

		hangingPawns := (pos.Pieces[PieceOf(Pawn, side)] & attackedByEnemy & ^defendedByAllied).Count()
		if hangingPawns > 0 {
			sc += Score(hangingPawns) * sideModifier[side] * HangingPawnThreat
		}
	}

	return
}

// ScaleFactor returns the scale factor for the endgame
func ScaleFactor(pos *Position) int {
	whitePawns := pos.Pieces[WhitePawn].Count()
	blackPawns := pos.Pieces[BlackPawn].Count()
	whiteMaterial := nonPawnMaterial(pos, White) + whitePawns
	blackMaterial := nonPawnMaterial(pos, Black) + blackPawns

	// Strong side by remaining material
	strong, strongPawns := Color(White), whitePawns
	if blackMaterial > whiteMaterial || (blackMaterial == whiteMaterial && blackPawns > whitePawns) {
		strong, strongPawns = Black, blackPawns
	}
	weak := strong.Opponent()

	// Opposite color bishops
	if pos.isOppositeColorBishops() {
		knights := (pos.Pieces[WhiteKnight] | pos.Pieces[BlackKnight]).Count()
		majors := (pos.Pieces[WhiteRook] | pos.Pieces[WhiteQueen] |
			pos.Pieces[BlackRook] | pos.Pieces[BlackQueen]).Count()

		if knights == 0 && majors == 0 {
			return min(ScaleNormal, ScalePureOCB+ScaleOCBPawnStep*strongPawns)
		}

		if knights > 0 && majors == 0 {
			return ScaleOCBKnights
		}

		if majors > 0 {
			return ScaleOCBMajors
		}
	}

	// Strong side has no pawns: it can't win without a decisive material lead
	if strongPawns == 0 && nonPawnMaterial(pos, strong)-nonPawnMaterial(pos, weak) <= 3 {
		return ScaleNoPawns
	}

	// Scale down as the number of pawns of the strong side reduces
	return min(ScaleNormal, 96+strongPawns*8)
}

// nonPawnMaterial returns the non pawn material of a side
func nonPawnMaterial(pos *Position, side Color) int {
	return pos.Pieces[PieceOf(Knight, side)].Count()*3 +
		pos.Pieces[PieceOf(Bishop, side)].Count()*3 +
		pos.Pieces[PieceOf(Rook, side)].Count()*5 +
		pos.Pieces[PieceOf(Queen, side)].Count()*9
}

// relativePsqtSquare returns the square number from the point of view of the side passed
func relativePsqtSquare(sq int, side Color) int {
	if side == White {
		return sq ^ 56
	}
	return sq
}

// RankRelativeToSide returns the relative rank from the point of view of the side
func RankRelativeToSide(from int, side Color) int {
	if side == White {
		return from / 8
	}
	return 7 - from/8
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

// CandidatePassedPawn returns a flag/type and the relative rank if we can create a passed pawn
func CandidatePassedPawn(pawn Bitboard, alliedPawns Bitboard, enemyPawns Bitboard, side Color) (flag int, rank int) {
	sq := Bsf(pawn)
	opponent := side.Opponent()
	stoppers := Bitboard(0)
	if side == White {
		stoppers = enemyPawns & (attacksFrontSpans[side][sq] | fillUp(bitboardFromIndex(sq+8)))
	} else {
		stoppers = enemyPawns & (attacksFrontSpans[side][sq] | fillDown(bitboardFromIndex(sq-8)))
	}

	if stoppers > 0 {
		threats := pawnAttacks(&pawn, side) & enemyPawns
		support := pawnAttacks(&pawn, opponent) & alliedPawns
		pushSquare := pawnPushesTable[side][sq]
		pushThreats := pawnAttacks(&pushSquare, side) & enemyPawns
		pushSupport := pawnAttacks(&pushSquare, opponent) & alliedPawns
		leftovers := stoppers ^ threats ^ pushThreats
		if leftovers == 0 && pushSupport.Count() >= pushThreats.Count() {
			rank := RankRelativeToSide(sq, side)
			if support.Count() >= threats.Count() {
				return 1, rank
			}
			return 0, rank
		}

	}
	return -1, 0
}
