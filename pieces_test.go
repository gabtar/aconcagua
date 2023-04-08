package main

import "testing"

// King moves tests

// TODO, tests more corner cases
func TestKingAttacks(t *testing.T){
  pos := InitialPosition()
  king, _ := pos.pieceAt("e1")

  expected := Bitboard(0b11100000101000)
  got := king.attacks(pos)

  if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingMovesToEmptySquares(t *testing.T) {
  pos := EmptyPosition()
  pos.addPiece(WHITE_KING, "e4")
  king, _ := pos.pieceAt("e4")

  expectedSquares := []string{"d3", "d4", "d5", "e3", "e5", "f3", "f4", "f5"}

  expected := sqaureToBitboard(expectedSquares)
  got := king.moves(pos)

  if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCannotMoveToAttackedSquare(t *testing.T) {
  pos := EmptyPosition()
  pos.addPiece(WHITE_KING, "e4")
  pos.addPiece(BLACK_KNIGHT, "c6")
  king, _ := pos.pieceAt("e4")

  // Cannot move to d4 or e5 because it's attacked by the black knight
  expectedSquares := []string{"d3", "d5", "e3", "f3", "f4", "f5"}

  expected := sqaureToBitboard(expectedSquares)
  got := king.moves(pos)

  if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

// Rook moves tests
func TestRookAttacksOnEmptyBoard(t *testing.T) {
  pos := EmptyPosition()
  pos.addPiece(BLACK_ROOK, "e4")
  rook, _ := pos.pieceAt("e4")

  expectedSquares := []string{"e1", "e2", "e3", "e5", "e6", "e7", "e8",
                              "a4", "b4", "c4", "d4", "f4", "g4", "h4"}

  expected := sqaureToBitboard(expectedSquares)
  got := rook.attacks(pos)

  if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookAttacksWithBlockedSquares(t *testing.T) {
  pos := EmptyPosition()
  pos.addPiece(BLACK_ROOK, "e4")
  pos.addPiece(WHITE_KNIGHT, "c4") // Knight blocking on c4
  rook, _ := pos.pieceAt("e4")

  expectedSquares := []string{"e1", "e2", "e3", "e5", "e6", "e7", "e8",
                              "c4", "d4", "f4", "g4", "h4"}

  expected := sqaureToBitboard(expectedSquares)
  got := rook.attacks(pos)

  if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}


func TestRookAttacksWithAllSquaresBlocked(t *testing.T) {
  pos := EmptyPosition()
  pos.addPiece(BLACK_ROOK, "b3")
  pos.addPiece(WHITE_KNIGHT, "b4") // Knight blocking on b4
  pos.addPiece(WHITE_KNIGHT, "b2") // Knight blocking on b5
  pos.addPiece(WHITE_KNIGHT, "a3") // Knight blocking on a3
  pos.addPiece(WHITE_KNIGHT, "c3") // Knight blocking on c3
  rook, _ := pos.pieceAt("b3")

  expectedSquares := []string{"b4", "b2", "a3", "c3"}

  expected := sqaureToBitboard(expectedSquares)
  got := rook.attacks(pos)

  if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
