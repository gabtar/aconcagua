package aconcagua

// pvLine is a principal variation line of a position
type pvLine []Move

// NewPvLine returns a new principal variation line with the capacity passed allocated
func NewPvLine(capacity int) pvLine {
	return make(pvLine, 0, capacity)
}

// insert inserts a move at the beginning of the pvLine
func (pv *pvLine) insert(move Move, branchPv *pvLine) {
	*pv = append([]Move{move}, *branchPv...)
}

// reset resets the pvLine
func (pv *pvLine) reset() {
	*pv = (*pv)[:0]
}

// String returns the string representation of the principal variation
func (pv *pvLine) String() string {
	moves := ""
	for _, m := range *pv {
		moves += m.String() + " "
	}
	moves = moves[:len(moves)-1]
	return moves
}
