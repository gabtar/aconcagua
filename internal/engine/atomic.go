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
// The special case should be the en passant capture, that the explosion occurs on the
// ep target square, i think. Need to check out this.
//
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
	explodedPieces [8]ExplodedPiece // 8 'potential' surrounding pieces per explosion
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

	if move.flag() != capture && move.flag() <= knightCapturePromotion {
		return
	}

	fromBB := bitboardFromIndex(move.to())
	adjacentSquares := kingAttacks(&fromBB)
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

// UnmakeMove reverts a move, restoring any exploded pieces
func (ap *AtomicPosition) UnmakeMove(move *Move) {
	explosion := ap.explosionHistory.pop()
	if move.flag() == capture || move.flag() >= knightCapturePromotion {
		for i := 0; i < explosion.count; i++ {
			sq, piece := explosion.explodedPieces[i].decode()
			ap.pos.AddPiece(piece, sq)
		}
	}

	ap.pos.UnmakeMove(move)
}
