package aconcagua

const MaxHistoryMoves = 255

// PositionHistory represents the history of the position and castle rigths
type PositionHistory struct {
	previousPosition     [MaxHistoryMoves * 2]uint64
	previousState        [MaxHistoryMoves]positionBefore
	previousCastleRigths [MaxHistoryMoves]castling
	currentIndex         int
}

// count returns the number of previous positions with the same hash
func (ph *PositionHistory) repetitionCount(hash uint64) int {
	count := 0
	for i := range MaxHistoryMoves {
		if ph.previousPosition[i] == hash {
			count++
		}
	}
	return count
}

// clear clears the position history
func (ph *PositionHistory) clear() {
	ph.previousState = [MaxHistoryMoves]positionBefore{}
	ph.previousCastleRigths = [MaxHistoryMoves]castling{}
	ph.currentIndex = 0
}

// add adds a previous position and a castle rigths to the position history
func (ph *PositionHistory) add(pb positionBefore, c castling, hash uint64) {
	ph.previousState[ph.currentIndex] = pb
	ph.previousCastleRigths[ph.currentIndex] = c
	ph.previousPosition[ph.currentIndex] = hash
	ph.currentIndex++
}

// pop returns the last position and castle rigths added to the position history
func (ph *PositionHistory) pop() (positionBefore, castling) {
	ph.currentIndex--
	ph.previousPosition[ph.currentIndex] = 0
	return ph.previousState[ph.currentIndex], ph.previousCastleRigths[ph.currentIndex]
}

// NewPositionHistory returns a new PositionHistory
func NewPositionHistory() *PositionHistory {
	return &PositionHistory{
		previousPosition:     [MaxHistoryMoves * 2]uint64{},
		previousState:        [MaxHistoryMoves]positionBefore{},
		previousCastleRigths: [MaxHistoryMoves]castling{},
		currentIndex:         0,
	}
}
