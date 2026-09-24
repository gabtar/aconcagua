package engine

const (
	MaxPhaseValue = 24
	MgScoreMask   = 0xffff
)

var (
	// Mobility arrays based on the number of squares a piece can attack
	QueenMobility = [28]Score{
		S(-21, -77), S(-18, -66), S(-38, -56), S(-51, -74), S(-61, 18), S(-36, 70),
		S(-31, 97), S(-24, 107), S(-21, 128), S(-17, 153), S(-13, 159), S(-8, 166),
		S(-4, 174), S(0, 174), S(2, 178), S(4, 187), S(6, 188), S(6, 197),
		S(7, 201), S(10, 203), S(17, 206), S(29, 193), S(46, 189), S(69, 174),
		S(93, 166), S(85, 173), S(43, 159), S(23, 153),
	}
	RookMobility = [15]Score{
		S(-33, -16), S(-32, 2), S(-12, 29), S(-7, 53), S(-1, 65), S(0, 74),
		S(3, 80), S(5, 87), S(8, 87), S(14, 90), S(17, 94), S(20, 97),
		S(27, 99), S(34, 96), S(41, 93),
	}
	BishopMobility = [14]Score{
		S(-42, -129), S(-57, -49), S(-27, 0), S(-15, 22), S(-4, 33), S(2, 41),
		S(9, 50), S(15, 55), S(18, 60), S(24, 61), S(29, 62), S(47, 54),
		S(56, 56), S(68, 46),
	}
	KnightMobility = [9]Score{
		S(-123, -157), S(-29, -26), S(-9, 9), S(1, 31), S(14, 41), S(17, 53),
		S(27, 55), S(38, 57), S(51, 53),
	}

	// Material Adjustment
	BishopPairBonus         = S(25, 69)
	RookOnOpenFileBonus     = S(34, 11)
	RookOnSemiOpenFileBonus = S(11, 19)
	RookOnSeventhRankBonus  = S(21, 23)
	QueenOnSeventhRankBonus = S(18, 24)
	KnightOutpostBonus      = S(38, 20)
	ConnectedKnightBonus    = S(0, -5)
	BishopOutpostBonus      = S(41, 0)

	// Pawn Structure
	DoubledPawnPenalty        = S(3, -12)
	IsolatedPawnPenalty       = S(-3, -2)
	BackwardPawnPenalty       = S(-2, -3)
	DefendedPawnBonus         = S(9, 8)
	ConnectedPawnBonus        = S(7, 4)
	PassedPawnsBonus          = [8]Score{S(0, 0), S(4, 11), S(-2, 18), S(-5, 45), S(19, 75), S(20, 144), S(5, 95), S(0, 0)}
	CandidatePassedPawnsBonus = [2][8]Score{
		{S(0, 0), S(-26, 5), S(-12, 8), S(7, 38), S(29, 56), S(48, 86), S(0, 65), S(0, 0)},
		{S(0, 0), S(-18, 1), S(-4, 21), S(7, 40), S(45, 74), S(28, 133), S(0, 97), S(0, 0)},
	}

	// King Safety
	PawnShield = [2][8]Score{
		{S(32, -9), S(29, -17), S(18, -11), S(13, -21), S(11, -28), S(7, -28), S(-1, 0), S(0, 0)},
		{S(0, 0), S(36, -15), S(34, -9), S(16, -21), S(13, -28), S(-3, -10), S(-4, -3), S(0, 0)},
	}
	PawnStorm = [2][2][8]Score{
		{
			{S(21, 27), S(-11, -11), S(-21, -1), S(-20, -7), S(-1, -10), S(6, -15), S(7, -14), S(0, 0)},
			{S(9, 21), S(-1, -5), S(2, -17), S(4, -20), S(5, -3), S(10, -12), S(29, 7), S(0, 0)},
		},
		{
			{S(0, 0), S(43, 33), S(-6, 0), S(0, -4), S(10, -9), S(16, -13), S(15, -11), S(0, 0)},
			{S(0, 0), S(17, 25), S(-2, -30), S(16, -13), S(24, -13), S(23, -6), S(22, 0), S(0, 0)},
		},
	}
	KingOnOpenFiles = [2]Score{S(-7, -18), S(-38, -15)}

	KnightAttackWeight   = S(25, 8)
	BishopAttackWeight   = S(25, 2)
	RookAttackWeight     = S(30, -4)
	QueenAttackWeight    = S(17, 21)
	KingZoneDefenseBonus = S(9, 7)

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
	attackedByPawns     [2]Bitboard
	backwardsPawns      [2]Bitboard
	passedPawns         [2]Bitboard
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
	ed.attackedByPawns = [2]Bitboard{0, 0}
	ed.backwardsPawns = [2]Bitboard{0, 0}
	ed.passedPawns = [2]Bitboard{0, 0}
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
	ed.attackedByPawns[White] = pawnAttacks(&pos.Pieces[WhitePawn], White)
	ed.attackedByPawns[Black] = pawnAttacks(&pos.Pieces[BlackPawn], Black)
	ed.backwardsPawns[White] = BackwardPawns(ed.pawns[White], ed.attackedByPawns[Black], White)
	ed.backwardsPawns[Black] = BackwardPawns(ed.pawns[Black], ed.attackedByPawns[White], Black)
	ed.passedPawns[White] = PassedPawns(ed.pawns[White], ed.pawns[Black], White)
	ed.passedPawns[Black] = PassedPawns(ed.pawns[Black], ed.pawns[White], Black)
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
	sq := squareRelativeToSide(from, side)
	sc += PieceSquaresScores[King][sq]

	kingFile, kingRank := from%8, from/8

	// Squares in front of the king on the king file and the two adjacent files
	frontMask := Bitboard(0)
	kingSquare := bitboardFromIndex(from)
	if side == White {
		frontMask = fillUp(kingSquare)
		if kingFile > 0 {
			frontMask |= fillUp(bitboardFromIndex(from - 1))
		}
		if kingFile < 7 {
			frontMask |= fillUp(bitboardFromIndex(from + 1))
		}
	} else {
		frontMask = fillDown(kingSquare)
		if kingFile > 0 {
			frontMask |= fillDown(bitboardFromIndex(from - 1))
		}
		if kingFile < 7 {
			frontMask |= fillDown(bitboardFromIndex(from + 1))
		}
	}

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
		defendedSquares := (KingZone[side][from] & ev.EvalData.attackedByPawns[side]).Count()
		sc += -ev.EvalData.kingAttackersWeight[opponent] + S(defendedSquares, 0)*KingZoneDefenseBonus
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
		rank := from / 8
		sq := squareRelativeToSide(from, side)
		sc += PieceSquaresScores[Queen][sq]

		attacks := queenAttacks(&nextQueen, pos.Sides[All])
		safeSquares := (attacks & ^ev.EvalData.attackedByPawns[side.Opponent()]).Count()
		sc += QueenMobility[safeSquares]

		enemyKingZone := KingZone[opponent][Bsf(ev.EvalData.kings[opponent])]
		if attacks&enemyKingZone != 0 {
			ev.EvalData.kingAttackersCount[side]++
			ev.EvalData.kingAttackersWeight[side] += QueenAttackWeight
		}

		relativeKingRank := squareRelativeToSide(Bsf(ev.EvalData.kings[opponent]), side) / 8
		if relativeKingRank == 7 && rank == 6 {
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
		rank := from / 8
		sq := squareRelativeToSide(from, side)
		sc += PieceSquaresScores[Rook][sq]

		attacks := rookAttacks(from, pos.Sides[All])
		safeSquares := (attacks & ^ev.EvalData.attackedByPawns[side.Opponent()]).Count()
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
		relativeKingRank := squareRelativeToSide(Bsf(ev.EvalData.kings[opponent]), side) / 8
		if relativeKingRank == 7 && rank == 6 {
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
		sq := squareRelativeToSide(from, side)
		sc += PieceSquaresScores[Bishop][sq]

		attacks := bishopAttacks(from, pos.Sides[All])
		safeSquares := (attacks & ^ev.EvalData.attackedByPawns[side.Opponent()]).Count()
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
		sq := squareRelativeToSide(from, side)
		sc += PieceSquaresScores[Knight][sq]

		attacks := knightAttacksTable[from]
		safeSquares := (attacks & ^ev.EvalData.attackedByPawns[side.Opponent()]).Count()
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

	for side := Color(White); side <= Black; side++ {
		opponent := side.Opponent()
		alliedPawns := pos.Pieces[PieceOf(Pawn, side)]
		enemyPawns := pos.Pieces[PieceOf(Pawn, opponent)]
		pawns := alliedPawns

		for pawns > 0 {
			nextPawn := pawns.NextBit()
			from := Bsf(nextPawn)
			file := from % 8
			sq := squareRelativeToSide(from, side)
			sc += sideModifier[side] * PieceSquaresScores[Pawn][sq]

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
			backward := ev.EvalData.backwardsPawns[side]&nextPawn > 0
			if backward {
				sc += sideModifier[side] * BackwardPawnPenalty
			}

			// Passed. A pawn whose path to promotion is not blocked nor attacked by enemy pawns
			if ev.EvalData.passedPawns[side]&nextPawn > 0 {
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

// squareRelativeToSide returns the square number from the point of view of the side passed
func squareRelativeToSide(sq int, side Color) int {
	if side == White {
		return sq ^ 56
	}
	return sq
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
