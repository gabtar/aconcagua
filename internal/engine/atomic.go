package engine

// TODO: Use fairly stockfish perft values to check move generation in variants

// Atomic Chess Rules:
// Explosions:
// When a capture occurs, an explosion happens on the captured piece's square.
// The capturing piece, the captured piece, and all surrounding non-pawn pieces
// are removed from the board.
// Pawn Captures:
// This is an special case of explosion. When a pawn is captured, both pawns are
// removed from the board, and the non pawn surrounding pieces are removed too
// The special case is also the en passant capture, that the explosion occurs on the
// ep target square. Need to check out this.
// King safety:
// Kings are not allowed to capture any piece, as this would cause
// their own king to explode
// Adjacent kings: Kings can be on adjacent squares without it being considered check,
// because neither can capture the other.
// Check and checkmate:
// - A move that results in an explosion that destroys the opponent's king results in an
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

// ExplosionHistory stores the history of explosions for all moves in the game,
// similar to how PositionHistory stores position states
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

// MakeMove applies a move to the position, including Atomic explosion logic
func (ap *AtomicPosition) MakeMove(move *Move) {
	ap.pos.MakeMove(move)

	if !isExplosion(move) {
		return
	}

	// Always remove the piece at exploded square since make move will not remove it
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
	if isExplosion(move) {
		explosion := ap.explosionHistory.pop()
		for i := range explosion.count {
			sq, piece := explosion.explodedPieces[i].decode()
			ap.pos.AddPiece(piece, sq)
		}
	}

	ap.pos.UnmakeMove(move)
}

// IsLegal returns if the move is legal in the current position
// NOTE:
// Legality moves on Atomic chess
// - Explode king. If the move explodes your own king, it is illegal.
// - Check King. If the move leaves your own king in check, it is illegal. Except if the only piece giving check is the oponent king.
// - King Capture. The king cannot capture any piece(will result in an explosion and remove your own king).
func (ap *AtomicPosition) IsLegal(move Move) bool {
	// Kings cannot capture any piece
	if pieceRole(ap.pos.PieceAt(move.from())) == King && move.flag() == capture {
		return false
	}

	side := ap.pos.Turn
	ap.MakeMove(&move)

	alliedKingCount := ap.pos.KingPosition(side).count()
	enemyKingCount := ap.pos.KingPosition(side.Opponent()).count()

	// If own king exploded, the move is illegal
	if alliedKingCount == 0 {
		ap.UnmakeMove(&move)
		return false
	}

	// If enemy king explodes, is legal and we won the game
	if enemyKingCount == 0 {
		ap.UnmakeMove(&move)
		return true
	}

	// Own king cannot be in check
	if ap.pos.Check(side) {
		ap.UnmakeMove(&move)
		return false
	}

	ap.UnmakeMove(&move)
	return true
}
