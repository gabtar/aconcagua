package search

import (
	"testing"

	"github.com/gabtar/aconcagua/board"
)

func TestFoundCheckmateMove(t *testing.T) {
	pos := board.From("3r4/pR6/2p5/2kb1N2/8/2B5/qP3PPP/3R2K1 w - - 1 3")
	_, bestMoves := BestMove(pos, 4)

	expected := "c3d4"
	got := bestMoves.ToUci()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestFoundCaptureTheQueen(t *testing.T) {
	pos := board.From("r1bqkbnr/7p/2p1p1p1/p1pp1p1Q/P4P2/3PP3/1PPBN1PP/RN3RK1 b kq - 1 9")
	_, bestMoves := BestMove(pos, 4)
	expected := "g6h5"
	got := bestMoves.ToUci()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestDontSacTheQueen(t *testing.T) {
	pos := board.From("r1b2bnr/6k1/1qp1p1p1/p1pp1p1p/P1P2P2/3PPQ2/1P2N1PP/RNB2RK1 b - - 6 13")
	// FIX: not working!!!!!
	_, bestMoves := BestMove(pos, 4)

	expected := "d5c4"
	got := bestMoves.ToUci()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
