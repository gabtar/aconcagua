package evaluation

import (
	"testing"

	"github.com/gabtar/aconcagua/board"
)

// Evaluation function testings

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

	expected := 150
	got := Evaluate(*pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBlackPawnNearQueening(t *testing.T) {
	pos := board.EmptyPosition()
	pos.AddPiece(board.BLACK_PAWN, "a2") // Pawn on 2rank +50

	expected := -150 // Black is negative
	got := Evaluate(*pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

// TODO: more test with pieces square tables...
func TestWhiteFirstMoveE4(t *testing.T) {
	pos := board.From("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1")

	expected := 40
	got := Evaluate(*pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestEvaluationEqualWithE4D5(t *testing.T) {
	pos := board.From("rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1")

	expected := 0
	got := Evaluate(*pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
