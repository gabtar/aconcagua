package aconcagua

import "strconv"

// Perft returns all the legal moves up to the passed depth
func (pos *Position) Perft(depth int) (nodes uint64) {
	if depth == 1 {
		ml := NewMoveList(255)
		pos.generateCaptures(&ml)
		pos.generateNonCaptures(&ml)
		return uint64(len(ml))
	}

	ml := NewMoveList(255)
	pos.generateCaptures(&ml)
	pos.generateNonCaptures(&ml)

	for i := range ml {
		pos.MakeMove(&ml[i])
		nodes += pos.Perft(depth - 1)
		pos.UnmakeMove(&ml[i])
	}

	return nodes
}

// Divide a variation of Perft, returns the perft of all moves in the current position
func (pos *Position) Divide(depth int) (divide string) {
	var totalNodes uint64 = 0

	ml := NewMoveList(255)
	pos.generateCaptures(&ml)
	pos.generateNonCaptures(&ml)

	for i := range ml {
		pos.MakeMove(&ml[i])
		nodes := pos.Perft(depth - 1)
		divide += ml[i].String() + " " + strconv.FormatUint(nodes, 10) + ","
		pos.UnmakeMove(&ml[i])
		totalNodes += nodes
	}

	divide += "\n" + strconv.FormatUint(totalNodes, 10)

	return
}
