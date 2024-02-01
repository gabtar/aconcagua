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

// add adds a new move to the principalVariation
func (pv *PrincipalVariation) add(m board.Move, d int) {
	pv.moves[pv.maxDepth-d] = m
}
