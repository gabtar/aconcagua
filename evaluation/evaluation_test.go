package evaluation

import (
	"testing"

	"github.com/gabtar/aconcagua/board"
)

// Evaluation function testings

func TestEvaluationWithEqualMaterial(t *testing.T) {
	pos := board.EmptyPosition()
	pos.AddPiece(board.WHITE_PAWN, "e2")
	pos.AddPiece(board.BLACK_PAWN, "e2")

	expected := 0.0
	got := Evaluate(*pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestEvaluationWithWhiteAdvantage(t *testing.T) {
	pos := board.From("5rk1/5ppp/4p3/8/8/2B5/5PPP/5RK1 w - - 0 1")

	expected := 230.0 // 1 bishop(white) vs 1 pawn(black) (330 - 100)
	got := Evaluate(*pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
