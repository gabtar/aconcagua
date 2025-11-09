package engine

const MaxLegalMoves = 100

// MoveList is a list of moves
type MoveList struct {
	moves  [MaxLegalMoves]Move
	scores [MaxLegalMoves]int
	length int
}

// NewMoveList returns a pointer to a MoveList
func NewMoveList() *MoveList {
	return &MoveList{}
}

// add adds a move to the list
func (ml *MoveList) add(move Move) {
	ml.moves[ml.length] = move
	ml.length++
}

// scoreCaptures scores the captures moves by static exchange evaluation
func (ml *MoveList) scoreCaptures(pos *Position) {
	for i := 0; i < ml.length; i++ {
		ml.scores[i] = pos.see(ml.moves[i].from(), ml.moves[i].to())
	}
}

// scoreNonCaptures scores the non captures moves by history score
func (ml *MoveList) scoreNonCaptures(ht *HistoryMovesTable, side Color, startIndex int) {
	for i := startIndex; i < ml.length; i++ {
		ml.scores[i] = ht[side][ml.moves[i].from()][ml.moves[i].to()]
	}
}

// getBestIndex returns the index and the score of the best move in the list
func (ml *MoveList) getBestIndex(start int) (index int) {
	if ml.length == start {
		return -1
	}

	bestIndex := start
	for i := start + 1; i < ml.length; i++ {
		if ml.scores[i] > ml.scores[bestIndex] {
			bestIndex = i
		}
	}
	return bestIndex
}

// swap moves the move at index to the end of the list, reducing the length
func (ml *MoveList) swap(index int) {
	ml.moves[ml.length-1], ml.moves[index] = ml.moves[index], ml.moves[ml.length-1]
	ml.scores[ml.length-1], ml.scores[index] = ml.scores[index], ml.scores[ml.length-1]
	ml.length--
}
