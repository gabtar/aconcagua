package aconcagua

const MaxMoves = 255

// movelist contains a list of chessMove for a determined position
type moveList struct {
	moves  [MaxMoves]Move
	length int
}

// add adds a move to the moveList
func (ml *moveList) add(move Move) {
	ml.moves[ml.length] = move
	ml.length++
}

// newMoveList returns a pointer to an empty moveList
func newMoveList() *moveList {
	return &moveList{
		moves:  [MaxMoves]Move{},
		length: 0,
	}
}

// sort sort the scores in the move list according to the scores array passed
func (ml *moveList) sort(scores []int) []int {
	for i := range ml.length - 1 {
		for j := range ml.length - i - 1 {
			if scores[j] < scores[j+1] {
				ml.moves[j], ml.moves[j+1] = ml.moves[j+1], ml.moves[j]
				scores[j], scores[j+1] = scores[j+1], scores[j]
			}
		}
	}
	return scores
}
