package main

import "fmt"

func main() {

  // Test direction function
  // bb1 := Bitboard(0b1 << 1)
  // bb2 := Bitboard(0b1)
  // bb1.Print()
  // bb2.Print()
  // fmt.Println(getDirection(bb2, bb1))

  // Check Nearest West!
  pos := EmptyPosition()
  pos.addPiece(BLACK_ROOK, "b4")
  pos.addPiece(BLACK_ROOK, "d4")
  pos.addPiece(BLACK_ROOK, "f4")

  piece, _ := pos.pieceAt("f4")

  fmt.Println("Attacks")
  piece.Attacks(pos).Print()


  // // Test new pos
  // pos := EmptyPosition()
  // pos.addPiece(WHITE_KING, "a1")
  // newPos := pos.RemovePiece(0b1)
  // newPos.bitboards[0].Print()
  //
  // fmt.Println(pos == &newPos)
  // pos.bitboards[0].Print()
  

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
