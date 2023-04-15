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

//
// func TestRays(t *testing.T) {
//   for i, v := range raysAttacks[NORTHWEST] {
//     if (i < 48 || i >= 56) {
//       continue
//     }
//     v.Print()
//   }
//
// }
