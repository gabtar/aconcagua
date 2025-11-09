package engine

import "strconv"

// Perft returns all the legal moves up to the passed depth
func (pos *Position) Perft(depth int) (nodes uint64) {
	if depth == 1 {
		ml := NewMoveList()
		pd := pos.generatePositionData()
		pos.generateCaptures(ml, &pd)
		pos.generateNonCaptures(ml, &pd)
		return uint64(ml.length)
	}

	ml := NewMoveList()
	pd := pos.generatePositionData()
	pos.generateCaptures(ml, &pd)
	pos.generateNonCaptures(ml, &pd)

	for i := range ml.length {
		pos.MakeMove(&ml.moves[i])
		nodes += pos.Perft(depth - 1)
		pos.UnmakeMove(&ml.moves[i])
	}

	return nodes
}

// Divide a variation of Perft, returns the perft of all moves in the current position
func (pos *Position) Divide(depth int) (divide string) {
	var totalNodes uint64 = 0

	ml := NewMoveList()
	pd := pos.generatePositionData()
	pos.generateCaptures(ml, &pd)
	pos.generateNonCaptures(ml, &pd)

	for i := range ml.length {
		pos.MakeMove(&ml.moves[i])
		nodes := pos.Perft(depth - 1)
		divide += ml.moves[i].String() + " " + strconv.FormatUint(nodes, 10) + ","
		pos.UnmakeMove(&ml.moves[i])
		totalNodes += nodes
	}

	divide += "\n" + strconv.FormatUint(totalNodes, 10)

	return
}
