package engine

const MaxHistoryMoves = 255

// PositionHistory represents the history of the position and castle rigths
type PositionHistory struct {
	previousPosition     [MaxHistoryMoves * 2]uint64
	previousState        [MaxHistoryMoves * 2]positionBefore
	previousCastleRigths [MaxHistoryMoves * 2]castlingRights
	moveCount            int
}

// isRepetition returns if the position has been repeated in the current search
func (ph *PositionHistory) isRepetition(hash uint64, halfmoveClock int) bool {
	// Calculate search limit based on halfmove clock
	// A repetition cannot never occur after a halfmove clock reset
	lastIrreversibleMove := max(ph.moveCount-halfmoveClock, 0)

	// Check all positions back to the last irreversible move
	// Only check positions with same side to move (every 2 plies)
	for i := ph.moveCount - 2; i >= lastIrreversibleMove; i -= 2 {
		if ph.previousPosition[i] == hash {
			return true
		}
	}

	return false
}

// clear clears the position history
func (ph *PositionHistory) clear() {
	for i := range MaxHistoryMoves * 2 {
		ph.previousState[i] = positionBefore(0)
		ph.previousCastleRigths[i] = noCastling
	}
	ph.moveCount = 0
}

// add adds a previous position and a castle rigths to the position history
func (ph *PositionHistory) add(pb positionBefore, c castlingRights, hash uint64) {
	ph.previousState[ph.moveCount] = pb
	ph.previousCastleRigths[ph.moveCount] = c
	ph.previousPosition[ph.moveCount] = hash
	ph.moveCount++
}

// pop returns the last position and castle rigths added to the position history
func (ph *PositionHistory) pop() (positionBefore, castlingRights) {
	ph.moveCount--
	ph.previousPosition[ph.moveCount] = 0
	return ph.previousState[ph.moveCount], ph.previousCastleRigths[ph.moveCount]
}

// NewPositionHistory returns a new PositionHistory
func NewPositionHistory() *PositionHistory {
	return &PositionHistory{
		previousPosition:     [MaxHistoryMoves * 2]uint64{},
		previousState:        [MaxHistoryMoves * 2]positionBefore{},
		previousCastleRigths: [MaxHistoryMoves * 2]castlingRights{},
		moveCount:            0,
	}
}
