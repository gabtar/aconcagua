package aconcagua

import (
	"testing"
)

func TestCheckingPieces(t *testing.T) {
	pos := EmptyPosition()

	pos.AddPiece(BlackKnight, "f3")
	pos.AddPiece(WhiteKing, "e1")

	expected := 1
	checkingPieces, _ := pos.CheckingPieces(White)
	got := checkingPieces.count()

	if expected != got {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGetRayPath(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackRook, "c4")
	pos.AddPiece(WhiteRook, "f4")
	from := bitboardFromCoordinates("c4")
	to := bitboardFromCoordinates("f4")

	expected := bitboardFromCoordinates("d4", "e4")
	got := getRayPath(&from, &to)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPinnedPiece(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackKing, "c7")
	pos.AddPiece(BlackRook, "c6")
	pos.AddPiece(WhiteRook, "c1")
	from := bitboardFromCoordinates("c6")

	expected := true
	got := from&pos.pinnedPieces(Black) > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPinnedPieceKnightFail(t *testing.T) {
	pos := From("rnQq1k1r/pp2bppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R b KQ - 1 8")
	from := bitboardFromCoordinates("b8")

	expected := false
	got := from&pos.pinnedPieces(Black) > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnAPositionWithPromotion(t *testing.T) {
	pos := From("3r2k1/5ppp/8/8/8/8/pp4PP/5R1K b - - 0 1")

	expected := 28
	got := pos.LegalMoves().length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnAPositionIllegalLongCastle(t *testing.T) {
	pos := From("6k1/5ppp/8/7q/7b/8/5PPP/RN2K2R w KQ - 1 1")

	expected := 18
	got := pos.LegalMoves().length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnAPositionWithDoubleEnPassantCaptures(t *testing.T) {
	pos := From("6k1/5bpp/8/1PpPN3/8/8/6PP/6K1 w - c6 0 1")

	expected := 19
	got := pos.LegalMoves().length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnMultiplePinsWithCheck(t *testing.T) {
	pos := From("8/1k3Rpp/1n6/3b4/8/5B2/6PP/1R4K1 b - - 0 1")

	expected := 5
	got := pos.LegalMoves().length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnMultiplePinsWithCheckTwo(t *testing.T) {
	pos := From("8/1k3Rpp/1n6/3b4/8/5B2/6PP/2R3K1 b - - 0 1")

	expected := 4 // 3 of king 1 block of the knight
	got := pos.LegalMoves().length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestFenSerializationFromPosition(t *testing.T) {
	pos := InitialPosition()

	expected := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBlackIsInCheckmate(t *testing.T) {

	pos := From("4R2k/r5pp/8/8/8/8/PPP5/1K6 b - - 0 1")

	expected := true
	got := pos.Checkmate(Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBlackIsNotInCheckmate(t *testing.T) {

	pos := From("4R2k/r5pp/8/8/8/1b6/PPP5/1K6 b - - 0 1")

	expected := false
	got := pos.Checkmate(Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBlackIsNotInCheckmate2(t *testing.T) {

	pos := From("4R2k/6pp/8/1b6/8/8/PPP5/1K6 b - - 0 1")

	expected := false
	got := pos.Checkmate(Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestCaptureUpdatesPosition(t *testing.T) {
	pos := From("7k/8/8/8/3p4/4P3/8/7K w - - 0 1")

	from := Bsf(bitboardFromCoordinates("e3"))
	to := Bsf(bitboardFromCoordinates("d4"))
	move := encodeMove(uint16(from), uint16(to), capture)

	pos.MakeMove(move)

	expected := "7k/8/8/8/3P4/8/8/7K b - - 0 1"
	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

// Test for unmake move
func TestUnmakeInNormalMove(t *testing.T) {
	pos := InitialPosition()

	from := Bsf(bitboardFromCoordinates("g1"))
	to := Bsf(bitboardFromCoordinates("f3"))

	move := encodeMove(uint16(from), uint16(to), quiet)

	expected := pos.Hash

	// Make and restore
	pos.MakeMove(move)
	pos.UnmakeMove(move)

	got := pos.Hash

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestUnmakeMoveInDoublePawnPush(t *testing.T) {
	pos := InitialPosition()

	from := Bsf(bitboardFromCoordinates("e2"))
	to := Bsf(bitboardFromCoordinates("e4"))

	move := encodeMove(uint16(from), uint16(to), doublePawnPush)

	expected := pos.ToFen() // Orifinal position

	pos.MakeMove(move)
	pos.UnmakeMove(move)

	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestUnmakeMoveQuietMove(t *testing.T) {
	pos := From("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1") // Position. 1. e4 (black to move)

	from := Bsf(bitboardFromCoordinates("e7"))
	to := Bsf(bitboardFromCoordinates("e5"))

	move := encodeMove(uint16(from), uint16(to), quiet)

	expected := pos.ToFen()

	pos.MakeMove(move)
	pos.UnmakeMove(move)

	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestUnmakeCapture(t *testing.T) {
	pos := From("6k1/6pp/1r3p2/8/4n3/4B1P1/5P1P/6K1 w - - 0 1")

	from := Bsf(bitboardFromCoordinates("e3"))
	to := Bsf(bitboardFromCoordinates("b6"))

	move := encodeMove(uint16(from), uint16(to), capture)

	expected := pos.ToFen()

	pos.MakeMove(move)
	pos.UnmakeMove(move)

	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestUnmakeCaptureThatChangesCastleRights(t *testing.T) {
	pos := From("6k1/1b4pp/5p2/8/8/4B1P1/5P1P/4K2R b K - 1 1")

	from := Bsf(bitboardFromCoordinates("b7"))
	to := Bsf(bitboardFromCoordinates("h1"))

	move := encodeMove(uint16(from), uint16(to), capture)

	expected := pos.ToFen()

	pos.MakeMove(move)
	pos.UnmakeMove(move)

	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestUnmakePromotionRestoresThePawnTo7thRank(t *testing.T) {
	pos := From("1kq5/ppr1P3/2p5/8/8/8/5PPP/4R1K1 w - - 1 1")

	from := Bsf(bitboardFromCoordinates("e7"))
	to := Bsf(bitboardFromCoordinates("e8"))

	move := encodeMove(uint16(from), uint16(to), queenPromotion)

	expected := pos.ToFen()

	pos.MakeMove(move)
	pos.UnmakeMove(move)

	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestUnmakeCastleForWhite(t *testing.T) {
	pos := From("5rk1/pbpq1ppp/1pnp1n2/4p2P/4P1P1/2NP1PN1/PPPQ4/R3K2R w KQ - 0 1")

	from := Bsf(bitboardFromCoordinates("e1"))
	to := Bsf(bitboardFromCoordinates("c1"))

	move := encodeMove(uint16(from), uint16(to), queensideCastle)
	expected := pos.ToFen()

	pos.MakeMove(move)
	pos.UnmakeMove(move)

	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestUnmakeCastleForBlack(t *testing.T) {
	pos := From("r3k3/pbpq1ppp/1pnp1n2/4p2P/4P1P1/2NP1PN1/PPPQ4/2KR3R b q - 0 1")

	from := Bsf(bitboardFromCoordinates("e8"))
	to := Bsf(bitboardFromCoordinates("c8"))

	move := encodeMove(uint16(from), uint16(to), queensideCastle)
	expected := pos.ToFen()

	pos.MakeMove(move)
	pos.UnmakeMove(move)

	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestUnmakeEnPassantCaptureForBlack(t *testing.T) {
	pos := From("5rk1/1q3ppp/4p3/3pN3/1Pp5/5Q2/5PPP/5RK1 b - b3 0 1")

	from := Bsf(bitboardFromCoordinates("c4"))
	to := Bsf(bitboardFromCoordinates("b3"))

	move := encodeMove(uint16(from), uint16(to), epCapture)
	expected := pos.ToFen()

	pos.MakeMove(move)
	pos.UnmakeMove(move)

	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestUnmakeEnPassantCaptureForWhite(t *testing.T) {
	pos := From("5rk1/pp3ppp/4pn2/2pP4/8/2P3P1/PP3PBP/4R1K1 w - c6 0 1")

	from := Bsf(bitboardFromCoordinates("d5"))
	to := Bsf(bitboardFromCoordinates("c6"))

	move := encodeMove(uint16(from), uint16(to), epCapture)
	expected := pos.ToFen()

	pos.MakeMove(move)
	pos.UnmakeMove(move)

	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMakeMoveWithPromotionCapture(t *testing.T) {
	pos := From("8/8/2kq1N2/2pp4/1p6/4R3/p4PPP/1N4K1 b - - 0 1")

	from := Bsf(bitboardFromCoordinates("a2"))
	to := Bsf(bitboardFromCoordinates("b1"))

	move := encodeMove(uint16(from), uint16(to), queenCapturePromotion)

	expected := "8/8/2kq1N2/2pp4/1p6/4R3/5PPP/1q4K1 w - - 0 2"

	pos.MakeMove(move)

	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestUnmakeMoveWithEpCapture(t *testing.T) {
	pos := From("5rk1/pp3ppp/4pn2/2pP4/8/2P3P1/PP3PBP/4R1K1 w - c6 0 1")

	from := Bsf(bitboardFromCoordinates("d5"))
	to := Bsf(bitboardFromCoordinates("c6"))

	move := encodeMove(uint16(from), uint16(to), epCapture)
	pos.MakeMove(move)

	pos.UnmakeMove(move)

	expected := "5rk1/pp3ppp/4pn2/2pP4/8/2P3P1/PP3PBP/4R1K1 w - c6 0 1"
	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}

func TestUnmakeMoveWithKinigtMatePromotion(t *testing.T) {
	pos := From("8/p7/1pkb4/2p5/8/6PP/5pNK/6BQ b - - 50 1")

	from := Bsf(bitboardFromCoordinates("f2"))
	to := Bsf(bitboardFromCoordinates("f1"))

	move := encodeMove(uint16(from), uint16(to), knightPromotion)

	pos.MakeMove(move)
	pos.UnmakeMove(move)

	expected := "8/p7/1pkb4/2p5/8/6PP/5pNK/6BQ b - - 50 1"
	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}
