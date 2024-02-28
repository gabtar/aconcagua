package board

import "testing"

// Tests that diferent moves produces the correct update on the board

func TestNormalMove(t *testing.T) {
	pos := InitialPosition()

	from := Bsf(bitboardFromCoordinate("e2"))
	to := Bsf(bitboardFromCoordinate("e4"))
	move := newMove().
		setFromSq(from).
		setToSq(to).
		setPiece(WhitePawn).
		setMoveType(Normal)

	pos.MakeMove(move)

	expected := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1"
	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestCastleMoveUpdate(t *testing.T) {
	pos := From("5rk1/5ppp/8/8/8/1bB5/4PPPP/4K2R w K - 0 1")

	from := Bsf(bitboardFromCoordinate("e1"))
	to := Bsf(bitboardFromCoordinate("g1"))
	move := newMove().
		setFromSq(from).
		setToSq(to).
		setPiece(WhiteKing).
		setMoveType(Castle)

	pos.MakeMove(move)

	expected := "5rk1/5ppp/8/8/8/1bB5/4PPPP/5RK1 b - - 1 1"
	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestEnPassantCaptureUpdate(t *testing.T) {
	pos := From("5rk1/5ppp/5b2/8/1pP5/3N4/5PPP/5RK1 b - c3 0 1")

	from := Bsf(bitboardFromCoordinate("b4"))
	to := Bsf(bitboardFromCoordinate("c3"))
	move := newMove().
		setFromSq(from).
		setToSq(to).
		setPiece(BlackPawn).
		setMoveType(EnPassant)

	pos.MakeMove(move)

	expected := "5rk1/5ppp/5b2/8/2P5/2pN4/5PPP/5RK1 w - - 0 2"
	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookCapture(t *testing.T) {
	pos := From("2r2rk1/5pbp/6p1/8/1pP5/1N5P/5PP1/1R3RK1 b - c3 0 1")

	from := Bsf(bitboardFromCoordinate("c8"))
	to := Bsf(bitboardFromCoordinate("c4"))
	move := newMove().
		setFromSq(from).
		setToSq(to).
		setPiece(BlackRook).
		setMoveType(Capture)

	pos.MakeMove(move)

	expected := "5rk1/5pbp/6p1/8/1pr5/1N5P/5PP1/1R3RK1 w - - 0 2"
	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestDoublePawnPush(t *testing.T) {
	pos := From("5rk1/5p1p/6p1/2N5/1pP5/2b4P/5PP1/1R3RK1 w - - 1 3")

	from := Bsf(bitboardFromCoordinate("f2"))
	to := Bsf(bitboardFromCoordinate("f4"))
	move := newMove().
		setFromSq(from).
		setToSq(to).
		setPiece(WhitePawn).
		setMoveType(PawnDoublePush)

	pos.MakeMove(move)

	expected := "5rk1/5p1p/6p1/2N5/1pP2P2/2b4P/6P1/1R3RK1 b - f3 0 3"
	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnPromotion(t *testing.T) {
	pos := From("1r3n1k/5Ppp/8/8/R7/6PP/1p5K/8 b - - 0 1")

	from := Bsf(bitboardFromCoordinate("b2"))
	to := Bsf(bitboardFromCoordinate("b1"))
	move := newMove().
		setFromSq(from).
		setToSq(to).
		setPiece(BlackPawn).
		setMoveType(Promotion).
		setPromotedTo(BlackQueen)

	pos.MakeMove(move)

	expected := "1r3n1k/5Ppp/8/8/R7/6PP/7K/1q6 w - - 0 2"
	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBuildMove(t *testing.T) {
	move := newMove().
		setFromSq(63).
		setToSq(9).
		setPiece(BlackBishop).
		setMoveType(Normal)

	expected := Normal
	got := move.MoveType()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}

func TestMoveBuilder(t *testing.T) {
	move := newMove().
		setFromSq(0).
		setToSq(8).
		setPiece(WhitePawn).
		setMoveType(Normal).
		setEpTargetBefore(bitboardFromIndex(56)).
		setRule50Before(10).
		setCastleRightsBefore(castling(0b1111)).
		SetScore(100)

	expectedFrom := 0
	gotFrom := move.from()

	if gotFrom != expectedFrom {
		t.Errorf("Expected: %v, got: %v", expectedFrom, gotFrom)
	}

	expectedTo := 8
	gotTo := move.To()

	if gotTo != expectedTo {
		t.Errorf("Expected: %v, got: %v", expectedTo, gotTo)
	}

	expectedPiece := int(WhitePawn)
	gotPiece := move.Piece()

	if gotPiece != expectedPiece {
		t.Errorf("Expected: %v, got: %v", expectedPiece, gotPiece)
	}

	expectedType := Normal
	gotType := move.MoveType()

	if gotType != expectedType {
		t.Errorf("Expected: %v, got: %v", expectedPiece, gotPiece)
	}

	expectedEpBefore := bitboardFromIndex(56)
	gotEpBefore := move.epTargetBefore()

	if gotEpBefore != expectedEpBefore {
		t.Errorf("Expected: %v, got: %v", expectedEpBefore, gotEpBefore)
	}

	expectedRule50Before := 10
	gotRule50Before := move.rule50Before()

	if gotRule50Before != expectedRule50Before {
		t.Errorf("Expected: %v, got: %v", expectedEpBefore, gotEpBefore)
	}

	expectedCastlingBefore := castling(0b1111)
	gotRuleCastlingBefore := move.castleRightsBefore()

	if gotRuleCastlingBefore != expectedCastlingBefore {
		t.Errorf("Expected: %v, got: %v", expectedEpBefore, gotEpBefore)
	}

	expectedScore := 100
	gotScore := move.Score()

	if gotScore != expectedScore {
		t.Errorf("Expected: %v, got: %v", expectedScore, gotScore)
	}
}
