package engine

const (
	MaxPhaseValue = 24
	MgScoreMask   = 0xffff
)

var (
	// Mobility arrays based on the number of squares a piece can attack
	QueenMobility = [28]Score{
		S(-21, -77), S(-18, -66), S(-39, -56), S(-50, -74), S(-50, 24), S(-26, 76),
		S(-22, 111), S(-17, 126), S(-15, 149), S(-12, 172), S(-9, 177), S(-5, 183),
		S(-2, 190), S(1, 188), S(3, 190), S(4, 195), S(6, 192), S(5, 197),
		S(7, 195), S(10, 191), S(16, 187), S(26, 168), S(35, 161), S(45, 143),
		S(63, 133), S(58, 140), S(29, 138), S(11, 134),
	}
	RookMobility = [15]Score{
		S(-30, -16), S(-30, 4), S(-12, 33), S(-6, 56), S(-1, 68), S(1, 77),
		S(2, 83), S(4, 89), S(8, 89), S(12, 92), S(15, 96), S(16, 100),
		S(19, 102), S(23, 100), S(29, 97),
	}
	BishopMobility = [14]Score{
		S(-36, -127), S(-53, -48), S(-24, 2), S(-13, 24), S(-3, 34), S(4, 42),
		S(10, 51), S(16, 55), S(17, 61), S(24, 61), S(30, 61), S(48, 53),
		S(57, 54), S(70, 44),
	}
	KnightMobility = [9]Score{
		S(-120, -148), S(-28, -26), S(-7, 10), S(3, 31), S(14, 41), S(17, 52),
		S(27, 53), S(38, 54), S(51, 50),
	}

	// Material Adjustment
	BishopPairBonus         = S(25, 65)
	RookOnOpenFileBonus     = S(32, 12)
	RookOnSemiOpenFileBonus = S(11, 21)
	RookOnSeventhRankBonus  = S(8, 20)
	QueenOnSeventhRankBonus = S(0, 2)
	KnightOutpostBonus      = S(38, 20)
	ConnectedKnightBonus    = S(0, -4)
	BishopOutpostBonus      = S(42, 0)

	// Pawn Structure
	DoubledPawnPenalty        = S(3, -13)
	IsolatedPawnPenalty       = S(-2, -2)
	BackwardPawnPenalty       = S(-2, -3)
	DefendedPawnBonus         = S(9, 8)
	ConnectedPawnBonus        = S(7, 4)
	PassedPawnsBonus          = [8]Score{S(0, 0), S(3, 11), S(-3, 17), S(-5, 45), S(20, 75), S(24, 145), S(12, 93), S(0, 0)}
	CandidatePassedPawnsBonus = [2][8]Score{
		{S(0, 0), S(-34, 9), S(-15, 10), S(7, 38), S(30, 58), S(56, 85), S(0, 65), S(0, 0)},
		{S(0, 0), S(-19, 1), S(-4, 21), S(6, 40), S(46, 74), S(50, 150), S(0, 97), S(0, 0)},
	}

	// King Safety
	PawnShield = [2][8]Score{
		{S(29, -12), S(27, -20), S(19, -14), S(13, -24), S(14, -32), S(8, -33), S(2, -4), S(0, 0)},
		{S(0, 0), S(36, -20), S(32, -18), S(17, -27), S(15, -36), S(-6, -19), S(0, -8), S(0, 0)},
	}
	PawnStorm = [2][2][8]Score{
		{
			{S(18, 24), S(-25, -25), S(-23, -5), S(-17, -10), S(1, -12), S(9, -16), S(9, -14), S(0, 0)},
			{S(17, 39), S(-3, -12), S(-2, -17), S(6, -20), S(6, -4), S(9, -11), S(32, 9), S(0, 0)},
		},
		{
			{S(0, 0), S(63, 22), S(-6, -5), S(2, -8), S(10, -12), S(16, -15), S(16, -12), S(0, 0)},
			{S(0, 0), S(31, 46), S(-10, -29), S(16, -12), S(23, -13), S(20, -5), S(41, -2), S(0, 0)},
		},
	}
	KingOnOpenFiles = [2]Score{S(2, -23), S(-32, -20)}

	KnightAttackWeight   = S(22, -6)
	BishopAttackWeight   = S(21, -9)
	RookAttackWeight     = S(26, -31)
	QueenAttackWeight    = S(6, -10)
	KingZoneDefenseBonus = S(10, 2)
	SafeQueenCheck       = S(16, 16)
	SafeRookCheck        = S(31, 14)
	SafeBishopCheck      = S(18, 18)
	SafeKnightCheck      = S(14, 18)

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
	// otherwise we will have overflow and wrong values
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

	phase := GetEvalPhase(pos)
	mgPhase := min(phase, MaxPhaseValue)
	egPhase := MaxPhaseValue - phase
	mg, eg := sc.Get()
	score = (mg*mgPhase + eg*egPhase) / MaxPhaseValue
	if pos.Turn == Black {
		score = -score
	}
	return
}

// evaluateKings returns the score of the Kings in the position for the side passed
func (ev *Evaluation) evaluateKings(pos *Position, side Color) (sc Score) {
	opponent := side.Opponent()
	king := pos.Pieces[PieceOf(King, side)]
	from := Bsf(king)
	sc += PieceSquaresScores[King][relativePsqtSquare(from, side)]

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
	if ev.EvalData.kingAttackersCount[opponent] > (1 - pos.Pieces[PieceOf(Queen, opponent)].Count()) {

		// Safe checks
		// Possible squares where an enemy piece can move to give check to our king
		defendedByPawns := ev.EvalData.attackedBy[side][Pawn]
		knightChecksThreats := knightAttacksTable[from]
		bishopChecksThreats := bishopAttacks(from, pos.Sides[All])
		rookChecksThreats := rookAttacks(from, pos.Sides[All])
		queenChecksThreats := queenAttacks(&king, pos.Sides[All])

		knightChecks := (knightChecksThreats & ^defendedByPawns & ev.EvalData.attackedBy[opponent][Knight]).Count()
		bishopChecks := (bishopChecksThreats & ^defendedByPawns & ev.EvalData.attackedBy[opponent][Bishop]).Count()
		rookChecks := (rookChecksThreats & ^defendedByPawns & ev.EvalData.attackedBy[opponent][Rook]).Count()
		queenChecks := (queenChecksThreats & ^defendedByPawns & ev.EvalData.attackedBy[opponent][Queen]).Count()

		defendedSquares := (KingZone[side][from] & defendedByPawns).Count()
		// Safe checks threaten our king: penalize the defender (the attacker gains
		// as much, mirroring the positive Safe*Check values).
		sc += -ev.EvalData.kingAttackersWeight[opponent] + Score(defendedSquares)*KingZoneDefenseBonus -
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
		if attacks&pos.Pieces[PieceOf(Knight, side)] > 0 {
			sc += ConnectedKnightBonus
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
				rank := from / 8
				if side == Black {
					rank = 7 - rank
				}
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
			rank := sq / 8
			if side == Black {
				rank = 7 - rank
			}
			if support.Count() >= threats.Count() {
				return 1, rank
			}
			return 0, rank
		}

	}
	return -1, 0
}
