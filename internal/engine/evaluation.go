package engine

const (
	MaxPhaseValue = 24

	MgScoreMask = 0xffff
)

// Score encodes the values for middlegame and endgame score of an evaluation feature
type Score int32

// NewScore creates an Score from mg and eg values
func NewScore(mg int, eg int) Score {
	return Score(int32(mg) + int32(uint16(eg))<<16)
}

// Get returns the mg and eg values of the score
func (sc Score) Get() (mg int, eg int) {
	mg = int(int16(sc & MgScoreMask))
	eg = int(int16((sc + 0x8000) >> 16))
	return
}

// Evaluation contains the elements for evaluation of a position
type Evaluation struct {
	Eval      EvalVector
	EvalData  EvalData
	PawnCache PawnHashTable
}

// EvalVector contains the different evaluation elements of a position
type EvalVector struct {
	mgMaterial [2]int // White and Black scores
	egMaterial [2]int
}

// EvalData contains positional data about the current position
type EvalData struct {
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
}

// clear clears the EvalData
func (ed *EvalData) clear() {
}

// init initializes the evaluation data
func (ed *EvalData) init(pos *Position) {
}

// GetEvalPhase returns the mg phase value of the position depending on the remaining material on it
func GetEvalPhase(pos *Position) int {
	return (pos.Pieces[WhiteQueen]|pos.Pieces[BlackQueen]).count()*4 +
		(pos.Pieces[WhiteRook]|pos.Pieces[BlackRook]).count()*2 +
		(pos.Pieces[WhiteBishop] | pos.Pieces[BlackBishop]).count() +
		(pos.Pieces[WhiteKnight] | pos.Pieces[BlackKnight]).count()
}

// Evaluate returns the static score of the position
func (ev *Evaluation) Evaluate(pos *Position) (score int) {
	var sc Score

	for piece, pieces := range pos.Pieces {
		role := pieceRole(piece)
		for pieces > 0 {
			side := Color(piece / 6)
			nextPiece := pieces.NextBit()
			from := Bsf(nextPiece)
			if side == White {
				from = from ^ 56
			}

			if side == White {
				sc += PiecesScore[role] + Psqt[role][from]
			} else {
				sc -= (PiecesScore[role] + Psqt[role][from])
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
