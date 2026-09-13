package engine

const (
	MaxPhaseValue = 24
	MgScoreMask   = 0xffff
)

var (
	// Mobility arrays based on the number of squares a piece can attack
	QueenMobility = [28]Score{
		S(-21, -77), S(-18, -66), S(-38, -56), S(-50, -74), S(-52, 23), S(-37, 83),
		S(-34, 113), S(-27, 124), S(-23, 140), S(-19, 162), S(-14, 167), S(-9, 173),
		S(-4, 181), S(1, 180), S(3, 184), S(6, 192), S(7, 193), S(7, 202),
		S(8, 205), S(11, 205), S(21, 204), S(37, 188), S(58, 180), S(71, 160),
		S(98, 154), S(46, 134), S(22, 132), S(10, 134),
	}
	RookMobility = [15]Score{
		S(-39, -17), S(-30, 4), S(-14, 28), S(-8, 46), S(-2, 58), S(1, 66),
		S(4, 72), S(7, 79), S(11, 80), S(17, 83), S(22, 87), S(26, 89),
		S(31, 91), S(37, 89), S(37, 88),
	}
	BishopMobility = [14]Score{
		S(-30, -134), S(-62, -55), S(-34, -7), S(-20, 17), S(-6, 27), S(3, 35),
		S(10, 45), S(17, 49), S(21, 55), S(26, 55), S(31, 56), S(46, 50),
		S(51, 51), S(62, 39),
	}
	KnightMobility = [9]Score{
		S(-133, -173), S(-24, -18), S(-6, 10), S(4, 32), S(15, 39), S(17, 47),
		S(29, 50), S(41, 51), S(53, 47),
	}

	// Material Adjustment
	BishopPairBonus         = S(22, 70)
	RookOnOpenFileBonus     = S(36, 7)
	RookOnSemiOpenFileBonus = S(14, 8)
	RookOnSeventhRankBonus  = S(19, 34)
	QueenOnSeventhRankBonus = S(15, 22)
	KnightOutpostBonus      = S(40, 22)
	ConnectedKnightBonus    = S(-1, 3)
	BishopOutpostBonus      = S(40, 0)

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
	outposts        [2]Bitboard
	attackedByPawns [2]Bitboard
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

	sc += ev.evaluateKings(pos, White) - ev.evaluateKings(pos, Black)
	sc += ev.evaluateQueens(pos, White) - ev.evaluateQueens(pos, Black)
	sc += ev.evaluateRooks(pos, White) - ev.evaluateRooks(pos, Black)
	sc += ev.evaluateBishops(pos, White) - ev.evaluateBishops(pos, Black)
	sc += ev.evaluateKnights(pos, White) - ev.evaluateKnights(pos, Black)
	sc += ev.evaluatePawns(pos, White) - ev.evaluatePawns(pos, Black)

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
	sq := squareRelativeToSide(Bsf(king), side)
	sc += PieceSquaresScores[King][sq]
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
func (ev *Evaluation) evaluatePawns(pos *Position, side Color) (sc Score) {
	pawns := pos.Pieces[PieceOf(Pawn, side)]
	for pawns > 0 {
		nextPawn := pawns.NextBit()
		sq := squareRelativeToSide(Bsf(nextPawn), side)
		sc += PieceSquaresScores[Pawn][sq]
	}
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
