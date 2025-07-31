package aconcagua

import "testing"

func TestEval(t *testing.T) {
	pos := NewPositionFromFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	ev := Evaluate(pos)

	if ev != 0 {
		t.Errorf("Expected: %v, got: %v", 0, ev)
	}
}

func TestPawnIsDoubled(t *testing.T) {
	pos := NewPositionFromFen("2k5/8/8/8/1PP3P1/6P1/8/4K3 w - - 0 1")
	pawnBB := bitboardFromCoordinates("g3")

	expected := true
	got := isDoubled(&pawnBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsNotDoubled(t *testing.T) {
	pos := NewPositionFromFen("2k5/8/8/8/1PP3P1/6P1/8/4K3 w - - 0 1")
	pawnBB := bitboardFromCoordinates("c4")

	expected := false
	got := isDoubled(&pawnBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsIsolated(t *testing.T) {
	pos := NewPositionFromFen("2k5/8/6p1/3p4/2pP4/2P5/3K4/8 b - - 0 1")
	pawnBB := bitboardFromCoordinates("g6")

	expected := true
	got := isIsolated(&pawnBB, pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsNotIsolated(t *testing.T) {
	pos := NewPositionFromFen("2k5/8/6p1/3p4/2pP4/2P5/3K4/8 b - - 0 1")
	pawnBB := bitboardFromCoordinates("c3")

	expected := false
	got := isIsolated(&pawnBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsBackward(t *testing.T) {
	pos := NewPositionFromFen("1k6/2p5/3p4/3P4/8/8/8/3K4 w - - 0 1")

	expected := true
	got := backwardsPawns(pos, Black)&pos.Bitboards[BlackPawn] > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsNotBackward(t *testing.T) {
	pos := NewPositionFromFen("1k6/2p5/3p4/3PP3/8/8/8/3K4 w - - 0 1")

	expected := false
	got := backwardsPawns(pos, White)&pos.Bitboards[WhitePawn] > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBackwardPawnsForWhite(t *testing.T) {
	pos := NewPositionFromFen("8/5p2/6p1/p1p3P1/P1P4P/1P6/8/8 w - - 0 1")

	expected := true
	got := backwardsPawns(pos, White)&pos.Bitboards[WhitePawn] > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsNotPassed(t *testing.T) {
	pos := NewPositionFromFen("1k6/2p5/3p4/3PP3/8/8/8/3K4 w - - 0 1")
	pawnBB := bitboardFromCoordinates("e5")

	expected := false
	got := passedPawns(pos, White)&pawnBB > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsPassed(t *testing.T) {
	pos := NewPositionFromFen("1k6/2p5/3pP3/3P4/8/8/8/3K4 w - - 0 1")
	pawnBB := bitboardFromCoordinates("e6")

	expected := true
	got := passedPawns(pos, White)&pawnBB > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPassedPawnOn7thRank(t *testing.T) {
	pos := NewPositionFromFen("1k6/2p4P/3p4/3P4/8/8/8/3K4 w - - 0 1")
	pawnBB := bitboardFromCoordinates("h7")

	expected := true
	got := passedPawns(pos, White)&pawnBB > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}

func TestBlackPawnIsNotPassed(t *testing.T) {
	pos := NewPositionFromFen("1k6/2p5/3p4/3PP3/8/8/8/3K4 w - - 0 1")
	pawnBB := bitboardFromCoordinates("c7")

	expected := false
	got := passedPawns(pos, Black)&pawnBB > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
