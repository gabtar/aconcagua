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

// mvvScore is the score to use for move ordering for each victim type piece
var mvvScore = [6]int{0, 3200, 1800, 1000, 1000, 300} // King, Queen, Rook, Bishop, Knight, Pawn

// scoreNoisy scores the captures/promotion moves for move ordering
func (ml *MoveList) scoreNoisy(pos *Position, nh *NoisyHistoryTable) {
	for i := 0; i < ml.length; i++ {
		attacker := pieceRole(pos.PieceAt(ml.moves[i].from()))
		victim := pos.getCapturedPiece(&ml.moves[i])
		victimIdx := pieceRole(victim)
		mvv := 0
		if victim == NoPiece { // quiet promotion
			victimIdx = 6 + ml.moves[i].flag() - knightPromotion
		} else {
			mvv = mvvScore[pieceRole(victim)]
		}
		ml.scores[i] = MaxHistoryBonus + mvv + nh[attacker][victimIdx][ml.moves[i].to()]
	}
}

// scoreQuiets scores the non captures moves by history score
func (ml *MoveList) scoreQuiets(qh *QuietHistoryTable, side Color, startIndex int) {
	for i := startIndex; i < ml.length; i++ {
		ml.scores[i] = qh[side][ml.moves[i].from()][ml.moves[i].to()]
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
