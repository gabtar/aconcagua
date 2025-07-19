package aconcagua

import "testing"

func TestAttackers(t *testing.T) {
	// White Knight(e4) attacks Black Pawn on d6
	// White Bishop(a3) attacks Black Pawn on d6
	pos := NewPositionFromFen("6k1/2p5/3p4/3P4/4N3/B7/8/6K1 w - - 0 1")

	expected := 2
	got := pos.attackers(43, White, ^pos.EmptySquares()).count()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGetLeastValuableAttacker(t *testing.T) {
	// White Knight(e4) attacks Black Pawn on d6
	// White Bishop(a3) attacks Black Pawn on d6
	pos := NewPositionFromFen("6k1/2p5/3p4/3P4/4N3/B7/8/6K1 w - - 0 1")
	attackersOnD6 := pos.attackers(43, White, ^pos.EmptySquares())

	expectedBB, expectedPiece := bitboardFromCoordinates("e4"), WhiteKnight
	gotBB, gotPiece := pos.getLeastValuableAttacker(attackersOnD6, White)

	if gotBB != expectedBB {
		t.Errorf("Expected: %v, got: %v", expectedBB, gotBB)
	}
	if gotPiece != expectedPiece {
		t.Errorf("Expected: %v, got: %v", expectedPiece, gotPiece)
	}
}

func TestSee(t *testing.T) {
	pos := NewPositionFromFen("6k1/2p5/3p4/3P4/4N3/B7/8/6K1 w - - 0 1") // Nxd6

	expected := -100
	got := pos.see(28, 43)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestSeeTwo(t *testing.T) {
	pos := NewPositionFromFen("1k1r4/1pp4p/p7/4p3/8/P5P1/1PP4P/2K1R3 w - - 0 1") // Rxe5

	expected := 100
	got := pos.see(4, 36)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestSeeThree(t *testing.T) {
	pos := NewPositionFromFen("1b4k1/2p5/3p4/3P4/4N3/B7/8/6K1 w - - 0 1") // Nxd6 now the bishop on b8 attacks by xrays the d6 square

	expected := -200
	got := pos.see(28, 43)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestSeeFour(t *testing.T) {
	pos := NewPositionFromFen("1k1r3q/1ppn3p/p4b2/4p3/8/P2N2P1/1PP1R1BP/2K1Q3 w - - 0 1")

	expected := -200
	got := pos.see(19, 36)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestSeeFive(t *testing.T) {
	// From mediocre chess guide: https://mediocrechess.blogspot.com/2007/03/guide-static-exchange-evaluation-see.html
	pos := NewPositionFromFen("7k/p7/1p6/8/8/1Q6/8/7K w - - 0 1")

	expected := -800
	got := pos.see(17, 41)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
