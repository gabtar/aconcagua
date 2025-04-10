package aconcagua

// PositionHistory represents the history of the position and castle rigths
type PositionHistory struct {
	previousPosition     [100]positionBefore
	previousCastleRigths [100]castling
	currentIndex         int
}

// add adds a previous position and a castle rigths to the position history
func (ph *PositionHistory) add(pp positionBefore, cr castling) {
	ph.previousPosition[ph.currentIndex] = pp
	ph.previousCastleRigths[ph.currentIndex] = cr
	ph.currentIndex++
}

// pop returns the last position and castle rigths added to the position history
func (ph *PositionHistory) pop() (pp positionBefore, cr castling) {
	ph.currentIndex--
	return ph.previousPosition[ph.currentIndex], ph.previousCastleRigths[ph.currentIndex]
}

// NewPositionHistory returns a new PositionHistory
func NewPositionHistory() *PositionHistory {
	return &PositionHistory{}
}
