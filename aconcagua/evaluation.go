package aconcagua

// TODO: Evaluation improvements...                Elo improvement(from base PSQT eval - not tunned). selfplay 50 games 1+0
// 0. New PSQT														✔																	0.0
// 1. Mobility                            ✔															+ 63.23
// 2. Pawn structure analysis
//		2.1 Doubled Pawns
//		2.2 Isolated Pawn
//		2.3 Backward Pawns
//		2.4 Passed Pawns
// 3. King safety
//		3.1 Pawn shield
//		3.2 Pawn storm
//    3.3 Open file
//    3.4 King attackers
// 4. Center control
// 5. Open files
// 6. Bishop pair bonus

// Evaluation is a vector containing the diferent evaluations of the position
type Evaluation struct {
	mgMaterial [2]int // PSQT + material weight [white, black]
	egMaterial [2]int
	mgMobility [2]int
	egMobility [2]int
	phase      int
}

// Evaluate returns the evaluation of the position
func Evaluate(pos *Position) int {
	ev := Evaluation{}

	for p, bb := range pos.Bitboards {
		color := p / 6

		for bb > 0 {
			fromBB := bb.NextBit()
			switch pieceRole(p) {
			case King:
				ev.evaluateKing(pos, &fromBB, Color(color))
			case Queen:
				ev.evaluateQueen(pos, &fromBB, Color(color))
			case Rook:
				ev.evaluateRook(pos, &fromBB, Color(color))
			case Bishop:
				ev.evaluateBishop(pos, &fromBB, Color(color))
			case Knight:
				ev.evaluateKnight(pos, &fromBB, Color(color))
			case Pawn:
				ev.evaluatePawn(pos, &fromBB, Color(color))
			}
		}
	}

	return ev.score(pos.Turn)
}

// score returns the score relative to the side passed
func (ev *Evaluation) score(side Color) int {
	opponent := side.Opponent()
	mgPhase := min(ev.phase, 24)
	egPhase := 24 - mgPhase

	mg := ev.mgMaterial[side] - ev.mgMaterial[opponent] + ev.mgMobility[side] - ev.mgMobility[opponent]
	eg := ev.egMaterial[side] - ev.egMaterial[opponent] + ev.egMobility[side] - ev.egMobility[opponent]

	return (mg*mgPhase + eg*egPhase) / 24
}

// evaluateKing returns the middlegame and endgame score of the king in the position
func (ev *Evaluation) evaluateKing(pos *Position, king *Bitboard, side Color) (mgScore int, egScore int) {
	sq := Bsf(*king)

	ev.mgMaterial[side] += middlegamePiecesScore[pieceColor(King, side)][sq]
	ev.egMaterial[side] += endgamePiecesScore[pieceColor(King, side)][sq]
	return
}

// TODO: extract to EvalTables or somehting else
var queenMobilityMg = [28]int{
	-60, -40, -20, -10, 0, 10, 20, 30, 35, 40, 45, 50, 55, 60, 65,
	70, 75, 80, 85, 90, 95, 100, 105, 110, 115, 120, 125, 130,
}
var queenMobilityEg = [28]int{
	-60, -40, -20, -10, 0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50,
	55, 60, 65, 70, 75, 80, 85, 90, 95, 100, 105, 110, 115,
}

// evaluateQueen returns the middlegame and endgame score of the queen in the position
func (ev *Evaluation) evaluateQueen(pos *Position, queen *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Queen, side)
	sq := Bsf(*queen)

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	attacksCount := Attacks(piece, *queen, ^pos.EmptySquares()).count()
	ev.mgMobility[side] += queenMobilityMg[attacksCount]
	ev.egMobility[side] += queenMobilityEg[attacksCount]

	ev.phase += 4

	return
}

var rookMobilityMg = [15]int{-60, -30, -10, 0, 10, 20, 30, 40, 45, 50, 55, 60, 65, 70, 75}
var rookMobilityEg = [15]int{-60, -30, -10, 5, 15, 25, 35, 45, 50, 55, 60, 65, 70, 75, 80}

// evaluateRook returns the middlegame and endgame score of the rook in the position
func (ev *Evaluation) evaluateRook(pos *Position, rook *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Rook, side)
	sq := Bsf(*rook)

	// TODO: open files bonus???
	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	attacksCount := Attacks(piece, *rook, ^pos.EmptySquares()).count()
	ev.mgMobility[side] += rookMobilityMg[attacksCount]
	ev.egMobility[side] += rookMobilityEg[attacksCount]

	ev.phase += 2

	return
}

var bishopMobilityMg = [14]int{-50, -25, -10, 0, 10, 20, 30, 40, 45, 50, 55, 60, 65, 70}
var bishopMobilityEg = [14]int{-50, -25, -10, 0, 5, 15, 25, 35, 40, 45, 50, 55, 60, 65}

// evaluateBishop returns the middlegame and endgame score of the bishop in the position
func (ev *Evaluation) evaluateBishop(pos *Position, bishop *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Bishop, side)
	sq := Bsf(*bishop)

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	attacksCount := Attacks(piece, *bishop, ^pos.EmptySquares()).count()
	ev.mgMobility[side] += bishopMobilityMg[attacksCount]
	ev.egMobility[side] += bishopMobilityEg[attacksCount]

	ev.phase += 1

	return
}

var knightMobilityMg = [9]int{-50, -20, 0, 10, 20, 30, 40, 45, 50}
var knightMobilityEg = [9]int{-50, -20, 0, 10, 20, 30, 35, 40, 45}

// evaluateKnight returns the middlegame and endgame score of the knight in the position
func (ev *Evaluation) evaluateKnight(pos *Position, knight *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Knight, side)
	sq := Bsf(*knight)

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	// TODO: extract to function?
	attacksCount := Attacks(piece, *knight, ^pos.EmptySquares()).count()
	ev.mgMobility[side] += knightMobilityMg[attacksCount]
	ev.egMobility[side] += knightMobilityEg[attacksCount]

	ev.phase += 1

	return
}

func (ev *Evaluation) evaluatePawn(pos *Position, pawn *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Pawn, side)
	sq := Bsf(*pawn)

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	return
}
