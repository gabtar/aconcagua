package aconcagua

const MaxMoves = 255

// movelist contains a list of chessMove for a determined position
type moveList struct {
	moves  [MaxMoves]chessMove
	length int
}

// add adds a move to the moveList
func (ml *moveList) add(move chessMove) {
	ml.moves[ml.length] = move
	ml.length++
}

// newMoveList returns a pointer to an empty moveList
func newMoveList() *moveList {
	return &moveList{
		moves:  [MaxMoves]chessMove{},
		length: 0,
	}
}
