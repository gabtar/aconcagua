package engine

const (
	MaxPhaseValue = 24
	MgScoreMask   = 0xffff
)

var (
	// Mobility arrays based on the number of squares a piece can attack
	QueenMobility = [28]Score{
		S(-21, -77), S(-18, -66), S(-38, -56), S(-50, -74), S(-47, 25), S(-29, 87),
		S(-26, 119), S(-21, 132), S(-19, 150), S(-16, 171), S(-12, 176), S(-8, 182),
		S(-5, 190), S(0, 190), S(1, 192), S(4, 198), S(5, 199), S(5, 207),
		S(7, 207), S(11, 204), S(23, 200), S(40, 181), S(60, 175), S(60, 148),
		S(94, 147), S(34, 122), S(17, 125), S(8, 130),
	}
	RookMobility = [15]Score{
		S(-39, -17), S(-29, 8), S(-3, 27), S(-4, 46), S(-4, 57), S(-2, 65),
		S(0, 73), S(3, 79), S(8, 81), S(17, 85), S(26, 88), S(30, 95),
		S(34, 100), S(39, 98), S(45, 95),
	}
	BishopMobility = [14]Score{
		S(-30, -136), S(-50, -48), S(-22, 4), S(-11, 30), S(0, 35), S(8, 40),
		S(15, 49), S(21, 52), S(23, 55), S(28, 54), S(31, 53), S(43, 46),
		S(45, 46), S(59, 33),
	}
	KnightMobility = [9]Score{
		S(-137, -177), S(-22, -19), S(-1, 15), S(9, 38), S(13, 43), S(12, 49),
		S(24, 52), S(34, 52), S(47, 45),
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
}

// init initializes the evaluation data
func (ed *EvalData) init(pos *Position) {
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
	var sc Score

	attackedByPawns := [2]Bitboard{
		pawnAttacks(&pos.Pieces[WhitePawn], White),
		pawnAttacks(&pos.Pieces[BlackPawn], Black),
	}

	for piece, pieces := range pos.Pieces {
		role := RoleOf(piece)
		for pieces > 0 {
			side := SideOf(piece)
			nextPiece := pieces.NextBit()
			from := Bsf(nextPiece)
			if side == White {
				from = from ^ 56
			}

			// Mobility. TODO: refactor this later. Just for testing new eval
			if role >= Queen && role <= Knight {
				attacks := Attacks(piece, nextPiece, pos.Sides[All])
				safeSquares := (attacks & ^attackedByPawns[side.Opponent()]).Count()
				switch role {
				case Queen:
					if side == White {
						sc += QueenMobility[safeSquares]
					} else {
						sc -= QueenMobility[safeSquares]
					}
				case Rook:
					if side == White {
						sc += RookMobility[safeSquares]
					} else {
						sc -= RookMobility[safeSquares]
					}
				case Bishop:
					if side == White {
						sc += BishopMobility[safeSquares]
					} else {
						sc -= BishopMobility[safeSquares]
					}
				case Knight:
					if side == White {
						sc += KnightMobility[safeSquares]
					} else {
						sc -= KnightMobility[safeSquares]
					}
				}
			}

			if side == White {
				sc += PieceSquaresScores[role][from]
			} else {
				sc -= PieceSquaresScores[role][from]
			}
		}
	}

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
