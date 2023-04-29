package main

import (
	"fmt"

	"github.com/gabtar/aconcagua/board"
)

func main() {

	// Tests for fen string construction
	initialPos := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
  // initialPos = "r6r/1b2k1bq/8/8/7B/8/8/R3K2R b KQ - 3 2"
  // initialPos = "7k/8/8/8/pPp5/8/8/7K b - b3 0 1"
	pos := board.From(initialPos)
	pos.Print()

	fmt.Println()
	fmt.Println("Initial moves for white: ", len(pos.LegalMoves(board.BLACK)))

}
