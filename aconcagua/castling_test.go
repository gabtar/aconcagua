package aconcagua

import "testing"

// Castling tests

func TestCastleFromFen(t *testing.T) {
	var c castling
	castlings := "KQqk"
	c.fromFen(castlings)

	expected := castling(0b1111)
	got := c

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestCastleFromFen2(t *testing.T) {
	var c castling
	castlings := "Qk"
	c.fromFen(castlings)

	expected := castling(0b0110)
	got := c

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestEmptyCastle(t *testing.T) {
	var c castling
	castlings := "-"
	c.fromFen(castlings)

	expected := castling(0)
	got := c

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestAddLongWhiteCastle(t *testing.T) {
	var c castling
	castlings := "-"

	c.fromFen(castlings)
	c.add(Q)

	expected := castling(0b0100)
	got := c

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestWhiteCanCastleLong(t *testing.T) {
	var c castling

	castlings := "Qkq"
	c.fromFen(castlings)

	expected := true
	got := c.canCastle(Q)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
