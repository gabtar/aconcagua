package board

import "strconv"

// Perft returns all the legal moves up to the passed depth
func (pos *Position) Perft(depth int) (nodes uint64) {
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

// The Divide command is often implemented as a variation of Perft, listing all moves and for each move, the perft of the decremented depth. However, some programs already give "divided" output for Perft. Below is output of Stockfish when computing perft 5 for start position:
// go perft 5
// a2a3: 181046
// b2b3: 215255
// c2c3: 222861
// d2d3: 328511
// e2e3: 402988
// f2f3: 178889
// g2g3: 217210
// h2h3: 181044
// a2a4: 217832
// b2b4: 216145
// c2c4: 240082
// d2d4: 361790
// e2e4: 405385
// f2f4: 198473
// g2g4: 214048
// h2h4: 218829
// b1a3: 198572
// b1c3: 234656
// g1f3: 233491
// g1h3: 198502
//
// Nodes searched: 4865609

// Divide, a variation of Perft, returns the perft of all moves in the current position
func (pos *Position) Divide(depth int) (divide string) {
	var totalNodes uint64 = 0

	for _, m := range pos.LegalMoves(pos.ToMove()) {
		pos.MakeMove(&m)
		nodes := pos.Perft(depth - 1)
		divide += m.ToUci() + ": " + strconv.FormatUint(nodes, 10) + ","
		pos.UnmakeMove(m)
		totalNodes += nodes
	}

	divide += "Nodes searched: " + strconv.FormatUint(totalNodes, 10)

	return
}
