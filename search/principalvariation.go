package search

import "github.com/gabtar/aconcagua/board"

// principalVariation stores the current best line when searching for best moves
type principalVariation struct {
	maxDepth int
	moves    []board.Move
}

// place sets a move in the principal variation at the specified depth
func (pV *principalVariation) place(depth int, move board.Move) {
	pV.moves[depth] = move
}

// newPrincipalVariation is a factory that returns a pointer to a principalVariation struct
func newPrincipalVariation(depth int) *principalVariation {
	return &principalVariation{
		maxDepth: depth,
		moves:    make([]board.Move, depth),
	}
}
