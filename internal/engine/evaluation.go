package engine

const (
	MaxPhaseValue = 24
	MgScoreMask   = 0xffff
)

var (
	// Mobility arrays based on the number of squares a piece can attack
	QueenMobility = [28]Score{
		S(-21, -77), S(-18, -66), S(-38, -56), S(-50, -74), S(-61, 19), S(-37, 75),
		S(-32, 102), S(-25, 111), S(-21, 130), S(-17, 156), S(-13, 162), S(-8, 169),
		S(-4, 178), S(1, 177), S(4, 181), S(6, 190), S(8, 192), S(8, 201),
		S(10, 206), S(12, 208), S(20, 211), S(32, 200), S(52, 193), S(76, 178),
		S(101, 169), S(76, 166), S(37, 151), S(19, 147),
	}
	RookMobility = [15]Score{
		S(-35, -16), S(-31, 4), S(-15, 33), S(-9, 55), S(-3, 67), S(0, 76),
		S(2, 82), S(5, 88), S(9, 89), S(16, 91), S(20, 95), S(24, 98),
		S(31, 100), S(38, 98), S(44, 95),
	}
	BishopMobility = [14]Score{
		S(-38, -129), S(-63, -46), S(-31, 0), S(-17, 23), S(-5, 34), S(3, 42),
		S(10, 51), S(18, 55), S(21, 61), S(28, 62), S(34, 63), S(53, 55),
		S(61, 59), S(71, 49),
	}
	KnightMobility = [9]Score{
		S(-126, -162), S(-29, -22), S(-8, 11), S(2, 33), S(15, 42), S(18, 54),
		S(28, 56), S(39, 59), S(53, 55),
	}

	// Material Adjustment
	BishopPairBonus         = S(25, 70)
	RookOnOpenFileBonus     = S(36, 10)
	RookOnSemiOpenFileBonus = S(15, 17)
	RookOnSeventhRankBonus  = S(20, 27)
	QueenOnSeventhRankBonus = S(18, 26)
	KnightOutpostBonus      = S(37, 21)
	ConnectedKnightBonus    = S(0, -4)
	BishopOutpostBonus      = S(43, 0)

	// Pawn Structure
	DoubledPawnPenalty        = S(2, -11)
	IsolatedPawnPenalty       = S(-3, -2)
	BackwardPawnPenalty       = S(0, -4)
	DefendedPawnBonus         = S(9, 8)
	ConnectedPawnBonus        = S(8, 4)
	PassedPawnsBonus          = [8]Score{S(0, 0), S(4, 11), S(-2, 18), S(-5, 44), S(19, 74), S(16, 130), S(-1, 93), S(0, 0)}
	CandidatePassedPawnsBonus = [2][8]Score{
		{S(0, 0), S(-17, 4), S(-11, 4), S(7, 37), S(29, 56), S(31, 82), S(0, 65), S(0, 0)},
		{S(0, 0), S(-15, 0), S(-4, 22), S(6, 41), S(41, 78), S(15, 122), S(0, 97), S(0, 0)},
	}

	// King Safety
	PawnShield = [2][8]Score{
		{S(22, -1), S(21, -8), S(10, -4), S(7, -13), S(3, -17), S(4, -14), S(-2, 3), S(0, 0)},
		{S(0, 0), S(24, 0), S(19, 0), S(10, -12), S(4, -13), S(-5, 0), S(-3, -1), S(0, 0)},
	}
	PawnStorm = [2][2][8]Score{
		{
			{S(20, 22), S(-4, -4), S(-21, 4), S(-15, -5), S(1, -8), S(8, -13), S(9, -12), S(0, 0)},
			{S(4, 10), S(0, -2), S(0, -13), S(10, -19), S(9, -4), S(11, -13), S(15, 0), S(0, 0)},
		},
		{
			{S(0, 0), S(23, 23), S(-15, 9), S(4, 0), S(13, -4), S(16, -8), S(16, -5), S(0, 0)},
			{S(0, 0), S(8, 12), S(-5, -22), S(20, -8), S(21, -6), S(13, 1), S(7, -1), S(0, 0)},
		},
	}
	KingOnOpenFiles = [2]Score{S(-16, -9), S(-17, -10)}

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
	kings           [2]Bitboard
	pawns           [2]Bitboard
	attackedByPawns [2]Bitboard
	backwardsPawns  [2]Bitboard
	passedPawns     [2]Bitboard
	outposts        [2]Bitboard
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
	sc += ev.evaluateKings(pos, White) - ev.evaluateKings(pos, Black)
	sc += ev.evaluateQueens(pos, White) - ev.evaluateQueens(pos, Black)
	sc += ev.evaluateRooks(pos, White) - ev.evaluateRooks(pos, Black)
	sc += ev.evaluateBishops(pos, White) - ev.evaluateBishops(pos, Black)
	sc += ev.evaluateKnights(pos, White) - ev.evaluateKnights(pos, Black)

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
		stormers := pos.Pieces[PieceOf(Pawn, side.Opponent())] & Files[file] & frontMask

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
			stormRank := NearestFromSide(stormers, side.Opponent()) / 8
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

		if nextBishop&ev.EvalData.outposts[side] > 0 {
			sc += BishopOutpostBonus
		}
	}
	return sc
}

// evaluateKnights returns the score of the Knights in the position for the side passed
func (ev *Evaluation) evaluateKnights(pos *Position, side Color) (sc Score) {
	knights := pos.Pieces[PieceOf(Knight, side)]
	for knights > 0 {
		nextKnight := knights.NextBit()
		from := Bsf(nextKnight)
		sq := squareRelativeToSide(from, side)
		sc += PieceSquaresScores[Knight][sq]

		attacks := knightAttacksTable[from]
		safeSquares := (attacks & ^ev.EvalData.attackedByPawns[side.Opponent()]).Count()
		sc += KnightMobility[safeSquares]

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
