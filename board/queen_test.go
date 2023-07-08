package board

// import "testing"

// Queen moves tests
// TODO: Not necessary now, not queenAttacksMethod...
// func TestQueenAttacksOnEmptyBoard(t *testing.T) {
// 	pos := EmptyPosition()
// 	pos.AddPiece(WHITE_QUEEN, "e4")
// 	queen, _ := pos.PieceAt("e4")
//
// 	expectedSquares := []string{"e1", "e2", "e3", "e5", "e6", "e7", "e8",
// 		"a4", "b4", "c4", "d4", "f4", "g4", "h4", "h1", "g2", "f3", "d5",
// 		"c6", "b7", "a8", "b1", "c2", "d3", "f5", "g6", "h7"}
//
// 	expected := squareToBitboard(expectedSquares)
// 	got := queen.Attacks(pos)
//
// 	if got != expected {
// 		t.Errorf("Expected: %v, got: %v", expected, got)
// 	}
// }

// func TestQueenAttacksWithBlockedSquares(t *testing.T) {
// 	pos := EmptyPosition()
// 	pos.AddPiece(BLACK_QUEEN, "d5")
// 	pos.AddPiece(WHITE_KNIGHT, "e6")
// 	pos.AddPiece(WHITE_BISHOP, "e4")
// 	pos.AddPiece(WHITE_BISHOP, "d4")
// 	pos.AddPiece(WHITE_BISHOP, "c4")
// 	pos.AddPiece(WHITE_ROOK, "c5")
// 	pos.AddPiece(WHITE_ROOK, "c6")
// 	queen, _ := pos.PieceAt("d5")
//
// 	expectedSquares := []string{"d8", "d7", "d6", "c6", "e6", "c5", "e5", "f5", "g5", "h5", "c4", "d4", "e4"}
//
// 	expected := squareToBitboard(expectedSquares)
// 	got := queen.Attacks(pos)
//
// 	if got != expected {
// 		t.Errorf("Expected: %v, got: %v", expected, got)
// 	}
// }

// func TestQueenMovesWithCaptures(t *testing.T) {
// 	pos := EmptyPosition()
// 	pos.AddPiece(WHITE_QUEEN, "h1")
// 	pos.AddPiece(BLACK_BISHOP, "h4") // Can move(capture) black bishop on h4
// 	pos.AddPiece(BLACK_KNIGHT, "e1") // Can move(capture) black knight on e1
// 	pos.AddPiece(BLACK_ROOK, "e4")   // Can move(capture) black queen on e4
// 	queen, _ := pos.PieceAt("h1")
// 	queenBB := queen.Square()
//
// 	expectedSquares := []string{"h2", "h3", "h4", "g1", "f1", "e1", "g2", "f3", "e4"}
//
// 	expected := squareToBitboard(expectedSquares)
// 	got := queen.Moves(pos)
//
// 	if got != expected {
// 		t.Errorf("Expected: %v, got: %v", expected, got)
// 	}
// }
//
// func TestQueenMovesWithBlockingPieces(t *testing.T) {
// 	pos := EmptyPosition()
// 	pos.AddPiece(WHITE_QUEEN, "b7")
// 	pos.AddPiece(WHITE_KNIGHT, "a8")
// 	pos.AddPiece(WHITE_KNIGHT, "b8")
// 	pos.AddPiece(WHITE_KNIGHT, "c8")
// 	pos.AddPiece(WHITE_ROOK, "a7")
// 	pos.AddPiece(WHITE_BISHOP, "a6")
// 	pos.AddPiece(WHITE_BISHOP, "b6")
// 	pos.AddPiece(WHITE_BISHOP, "e7")
// 	pos.AddPiece(WHITE_ROOK, "d5")
//
// 	queen, _ := pos.PieceAt("b7")
//
// 	// Only legal moves (not blocked by white pieces)
// 	expectedSquares := []string{"c6", "c7", "d7"}
//
// 	expected := squareToBitboard(expectedSquares)
// 	got := queen.Moves(pos)
//
// 	if got != expected {
// 		t.Errorf("Expected: %v, got: %v", expected, got)
// 	}
// }
//
// func TestQueenMoveWhenCanBlockCheck(t *testing.T) {
// 	pos := EmptyPosition()
// 	pos.AddPiece(BLACK_QUEEN, "g8")
// 	pos.AddPiece(WHITE_ROOK, "d8")
// 	pos.AddPiece(BLACK_KING, "d1")
// 	queen, _ := pos.PieceAt("g8")
//
// 	// Black king is in check, so only legal moves are Qxd1(capture the rook) or Qd5(block)
// 	expectedSquares := []string{"d8", "d5"}
//
// 	expected := squareToBitboard(expectedSquares)
// 	got := queen.Moves(pos)
//
// 	if got != expected {
// 		t.Errorf("Expected: %v, got: %v", expected, got)
// 	}
// }
//
// func TestQueenMovesWhenKingInDoubleCheck(t *testing.T) {
// 	pos := EmptyPosition()
// 	pos.AddPiece(BLACK_QUEEN, "c6")
// 	pos.AddPiece(WHITE_ROOK, "e1") // Gives check to the black king on e8
// 	pos.AddPiece(WHITE_ROOK, "a8") // Gives check to the black king on e8
// 	pos.AddPiece(BLACK_KING, "e8")
// 	rook, _ := pos.PieceAt("c6")
//
// 	expectedSquares := []string{} // The queen cannot move at all because of the double check in own king
//
// 	expected := squareToBitboard(expectedSquares)
// 	got := rook.Moves(pos)
//
// 	if got != expected {
// 		t.Errorf("Expected: %v, got: %v", expected, got)
// 	}
// }
//
// func TestQueenMovesWhenKingInDoubleCheckTest2(t *testing.T) {
// 	pos := EmptyPosition()
// 	pos.AddPiece(BLACK_QUEEN, "c6")
// 	pos.AddPiece(WHITE_ROOK, "d8") // Gives check to the black king on e8
// 	pos.AddPiece(WHITE_ROOK, "a8") // Gives check to the black king on e8(by xrays)
// 	pos.AddPiece(BLACK_KING, "e8")
// 	queen, _ := pos.PieceAt("c6")
//
// 	expectedSquares := []string{} // The queen cannot move at all because of the double check in own king
//
// 	expected := squareToBitboard(expectedSquares)
// 	got := queen.Moves(pos)
//
// 	if got != expected {
// 		t.Errorf("Expected: %v, got: %v", expected, got)
// 	}
// }
//
// func TestQueenMovesWhenTheQueenIsPinned(t *testing.T) {
// 	pos := EmptyPosition()
// 	pos.AddPiece(BLACK_KING, "e8")
// 	pos.AddPiece(BLACK_ROOK, "e4")
// 	pos.AddPiece(WHITE_ROOK, "e2")
// 	pos.AddPiece(WHITE_ROOK, "e3")
//
// 	rook, _ := pos.PieceAt("e4")
// 	// The rook can only move along the e file, becuase it's pinned if moves
// 	// along the 4 rank, the king will be in check!
// 	expectedSquares := []string{"e3", "e5", "e6", "e7"}
//
// 	expected := squareToBitboard(expectedSquares)
// 	got := rook.Moves(pos)
//
// 	if got != expected {
// 		t.Errorf("Expected: %v, got: %v", expected, got)
// 	}
//
// }
