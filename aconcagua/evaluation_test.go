package aconcagua

import "testing"

func TestEval(t *testing.T) {
	pos := NewPositionFromFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	ev := Evaluate(pos)

	if ev != 0 {
		t.Errorf("Expected: %v, got: %v", 0, ev)
	}
}
