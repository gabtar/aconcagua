package aconcagua

import "testing"

// Castling tests

func TestCastleFromFen(t *testing.T) {
	var c castlingRights
	castlings := "KQqk"
	c.fromFen(castlings)

	expected := castlingRights(0b1111)
	got := c

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestCastleFromFen2(t *testing.T) {
	var c castlingRights
	castlings := "Qk"
	c.fromFen(castlings)

	expected := castlingRights(0b0110)
	got := c

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestEmptyCastle(t *testing.T) {
	var c castlingRights
	castlings := "-"
	c.fromFen(castlings)

	expected := castlingRights(0)
	got := c

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestAddLongWhiteCastle(t *testing.T) {
	var c castlingRights
	castlings := "-"

	c.fromFen(castlings)
	c.add(Q)

	expected := castlingRights(0b0100)
	got := c

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestWhiteCanCastleLong(t *testing.T) {
	var c castlingRights

	castlings := "Qkq"
	c.fromFen(castlings)

	expected := true
	got := c.canCastle(Q)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestWhiteCannotCastleLongIfBlocked(t *testing.T) {
	pos := From("r3k2r/8/8/8/8/8/3PP3/R1B1K2R w KQkq - 0 1")

	expected := false
	got := pos.canCastleLong(White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestWhiteCanCastleLong960(t *testing.T) {

}
