package aconcagua

// PVLine is a principal variation line of a position
type PVLine struct {
	moves  []Move
	length int
}

// reset resets the PVLine length
func (pv *PVLine) reset() {
	pv.length = 0
}

// prepend prepends a move to the PVLine
func (pv *PVLine) prepend(move Move, branchPV *PVLine) {
	pv.moves[0] = move
	if branchPV.length > 0 {
		copy(pv.moves[1:], branchPV.moves[:branchPV.length])
	}
	pv.length = branchPV.length + 1
}

// String returns a string representation of the PVLine
func (pv *PVLine) String() string {
	if pv.length == 0 {
		return ""
	}

	moves := ""
	for i := range pv.length {
		moves += pv.moves[i].String() + " "
	}
	moves = moves[:len(moves)-1]
	return moves
}

// PVTable is a principal variation table for holding PVLines during the search
type PVTable []PVLine

// reset resets the pv line at the given plies
func (pvTable PVTable) reset(plies int) {
	pvTable[plies].reset()
}

// NewPVTable returns a new PVTable
func NewPVTable(depth int) PVTable {
	table := make(PVTable, depth)
	for i := range table {
		table[i] = PVLine{moves: make([]Move, depth), length: 0}
	}
	return table
}
