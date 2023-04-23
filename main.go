package main

import (
	"fmt"

	"github.com/gabtar/aconcagua/board"
)

func main() {

	// Test new pos
	pos := board.EmptyPosition()
	pos.AddPiece(board.WHITE_KING, "a1")
	newPos := pos.RemovePiece(0b1)

	fmt.Println(pos == &newPos)

	// pos := InitialPosition()
	// pos.addPiece(WHITE_PAWN, "f3")
	//
	// fmt.Println("Occupied squares")
	// fmt.Println("------------------")
	// occupied := ^pos.emptySquares()
	// occupied.Print()
	//
	// piece, err := pos.pieceAt("g1")
	// if err != nil {
	//   fmt.Println(err)
	//   return
	// }
	//
	// fmt.Println("Attacks of the Knight at g1")
	// fmt.Println("------------------")
	// piece.attacks(pos).Print()
	//
	// fmt.Println("Moves of the Knight at g1(f3 blocked by own pawn)")
	// fmt.Println("------------------")
	// piece.moves(pos).Print()
	//
	// fmt.Println("Attacked squares by white")
	// fmt.Println("------------------")
	// att := pos.attackedSquares(WHITE)
	// att.Print()
}
