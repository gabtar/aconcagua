package search

import "github.com/gabtar/aconcagua/board"

// PrincipalVariation stores the current best line when searching for best moves
type PrincipalVariation struct {
	maxDepth int
	moves    []board.Move
}

// newPrincipalVariation is a factory that returns a pointer to a principalVariation struct
func newPrincipalVariation(depth int) *PrincipalVariation {
	return &PrincipalVariation{
		maxDepth: depth,
		moves:    make([]board.Move, depth),
	}
}

// add adds a new move to the principalVariation at the specific depth
func (pv *PrincipalVariation) add(m board.Move, depth int) {
	pv.moves[pv.maxDepth-depth] = m
}

// clear resets the principal variation moves
func (pv *PrincipalVariation) clear() {
	pv.moves = pv.moves[:0]
}

// String returns the string representation of the principal variation moves
func (pv *PrincipalVariation) String() string {
	list := ""
	for _, m := range pv.moves[:pv.maxDepth] {
		list += m.ToUci() + " "
	}
	return list
}
