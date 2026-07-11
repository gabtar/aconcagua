package engine

import (
	"strings"
	"testing"
)

func TestCheckingPieces(t *testing.T) {
	pos := NewPosition()

	pos.AddPiece(BlackKnight, f3)
	pos.AddPiece(WhiteKing, e1)

	expected := 1
	checkingPieces, _ := pos.CheckingPieces(White)
	got := checkingPieces.count()

	if expected != got {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGetRayPath(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(BlackRook, c4)
	pos.AddPiece(WhiteRook, f4)
	from := bitboardFromCoordinates("c4")
	to := bitboardFromCoordinates("f4")

	expected := bitboardFromCoordinates("d4", "e4")
	got := getRayPath(&from, &to)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPinnedPiece(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(BlackKing, c7)
	pos.AddPiece(BlackRook, c6)
	pos.AddPiece(WhiteRook, c1)
	from := bitboardFromCoordinates("c6")

	expected := true
	got := from&pos.pinnedPieces(Black) > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPinnedPieceKnightFail(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("rnQq1k1r/pp2bppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R b KQ - 1 8")
	from := bitboardFromCoordinates("b8")

	expected := false
	got := from&pos.pinnedPieces(Black) > 0

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnAPositionWithPromotion(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("3r2k1/5ppp/8/8/8/8/pp4PP/5R1K b - - 0 1")

	ml := NewMoveList()
	pd := pos.generatePositionData()
	pos.generateNoisy(ml, &pd)
	pos.generateQuiets(ml, &pd)

	expected := 28
	got := ml.length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnAPositionIllegalLongCastle(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("6k1/5ppp/8/7q/7b/8/5PPP/RN2K2R w KQ - 1 1")

	ml := NewMoveList()
	pd := pos.generatePositionData()
	pos.generateNoisy(ml, &pd)
	pos.generateQuiets(ml, &pd)

	expected := 18
	got := ml.length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnAPositionWithDoubleEnPassantCaptures(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("6k1/5bpp/8/1PpPN3/8/8/6PP/6K1 w - c6 0 1")

	ml := NewMoveList()
	pd := pos.generatePositionData()
	pos.generateNoisy(ml, &pd)
	pos.generateQuiets(ml, &pd)

	expected := 19
	got := ml.length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnMultiplePinsWithCheck(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("8/1k3Rpp/1n6/3b4/8/5B2/6PP/1R4K1 b - - 0 1")

	ml := NewMoveList()
	pd := pos.generatePositionData()
	pos.generateNoisy(ml, &pd)
	pos.generateQuiets(ml, &pd)

	expected := 5
	got := ml.length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnMultiplePinsWithCheckTwo(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("8/1k3Rpp/1n6/3b4/8/5B2/6PP/2R3K1 b - - 0 1")

	ml := NewMoveList()
	pd := pos.generatePositionData()
	pos.generateNoisy(ml, &pd)
	pos.generateQuiets(ml, &pd)

	expected := 4 // 3 of king 1 block of the knight
	got := ml.length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestFenSerializationFromPosition(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	expected := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestCaptureUpdatesPosition(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("7k/8/8/8/3p4/4P3/8/7K w - - 0 1")

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
	pos := NewPosition()

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
	pos := NewPosition()

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
	pos := NewPosition()
	pos.LoadFromFenString("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1") // Position. 1. e4 (black to move)

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
	pos := NewPosition()
	pos.LoadFromFenString("6k1/6pp/1r3p2/8/4n3/4B1P1/5P1P/6K1 w - - 0 1")

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
	pos := NewPosition()
	pos.LoadFromFenString("6k1/1b4pp/5p2/8/8/4B1P1/5P1P/4K2R b K - 1 1")

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
	pos := NewPosition()
	pos.LoadFromFenString("1kq5/ppr1P3/2p5/8/8/8/5PPP/4R1K1 w - - 1 1")

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
	pos := NewPosition()
	pos.LoadFromFenString("5rk1/pbpq1ppp/1pnp1n2/4p2P/4P1P1/2NP1PN1/PPPQ4/R3K2R w KQ - 0 1")

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
	pos := NewPosition()
	pos.LoadFromFenString("r3k3/pbpq1ppp/1pnp1n2/4p2P/4P1P1/2NP1PN1/PPPQ4/2KR3R b q - 0 1")

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
	pos := NewPosition()
	pos.LoadFromFenString("5rk1/1q3ppp/4p3/3pN3/1Pp5/5Q2/5PPP/5RK1 b - b3 0 1")

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
	pos := NewPosition()
	pos.LoadFromFenString("5rk1/pp3ppp/4pn2/2pP4/8/2P3P1/PP3PBP/4R1K1 w - c6 0 1")

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
	pos := NewPosition()
	pos.LoadFromFenString("8/8/2kq1N2/2pp4/1p6/4R3/p4PPP/1N4K1 b - - 0 1")

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
	pos := NewPosition()
	pos.LoadFromFenString("5rk1/pp3ppp/4pn2/2pP4/8/2P3P1/PP3PBP/4R1K1 w - c6 0 1")

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
	pos := NewPosition()
	pos.LoadFromFenString("8/p7/1pkb4/2p5/8/6PP/5pNK/6BQ b - - 50 1")

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

func TestInsuficientMaterial(t *testing.T) {
	testCases := []struct {
		name                string
		fen                 string
		insuficientMaterial bool
	}{
		{"pawn and king vs king", "8/1k6/8/8/3P4/3K4/8/8 w - - 0 1", false},
		{"bishop and king vs bishop and king", "8/1k6/1b6/8/8/2BK4/8/8 w - - 0 1", false},
		{"queen and king vs bishop and king", "8/1kb5/8/8/3Q4/3K4/8/8 w - - 0 1", false},
		{"lone king vs lone king", "8/1k6/8/8/8/3K4/8/8 w - - 0 1", true},
		{"2 knights and king vs 1 knight and king", "8/1kn5/8/8/3NN3/3K4/8/8 w - - 0 1", false},
		{"knight and king vs king", "8/1k6/8/4N3/4K3/8/8/8 w - - 0 1", true},
		{"bishop and king vs king", "8/1kb5/8/8/4K3/8/8/8 w - - 0 1", true},
	}

	for _, tc := range testCases {
		pos := NewPosition()
		t.Run(tc.name, func(t *testing.T) {
			pos.LoadFromFenString(tc.fen)
			got := pos.insuficientMaterial()
			if got != tc.insuficientMaterial {
				t.Errorf("Case: %v, expected: %v, got: %v", tc.name, tc.insuficientMaterial, got)
			}
		})
	}

}

// 960 castles make and unmake moves tests
func TestMakeMoveWithCastling960(t *testing.T) {
	// chess 960 initial pos 599
	// rqbnkrnb/pppppppp/8/8/8/8/PPPPPPPP/RQBNKRNB w KQkq - 0 1

	pos := NewPosition()
	pos.LoadFromFenString("rq2krn1/pp1b1pbp/2n3p1/4p3/8/3PN1P1/PPP1NP1P/RQB1KR1B w KQkq - 0 9")
	pos.castling = *NewCastling(4, 5, 0)
	pos.castling.chess960 = true
	pos.castling.castlingRights = KQkq

	move := encodeMove(uint16(4), uint16(5), kingsideCastle)
	pos.MakeMove(move)

	// fen "after" white short castle
	// rq2krn1/pp1b1pbp/2n3p1/4p3/8/3PN1P1/PPP1NP1P/RQB2RKB b kq - 1 9
	expected := "rq2krn1/pp1b1pbp/2n3p1/4p3/8/3PN1P1/PPP1NP1P/RQB2RKB b kq - 1 9"
	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestUnMakeMoveWithCastling960(t *testing.T) {
	// chess 960 initial pos 599
	// rqbnkrnb/pppppppp/8/8/8/8/PPPPPPPP/RQBNKRNB w KQkq - 0 1

	pos := NewPosition()
	pos.LoadFromFenString("rq2krn1/pp1b1pbp/2n3p1/4p3/8/3PN1P1/PPP1NP1P/RQB1KR1B w KQkq - 0 9")
	pos.castling = *NewCastling(4, 5, 0)
	pos.castling.chess960 = true
	pos.castling.castlingRights = KQkq

	move := encodeMove(uint16(4), uint16(5), kingsideCastle)
	pos.MakeMove(move)
	pos.UnmakeMove(move)

	expected := "rq2krn1/pp1b1pbp/2n3p1/4p3/8/3PN1P1/PPP1NP1P/RQB1KR1B w KQkq - 0 9"
	got := pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestColorModifier(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("8/8/8/8/8/8/8/8 w - - 0 1")

	expected := 1
	got := pos.Turn.Modifier()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestThreefoldRepetitions(t *testing.T) {
	testCases := []struct {
		fen                   string
		moves                 string
		isThreefoldRepetition bool
	}{
		// BUG #1: Halfmove clock not reset after double pawn pushes
		// BUG #2: If ep square is not reachable by enemy pawns, there  no ep avaialbe. So do not hash the zobrist ep square key.
		// Warning; PV continues after threefold repetition - move a4a5 from Aconcagua DEV
		// Info; info depth 13 seldepth 21 score cp -60 nodes 22290 nps 3026546 hashfull 5 time 25 pv d7e6 c2d3 e6f7 g2g4 a4a3 d3e4 a3a4 e4d3 a4a5 d3e4 a5a4 e4d3 a4a5 d3d4 a5b5 d4e4
		// Info; info depth 13 seldepth 21 score cp -60 nodes 22290 nps 3026546 hashfull 5 time 25 pv d7e6 c2d3 e6f7 g2g4 a4a3 d3e4 a3a4 e4d3 a4a5 d3e4 a5a4 (e4d3 - this is the draw by repetition) a4a5 d3d4 a5b5 d4e4
		// Position; fen r1bqkb1r/pppp1ppp/2n2n2/4P3/4P3/5N2/PPP2PPP/RNBQKB1R b KQkq - 0 4
		// Moves; f6g4 f1e2 f8c5 e1g1 g4e5 b1c3 e5f3 e2f3 e8g8 c3d5 c6e5 c1e3 d7d6 f3e2 c5e3 d5e3 f8e8 f2f4 e5d7 d1d4 d7c5 e4e5 c8d7 a1d1 a7a5 e5d6 c7d6 e2f3 a8a6 e3c4 c5e6 d4d2 d7c6 f4f5 e6g5 f3c6 a6c6 c4d6 d8b6 g1h1 e8d8 d2g5 c6d6 d1d6 b6d6 g5h4 f7f6 h4e4 d6c6 e4c6 b7c6 h1g1 d8d2 f1c1 g8f7 b2b3 f7e7 c1e1 e7d6 e1e6 d6d7 e6e3 d2c2 e3g3 g7g5 f5g6 h7g6 g3g6 c2c1 g1f2 c1c2 f2f1 c2c1 f1e2 c1c2 e2d3 c2a2 h2h4 a2a3 d3c2 a5a4 b3a4 a3a4 h4h5
		{"r1bqkb1r/pppp1ppp/2n2n2/4P3/4P3/5N2/PPP2PPP/RNBQKB1R b KQkq - 0 4", "f6g4 f1e2 f8c5 e1g1 g4e5 b1c3 e5f3 e2f3 e8g8 c3d5 c6e5 c1e3 d7d6 f3e2 c5e3 d5e3 f8e8 f2f4 e5d7 d1d4 d7c5 e4e5 c8d7 a1d1 a7a5 e5d6 c7d6 e2f3 a8a6 e3c4 c5e6 d4d2 d7c6 f4f5 e6g5 f3c6 a6c6 c4d6 d8b6 g1h1 e8d8 d2g5 c6d6 d1d6 b6d6 g5h4 f7f6 h4e4 d6c6 e4c6 b7c6 h1g1 d8d2 f1c1 g8f7 b2b3 f7e7 c1e1 e7d6 e1e6 d6d7 e6e3 d2c2 e3g3 g7g5 f5g6 h7g6 g3g6 c2c1 g1f2 c1c2 f2f1 c2c1 f1e2 c1c2 e2d3 c2a2 h2h4 a2a3 d3c2 a5a4 b3a4 a3a4 h4h5 d7e6 c2d3 e6f7 g2g4 a4a3 d3e4 a3a4 e4d3 a4a5 d3e4 a5a4 e4d3", true},
		{"r2q1rk1/pp1b1p2/2n1pn1Q/2bp4/8/2PBPN2/PP1N1PPP/R3K2R b KQ - 0 12", "e6e5 h6g5 g8h8 g5h6 h8g8 h6g5 g8h8 g5h6 h8g8", true},
		{"5k2/6p1/pQ3p1p/1p6/4q3/2P3PK/1P5P/8 b - - 6 36", "e4f5 h3g2 f5e4 g2h3 e4f5 h3h4 f5e4 h4h3", true},
		{"5k2/6p1/pQ3p1p/1p6/4q3/2P3PK/1P5P/8 b - - 6 36", "e4f5 h3g2 f5e4 g2h3 e4f5 h3h4 f5e4", false},
		{"2b2rk1/pr5p/2pqp1p1/2p1R3/4R3/P1NP4/1P1Q1PPP/6K1 w - - 3 23", "c3a4 b7b5 a4c3 b5b7 c3a4 b7b5 a4c3 b5b7", true},
	}

	for _, test := range testCases {
		pos := NewPosition()
		pos.LoadFromFenString(test.fen)
		moves := strings.Split(test.moves, " ")
		pos.LoadMoves(moves...)

		expected := test.isThreefoldRepetition
		got := pos.isThreefoldRepetition()

		if got != expected {
			t.Errorf("Expected: %v got: %v", expected, got)
		}

	}
}
