package search

import "github.com/gabtar/aconcagua/board"

// PrincipalVariation stores the current best line when searching for best moves
type PrincipalVariation struct {
	maxDepth int
	moves    []board.Move
}

// place sets a move in the principal variation at the specified depth
func (pV *PrincipalVariation) place(depth int, move board.Move) {
	pV.moves[depth] = move
}

// newPrincipalVariation is a factory that returns a pointer to a principalVariation struct
func newPrincipalVariation(depth int) *PrincipalVariation {
	return &PrincipalVariation{
		maxDepth: depth,
		moves:    make([]board.Move, depth),
	}
}
