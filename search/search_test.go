package search

import (
	"testing"

	"github.com/gabtar/aconcagua/board"
)

func TestFoundCheckmateMove(t *testing.T) {
	pos := board.From("3r4/pR6/2p5/2kb1N2/8/2B5/qP3PPP/3R2K1 w - - 1 3")
	_, bestMoves := BestMove(pos, 4)
	expected := "c3d4"
	got := bestMoves[0].ToUci()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}
