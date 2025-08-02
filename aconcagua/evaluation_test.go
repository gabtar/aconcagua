package aconcagua

import (
	"fmt"
	"testing"
)

func TestEval(t *testing.T) {
	pos := NewPositionFromFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	ev := Evaluate(pos)

	if ev != 0 {
		t.Errorf("Expected: %v, got: %v", 0, ev)
	}
}

func TestDoubledPawns(t *testing.T) {
	pos := NewPositionFromFen("2k5/8/8/8/1PP3P1/6P1/8/4K3 w - - 0 1")

	expected := 1
	got := doubledPawns(pos, White).count()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestTripledPawnsOnGFile(t *testing.T) {
	pos := NewPositionFromFen("2k5/6p1/6p1/6p1/1PP3P1/6P1/8/4K3 w - - 0 1")

	expected := 2
	got := doubledPawns(pos, Black).count()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}

func TestPawnIsIsolated(t *testing.T) {
	pos := NewPositionFromFen("2k5/8/6p1/3p4/2pP4/2P5/3K4/8 b - - 0 1")
	pawnBB := bitboardFromCoordinates("g6")

	expected := pawnBB
	got := isolatedPawns(pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsNotIsolated(t *testing.T) {
	pos := NewPositionFromFen("2k5/8/8/3p4/2pP4/2P5/3K4/8 b - - 0 1")

	expected := Bitboard(0)
	got := isolatedPawns(pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsBackward(t *testing.T) {
	pos := NewPositionFromFen("1k6/2p5/3p4/3P4/8/8/8/3K4 w - - 0 1")

	expected := true
	got := backwardPawns(pos, Black)&pos.Bitboards[BlackPawn] > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsNotBackward(t *testing.T) {
	pos := NewPositionFromFen("1k6/2p5/3p4/3PP3/8/8/8/3K4 w - - 0 1")

	expected := false
	got := backwardPawns(pos, White)&pos.Bitboards[WhitePawn] > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBackwardPawnsForWhite(t *testing.T) {
	pos := NewPositionFromFen("8/5p2/6p1/p1p3P1/P1P4P/1P6/8/8 w - - 0 1")

	expected := true
	got := backwardPawns(pos, White)&pos.Bitboards[WhitePawn] > 0

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

func TestPawnStructureEvaluation(t *testing.T) {
	pos := NewPositionFromFen("1k6/2p4P/2Pp4/1P3p2/1P3P2/8/8/3K4 w - - 0 1")
	// 1 doubled pawn b file
	// 2 isolated pawn g, and h files
	// 1 passed pawn on 7th rank h file
	// 1 backward pawns f file

	ev := Evaluation{}

	ev.evaluatePawnStructure(pos, White)

	fmt.Println(backwardPawns(pos, White).count())
	fmt.Println(isolatedPawns(pos, White).count())
	fmt.Println(passedPawns(pos, White).count())
	fmt.Println(doubledPawns(pos, White).count())

	expected := DoubledPawnPenaltyEg + 2*IsolatedPawnPenaltyEg + 1*BackwardPawnPenaltyEg + 100
	got := ev.egPawnStructure[White]

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
