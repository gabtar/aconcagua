package aconcagua

import "strconv"

// Perft returns all the legal moves up to the passed depth
func (pos *Position) Perft(depth int) (nodes uint64) {
	if depth == 0 {
		return 1
	}

	if depth == 1 {
		return uint64(pos.LegalMoves().length)
	}

	moveList := pos.LegalMoves()

	for i := 0; i < moveList.length; i++ {
		pos.MakeMove(&moveList.moves[i])
		nodes += pos.Perft(depth - 1)
		pos.UnmakeMove(&moveList.moves[i])
	}

	return nodes
}

// Divide a variation of Perft, returns the perft of all moves in the current position
func (pos *Position) Divide(depth int) (divide string) {
	var totalNodes uint64 = 0

	moveList := pos.LegalMoves()

	for i := 0; i < moveList.length; i++ {
		pos.MakeMove(&moveList.moves[i])
		nodes := pos.Perft(depth - 1)
		divide += moveList.moves[i].String() + " " + strconv.FormatUint(nodes, 10) + ","
		pos.UnmakeMove(&moveList.moves[i])
		totalNodes += nodes
	}

	divide += "\n" + strconv.FormatUint(totalNodes, 10)

	return
}
