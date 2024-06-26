package aconcagua

import (
	"testing"
)

// Evaluation function testings

func TestSingleKingsOnEndGameEvaluation(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackKing, "d4") // King on d4 +27
	pos.AddPiece(WhiteKing, "e1") // King on e1 -28

	expected := -55 // Black king is "better", because it's on the center
	got := Eval(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestWhitePawnNearQueening(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(WhitePawn, "d7") // Pawn on 7rank +134 (endgame) + 94 pawn value

	expected := 134 + 94
	got := Eval(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBlackPawnNearQueening(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackPawn, "a2") // Pawn on 2rank +178 (a7 from white's view)

	expected := -(178 + 94) // Negative because is white to move
	got := Eval(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestWhiteFirstMoveE4(t *testing.T) {
	pos := From("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1")

	expected := +17 - (-15) // Pawn on e2(-15 penalty) and on e4 (+17) (middlegame table)
	got := Eval(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestEvaluationEqualWithE4D5(t *testing.T) {
	pos := From("rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1")

	expected := -3
	got := Eval(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
