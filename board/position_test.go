package board

import (
	"testing"
)

// Position tests

func TestCheckingPieces(t *testing.T) {
	pos := EmptyPosition()

	pos.AddPiece(BlackKnight, "f3")
	pos.AddPiece(WhiteKing, "e1")

	expected := 1
	got := pos.CheckingPieces(White).count()

	if expected != got {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGetDirectionNorth(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackKing, "e1")
	pos.AddPiece(BlackRook, "e8")
	// king, _ := pos.PieceAt("e1")
	// rook, _ := pos.PieceAt("e8")

	expected := NORTH
	got := getDirection(squareToBitboard([]string{"e8"}), squareToBitboard([]string{"e1"})) // king -> rook == NORTH

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGetDirectionSouth(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackKing, "e1")
	pos.AddPiece(BlackRook, "e8")

	expected := SOUTH
	got := getDirection(squareToBitboard([]string{"e1"}), squareToBitboard([]string{"e8"})) // rook -> king == SOUTH

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGetDirectionSouthWest(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackKing, "e4")
	pos.AddPiece(BlackRook, "d3")

	expected := SOUTHWEST
	got := getDirection(squareToBitboard([]string{"d3"}), squareToBitboard([]string{"e4"})) // king -> rook == SOUTHWEST

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGetRayPath(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackRook, "c4")
	pos.AddPiece(WhiteRook, "f4")

	expectedSquares := []string{"d4", "e4"}

	expected := squareToBitboard(expectedSquares)
	got := getRayPath(squareToBitboard([]string{"c4"}), squareToBitboard([]string{"f4"}))

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPinnedPiece(t *testing.T) {
	pos := EmptyPosition()
	pos.AddPiece(BlackKing, "c7")
	pos.AddPiece(BlackRook, "c6")
	pos.AddPiece(WhiteRook, "c1")
	blackRook, _ := pos.PieceAt("c6")

	expected := true
	got := isPinned(squareToBitboard([]string{"c6"}), pieceColor[blackRook], pos)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestShortLegalCastleForWhite(t *testing.T) {
	pos := From("8/8/8/8/8/8/8/4K2R w K - 0 1")

	expected := 1
	got := len(pos.legalCastles(White))

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestShortIlegalCastleForWhite(t *testing.T) {
	pos := From("8/8/8/2b5/8/8/8/4K2R w K - 0 1")

	expected := 0
	got := len(pos.legalCastles(White))

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLongIllegalCastleForBlack(t *testing.T) {
	pos := From("rn2k3/8/8/8/8/8/8/8 w q - 1 1")

	expected := 0
	got := len(pos.legalCastles(Black))

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestEnPassantMoves(t *testing.T) {
	// Both white pawns from c5 and e5 can capture en passant to d6
	pos := From("3k4/8/8/2PpP3/8/8/8/3K4 w - d6 0 1")

	expected := 2
	got := len(pos.legalEnPassant(White))

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesFromInitialPosition(t *testing.T) {
	pos := From("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	expected := 20
	got := len(pos.LegalMoves(White))

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnAPositionWithPromotion(t *testing.T) {
	pos := From("3r2k1/5ppp/8/8/8/8/pp4PP/5R1K b - - 0 1")

	expected := 28
	got := len(pos.LegalMoves(Black))

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnAPositionIllegalLongCastle(t *testing.T) {
	pos := From("6k1/5ppp/8/7q/7b/8/5PPP/RN2K2R w KQ - 1 1")

	expected := 18
	got := len(pos.LegalMoves(White))

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnAPositionWithDoubleEnPassantCaptures(t *testing.T) {
	pos := From("6k1/5bpp/8/1PpPN3/8/8/6PP/6K1 w - c6 0 1")

	expected := 19
	got := len(pos.LegalMoves(White))

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnMultiplePinsWithCheck(t *testing.T) {
	pos := From("8/1k3Rpp/1n6/3b4/8/5B2/6PP/1R4K1 b - - 0 1")

	expected := 5
	got := len(pos.LegalMoves(Black))

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestLegalMovesOnMultiplePinsWithCheckTwo(t *testing.T) {
	pos := From("8/1k3Rpp/1n6/3b4/8/5B2/6PP/2R3K1 b - - 0 1")

	expected := 4 // 3 of king 1 block of the knight
	got := len(pos.LegalMoves(Black))

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

// TODO: checkmate and stealmate position tests!!!!
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

func TestBlackIsInStealmate(t *testing.T) {

	pos := From("7k/6pn/6P1/3B4/7Q/7p/PPP4R/1K6 b - - 0 1")

	expected := true
	got := pos.Stealmate(Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBlackIsNotInStealmate(t *testing.T) {

	pos := From("7k/6pn/6P1/3B4/p6Q/7p/PPP4R/1K6 b - - 0 1")

	expected := false
	got := pos.Stealmate(Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestInsuficientMaterialKingVsKing(t *testing.T) {
	pos := From("8/8/3k4/8/1K6/8/8/8 w - - 0 1")

	expected := true
	got := pos.InsuficientMaterial()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestInsuficientMaterialKingAndKnightVsKing(t *testing.T) {
	pos := From("8/8/3k4/8/1K6/1N6/8/8 w - - 0 1")

	expected := true
	got := pos.InsuficientMaterial()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestInsuficientMaterialKingAndBishopVsKingAndKnight(t *testing.T) {
	pos := From("K7/8/3B4/4n3/8/8/7k/8 w - - 0 1")

	expected := true
	got := pos.InsuficientMaterial()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestNotInsuficientMaterial(t *testing.T) {
	pos := From("8/6R1/8/8/8/8/K6k/8 w - - 0 1")

	expected := false
	got := pos.InsuficientMaterial()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestCaptureUpdatesPosition(t *testing.T) {
	pos := From("7k/8/8/8/3p4/4P3/8/7K w - - 0 1")
	// move := &Move{from: "e3", to: "d4", piece: WHITE_PAWN, moveType: CAPTURE}

	from := Bsf(squareToBitboard([]string{"e3"}))
	to := Bsf(squareToBitboard([]string{"d4"}))
	move := MoveEncode(from, to, int(WhitePawn), 0, CAPTURE)

	newPos := pos.MakeMove(&move)

	expected := "7k/8/8/8/3P4/8/8/7K b - - 0 1"
	got := newPos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

// TODO: Test moves sequences that generates the same position gets the same zorbitst hash
func TestZobristUpdate(t *testing.T) {
	pos := InitialPosition()

	from := Bsf(squareToBitboard([]string{"g1"}))
	to := Bsf(squareToBitboard([]string{"f3"}))
	move1 := MoveEncode(from, to, int(WhitePawn), 0, NORMAL)

	from = Bsf(squareToBitboard([]string{"b7"}))
	to = Bsf(squareToBitboard([]string{"c5"}))
	move2 := MoveEncode(from, to, int(BlackPawn), 0, NORMAL)

	from = Bsf(squareToBitboard([]string{"b1"}))
	to = Bsf(squareToBitboard([]string{"c3"}))
	move3 := MoveEncode(from, to, int(WhiteKnight), 0, NORMAL)

	from = Bsf(squareToBitboard([]string{"g8"}))
	to = Bsf(squareToBitboard([]string{"f6"}))
	move4 := MoveEncode(from, to, int(BlackKnight), 0, NORMAL)

	// Normal order 1 2 3 4
	// move1 := &Move{from: "g1", to: "f3", piece: WHITE_PAWN, moveType: NORMAL}
	// move2 := &Move{from: "b8", to: "c6", piece: BLACK_PAWN, moveType: NORMAL}
	// move3 := &Move{from: "b1", to: "c3", piece: WHITE_KNIGHT, moveType: NORMAL}
	// move4 := &Move{from: "g8", to: "f6", piece: BLACK_KNIGHT, moveType: NORMAL}

	pos1 := pos.MakeMove(&move1)
	pos2 := pos1.MakeMove(&move2)
	pos3 := pos2.MakeMove(&move3)
	pos4 := pos3.MakeMove(&move4)

	// Invert move order 3 4 1 2 -> Gets the same fen -> "r1bqkb1r/pppppppp/2n2n2/8/8/2N2N2/PPPPPPPP/R1BQKB1R w KQkq - 2 3"
	pos5 := pos.MakeMove(&move3)
	pos6 := pos5.MakeMove(&move4)
	pos7 := pos6.MakeMove(&move1)
	pos8 := pos7.MakeMove(&move2)

	expected := pos4.Zobrist
	got := pos8.Zobrist

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
