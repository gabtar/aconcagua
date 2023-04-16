package board

import (
	"testing"
)

// Position tests

func TestCheckingPieces(t *testing.T) {
	pos := EmptyPosition()

	pos.AddPiece(BLACK_KNIGHT, "f3")
	pos.AddPiece(WHITE_KING, "e1")

	expected := pos.CheckingPieces(WHITE)
	got, _ := pos.PieceAt("f3")

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

func TestGetDirection(t *testing.T) {
  pos := EmptyPosition()
  pos.AddPiece(BLACK_KING, "e1")
  pos.AddPiece(BLACK_ROOK, "e8")
  king, _ := pos.PieceAt("e1")
  rook, _ := pos.PieceAt("e8")

	expected := NORTH
	got := getDirection(rook.Square(), king.Square()) // rook is in NORTH of the king

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
