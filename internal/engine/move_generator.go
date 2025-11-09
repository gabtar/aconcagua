package engine

// MoveGenerator stages
const (
	HashMoveStage = iota
	GenerateCapturesStage
	CapturesStage
	FirstKillerStage
	SecondKillerStage
	GenerateNonCapturesStage
	NonCapturesStage
	BadCapturesStage
	EndStage
)

// MoveGenerator implements a staged move generator for a given position
type MoveGenerator struct {
	stage                      int
	moveNumber                 int // the move count selected so far
	pos                        *Position
	pd                         PositionData
	hashMove, killer1, killer2 *Move
	historyMoves               *HistoryMovesTable
	moves                      *MoveList
	badCapLength               int // To track the number of bad captures
}

// NewMoveGenerator returns a new move generator
func NewMoveGenerator(pos *Position, hashMove *Move, killer1 *Move, killer2 *Move, historyMoves *HistoryMovesTable) *MoveGenerator {
	return &MoveGenerator{
		stage:        HashMoveStage,
		pos:          pos,
		hashMove:     hashMove,
		killer1:      killer1,
		killer2:      killer2,
		moveNumber:   -1, // NOTE: initialize with -1 to make the first move selected to have moveNumber = 0
		historyMoves: historyMoves,
		moves:        NewMoveList(),
	}
}

// nextMove return the nextMove move of the position
func (mg *MoveGenerator) nextMove() (move Move) {
	mg.moveNumber++
	switch mg.stage {
	case HashMoveStage:
		mg.stage = GenerateCapturesStage
		if *mg.hashMove != NoMove {
			return *mg.hashMove
		}
		fallthrough
	case GenerateCapturesStage:
		mg.stage = CapturesStage
		mg.pd = mg.pos.generatePositionData()
		mg.pos.generateCaptures(mg.moves, &mg.pd)
		mg.moves.scoreCaptures(mg.pos)
		fallthrough
	case CapturesStage:
		move = mg.nextGoodCapture()
		if move != NoMove {
			return move
		}
		mg.badCapLength = mg.moves.length // keep track of the number of (bad) captures left
		mg.stage = FirstKillerStage
		fallthrough
	case FirstKillerStage:
		mg.stage = SecondKillerStage
		move = *mg.killer1
		if move != NoMove && move != *mg.hashMove && mg.isLegal(move) {
			return move
		}
		fallthrough
	case SecondKillerStage:
		mg.stage = GenerateNonCapturesStage
		move = *mg.killer2
		if move != NoMove && move != *mg.hashMove && mg.isLegal(move) {
			return move
		}
		fallthrough
	case GenerateNonCapturesStage:
		mg.stage = NonCapturesStage
		mg.pos.generateNonCaptures(mg.moves, &mg.pd)
		mg.moves.scoreNonCaptures(mg.historyMoves, mg.pos.Turn, mg.badCapLength)
		fallthrough
	case NonCapturesStage:
		move = mg.nextNonCapture()
		if move != NoMove {
			return move
		}
		mg.stage = BadCapturesStage
		fallthrough
	case BadCapturesStage:
		move = mg.nextBadCapture()
		if move != NoMove {
			return move
		}
		mg.stage = EndStage
	case EndStage:
		return NoMove
	}
	return
}

// nextGoodCapture returns the next good capture move or NoMove
func (mg *MoveGenerator) nextGoodCapture() (move Move) {
	for {
		idx := mg.moves.getBestIndex(0)
		// Picks a move unless its the first bad capture or out of moves
		if idx < 0 || mg.moves.scores[idx] < 0 {
			return NoMove
		}
		move = mg.moves.moves[idx]
		mg.moves.swap(idx)

		if move != *mg.hashMove {
			return move
		}
	}
}

// nextNonCapture returns the next non capture move or NoMove
func (mg *MoveGenerator) nextNonCapture() (move Move) {
	for {
		idx := mg.moves.getBestIndex(mg.badCapLength)
		// Picks the move unless it is a bad capture
		if idx < mg.badCapLength {
			return NoMove
		}
		move = mg.moves.moves[idx]
		mg.moves.swap(idx)

		if move != *mg.hashMove && move != *mg.killer1 && move != *mg.killer2 {
			return move
		}
	}
}

// nextBadCapture returns the next bad capture move or NoMove
func (mg *MoveGenerator) nextBadCapture() (move Move) {
	for {
		idx := mg.moves.getBestIndex(0)
		if idx < 0 {
			return NoMove
		}
		move = mg.moves.moves[idx]
		mg.moves.swap(idx)

		if move != *mg.hashMove {
			return move
		}
	}
}

// generateCaptures generates all captures in the position and stores them in the move list
func (pos *Position) generateCaptures(ml *MoveList, pd *PositionData) {
	bitboards := pos.getBitboards(pos.Turn)

	for piece, bb := range bitboards {
		for bb > 0 {
			pieceBB := bb.NextBit()
			switch piece {
			case King:
				genMovesFromTargets(&pieceBB, kingMoves(&pieceBB, pos, pos.Turn)&pd.enemies, ml, pd)
			case Queen:
				genMovesFromTargets(&pieceBB, (rookMoves(&pieceBB, pd)|bishopMoves(&pieceBB, pd))&pd.enemies, ml, pd)
			case Rook:
				genMovesFromTargets(&pieceBB, rookMoves(&pieceBB, pd)&pd.enemies, ml, pd)
			case Bishop:
				genMovesFromTargets(&pieceBB, bishopMoves(&pieceBB, pd)&pd.enemies, ml, pd)
			case Knight:
				genMovesFromTargets(&pieceBB, knightMoves(&pieceBB, pd)&pd.enemies, ml, pd)
			case Pawn:
				genPawnCapturesMoves(&pieceBB, pos.Turn, ml, pd)
			}
		}
	}
	genEnPassantCaptures(pos, pos.Turn, ml)
}

// generateNonCaptures generates all non captures in the position and stores them in the move list
func (pos *Position) generateNonCaptures(ml *MoveList, pd *PositionData) {
	bitboards := pos.getBitboards(pos.Turn)

	for piece, bb := range bitboards {
		for bb > 0 {
			pieceBB := bb.NextBit()
			switch piece {
			case King:
				genMovesFromTargets(&pieceBB, kingMoves(&pieceBB, pos, pos.Turn)&^pd.enemies, ml, pd)
				genCastleMoves(pos, ml)
			case Queen:
				genMovesFromTargets(&pieceBB, (rookMoves(&pieceBB, pd)|bishopMoves(&pieceBB, pd))&^pd.enemies, ml, pd)
			case Rook:
				genMovesFromTargets(&pieceBB, rookMoves(&pieceBB, pd)&^pd.enemies, ml, pd)
			case Bishop:
				genMovesFromTargets(&pieceBB, bishopMoves(&pieceBB, pd)&^pd.enemies, ml, pd)
			case Knight:
				genMovesFromTargets(&pieceBB, knightMoves(&pieceBB, pd)&^pd.enemies, ml, pd)
			case Pawn:
				genPawnMovesFromTarget(&pieceBB, pawnMoves(&pieceBB, pd, pos.Turn)&^pd.enemies, pos.Turn, ml, pd)
			}
		}
	}
}

// kingMoves returns a bitboard with the legal moves of the king from the bitboard passed
func kingMoves(k *Bitboard, pos *Position, side Color) (moves Bitboard) {
	pos.RemovePiece(pieceColor(King, side), *k)
	attackedSquares := pos.AttackedSquares(side.Opponent()) // to check attacks rays (behind) the king he is actually blocking
	pos.AddPiece(pieceColor(King, side), Bsf(*k))

	moves = kingAttacks(k) & ^attackedSquares & ^pos.pieces[side]
	return
}

// rookMoves returns a bitboard with the legal moves of the rook from the bitboard passed
func rookMoves(r *Bitboard, pd *PositionData) (moves Bitboard) {
	return rookAttacks(Bsf(*r), pd.allies|pd.enemies) &
		^pd.allies &
		pd.checkRestrictedSquares &
		pinRestrictedSquares(*r, pd.kingPosition, pd.pinnedPieces)
}

// bishopMoves returns a bitboard with the legal moves of the bishop from the bitboard passed
func bishopMoves(b *Bitboard, pd *PositionData) (moves Bitboard) {
	return bishopAttacks(Bsf(*b), pd.allies|pd.enemies) &
		^pd.allies &
		pd.checkRestrictedSquares &
		pinRestrictedSquares(*b, pd.kingPosition, pd.pinnedPieces)
}

// knightMoves returns a bitboard with the legal moves of the knight from the bitboard passed
func knightMoves(k *Bitboard, pd *PositionData) (moves Bitboard) {
	// If the knight is pinned it can move at all
	if *k&pd.pinnedPieces > 0 {
		return Bitboard(0)
	}
	return knightAttacksTable[Bsf(*k)] & ^pd.allies & pd.checkRestrictedSquares
}

// pawnMoves returns a Bitboard with the squares a pawn can move to in the passed position
func pawnMoves(p *Bitboard, pd *PositionData, side Color) (moves Bitboard) {
	posibleCaptures := pawnAttacks(p, side) & pd.enemies
	posiblesMoves := Bitboard(0)
	emptySquares := ^(pd.allies | pd.enemies)

	if side == White {
		singleMove := *p << 8 & emptySquares
		firstPawnMoveAvailable := (*p & ranks[1]) << 16 & (singleMove << 8) & emptySquares
		posiblesMoves = singleMove | firstPawnMoveAvailable
	} else {
		singleMove := *p >> 8 & emptySquares
		firstPawnMoveAvailable := (*p & ranks[6]) >> 16 & (singleMove >> 8) & emptySquares
		posiblesMoves = singleMove | firstPawnMoveAvailable
	}

	moves = (posibleCaptures | posiblesMoves) & pd.checkRestrictedSquares &
		pinRestrictedSquares(*p, pd.kingPosition, pd.pinnedPieces)
	return
}

// potentialEpCapturers returns a bitboard with the potential pawn that can caputure enPassant
func potentialEpCapturers(pos *Position, side Color) (epCaptures Bitboard) {
	epShift := pos.enPassantTarget >> 8
	if side == Black {
		epShift = epShift << 16
	}
	notInHFile := epShift & ^(epShift & files[7])
	notInAFile := epShift & ^(epShift & files[0])

	epCaptures |= pos.getBitboards(side)[Pawn] & (notInAFile>>1 | notInHFile<<1)
	return
}

// lastRank returns the rank of the last rank for the side passed
func lastRank(side Color) (rank Bitboard) {
	if side == White {
		rank = ranks[7]
	} else {
		rank = ranks[0]
	}
	return
}

// genMovesFromTargets generates the moves from the square passed to the targets passed in a MoveList
func genMovesFromTargets(from *Bitboard, targets Bitboard, ml *MoveList, pd *PositionData) {
	for targets > 0 {
		toSquare := targets.NextBit()
		flag := quiet
		if toSquare&pd.enemies > 0 {
			flag = capture
		}
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
}

// genCastleMoves generates the castles moves availabes in the move list
func genCastleMoves(pos *Position, ml *MoveList) {
	fromSq := pos.castling.kingsStartSquare[pos.Turn]
	flipModifier := 2 - (int(pos.Turn) + 1)
	kingsideCastleTo := 62 ^ (flipModifier * 56) // Flip to g1 or g8 depending on current side
	queensideCastleTo := 58 ^ (flipModifier * 56)
	if pos.castling.chess960 {
		kingsideCastleTo = pos.castling.rooksStartSquare[pos.Turn][0]
		queensideCastleTo = pos.castling.rooksStartSquare[pos.Turn][1]
	}

	if pos.canCastle(pos.Turn, kingsideCastle) {
		ml.add(*encodeMove(uint16(fromSq), uint16(kingsideCastleTo), kingsideCastle))
	}
	if pos.canCastle(pos.Turn, queensideCastle) {
		ml.add(*encodeMove(uint16(fromSq), uint16(queensideCastleTo), queensideCastle))
	}
}

// genPawnPromotions generates the pawn promotions in the move list
func genPawnPromotions(from *Bitboard, to *Bitboard, ml *MoveList, isCapture bool) {
	if isCapture {
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(*to)), knightCapturePromotion))
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(*to)), bishopCapturePromotion))
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(*to)), rookCapturePromotion))
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(*to)), queenCapturePromotion))
	} else {
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(*to)), knightPromotion))
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(*to)), bishopPromotion))
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(*to)), rookPromotion))
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(*to)), queenPromotion))
	}
}

// genPawnMovesFromTarget generates the pawn moves in the move list
func genPawnMovesFromTarget(from *Bitboard, targets Bitboard, side Color, ml *MoveList, pd *PositionData) {
	for targets > 0 {
		toSquare := targets.NextBit()
		isPromotion := lastRank(side) & toSquare

		switch {
		case isPromotion > 0 && pd.enemies&toSquare > 0: // Promo Capture
			genPawnPromotions(from, &toSquare, ml, true)
		case isPromotion > 0: // Promotion
			genPawnPromotions(from, &toSquare, ml, false)
		case pd.enemies&toSquare > 0: // Capture
			ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), capture))
		case Bsf(toSquare)-Bsf(*from) == 16 || Bsf(*from)-Bsf(toSquare) == 16: // Double Push
			ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), doublePawnPush))
		default: // Quiet
			ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), quiet))
		}
	}
}

// genPawnCapturesMoves generates the pawn captures in the move list
func genPawnCapturesMoves(from *Bitboard, side Color, ml *MoveList, pd *PositionData) {
	toSquares := pawnMoves(from, pd, side) & pd.enemies

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		isPromotion := lastRank(side) & toSquare

		switch {
		case isPromotion > 0 && pd.enemies&toSquare > 0: // Promo Capture
			genPawnPromotions(from, &toSquare, ml, true)
		default: // Capture
			ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), capture))
		}
	}
}

// genEnPassantCaptures generates the enPassant captures on the move list
func genEnPassantCaptures(pos *Position, side Color, ml *MoveList) {
	if pos.enPassantTarget == 0 {
		return
	}
	from := potentialEpCapturers(pos, side)

	for from > 0 {
		fromBB := from.NextBit()
		move := encodeMove(uint16(Bsf(fromBB)), uint16(Bsf(pos.enPassantTarget)), epCapture)

		pos.MakeMove(move)
		if !pos.Check(side) {
			ml.add(*move)
		}
		pos.UnmakeMove(move)
	}
}

// checkRestrictedSquares returns a bitboard with the squares that are allowed to move when in check
func checkRestrictedSquares(king Bitboard, checkingSliders Bitboard, checkingNonSliders Bitboard) (allowedSquares Bitboard) {
	checkingPieces := checkingSliders | checkingNonSliders
	if checkingPieces.count() == 0 {
		return AllSquares
	}

	if checkingPieces == checkingSliders && checkingPieces.count() == 1 {
		return getRayPath(&checkingPieces, &king) | checkingPieces
	}

	if checkingPieces.count() == 1 {
		return checkingPieces
	}

	return
}

// pinRestrictedSquares returns a bitboard with the squares allowed to move when the piece is pinned
func pinRestrictedSquares(piece Bitboard, king Bitboard, pinnedPieces Bitboard) (restrictedSquares Bitboard) {
	if pinnedPieces&piece > 0 {
		direction := directions[Bsf(piece)][Bsf(king)]
		return raysDirection(king, direction)
	}
	return AllSquares
}

// isLegal returns if the passed move is legal in the current position
func (mg *MoveGenerator) isLegal(move Move) bool {
	if move == NoMove {
		return false
	}

	side := mg.pos.Turn
	from, to := move.from(), move.to()
	fromBB := bitboardFromIndex(from)
	toBB := bitboardFromIndex(to)
	if fromBB&mg.pd.allies == 0 {
		return false
	}
	if toBB&(mg.pd.allies|mg.pd.enemies) != 0 {
		return false
	}

	flag := move.flag()
	piece := pieceRole(mg.pos.PieceAt(from))
	if flag == kingsideCastle || flag == queensideCastle {
		return mg.pos.canCastle(side, flag)
	}
	if flag == doublePawnPush && piece != Pawn {
		return false
	}
	if piece == Pawn && flag == quiet && abs(to-from) == 16 {
		return false
	}

	legalMoves := Bitboard(0)
	switch piece {
	case Knight:
		legalMoves = knightMoves(&fromBB, &mg.pd)
	case Bishop:
		legalMoves = bishopMoves(&fromBB, &mg.pd)
	case Rook:
		legalMoves = rookMoves(&fromBB, &mg.pd)
	case Queen:
		legalMoves = (rookMoves(&fromBB, &mg.pd) | bishopMoves(&fromBB, &mg.pd))
	case King:
		legalMoves = kingMoves(&fromBB, mg.pos, side)
	case Pawn:
		legalMoves = pawnMoves(&fromBB, &mg.pd, side)
	default:
		return false
	}
	return (toBB & legalMoves) != 0
}
