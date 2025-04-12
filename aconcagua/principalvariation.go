package aconcagua

// PV stores the Principal Variation line during a search
type PV []Move

// newPV returns a pointer to a new PV
func newPV() *PV {
	return &PV{}
}

// insert adds a move at the start of the principal variation
func (pv *PV) insert(move Move, branchPv *PV) {
	*pv = append([]Move{move}, *branchPv...)
}

// moveAt returns a chessMove and a boolean if its a move at the ply passed in the principal variation
func (pv *PV) moveAt(ply int) (Move, bool) {
	if len(*pv) > ply {
		return (*pv)[ply], true
	}
	return Move(0), false
}

// String returns the string representation of the principal variation
func (pv *PV) String() string {
	moves := ""
	for _, m := range *pv {
		moves += m.String() + " "
	}
	moves = moves[:len(moves)-1]
	return moves
}
