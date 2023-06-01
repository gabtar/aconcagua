package board

import "testing"

// Tests that diferent moves produces the correct update on the board

func TestNormalMove(t *testing.T) {
	pos := InitialPosition()
	move := &Move{from: "e2", to: "e4", piece: WHITE_PAWN, moveType: NORMAL}
	newPos := pos.MakeMove(move)

	expected := "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1"
	got := newPos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestCastleMoveUpdate(t *testing.T) {
	pos := From("5rk1/5ppp/8/8/8/1bB5/4PPPP/4K2R w K - 0 1")
	move := &Move{from: "e1", to: "g1", piece: WHITE_KING, moveType: CASTLE}
	newPos := pos.MakeMove(move)

	expected := "5rk1/5ppp/8/8/8/1bB5/4PPPP/5RK1 b - - 1 1"
	got := newPos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestEnPassantCaptureUpdate(t *testing.T) {
	pos := From("5rk1/5ppp/5b2/8/1pP5/3N4/5PPP/5RK1 b - c3 0 1")
	move := &Move{from: "b4", to: "c3", piece: BLACK_PAWN, moveType: EN_PASSANT}
	newPos := pos.MakeMove(move)

	expected := "5rk1/5ppp/5b2/8/8/2pN4/5PPP/5RK1 w - - 0 2"
	got := newPos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookCapture(t *testing.T) {
	pos := From("2r2rk1/5pbp/6p1/8/1pP5/1N5P/5PP1/1R3RK1 b - c3 0 1")
	move := &Move{from: "c8", to: "c4", piece: BLACK_ROOK, moveType: CAPTURE}
	newPos := pos.MakeMove(move)

	expected := "5rk1/5pbp/6p1/8/1pr5/1N5P/5PP1/1R3RK1 w - - 0 2"
	got := newPos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestDoublePawnPush(t *testing.T) {
	pos := From("5rk1/5p1p/6p1/2N5/1pP5/2b4P/5PP1/1R3RK1 w - - 1 3")
	move := &Move{from: "f2", to: "f4", piece: WHITE_PAWN, moveType: PAWN_DOUBLE_PUSH}
	newPos := pos.MakeMove(move)

	expected := "5rk1/5p1p/6p1/2N5/1pP2P2/2b4P/6P1/1R3RK1 b - f3 0 3"
	got := newPos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnPromotion(t *testing.T) {
	pos := From("1r3n1k/5Ppp/8/8/R7/6PP/1p5K/8 b - - 0 1")
	move := &Move{from: "b2", to: "b1", piece: BLACK_PAWN, moveType: PROMOTION, promotedTo: BLACK_QUEEN}
	newPos := pos.MakeMove(move)

	expected := "1r3n1k/5Ppp/8/8/R7/6PP/7K/1q6 w - - 0 2"
	got := newPos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
