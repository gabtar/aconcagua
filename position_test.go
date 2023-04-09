package main

import (
	"testing"
)

// Position tests

func TestCheckingPieces(t *testing.T){
  pos := EmptyPosition()

  pos.addPiece(BLACK_KNIGHT, "f3")
  pos.addPiece(WHITE_KING, "e1")

  expected := pos.checkingPieces(WHITE)
  got, _ := pos.pieceAt("f3")

  included := false
  for _, piece := range expected {
    // TODO, need to check properly equality on interfaces!
    if piece.Attacks(pos) == got.Attacks(pos) {
      included = true
    }
  }

  if !included {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
