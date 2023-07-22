package board

import "testing"

// Tests that diferent moves produces the correct update on the board

func TestNormalMove(t *testing.T) {
	pos := InitialPosition()

	from := Bsf(squareToBitboard([]string{"e2"}))
	to := Bsf(squareToBitboard([]string{"e4"}))
	move := MoveEncode(from, to, int(WhitePawn), 0, NORMAL, 0, 0)

	newPos := pos.MakeMove(&move)

	expected := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1"
	got := newPos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestCastleMoveUpdate(t *testing.T) {
	pos := From("5rk1/5ppp/8/8/8/1bB5/4PPPP/4K2R w K - 0 1")

	from := Bsf(squareToBitboard([]string{"e1"}))
	to := Bsf(squareToBitboard([]string{"g1"}))
	move := MoveEncode(from, to, int(WhiteKing), 0, CASTLE, 0, 0)

	newPos := pos.MakeMove(&move)

	expected := "5rk1/5ppp/8/8/8/1bB5/4PPPP/5RK1 b - - 1 1"
	got := newPos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestEnPassantCaptureUpdate(t *testing.T) {
	pos := From("5rk1/5ppp/5b2/8/1pP5/3N4/5PPP/5RK1 b - c3 0 1")

	from := Bsf(squareToBitboard([]string{"b4"}))
	to := Bsf(squareToBitboard([]string{"c3"}))
	move := MoveEncode(from, to, int(BlackPawn), 0, EN_PASSANT, 0, 0)

	newPos := pos.MakeMove(&move)

	expected := "5rk1/5ppp/5b2/8/8/2pN4/5PPP/5RK1 w - - 0 2"
	got := newPos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookCapture(t *testing.T) {
	pos := From("2r2rk1/5pbp/6p1/8/1pP5/1N5P/5PP1/1R3RK1 b - c3 0 1")

	from := Bsf(squareToBitboard([]string{"c8"}))
	to := Bsf(squareToBitboard([]string{"c4"}))
	move := MoveEncode(from, to, int(BlackRook), 0, CAPTURE, 0, 0)

	newPos := pos.MakeMove(&move)

	expected := "5rk1/5pbp/6p1/8/1pr5/1N5P/5PP1/1R3RK1 w - - 0 2"
	got := newPos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestDoublePawnPush(t *testing.T) {
	pos := From("5rk1/5p1p/6p1/2N5/1pP5/2b4P/5PP1/1R3RK1 w - - 1 3")

	from := Bsf(squareToBitboard([]string{"f2"}))
	to := Bsf(squareToBitboard([]string{"f4"}))
	move := MoveEncode(from, to, int(WhitePawn), 0, PAWN_DOUBLE_PUSH, 0, 0)

	newPos := pos.MakeMove(&move)

	expected := "5rk1/5p1p/6p1/2N5/1pP2P2/2b4P/6P1/1R3RK1 b - f3 0 3"
	got := newPos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnPromotion(t *testing.T) {
	pos := From("1r3n1k/5Ppp/8/8/R7/6PP/1p5K/8 b - - 0 1")

	from := Bsf(squareToBitboard([]string{"b2"}))
	to := Bsf(squareToBitboard([]string{"b1"}))
	move := MoveEncode(from, to, int(BlackPawn), int(BlackQueen), PROMOTION, 0, 0)

	newPos := pos.MakeMove(&move)

	expected := "1r3n1k/5Ppp/8/8/R7/6PP/7K/1q6 w - - 0 2"
	got := newPos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveEncode(t *testing.T) {
	move := MoveEncode(63, 9, int(BlackBishop), 0, NORMAL, 0, 0)

	expected := NORMAL
	got := move.MoveType()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}
