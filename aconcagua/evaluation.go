package aconcagua

// TODO: New Evaluation Tests (Handcrafted/'Intuitive' adjustments) - Test each featuere against w/ 400 games 30s+1s TC (Blitz_Testing_4moves.epd)
// LOS means likehood of superiority - the engine is better than the opponent
// | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// | Feature                      | Implemented |       Elo vs PSQT only(LOS %) |   Elo vs Aconcagua-v3.0.0 (LOS %)   |   Observation																														|
// | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// | 0. New PSQT								  |			✔				|  		         0.0	     				| -213.60±36.24 (PeSTO PSQT) (0.00 %) | From SimpleEvaluation Function w/ minor adjustments											|
// | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// | 1. Mobility v1               |     ✔				|											----- Tests Aborted ----                        | Badresults, so aborted																									|
// | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// | 1. Mobility v2               |     ✔				|	  54.29±30.33 (LOS: 99.98 %)  | -134.95 +/- 34.92 (LOS: 0.00 %)     | I should try to add mobility for King endgame only											|
// | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// | 2. Pawn structure analysis individual tests are with 15s+1s TC(300 games)                                        | 																																				|
// | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// |  	2.1 Doubled Pawns  5/15  |      ✔      |     -3.47±27.61 (40.24 %)      |  ---------------------------------  | penalty mg -5 / eg -15																									|
// |  	2.1 Doubled Pawns  20/25 |      ✔      |     -24.36±37.21 (9.81 %)      |  ---------------------------------  |																																					|
// |  	2.1 Doubled Pawns   4/6  |      ✔      |     9.27±34.83  (69.97 %) *    |  ---------------------------------  | Can't conclude they are better, but is the best result  								|
// |  	2.2 Isolated Pawns  5/15 |      ✔      |    5.79±30.27 (64.65%)    *    |  ---------------------------------  |																																					|
// |  	2.3 Backward Pawns 5/25  |      ✔      |   -53.70±35.5  (0.13 %)        |  ---------------------------------  | Bad implementaion of backward pawns																			|
// |  	2.3 BackwardPawnsv2 5/25 |      ✔      |    -88.74±37.92 (0.00 %)       |  ---------------------------------  | Chess programing wiki routine. Check pawn by pawn.          						|
// |  	                         |             |                                |                                     | Bug with black pawns!!! 															      						|
// |  	2.3 Backward Pawnsv3 5/5 |      ✔      |    1.74±28.03 (54.84%)         |  ---------------------------------  | Calculate all backward pawns at once. Fix bug. Adjust penalties					|
// |  	2.3 Backward Pawnsv3 6/12|      ✔      |    -13.90±34.82 (21.59%)       |  ---------------------------------  | Use precalculated attacksFrontSpans                    	  							|
// |  	2.3 Backward Pawnsv3 3/8 |      ✔      |    32.52±31.38 (97.98 %) *     |  ---------------------------------  |                                                        	  							|
// |  	2.4 Passed Pawns 20/40   |      ✔      |    -19.71±31.90 (11,18 %)      |  ---------------------------------  | Fixed bonus for passed pawns																						|
// |  	2.4 Passed Pawns incr    |      ✔      |      31.35±33.59 (96.75%)      |  ---------------------------------  | Incremental bonus on ranks to go to promotion (2-7) 10, 20, 30, 40, 50  |
// |  	2.4 Passed Pawns incr2   |      ✔      |     37.20±33.95 (98.51%) *     |  ---------------------------------  | Huge bonus near promotion (2-7) 10, 20, 30, 60, 100                     |
// | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// |  	All Pawn Improvements    |      ✔      |     52.51±28.47  (99.99 %)     |                                     | Penalties/Bonus are best results in individual tests(400games 30s+1s TC)|
// | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
// | 3. King safety
// |  	3.1 Pawn shield
// |  	3.2 Pawn storm
// |    3.3 Open file
// |    3.4 King attackers
// | 4. Center control
// | 5. Open files
// | 6. Bishop pair bonus

// Evaluation is a vector containing the diferent evaluations of the position
type Evaluation struct {
	mgMaterial [2]int // PSQT + material weight [white, black]
	egMaterial [2]int
	mgMobility [2]int
	egMobility [2]int
	// TODO: use a pawn hash table to eval pawn structure/so it can be cached?
	// mgPawnStrucutre [2]int
	// egPawnStructure [2]int
	phase int
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

	// TODO: use a separate pawn structure evaluation
	// Total Backwards pawns for each side
	whiteBackwardPawns := backwardsPawns(pos, White)
	blackBackwardPawns := backwardsPawns(pos, Black)
	ev.mgMaterial[White] -= whiteBackwardPawns.count() * 3
	ev.egMaterial[Black] -= blackBackwardPawns.count() * 8

	// Passed Pawns
	whitePassedPawns := passedPawns(pos, White)
	blackPassedPawns := passedPawns(pos, Black)

	// Passed pawn bonus depending on the ranks to go to promotion
	passedPawnBonus := [8]int{0, 0, 10, 20, 30, 60, 100, 0}
	for whitePassedPawns > 0 {
		fromBB := whitePassedPawns.NextBit()
		sq := Bsf(fromBB)
		ev.mgMaterial[White] += passedPawnBonus[sq/8]
		ev.egMaterial[White] += passedPawnBonus[sq/8]
	}

	for blackPassedPawns > 0 {
		fromBB := blackPassedPawns.NextBit()
		sq := Bsf(fromBB)
		ev.mgMaterial[Black] += passedPawnBonus[7-sq/8]
		ev.egMaterial[Black] += passedPawnBonus[7-sq/8]
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

// evaluateQueen returns the middlegame and endgame score of the queen in the position
func (ev *Evaluation) evaluateQueen(pos *Position, queen *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Queen, side)
	sq := Bsf(*queen)

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	// attacksCount := Attacks(piece, *queen, ^pos.EmptySquares()).count()
	// ev.mgMobility[side] += (attacksCount - 7) * 5
	// ev.egMobility[side] += (attacksCount - 7) * 3

	ev.phase += 4

	return
}

// evaluateRook returns the middlegame and endgame score of the rook in the position
func (ev *Evaluation) evaluateRook(pos *Position, rook *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Rook, side)
	sq := Bsf(*rook)

	// TODO: open files bonus???
	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	// attacksCount := Attacks(piece, *rook, ^pos.EmptySquares()).count()
	// ev.mgMobility[side] += (attacksCount - 5) * 3
	// ev.egMobility[side] += (attacksCount - 5) * 3

	ev.phase += 2

	return
}

// evaluateBishop returns the middlegame and endgame score of the bishop in the position
func (ev *Evaluation) evaluateBishop(pos *Position, bishop *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Bishop, side)
	sq := Bsf(*bishop)

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	// attacksCount := Attacks(piece, *bishop, ^pos.EmptySquares()).count()
	// ev.mgMobility[side] += (attacksCount - 5) * 3
	// ev.egMobility[side] += (attacksCount - 5) * 4

	ev.phase += 1

	return
}

// evaluateKnight returns the middlegame and endgame score of the knight in the position
func (ev *Evaluation) evaluateKnight(pos *Position, knight *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Knight, side)
	sq := Bsf(*knight)

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	// TODO: extract to function?
	// attacksCount := Attacks(piece, *knight, ^pos.EmptySquares()).count()
	// ev.mgMobility[side] += (attacksCount - 3) * 3
	// ev.egMobility[side] += (attacksCount - 3) * 4

	ev.phase += 1

	return
}

// evaluatePawn returns the middlegame and endgame score of the pawn in the position
func (ev *Evaluation) evaluatePawn(pos *Position, pawn *Bitboard, side Color) (mgScore int, egScore int) {
	piece := pieceColor(Pawn, side)
	sq := Bsf(*pawn)

	// NOTE: applies the penalty for each pawn. So, if there are 3 pawns on the same file, the penalty is 3*penalty
	if isDoubled(pawn, pos, side) {
		ev.mgMaterial[side] -= 4
		ev.egMaterial[side] -= 6
	}
	//
	if isIsolated(pawn, pos, side) {
		ev.mgMaterial[side] -= 5
		ev.egMaterial[side] -= 15
	}
	//
	// if backwardsPawns(pos, side)&*pawn > 0 {
	// 	ev.mgMaterial[side] -= 5
	// 	ev.egMaterial[side] -= 5
	// }
	//
	// if isPassed(pawn, pos, side) {
	// 	ev.mgMaterial[side] += 20
	// 	ev.egMaterial[side] += 30
	// }

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	return
}

// isDoubled returns true if the pawn is doubled
func isDoubled(pawn *Bitboard, pos *Position, side Color) bool {
	file := Bsf(*pawn) % 8
	filePawns := files[file] & pos.Bitboards[pieceColor(Pawn, side)]

	if filePawns.count() > 1 {
		return true
	}
	return false
}

// isIsolated returns true if the pawn is isolated
func isIsolated(pawn *Bitboard, pos *Position, side Color) bool {
	file := Bsf(*pawn) % 8

	if IsolatedAdjacentFilesMask[file]&pos.Bitboards[pieceColor(Pawn, side)] == 0 {
		return true
	}
	return false
}

// IsolatedAdjacentFilesMask contains the adjacent files of a pawn to test if it is isolated
var IsolatedAdjacentFilesMask = [8]Bitboard{
	files[1],
	files[0] | files[2],
	files[1] | files[3],
	files[2] | files[4],
	files[3] | files[5],
	files[4] | files[6],
	files[5] | files[7],
	files[6],
}

// attacksFrontSpans is a precalculated table containing the front attack spans for each square
// front attack spans includes the attacked squares itself, thus it is like a fill of attacked squares in the appropriate direction
// front attack span for pawn on d4
// . . 1 . 1 . . .
// . . 1 . 1 . . .
// . . 1 . 1 . . .
// . . 1 . 1 . . .
// . . . w . . . .
// . . . . . . . .
// . . . . . . . .
// . . . . . . . .
var attacksFrontSpans [2][64]Bitboard

func generateAttacksFrontSpans() (attacksFrontSpans [2][64]Bitboard) {

	for sq := range 64 {
		file, rank := sq%8, sq/8
		eastFront, westFront := rank*8+file+1, rank*8+file-1

		if file < 7 {
			attacksFrontSpans[White][sq] |= rayAttacks[North][eastFront]
			attacksFrontSpans[Black][sq] |= rayAttacks[South][eastFront]
		}
		if file > 0 {
			attacksFrontSpans[White][sq] |= rayAttacks[North][westFront]
			attacksFrontSpans[Black][sq] |= rayAttacks[South][westFront]
		}

	}

	return
}

// backwardsPawns returns a bitboard with the pawns that are backwards
func backwardsPawns(pos *Position, side Color) Bitboard {
	// From wikipedia
	// In chess, a backward pawn is a pawn that is behind all pawns of the same color on the adjacent files and cannot be safely advanced

	// v2 from chess programing wiki definition
	// We may use a more general definition of backwardness, or better a pre-condition for backwardness, to consider certain subsets in further processing. All pawns, which stop is not member of own front-attackspans but controlled by a sentry are considered backward here, no matter if they are member of a ram or lever or more advanced mutually backward
	// 	U64 wBackward(U64 wpawns, U64 bpawns) {
	//    U64 stops = wpawns << 8;
	//    U64 wAttackSpans = wEastAttackFrontSpans(wpawns)
	//                     | wWestAttackFrontSpans(wpawns);
	//    U64 bAttacks     = bPawnEastAttacks(bpawns)
	//                     | bPawnWestAttacks(bpawns);
	//    return (stops & bAttacks & ~wAttackSpans) >> 8;
	// }

	// calculate front attacks spans
	// 	white attack
	// frontspan example from chessprogramingwiki
	// . . 1 . 1 . . .
	// . . 1 . 1 . . .
	// . . 1 . 1 . . .
	// . . 1 . 1 . . .
	// . . . w . . . .
	// . . . . . . . .
	// . . . . . . . .
	// . . . . . . . .

	pawns := pos.Bitboards[pieceColor(Pawn, side)]
	stops := pawns << 8
	if side == Black {
		stops = pawns >> 8
	}

	// TODO: create a 'table' with cached front spans
	// attackFrontSpans := Bitboard(0)
	// for pawns > 0 {
	// 	currPawn := pawns.NextBit()
	// 	frontDirection := North
	// 	if side == Black {
	// 		frontDirection = South
	// 	}
	//
	// 	file, rank := Bsf(currPawn)%8, Bsf(currPawn)/8
	// 	eastFront, westFront := rank*8+file+1, rank*8+file-1
	// 	if file < 7 {
	// 		attackFrontSpans |= rayAttacks[frontDirection][eastFront]
	// 	}
	// 	if file > 0 {
	// 		attackFrontSpans |= rayAttacks[frontDirection][westFront]
	// 	}
	// }
	attackFrontSpans := Bitboard(0)
	for pawns > 0 {
		pawn := pawns.NextBit()
		attackFrontSpans |= attacksFrontSpans[side][Bsf(pawn)]
	}

	enemyPawns := pos.Bitboards[pieceColor(Pawn, side.Opponent())]
	enemyPawnsAttacks := pawnAttacks(&enemyPawns, side.Opponent())

	// (^attackFrontSpans).Print()
	// enemyPawnsAttacks.Print()
	// (stops & enemyPawnsAttacks & ^attackFrontSpans).Print()

	if side == White {
		return (stops & enemyPawnsAttacks & ^attackFrontSpans) >> 8
	} else {
		return (stops & enemyPawnsAttacks & ^attackFrontSpans) << 8
	}

	// This impl. is wrong....
	// file := Bsf(*pawn) % 8
	// rank := Bsf(*pawn) / 8
	// backwardDirection := South
	// backwardLeftSq, backwardRightSq := rank*8+file-1, rank*8+file+1
	// upSq := *pawn << 8
	// if side == Black {
	// 	backwardLeftSq, backwardRightSq = rank*8+file-1, rank*8+file+1
	// 	upSq = *pawn >> 8
	// 	backwardDirection = North
	// }
	//
	// alliedPawns := pos.Bitboards[pieceColor(Pawn, side)]
	// adjacentBackward := Bitboard(0)
	// if file < 7 {
	// 	adjacentBackward |= rayAttacks[backwardDirection][backwardRightSq]
	// }
	// if file > 0 {
	// 	adjacentBackward |= rayAttacks[backwardDirection][backwardLeftSq]
	// }
	//
	// enemyPawns := pos.Bitboards[pieceColor(Pawn, side.Opponent())]
	// enemyPawnsAttacks := Attacks(pieceColor(Pawn, side.Opponent()), enemyPawns, ^pos.EmptySquares())
	//
	// if adjacentBackward&alliedPawns == 0 && enemyPawnsAttacks&upSq > 0 {
	// 	return true
	// }

	// return false
}

// passedPawns returns a bitboard with the passed pawns for the side
func passedPawns(pos *Position, side Color) (passedPawns Bitboard) {
	alliedPawns := pos.Bitboards[pieceColor(Pawn, side)]
	enemyPawns := pos.Bitboards[pieceColor(Pawn, side.Opponent())]
	direction := North
	if side == Black {
		direction = South
	}

	for alliedPawns > 0 {
		pawn := alliedPawns.NextBit()
		frontAndAdjacentSquares := attacksFrontSpans[side][Bsf(pawn)] | rayAttacks[direction][Bsf(pawn)]

		if frontAndAdjacentSquares&enemyPawns == 0 {
			passedPawns |= pawn
		}
	}

	return
}
