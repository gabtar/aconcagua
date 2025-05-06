package aconcagua

import "testing"

func TestStartingPositionPolyGlotKey(t *testing.T) {
	pos := InitialPosition()
	expected := uint64(0x463b96181691fc9c)
	got := PolyGlotKeyFromPosition(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPosition2(t *testing.T) {
	// position after e2e4
	pos := From("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")
	expected := uint64(0x823c9b50fd114196)
	got := PolyGlotKeyFromPosition(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPosition3(t *testing.T) {
	// position after e2e4 d75
	pos := From("rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2")

	expected := uint64(0x0756b94461c50fb0)
	got := PolyGlotKeyFromPosition(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPosition4(t *testing.T) {
	// position after e2e4 d7d5 e4e5
	pos := From("rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR b KQkq - 0 2")

	expected := uint64(0x662fafb965db29d4)
	got := PolyGlotKeyFromPosition(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPosition5(t *testing.T) {
	// position after e2e4 d7d5 e4e5 f7f5
	pos := From("rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3")

	expected := uint64(0x22a48b5a8e47ff78)
	got := PolyGlotKeyFromPosition(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPosition6(t *testing.T) {
	// position after e2e4 d7d5 e4e5 f7f5 e1e2
	pos := From("rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPPKPPP/RNBQ1BNR b kq - 0 3")

	expected := uint64(0x652a607ca3f242c1)
	got := PolyGlotKeyFromPosition(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPosition7(t *testing.T) {
	// position after e2e4 d7d5 e4e5 f7f5 e1e2 e8f7
	pos := From("rnbq1bnr/ppp1pkpp/8/3pPp2/8/8/PPPPKPPP/RNBQ1BNR w - - 0 4")

	expected := uint64(0x00fdd303c946bdd9)
	got := PolyGlotKeyFromPosition(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPosition8(t *testing.T) {
	// position after a2a4 b7b5 h2h4 b5b4 c2c4
	pos := From("rnbqkbnr/p1pppppp/8/8/PpP4P/8/1P1PPPP1/RNBQKBNR b KQkq c3 0 3")

	expected := uint64(0x3c8123ea7b067637)
	got := PolyGlotKeyFromPosition(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPosition9(t *testing.T) {
	// position after a2a4 b7b5 h2h4 b5b4 c2c4 b4c3 a1a3
	pos := From("rnbqkbnr/p1pppppp/8/8/P6P/R1p5/1P1PPPP1/1NBQKBNR b Kkq - 0 4")

	expected := uint64(0x5c3f9b829b279560)
	got := PolyGlotKeyFromPosition(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPosition10(t *testing.T) {
	pos := From("3R2rR/5k2/r2b1p2/p3p3/1p1p4/3P2B1/1P4K1/8 w - - 6 50")

	expected := uint64(0xe43c91de37bce2b4)
	got := PolyGlotKeyFromPosition(pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
