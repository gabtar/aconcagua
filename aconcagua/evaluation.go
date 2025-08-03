package aconcagua

const (
	DoubledPawnPenaltyMg  = -8
	DoubledPawnPenaltyEg  = -12
	IsolatedPawnPenaltyMg = -5
	IsolatedPawnPenaltyEg = -15
	BackwardPawnPenaltyMg = -3
	BackwardPawnPenaltyEg = -8
)

// Evaluation is a vector containing the diferent evaluations of the position
type Evaluation struct {
	mgMaterial      [2]int // PSQT + material weight [white, black]
	egMaterial      [2]int
	mgMobility      [2]int
	egMobility      [2]int
	mgPawnStrucutre [2]int
	egPawnStructure [2]int
	phase           int
}

// Evaluate returns the evaluation of the position
func Evaluate(pos *Position) int {
	ev := Evaluation{}

	for p, bb := range pos.Bitboards {
		color := Color(p / 6)

		for bb > 0 {
			fromBB := bb.NextBit()
			switch pieceRole(p) {
			case King:
				ev.evaluateKing(pos, &fromBB, color)
			case Queen:
				ev.evaluateQueen(pos, &fromBB, color)
			case Rook:
				ev.evaluateRook(pos, &fromBB, color)
			case Bishop:
				ev.evaluateBishop(pos, &fromBB, color)
			case Knight:
				ev.evaluateKnight(pos, &fromBB, color)
			case Pawn:
				ev.evaluatePawn(pos, &fromBB, color)
			}
		}
	}

	// Pawn structure evaluation
	// TODO: use a pawn hash table
	// ev.evaluatePawnStructure(pos, White)
	// ev.evaluatePawnStructure(pos, Black)

	return ev.score(pos.Turn)
}

// score returns the score relative to the side passed
func (ev *Evaluation) score(side Color) int {
	opponent := side.Opponent()
	mgPhase := min(ev.phase, 24)
	egPhase := 24 - mgPhase

	mg := ev.mgMaterial[side] - ev.mgMaterial[opponent] + ev.mgMobility[side] - ev.mgMobility[opponent] + ev.mgPawnStrucutre[side] - ev.mgPawnStrucutre[opponent]
	eg := ev.egMaterial[side] - ev.egMaterial[opponent] + ev.egMobility[side] - ev.egMobility[opponent] + ev.egPawnStructure[side] - ev.egPawnStructure[opponent]

	return (mg*mgPhase + eg*egPhase) / 24
}

// evaluateKing returns the middlegame and endgame score of the king in the position
func (ev *Evaluation) evaluateKing(pos *Position, king *Bitboard, side Color) {
	sq := Bsf(*king)

	ev.mgMaterial[side] += middlegamePiecesScore[pieceColor(King, side)][sq]
	ev.egMaterial[side] += endgamePiecesScore[pieceColor(King, side)][sq]

	// Pawn Shield only in middlegame ???
	// In the endgame i would prefer king mobility
	// ev.mgMaterial[side] += pawnShieldScore(king, pos, side)

	// Pawn Storm
	// TODO: not sure how make it work yet
	// pawnStorm := pawnStormScore(king, pos, side)
	// ev.mgMaterial[side] += pawnStorm
	// // ev.egMaterial[side] += pawnStorm

	kingZoneAttackers := kingZoneAttackersPenalty(king, pos, side)
	ev.mgMaterial[side] += kingZoneAttackers
	ev.egMaterial[side] += kingZoneAttackers

	return
}

func kingZoneAttackersPenalty(king *Bitboard, pos *Position, side Color) (score int) {
	attackersWeight := [6]int{0, -10, -5, -3, -3, -1}
	// for each square the enemy piece attacks the king zone(king ring), multiply the weight of the piece
	enemies := pos.getBitboards(side.Opponent())
	blocks := ^pos.EmptySquares()
	kingZone := Attacks(pieceColor(King, side), *king, blocks) | *king

	for p, bb := range enemies {
		for bb > 0 {
			fromBB := bb.NextBit()
			attacks := Attacks(pieceColor(p, side.Opponent()), fromBB, blocks)

			if attacks&kingZone > 0 {
				score += attackersWeight[p] * (attacks & kingZone).count()
			}
		}
	}
	return
}

// kingCastleZones contains the squares where the king is located when he has castled during a game
var kingCastleZones [2][2]Bitboard = [2][2]Bitboard{
	{bitboardFromCoordinates("f1", "g1", "h1"), bitboardFromCoordinates("a1", "b1", "c1")}, // white king { shortcastle, longcastle }
	{bitboardFromCoordinates("f8", "g8", "h8"), bitboardFromCoordinates("a8", "b8", "c8")}, // black king { shortcastle, longcastle }
}

// pawnShieldScore returns a bonus/penalty when the king has castled depending on the pawns protection
func pawnShieldScore(king *Bitboard, pos *Position, side Color) (score int) {
	if *king&kingCastleZones[side][0] == 0 && *king&kingCastleZones[side][1] == 0 {
		return
	}
	alliedPawns := pos.Bitboards[pieceColor(Pawn, side)]

	sq := Bsf(*king)
	file, rank := sq%8, sq/8

	for f := file - 1; f <= file+1; f++ {
		if f < 0 || f > 7 {
			continue
		}
		pawn := alliedPawns & files[f]
		if pawn > 0 {
			rankDiff := abs(rank - Bsf(pawn)/8)
			if rankDiff > 2 {
				continue
			}
			score += 20 / rankDiff
		} else {
			score -= 20
		}
	}
	return score
}

// pawnStormScore returns a bonus/penalty when the king has castled depending on the pawns protection
func pawnStormScore(king *Bitboard, pos *Position, side Color) (score int) {
	// If the enemy pawns are near to the king, there might be a threat of opening a file, even if the pawn shield is intact. Penalties for storming enemy pawns must be lower than penalties for (semi)open files, otherwise the pawn storm might backfire, resulting in a blockage.
	alliedPawns := pos.Bitboards[pieceColor(Pawn, side)]
	enemyPawns := pos.Bitboards[pieceColor(Pawn, side.Opponent())]

	sq := Bsf(*king)
	file, rank := sq%8, sq/8
	direction := North
	if side == Black {
		direction = South
	}

	// if all pawns are locked, it's not a threat, cannot open files to attack the king...
	adjacentFiles := attacksFrontSpans[side][sq] | rayAttacks[direction][sq]
	if side == White {
		stoppers := (adjacentFiles & alliedPawns) << 8
		if enemyPawns&adjacentFiles == stoppers {
			return
		}
	} else {
		stoppers := (adjacentFiles & alliedPawns) >> 8
		if enemyPawns&adjacentFiles == stoppers {
			return
		}
	}

	for f := file - 1; f <= file+1; f++ {
		if f < 0 || f > 7 {
			continue
		}
		pawn := enemyPawns & files[f]
		if pawn > 0 {
			rankDiff := abs(rank - Bsf(pawn)/8)
			if rankDiff > 3 || rankDiff == 0 {
				continue
			}
			score -= 15 / rankDiff
		}
	}

	return
}

// evaluateQueen returns the middlegame and endgame score of the queen in the position
func (ev *Evaluation) evaluateQueen(pos *Position, queen *Bitboard, side Color) {
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
func (ev *Evaluation) evaluateRook(pos *Position, rook *Bitboard, side Color) {
	piece := pieceColor(Rook, side)
	sq := Bsf(*rook)

	// TODO: open file bonus???
	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	// attacksCount := Attacks(piece, *rook, ^pos.EmptySquares()).count()
	// ev.mgMobility[side] += (attacksCount - 5) * 3
	// ev.egMobility[side] += (attacksCount - 5) * 3

	ev.phase += 2

	return
}

// evaluateBishop returns the middlegame and endgame score of the bishop in the position
func (ev *Evaluation) evaluateBishop(pos *Position, bishop *Bitboard, side Color) {
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
func (ev *Evaluation) evaluateKnight(pos *Position, knight *Bitboard, side Color) {
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
func (ev *Evaluation) evaluatePawn(pos *Position, pawn *Bitboard, side Color) {
	piece := pieceColor(Pawn, side)
	sq := Bsf(*pawn)

	ev.mgMaterial[side] += middlegamePiecesScore[piece][sq]
	ev.egMaterial[side] += endgamePiecesScore[piece][sq]

	return
}

// evaluatePawnStructure evaluates the pawn structure for the side in the position passed
func (ev *Evaluation) evaluatePawnStructure(pos *Position, side Color) {
	doubledPawns := doubledPawns(pos, side)
	ev.mgPawnStrucutre[side] += doubledPawns.count() * DoubledPawnPenaltyMg
	ev.egPawnStructure[side] += doubledPawns.count() * DoubledPawnPenaltyEg

	isolatedPawns := isolatedPawns(pos, side)
	ev.mgPawnStrucutre[side] += isolatedPawns.count() * IsolatedPawnPenaltyMg
	ev.egPawnStructure[side] += isolatedPawns.count() * IsolatedPawnPenaltyEg

	backwardPawns := backwardPawns(pos, side)
	ev.mgPawnStrucutre[side] += backwardPawns.count() * BackwardPawnPenaltyMg
	ev.egPawnStructure[side] += backwardPawns.count() * BackwardPawnPenaltyEg

	passedPawns := passedPawns(pos, side)
	passedPawnBonus := [8]int{0, 0, 10, 20, 30, 60, 100, 0}
	for passedPawns > 0 {
		fromBB := passedPawns.NextBit()
		sq := Bsf(fromBB)
		rank := sq / 8
		if side == Black {
			rank = 7 - rank
		}

		ev.mgPawnStrucutre[side] += passedPawnBonus[rank]
		ev.egPawnStructure[side] += passedPawnBonus[rank]
	}
}

// doubledPawns returns a bitboard with the files with more than 1 pawn
func doubledPawns(pos *Position, side Color) Bitboard {
	doubledPawns := Bitboard(0)
	pawns := pos.Bitboards[pieceColor(Pawn, side)]

	for file := range 8 {
		pawnsInFile := pawns & files[file]
		if pawnsInFile.count() > 1 {
			pawnsInFile.NextBit() // removes one to not double count the penalty
			doubledPawns |= pawnsInFile
		}
	}
	return doubledPawns
}

// isolatedPawns a bitboard with the isolated pawns for the side
func isolatedPawns(pos *Position, side Color) Bitboard {
	isolatedPawns := Bitboard(0)
	pawns := pos.Bitboards[pieceColor(Pawn, side)]

	for file := range 8 {
		if IsolatedAdjacentFilesMask[file]&pawns == 0 {
			isolatedPawns |= files[file] & pawns
		}
	}
	return isolatedPawns
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

// backwardPawns returns a bitboard with the pawns that are backwards
// A backward pawn is a pawn that is not member of own front-attackspans but controlled by a sentry (definition from CPW)
func backwardPawns(pos *Position, side Color) Bitboard {
	pawns := pos.Bitboards[pieceColor(Pawn, side)]
	stops := pawns << 8
	if side == Black {
		stops = pawns >> 8
	}

	attackFrontSpans := Bitboard(0)
	for pawns > 0 {
		pawn := pawns.NextBit()
		attackFrontSpans |= attacksFrontSpans[side][Bsf(pawn)]
	}

	enemyPawns := pos.Bitboards[pieceColor(Pawn, side.Opponent())]
	enemyPawnsAttacks := pawnAttacks(&enemyPawns, side.Opponent())

	if side == White {
		return (stops & enemyPawnsAttacks & ^attackFrontSpans) >> 8
	} else {
		return (stops & enemyPawnsAttacks & ^attackFrontSpans) << 8
	}
}

// passedPawns returns a bitboard with the passed pawns for the side
// A passed pawn is a pawn whose path to promotion is not blocke nor attacked by the enemy pawns
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
