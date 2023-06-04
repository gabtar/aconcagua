package evaluation

import (
	"testing"

	"github.com/gabtar/aconcagua/board"
)

// Evaluation function testings

func TestEvaluationWithEqualMaterial(t *testing.T) {
	// pos := board.EmptyPosition()
	// pos.AddPiece(board.WHITE_PAWN, "e2")
	// pos.AddPiece(board.BLACK_PAWN, "e2")
	//
	// expected := 0
	// got := Evaluate(*pos)
	//
	// if got != expected {
	// 	t.Errorf("Expected: %v, got: %v", expected, got)
	// }
}

//
// func TestEvaluationWithWhiteAdvantage(t *testing.T) {
// 	pos := board.From("5rk1/5ppp/4p3/8/8/2B5/5PPP/5RK1 w - - 0 1")
//
// 	expected := 230 // 1 bishop(white) vs 1 pawn(black) (330 - 100)
// 	got := Evaluate(*pos)
//
// 	if got != expected {
// 		t.Errorf("Expected: %v, got: %v", expected, got)
// 	}
// }

func TestSingleKingsOnEndGameEvaluation(t *testing.T) {
	pos := board.EmptyPosition()
	pos.AddPiece(board.BLACK_KING, "d4") // King on d4 +40
	pos.AddPiece(board.WHITE_KING, "e1") // King on e1 -30

	expected := -70 // Black king is "better", because it's on the center
	got := Evaluate(*pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestWhitePawnNearQueening(t *testing.T) {
	pos := board.EmptyPosition()
	pos.AddPiece(board.WHITE_PAWN, "d7") // Pawn on 7rank +50

	whitePawnsBB := pos.Bitboards(board.WHITE)[5]

	expected := 150
	got := pawnScore(whitePawnsBB, 'w')

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBlackPawnNearQueening(t *testing.T) {
	pos := board.EmptyPosition()
	pos.AddPiece(board.BLACK_PAWN, "a2") // Pawn on 2rank +50

	whitePawnsBB := pos.Bitboards(board.BLACK)[5]

	expected := 150
	got := pawnScore(whitePawnsBB, 'b')

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
