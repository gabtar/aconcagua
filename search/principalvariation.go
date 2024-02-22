package search

import "github.com/gabtar/aconcagua/board"

// PrincipalVariation stores the current best line when searching for best moves
type PrincipalVariation []board.Move

// newPrincipalVariation is a factory that returns a pointer to a principalVariation struct
func newPrincipalVariation() *PrincipalVariation {
	return &PrincipalVariation{}
}

// insert adds a move at the begginning of the principal variation
func (pv *PrincipalVariation) insert(move board.Move, branchPv *PrincipalVariation) {
	*pv = append([]board.Move{move}, *branchPv...)
}

func (pv *PrincipalVariation) moveAt(ply int) (board.Move, bool) {
	if len(*pv) > ply {
		return (*pv)[ply], true
	}
	return board.Move(0), false
}

// clear resets the principal variation moves
func (pv *PrincipalVariation) clear() {
	*pv = (*pv)[:0]
}

// String returns the string representation of the principal variation moves
func (pv *PrincipalVariation) String() string {
	list := ""
	for _, m := range *pv {
		list += m.ToUci() + " "
	}
	return list
}
