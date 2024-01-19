package board

import "strconv"

// Perft returns all the legal moves up to the passed depth
func (pos *Position) Perft(depth int) (nodes uint64) {
	if depth == 0 {
		return 1
	}

	if depth == 1 {
		return uint64(len(pos.LegalMoves(pos.ToMove())))
	}

	moves := pos.LegalMoves(pos.ToMove())

	for _, move := range moves {
		pos.MakeMove(&move)
		nodes += pos.Perft(depth - 1)
		pos.UnmakeMove(move)
	}

	return nodes
}

// Divide, a variation of Perft, returns the perft of all moves in the current position
func (pos *Position) Divide(depth int) (divide string) {
	var totalNodes uint64 = 0

	for _, m := range pos.LegalMoves(pos.ToMove()) {
		pos.MakeMove(&m)
		nodes := pos.Perft(depth - 1)
		divide += m.ToUci() + " " + strconv.FormatUint(nodes, 10) + ","
		pos.UnmakeMove(m)
		totalNodes += nodes
	}

	divide += "\n" + strconv.FormatUint(totalNodes, 10)

	return
}
