package aconcagua

const (
	QueenMobilityBonusMg  = 3
	QueenMobilityBonusEg  = 1
	QueenMobilityBase     = 7
	RookMobilityBonusMg   = 2
	RookMobilityBonusEg   = 2
	RookMobilityBase      = 4
	BishopMobilityBonusMg = 2
	BishopMobilityBonusEg = 3
	BishopMobilityBase    = 4
	KnightMobilityBonusMg = 1
	KnightMobilityBonusEg = 2
	KnightMobilityBase    = 2
)

// Evaluation contains the diferent evaluation elements of a position
type Evaluation struct {
	mgMaterial [2]int // White and Black scores
	egMaterial [2]int
	mgMobility [2]int
	egMobility [2]int
	phase      int
}

// NewEvaluation returns a new Evaluation
func NewEvaluation() *Evaluation {
	return &Evaluation{}
}

// clear clears the evaluation
func (ev *Evaluation) clear() {
	ev.mgMaterial = [2]int{0, 0}
	ev.egMaterial = [2]int{0, 0}
	ev.mgMobility = [2]int{0, 0}
	ev.egMobility = [2]int{0, 0}
	ev.phase = 0
}

// score returns the score relative to the side
func (ev *Evaluation) score(side Color) int {
	opponent := side.Opponent()

	mg := ev.mgMaterial[side] - ev.mgMaterial[opponent]
	eg := ev.egMaterial[side] - ev.egMaterial[opponent]
	mg += ev.mgMobility[side] - ev.mgMobility[opponent]
	eg += ev.egMobility[side] - ev.egMobility[opponent]

	mgPhase := min(ev.phase, 62)
	egPhase := 62 - mgPhase

	return (mg*mgPhase + eg*egPhase) / 62
}

// evaluateKing evaluates the score of a king
func (ev *Evaluation) evaluateKing(from int, side Color) {
	piece := pieceColor(King, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]
}

// evaluateQueen evaluates the score of a queen
func (ev *Evaluation) evaluateQueen(from int, blocks Bitboard, enemyPawns Bitboard, side Color) {
	piece := pieceColor(Queen, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	enemyPawnsAttacks := pawnAttacks(&enemyPawns, side.Opponent())
	squares := (attacks & ^enemyPawnsAttacks).count()

	ev.mgMobility[side] += (squares - QueenMobilityBase) * QueenMobilityBonusMg
	ev.egMobility[side] += (squares - QueenMobilityBase) * QueenMobilityBonusEg

	ev.phase += 9
}

// evaluateRook evaluates the score of a rook
func (ev *Evaluation) evaluateRook(from int, blocks Bitboard, enemyPawns Bitboard, side Color) {
	piece := pieceColor(Rook, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	enemyPawnsAttacks := pawnAttacks(&enemyPawns, side.Opponent())
	squares := (attacks & ^enemyPawnsAttacks).count()

	ev.mgMobility[side] += (squares - RookMobilityBase) * RookMobilityBonusMg
	ev.egMobility[side] += (squares - RookMobilityBase) * RookMobilityBonusEg

	ev.phase += 5
}

// evaluateBishop evaluates the score of a bishop
func (ev *Evaluation) evaluateBishop(from int, blocks Bitboard, enemyPawns Bitboard, side Color) {
	piece := pieceColor(Bishop, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	enemyPawnsAttacks := pawnAttacks(&enemyPawns, side.Opponent())
	squares := (attacks & ^enemyPawnsAttacks).count()

	ev.mgMobility[side] += (squares - BishopMobilityBase) * BishopMobilityBonusMg
	ev.egMobility[side] += (squares - BishopMobilityBase) * BishopMobilityBonusEg

	ev.phase += 3
}

// evaluateKnight evaluates the score of a knight
func (ev *Evaluation) evaluateKnight(from int, blocks Bitboard, enemyPawns Bitboard, side Color) {
	piece := pieceColor(Knight, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]

	fromBB := bitboardFromIndex(from)
	attacks := Attacks(piece, fromBB, blocks)
	enemyPawnsAttacks := pawnAttacks(&enemyPawns, side.Opponent())
	squares := (attacks & ^enemyPawnsAttacks).count()

	ev.mgMobility[side] += (squares - KnightMobilityBase) * KnightMobilityBonusMg
	ev.egMobility[side] += (squares - KnightMobilityBase) * KnightMobilityBonusEg

	ev.phase += 3
}

// evaluatePawn evaluates the score of a pawn
func (ev *Evaluation) evaluatePawn(from int, side Color) {
	piece := pieceColor(Pawn, side)
	ev.mgMaterial[side] += middlegamePiecesScore[piece][from]
	ev.egMaterial[side] += endgamePiecesScore[piece][from]
}

func (pos *Position) Evaluate() int {
	pos.evaluation.clear()
	blocks := ^pos.EmptySquares()

	for piece, bb := range pos.Bitboards {
		color := Color(piece / 6)

		for bb > 0 {
			bb := bb.NextBit()
			sq := Bsf(bb)

			switch pieceRole(piece) {
			case King:
				pos.evaluation.evaluateKing(sq, color)
			case Queen:
				pos.evaluation.evaluateQueen(sq, blocks, pos.Bitboards[pieceColor(Pawn, color.Opponent())], color)
			case Rook:
				pos.evaluation.evaluateRook(sq, blocks, pos.Bitboards[pieceColor(Pawn, color.Opponent())], color)
			case Bishop:
				pos.evaluation.evaluateBishop(sq, blocks, pos.Bitboards[pieceColor(Pawn, color.Opponent())], color)
			case Knight:
				pos.evaluation.evaluateKnight(sq, blocks, pos.Bitboards[pieceColor(Pawn, color.Opponent())], color)
			case Pawn:
				pos.evaluation.evaluatePawn(sq, color)
			}
		}
	}

	return pos.evaluation.score(pos.Turn)
}
