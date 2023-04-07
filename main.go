package main

func main() {

  // fmt.Println("Ranks")
  // for _, rank := range ranks {
  //   fmt.Println("-------------------")
  //   rank.Print()
  // }
  for i := 56; i < 64; i++ {
    Bitboard(raysAttacks[NORTH][i]).Print()
  }
  

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
