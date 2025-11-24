package engine

// TODO: Use fairly stockfish perft values to check move generation in variants

// Atomic Chess Rules:
// - Explosions:
// When a capture occurs, an explosion happens on the captured piece's square.
// The capturing piece, the captured piece, and all surrounding non-pawn pieces
// are removed from the board.
// - En Passant Captures:
// The explosion occurs on the ep target square. Need to check out this.
// - King safety:
// Kings are not allowed to capture any piece, as this would cause
// their own king to explode
// - Adjacent kings: Kings can be on adjacent squares without it being considered check,
// because neither can capture the other.
// - Check and checkmate:
// A move that results in an explosion that destroys the opponent's king results in an
/// immediate victory.
// - Checkmate is still a way to win.
// - It is illegal to make a capture that would result in your own king being blown up by
// the subsequent explosion, even if it means a checkmate

// ExplodedPiece encodes a piece type and the square it was on into a single 16-bit integer.
// The first 6 bits are for the square (0-63), and the next 4 bits are for the piece type (0-11).
type ExplodedPiece uint16

// encodeExplodedPiece creates an ExplodedPiece from a square and a piece
func encodeExplodedPiece(square int, piece int) ExplodedPiece {
	return ExplodedPiece(square) | (ExplodedPiece(piece) << 6)
}

// decode extracts the square and piece from an ExplodedPiece
func (ep ExplodedPiece) decode() (int, int) {
	square := int(ep & 0b111111)
	piece := int((ep >> 6) & 0b1111)
	return square, piece
}

// Explosion stores the list of pieces that exploded in a single move
type Explosion struct {
	explodedPieces [9]ExplodedPiece // 8 'potential' surrounding pieces per explosion + the target piece
	count          int
}

// add adds and exploded piece to the explosion
func (e *Explosion) add(piece ExplodedPiece) {
	e.explodedPieces[e.count] = piece
	e.count++
}

// clear clears the explosion
func (e *Explosion) clear() {
	for i := range e.explodedPieces {
		e.explodedPieces[i] = 0
	}
	e.count = 0
}

// ExplosionHistory stores the history of explosions during a search
type ExplosionHistory struct {
	history      [MaxHistoryMoves * 2]Explosion
	currentIndex int
}

// add records an explosion event in the history
func (eh *ExplosionHistory) add(piece int, square int) {
	eh.history[eh.currentIndex].add(encodeExplodedPiece(square, piece))
}

// increment increments the index of the history
func (eh *ExplosionHistory) increment() {
	eh.currentIndex++
}

// pop retrieves and removes the last explosion event from the history
func (eh *ExplosionHistory) pop() Explosion {
	eh.currentIndex--
	return eh.history[eh.currentIndex]
}

// clear resets the explosion history
func (eh *ExplosionHistory) clear() {
	for i := range eh.history {
		for j := range eh.history[i].explodedPieces {
			eh.history[i].explodedPieces[j] = 0
		}
		eh.history[i].count = 0
	}
	eh.currentIndex = 0
}

// NewExplosionHistory creates and returns a new ExplosionHistory
func NewExplosionHistory() *ExplosionHistory {
	return &ExplosionHistory{}
}

// AtomicPosition represents a position in an Atomic chess game
type AtomicPosition struct {
	pos              Position
	explosionHistory *ExplosionHistory
}

// NewAtomicPosition creates a new AtomicPosition.
func NewAtomicPosition(pos Position) *AtomicPosition {
	return &AtomicPosition{
		pos:              pos,
		explosionHistory: NewExplosionHistory(),
	}
}

// Evaluate evaluates the current position for the Atomic variant
func (ap *AtomicPosition) Evaluate() int {
	// TODO: Implement a specific evaluation for Atomic chess.
	// Higher penalties for pieces that may cause explosion near the opponent's king
	return ap.pos.Evaluate()
}

// MakeMove makes a move in the current position for Atomic chess
func (ap *AtomicPosition) MakeMove(move *Move) {
	ap.pos.MakeMove(move)

	if !isExplosion(move) {
		ap.explosionHistory.increment()
		return
	}

	// Remove the piece at exploded square since (standard) make move will not remove it
	toBB := bitboardFromIndex(move.to())
	piece := ap.pos.PieceAt(move.to())
	ap.pos.RemovePiece(piece, toBB)
	ap.explosionHistory.add(piece, move.to())

	adjacentSquares := kingAttacks(&toBB)
	for adjacentSquares > 0 {
		bb := adjacentSquares.NextBit()
		sq := Bsf(bb)
		piece := ap.pos.PieceAt(sq)

		if piece != NoPiece && pieceRole(piece) != Pawn {
			// Check if the explosion will affect rooks for updating castling rights
			if sq == h8 || sq == a8 || sq == h1 || sq == a1 {
				ap.pos.updateCastleRights(ap.pos.castling.updateCastleRights(move.from(), sq))
			}

			ap.pos.RemovePiece(piece, bb)
			ap.explosionHistory.add(piece, sq)
		}
	}
	ap.explosionHistory.increment()
}

// isExplosion returns if the move results in an explosion in atomic chess
func isExplosion(move *Move) bool {
	return move.flag() == capture || move.flag() == epCapture || move.flag() >= knightCapturePromotion
}

// UnmakeMove reverts a move, restoring any exploded pieces
func (ap *AtomicPosition) UnmakeMove(move *Move) {
	explosion := ap.explosionHistory.pop()
	if isExplosion(move) {
		for i := range explosion.count {
			sq, piece := explosion.explodedPieces[i].decode()
			ap.pos.AddPiece(piece, sq)
		}
	}
	ap.explosionHistory.history[ap.explosionHistory.currentIndex].clear()
	ap.pos.UnmakeMove(move)
}

// GenerateCaptures generates all pseudo legal captures in the position
func (ap *AtomicPosition) GenerateCaptures(ml *MoveList, pd *PositionData) {
	// Need to check all posibles moves, we cannot restrict by check, since check in Atomic chess works different
	bitboards := ap.pos.getBitboards(ap.pos.Turn)

	for piece, bb := range bitboards {
		for bb > 0 {
			pieceBB := bb.NextBit()
			switch piece {
			case Queen:
				genMovesFromTargets(&pieceBB, (rookAttacks(Bsf(pieceBB), pd.allies|pd.enemies)|bishopAttacks(Bsf(pieceBB), pd.allies|pd.enemies))&pd.enemies, ml, pd)
			case Rook:
				genMovesFromTargets(&pieceBB, rookAttacks(Bsf(pieceBB), pd.allies|pd.enemies)&pd.enemies, ml, pd)
			case Bishop:
				genMovesFromTargets(&pieceBB, bishopAttacks(Bsf(pieceBB), pd.allies|pd.enemies)&pd.enemies, ml, pd)
			case Knight:
				genMovesFromTargets(&pieceBB, knightAttacksTable[Bsf(pieceBB)]&pd.enemies, ml, pd)
			case Pawn:
				genPawnCapturesFromTargets(&pieceBB, pawnAttacks(&pieceBB, ap.pos.Turn)&pd.enemies, ap.pos.Turn, ml, pd)
			}
		}
	}
	genPosibleEpCaptures(&ap.pos, ap.pos.Turn, ml)

	filterIllegalMoves(ap, ml)
}

// GenerateNonCaptures generates all pseudo legal non-captures in the position
func (ap *AtomicPosition) GenerateNonCaptures(ml *MoveList, pd *PositionData) {
	// Need to check all posibles moves, we cannot restrict by check, since check in Atomic chess works different
	bitboards := ap.pos.getBitboards(ap.pos.Turn)
	blocks := pd.allies | pd.enemies

	for piece, bb := range bitboards {
		for bb > 0 {
			pieceBB := bb.NextBit()
			switch piece {
			case King:
				genMovesFromTargets(&pieceBB, kingAttacks(&pieceBB) & ^blocks, ml, pd)
				genCastleMoves(&ap.pos, ml)
			case Queen:
				genMovesFromTargets(&pieceBB, (rookAttacks(Bsf(pieceBB), blocks)|bishopAttacks(Bsf(pieceBB), blocks))&^blocks, ml, pd)
			case Rook:
				genMovesFromTargets(&pieceBB, rookAttacks(Bsf(pieceBB), blocks)&^blocks, ml, pd)
			case Bishop:
				genMovesFromTargets(&pieceBB, bishopAttacks(Bsf(pieceBB), blocks)&^blocks, ml, pd)
			case Knight:
				genMovesFromTargets(&pieceBB, knightAttacksTable[Bsf(pieceBB)]&^blocks, ml, pd)
			case Pawn:
				emptySquares := ^blocks
				genPawnMovesFromTarget(&pieceBB, pawnPushes(&pieceBB, &emptySquares, ap.pos.Turn)&^blocks, ap.pos.Turn, ml, pd)
			}
		}
	}

	filterIllegalMoves(ap, ml)
}

// filterIllegalMoves filters out illegal moves from the move list
func filterIllegalMoves(ap *AtomicPosition, ml *MoveList) {
	for i := ml.length - 1; i >= 0; i-- { // NOTE: Iterating backwards avoids affecting the indices already visited, so we dont skip moves
		if !ap.IsLegal(ml.moves[i]) {
			ml.moves[i], ml.moves[ml.length-1] = ml.moves[ml.length-1], NoMove
			ml.scores[i], ml.scores[ml.length-1] = ml.scores[ml.length-1], 0
			ml.length--
		}
	}
}

// GetPositionData returns the position data for the current position
func (ap *AtomicPosition) GetPositionData() PositionData {
	return ap.pos.generatePositionData()
}

// IsLegal returns if the move is legal in the current position for Atomic chess
func (ap *AtomicPosition) IsLegal(move Move) bool {
	side := ap.pos.Turn
	ap.MakeMove(&move)

	alliedKingCount := ap.pos.KingPosition(side).count()
	enemyKingCount := ap.pos.KingPosition(side.Opponent()).count()

	// If our own king has exploded, the move is illegal, we have been checkmated
	if alliedKingCount == 0 {
		ap.UnmakeMove(&move)
		return false
	}

	// If we explode enemy king, its a checkmate and the move is illegal
	if enemyKingCount == 0 && isExplosion(&move) {
		ap.UnmakeMove(&move)
		return true
	}

	// Our king can be in check as long as we can explode the opponent's king in the next move
	// But only if we are not already in check
	// And not placing the king in check
	if ap.pos.Check(side) {
		canAtomicCheckmate := ap.canAtomicCheckmateOpponent(side.Opponent())
		if canAtomicCheckmate {
			ap.UnmakeMove(&move)
			return true
		}

		// Special case check: Adjacent Kings
		// We can be in check if our king is adjacent to the enemy king, because cannot be captured,
		// since explosion will blow up both kings, and its illegal
		ourKing := ap.pos.KingPosition(side)
		enemyKing := ap.pos.KingPosition(side.Opponent())
		if !canAtomicCheckmate && ourKing&kingAttacks(&enemyKing) > 0 {
			ap.UnmakeMove(&move)
			return true
		}

		// Else the move is illegal, cannot be in regular check
		ap.UnmakeMove(&move)
		return false
	}

	ap.UnmakeMove(&move)
	return true
}

// canAtomicCheckmateOpponent returns if the opponent can be checkmated by explosion in the next move
func (ap *AtomicPosition) canAtomicCheckmateOpponent(side Color) bool {
	enemyKing := ap.pos.KingPosition(side.Opponent())

	// Check first if the king is not in check
	attacks := ap.pos.attackers(Bsf(enemyKing), side, ^ap.pos.EmptySquares())
	if attacks > 0 {
		return false
	}

	ourKing := ap.pos.KingPosition(side)
	enemyKingZone := kingAttacks(&enemyKing) | enemyKing
	explodableTargets := enemyKingZone & ap.pos.pieces[side.Opponent()]

	// Need to check if a capture will explode opponent king and not both Kings at the same time(is illegal)
	for explodableTargets > 0 {
		bb := explodableTargets.NextBit()
		targetSq := Bsf(bb)
		explosionZone := kingAttacks(&bb) | bb

		if ap.pos.attackers(targetSq, side, ^ap.pos.EmptySquares()) > 0 && explosionZone&ourKing == 0 {
			return true
		}
	}
	return false
}

// genPosibleEpCaptures generates all posible en passant captures, even if they are illegal in standard chess
func genPosibleEpCaptures(pos *Position, side Color, ml *MoveList) {
	if pos.enPassantTarget == 0 {
		return
	}
	from := potentialEpCapturers(pos, side)

	for from > 0 {
		fromBB := from.NextBit()
		move := encodeMove(uint16(Bsf(fromBB)), uint16(Bsf(pos.enPassantTarget)), epCapture)
		ml.add(*move)
	}
}
