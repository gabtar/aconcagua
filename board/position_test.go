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

func TestGetDirectionNorth(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_KING, "e1")
	pos.AddPiece(BLACK_ROOK, "e8")
	king, _ := pos.PieceAt("e1")
	rook, _ := pos.PieceAt("e8")

	expected := NORTH
	got := getDirection(rook.Square(), king.Square()) // king -> rook == NORTH

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGetDirectionSouth(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_KING, "e1")
	pos.AddPiece(BLACK_ROOK, "e8")
	king, _ := pos.PieceAt("e1")
	rook, _ := pos.PieceAt("e8")

	expected := SOUTH
	got := getDirection(king.Square(), rook.Square()) // rook -> king == SOUTH

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGetDirectionSouthWest(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_KING, "e4")
	pos.AddPiece(BLACK_ROOK, "d3")
	king, _ := pos.PieceAt("e4")
	rook, _ := pos.PieceAt("d3")

	expected := SOUTHWEST
	got := getDirection(rook.Square(), king.Square()) // king -> rook == SOUTHWEST

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGetRayPath(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_ROOK, "c4")
	pos.AddPiece(WHITE_ROOK, "f4")
	black_rook, _ := pos.PieceAt("c4")
	white_rook, _ := pos.PieceAt("f4")

	expectedSquares := []string{"d4", "e4"}

	expected := sqaureToBitboard(expectedSquares)
	got := getRayPath(black_rook.Square(), white_rook.Square())

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}

func TestPinnedPiece(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BLACK_KING, "c7")
	pos.AddPiece(BLACK_ROOK, "c6")
	pos.AddPiece(WHITE_ROOK, "c1")
	blackRook, _ := pos.PieceAt("c6")

	expected := true
	got := isPinned(blackRook.Square(), blackRook.Color(), pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}
