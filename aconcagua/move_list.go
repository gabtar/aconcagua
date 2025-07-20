package aconcagua

// MoveList is a list of chess moves for a given position of chess
type MoveList []Move

// NewMoveList returns a new moveList
func NewMoveList(cap int) MoveList {
	return make(MoveList, 0, cap)
}

// add adds a move to the moveList
func (ml *MoveList) add(move Move) {
	*ml = append(*ml, move)
}

// pickFirst returns the first move of the list
func (ml *MoveList) pickFirst() *Move {
	if len(*ml) == 0 {
		m := NoMove
		return &m
	}

	move := (*ml)[0]
	*ml = (*ml)[1:]
	return &move
}

// sort sorts the moveList by scores passed
func (ml *MoveList) sort(scores []int) {
	for i := range len(*ml) - 1 {
		for j := range len(*ml) - i - 1 {
			if scores[j] < scores[j+1] {
				(*ml)[j], (*ml)[j+1] = (*ml)[j+1], (*ml)[j]
				scores[j], scores[j+1] = scores[j+1], scores[j]
			}
		}
	}
}
