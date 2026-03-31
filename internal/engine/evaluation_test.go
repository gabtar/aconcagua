package engine

import (
	"fmt"
	"testing"
)

func TestEval(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	ev := NewEvaluation(DefaultPawnHashTableSizeInMb)

	if ev.Evaluate(pos) != TempoBonus {
		t.Errorf("Expected: %v, got: %v", 0, ev)
	}
}

func TestDoubledPawns(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("2k5/8/8/8/1PP3P1/6P1/8/4K3 w - - 0 1")

	expected := 1
	got := DoubledPawns(pos, White).count()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestTripledPawnsOnGFile(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("2k5/6p1/6p1/6p1/1PP3P1/6P1/8/4K3 w - - 0 1")

	expected := 2
	got := DoubledPawns(pos, Black).count()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}

func TestPawnIsIsolated(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("2k5/8/6p1/3p4/2pP4/2P5/3K4/8 b - - 0 1")
	pawnBB := bitboardFromCoordinates("g6")

	expected := pawnBB
	got := IsolatedPawns(pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsNotIsolated(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("2k5/8/8/3p4/2pP4/2P5/3K4/8 b - - 0 1")

	expected := Bitboard(0)
	got := IsolatedPawns(pos, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsBackward(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("1k6/2p5/3p4/3P4/8/8/8/3K4 w - - 0 1")

	expected := true
	got := BackwardPawns(pos.Bitboards[BlackPawn], pawnAttacks(&pos.Bitboards[WhitePawn], White), Black)&pos.Bitboards[BlackPawn] > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsNotBackward(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("1k6/2p5/3p4/3PP3/8/8/8/3K4 w - - 0 1")

	expected := false
	got := BackwardPawns(pos.Bitboards[WhitePawn], pawnAttacks(&pos.Bitboards[BlackPawn], Black), White)&pos.Bitboards[WhitePawn] > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBackwardPawnsForWhite(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("8/5p2/6p1/p1p3P1/P1P4P/1P6/8/8 w - - 0 1")

	expected := true
	got := BackwardPawns(pos.Bitboards[WhitePawn], pawnAttacks(&pos.Bitboards[BlackPawn], Black), White)&pos.Bitboards[WhitePawn] > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsNotPassed(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("1k6/2p5/3p4/3PP3/8/8/8/3K4 w - - 0 1")
	pawnBB := bitboardFromCoordinates("e5")

	expected := false
	got := PassedPawns(pos.Bitboards[WhitePawn], pos.Bitboards[BlackPawn], White)&pawnBB > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsPassed(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("1k6/2p5/3pP3/3P4/8/8/8/3K4 w - - 0 1")
	pawnBB := bitboardFromCoordinates("e6")

	expected := true
	got := PassedPawns(pos.Bitboards[WhitePawn], pos.Bitboards[BlackPawn], White)&pawnBB > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPassedPawnOn7thRank(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("1k6/2p4P/3p4/3P4/8/8/8/3K4 w - - 0 1")
	pawnBB := bitboardFromCoordinates("h7")

	expected := true
	got := PassedPawns(pos.Bitboards[WhitePawn], pos.Bitboards[BlackPawn], White)&pawnBB > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}

func TestBlackPawnIsNotPassed(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("1k6/2p5/3p4/3PP3/8/8/8/3K4 w - - 0 1")
	pawnBB := bitboardFromCoordinates("c7")

	expected := false
	got := PassedPawns(pos.Bitboards[BlackPawn], pos.Bitboards[WhitePawn], Black)&pawnBB > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnStructureEvaluation(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("1k6/2p4P/2Pp4/1P3p2/1P3P2/8/8/3K4 w - - 0 1")
	// 1 doubled pawn b file
	// 2 isolated pawn g, and h files
	// 1 passed pawn on 7th rank h file
	// 1 backward pawns f file

	ev := EvalVector{}
	ev.evaluatePawnStructure(pos, pawnAttacks(&pos.Bitboards[BlackPawn], Black), White)

	expected := DoubledPawnPenaltyEg + 2*IsolatedPawnPenaltyEg + 1*BackwardPawnPenaltyEg + PassedPawnsBonusEg[6]
	got := ev.egPawnStructure[White]

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestOutpostSquares(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("4r1k1/5ppp/p2p4/3Nb3/2P5/6P1/5P1P/4R1K1 w - - 0 1") // d5 is an outpost

	expected := bitboardFromIndex(d5)
	got := OutpostSquares(pos.Bitboards[WhitePawn], pos.Bitboards[BlackPawn], White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingSafety(t *testing.T) {
	pos := NewPosition()
	// Black king zone attacks:
	// A knight on g7 attacks a pawn on f5 and the e6 square in the black king zone
	// A Bishop on b1 attacks the pawn on f5 and the e4 square in the black king zone
	// The White Queen on h8, is not attacking. Just to met the safety condition (>= two attackers and one is a queen)
	// The black pawns defends g4 and e4 in the king zone
	// Safety = KnightAttackWeight + BishopAttackWeight - 2 * KingZoneDefenseBonus
	pos.LoadFromFenString("7Q/6N1/8/4kpp1/8/8/8/1B4K1 w - - 0 1")

	ev := NewEvaluation(DefaultPawnHashTableSizeInMb)
	ev.Evaluate(pos)

	got := ev.Eval.mgKingSafety[Black]
	expected := -(KnightAttackWeight*2 + BishopAttackWeight*2) + 2*KingZoneDefenseBonus

	fmt.Println(ev.Eval.kingAttacksWeight[White])

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
