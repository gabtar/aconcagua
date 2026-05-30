package engine

const MaxHistoryMoves = 255

// PositionHistory represents the history of the position and castle rigths
type PositionHistory struct {
	previousPosition     [MaxHistoryMoves * 2]uint64
	previousState        [MaxHistoryMoves * 2]positionBefore
	previousCastleRigths [MaxHistoryMoves * 2]castlingRights
	moveCount            int
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
