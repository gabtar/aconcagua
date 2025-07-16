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

func TestWhiteCanCastleShort960(t *testing.T) {
	// starts from ches960 - 484 position
	// qbbrnknr/pppppppp/8/8/8/8/PPPPPPPP/QBBRNKNR w KQkq - 0 1
	pos := From("qb1rnrk1/ppp1n1pp/3pbp2/8/2PQP3/1P3N2/PB3PPP/1B1RNK1R w KQ - 1 8")
	pos.castling = *NewCastling(5, 7, 3)
	pos.castling.castlingRights = KQ

	expected := true
	got := pos.canCastleShort(White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBlackCanCastleShort960(t *testing.T) {
	// starts from ches960 - 484 position
	// qbbrnknr/pppppppp/8/8/8/8/PPPPPPPP/QBBRNKNR w KQkq - 0 1

	pos := From("qbbrnk1r/ppp1n1pp/3p1p2/8/3QP3/1P3N2/PBP2PPP/1B1RNK1R b KQkq - 1 6")
	pos.castling = *NewCastling(5, 7, 3)
	pos.castling.castlingRights = KQkq

	expected := true
	got := pos.canCastleShort(Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestWhiteCanCastleLong960(t *testing.T) {
	// starts from ches960 - 484 position
	// qbbrnknr/pppppppp/8/8/8/8/PPPPPPPP/QBBRNKNR w KQkq - 0 1
	pos := From("qb1rnrk1/ppp1n1pp/3pbp2/8/2PQP3/1P3N2/PB3PPP/1B1R1K1R w KQ - 1 8")
	pos.castling = *NewCastling(5, 7, 3)
	pos.castling.castlingRights = KQ

	expected := true
	got := pos.canCastleLong(White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestWhiteCannotCastleLong960IfSquaresPassingAreAttacked(t *testing.T) {
	// starts from ches960 - 484 position
	// qbbrnknr/pppppppp/8/8/8/8/PPPPPPPP/QBBRNKNR w KQkq - 0 1
	pos := From("qb1rnrk1/ppp1n1pp/3p1p2/8/2PQP3/1P3N2/PB1b1PPP/1B1R1K1R w KQ - 1 8")
	pos.castling = *NewCastling(5, 7, 3)
	pos.castling.castlingRights = KQ

	expected := false
	got := pos.canCastleLong(White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestWhiteCannotCastleShort960IfPathIsBlocked(t *testing.T) {
	pos := From("qbbrnk1r/ppp1n1pp/3p1p2/8/3QP3/1P3N2/PBP2PPP/1B1RNK1R w KQkq - 1 6")
	pos.castling = *NewCastling(5, 7, 3)
	pos.castling.castlingRights = KQkq

	expected := false
	got := pos.canCastleLong(White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestWhiteCannotCastleShort960IfKingInCheck(t *testing.T) {
	pos := From("qbbrnk1r/ppp1n1pp/3p1p2/1b6/3QP3/1P3N2/PBP2PPP/1B1RNK1R w KQkq - 1 6")
	pos.castling = *NewCastling(5, 7, 3)
	pos.castling.castlingRights = KQkq

	expected := false
	got := pos.canCastleShort(White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
