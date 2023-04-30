package main

import (
	"fmt"

	"github.com/gabtar/aconcagua/board"
)

func main() {
  initialPos := "r3r1k1/pp3pbp/1qp1b1p1/2B5/2BP4/Q1n2N2/P4PPP/3R1K1R w - - 4 18" // 41 legal moves for white
	pos := board.From(initialPos)
	pos.Print()

	legalMoves := pos.LegalMoves(board.WHITE)
	fmt.Println()
	fmt.Println("Moves for white: ", len(legalMoves))

  // Show all moves
	for _, move := range pos.LegalMoves(board.WHITE) {
		fmt.Println(move)
	}

}
