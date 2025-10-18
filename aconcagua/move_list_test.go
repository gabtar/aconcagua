package aconcagua

import (
	"testing"
)

func TestListOfMovesAddMove(t *testing.T) {
	lm := MoveList{}

	move := encodeMove(0, 8, quiet)
	lm.add(*move)

	expected := 1
	got := lm.length

	if expected != got {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestListOfMovesScoreCaptures(t *testing.T) {
	lm := MoveList{}

	pos := NewPosition()
	pos.LoadFromFenString("4k3/8/3r1p2/4P3/4K3/8/8/8 w - - 0 1")
	m1 := encodeMove(36, 43, capture) // Pawn takes rook. Best move
	m2 := encodeMove(36, 45, capture) // Pawn takes pawn
	lm.add(*m1)
	lm.add(*m2)

	score1 := pos.see(m1.from(), m1.to())
	score2 := pos.see(m2.from(), m2.to())

	lm.scoreCaptures(pos)

	if lm.scores[0] != score1 || lm.scores[1] != score2 {
		t.Errorf("Expected: %v, got: %v", []int{score1, score2}, lm.scores)
	}
}

func TestListOfMovesScoreNonCaptures(t *testing.T) {
	lm := MoveList{}

	pos := NewPosition()
	pos.LoadFromFenString("8/6k1/2P2pp1/5P2/5K2/2P5/8/8 w - - 0 1")
	m1 := encodeMove(37, 46, capture) // Capture, non quiet
	m2 := encodeMove(18, 26, quiet)   // Pawn advance to 4 rank
	m3 := encodeMove(42, 50, quiet)   // Pawn advance to 7 rank

	lm.add(*m1)
	lm.add(*m2)
	lm.add(*m3)

	hm := HistoryMovesTable{}
	hm[White][18][26] = 10
	hm[White][42][50] = 100 // Pawn advance to 7 rank is historically better

	lm.scoreNonCaptures(&hm, White, 1) // starts from non capture moves

	if lm.scores[1] != 10 || lm.scores[2] != 100 {
		t.Errorf("Expected: %v, got: %v", []int{0, 10, 100}, lm.scores)
	}
}

func TestListOfMovesGetBestMoveIndex(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("4k3/8/3p1q2/4P3/4K3/8/8/8 w - - 0 1")

	m1 := encodeMove(36, 43, capture) // Pawn takes pawn
	m2 := encodeMove(36, 45, capture) // Pawn takes queen. Best Move!
	lm := MoveList{}
	lm.add(*m1)
	lm.add(*m2)

	lm.scoreCaptures(pos)

	expected := 1
	got := lm.getBestIndex(0)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
