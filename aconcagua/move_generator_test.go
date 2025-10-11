package aconcagua

import "testing"

func TestMoveGeneratorHasNextMove(t *testing.T) {
	pos := NewPosition()
	hashMove := encodeMove(0, 0, quiet)
	killers := Killer{NoMove, NoMove}
	mg := NewMoveGenerator(pos, hashMove, &killers[0], &killers[1], &HistoryMovesTable{})

	got := mg.nextMove()

	if got != NoMove {
		t.Errorf("Expected: %v, got: %v", NoMove, got)
	}
}

func TestMoveGeneratorNotHasNextMove(t *testing.T) {
	pos := NewPosition()
	hashMove := NoMove
	killers := Killer{NoMove, NoMove}
	mg := NewMoveGenerator(pos, &hashMove, &killers[0], &killers[1], &HistoryMovesTable{})
	mg.stage = EndStage

	expected := NoMove
	got := mg.nextMove()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveGeneratorCreatesCaptures(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("1b4k1/5pp1/3r3p/4P3/5PN1/3RK3/8/8 w - - 0 1") // Only 3 captures
	mg := NewMoveGenerator(pos, nil, nil, nil, nil)
	move := NoMove
	mg.hashMove = &move

	expected := *encodeMove(36, 43, capture) // Best capture. Pawn takes rook
	got := mg.nextMove()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveGeneratorCreatesNonCaptures(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("1b4k1/5pp1/3r3p/4P3/5PN1/3RK3/8/8 w - - 0 1") // Only 3 captures
	noMove := NoMove
	mg := NewMoveGenerator(pos, &noMove, &noMove, &noMove, &HistoryMovesTable{})
	mg.stage = FirstKillerStage // NOTE: Non captures are generated in killers stage to validate legaliy of killers
	move := NoMove
	mg.hashMove = &move

	mg.pd = mg.pos.generatePositionData()
	mg.nextMove()

	expected := 16 // 17 - 1 (the one picked)
	got := len(mg.nonCaptures)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveGeneratorGetsAllMoves(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("1b4k1/5pp1/3r3p/4P3/5PN1/3RK3/8/8 w - - 0 1") // 3 captures + 17 non capt
	hashMove := NoMove

	ml := NewMoveList(20)
	pd := pos.generatePositionData()
	pos.generateNonCaptures(&ml, &pd)
	killers := Killer{ml[0], ml[5]}

	mg := NewMoveGenerator(pos, &hashMove, &killers[0], &killers[1], &HistoryMovesTable{})
	for mg.nextMove() != NoMove {
	}

	expected := 20
	got := mg.moveNumber

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestStraightPinnedPieces(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKing, e1)
	pos.AddPiece(BlackRook, e8)
	pos.AddPiece(WhitePawn, e4)

	expected := bitboardFromCoordinates("e4")
	got := pos.pinnedPieces(White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestDiagonalPinnedPieces(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(BlackKing, e8)
	pos.AddPiece(BlackPawn, d7)
	pos.AddPiece(BlackKnight, f7)
	pos.AddPiece(WhiteQueen, h5)
	pos.AddPiece(WhiteBishop, a4)

	expected := bitboardFromCoordinates("f7", "d7")
	got := pos.pinnedPieces(Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestDiagonalNoPinnedPieces(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(BlackKing, e8)
	pos.AddPiece(BlackPawn, d7)
	pos.AddPiece(BlackPawn, c6)
	pos.AddPiece(WhiteBishop, a4)

	expected := Bitboard(0)
	got := pos.pinnedPieces(Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPinnedPiecesOnBlack(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("7k/6pn/6P1/3B4/7Q/7p/PPP4R/1K6 b - - 0 1")

	expected := Bitboard(1 << 55) // knight on h7
	got := pos.pinnedPieces(Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestCheckRestrictedSquares(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("8/8/8/8/1k4PP/1bp1r2K/8/5N2 w - - 0 1")
	checkingSliders := pos.Bitboards[BlackRook] // Black Rook on e3
	checkingNonSliders := Bitboard(0)

	expected := bitboardFromCoordinates("e3", "f3", "g3")
	got := checkRestrictedSquares(pos.KingPosition(White), checkingSliders, checkingNonSliders)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPinRestrictedSquares(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("2br2k1/5pp1/5p2/R4BP1/5PKP/8/8/8 w - - 0 1") // bishop on f5 is pinned can only move along the h3-c8 diagonal

	piece := pos.Bitboards[WhiteBishop]
	king := pos.KingPosition(White)
	pinnedPieces := pos.pinnedPieces(White)

	expected := bitboardFromCoordinates("h3", "g4", "f5", "e6", "d7", "c8")
	got := pinRestrictedSquares(piece, king, pinnedPieces)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

// King tests
func TestKingAttacks(t *testing.T) {
	kingBB := bitboardFromCoordinates("e1")

	expected := Bitboard(0b11100000101000)
	got := kingAttacks(&kingBB) // The king defends... all pieces around him

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingMovesToEmptySquares(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKing, e4)
	kingBB := bitboardFromCoordinates("e4")

	expected := bitboardFromCoordinates("d3", "d4", "d5", "e3", "e5", "f3", "f4", "f5")
	got := kingMoves(&kingBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCannotMoveToAttackedSquare(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKing, e4)
	pos.AddPiece(BlackKnight, c6)
	kingBB := bitboardFromCoordinates("e4")

	// Cannot move to d4 or e5 because it's attacked by the black knight
	expected := bitboardFromCoordinates("d3", "d5", "e3", "f3", "f4", "f5")
	got := kingMoves(&kingBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingMovesWhenInCheck(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKing, e1)
	pos.AddPiece(BlackRook, h1)
	kingBB := bitboardFromCoordinates("e1")

	// Can only move to the second rank, becuase first rank is attacked by the rook, by x rays
	expected := bitboardFromCoordinates("d2", "e2", "f2")
	got := kingMoves(&kingBB, pos, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingValidMoves(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKing, e1)
	pos.AddPiece(BlackRook, h1)
	kingBB := bitboardFromCoordinates("e1")

	expected := 3
	got := kingMoves(&kingBB, pos, White).count()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCanCastleShort(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKing, e1)
	pos.AddPiece(WhiteRook, h1)
	pos.castling.castlingRights = Kk

	expected := true
	got := pos.canCastle(White, kingsideCastle)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCanCastleShortTwo(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("rnb2k1r/pp1Pbppp/2p5/q7/2B5/8/PPPQNnPP/RNB1K2R w KQ - 3 9")
	pos.castling.castlingRights = Kk

	expected := true
	got := pos.canCastle(White, kingsideCastle)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCannotCastleShortIfPathBlocked(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKing, e1)
	pos.AddPiece(WhiteRook, h1)
	pos.AddPiece(WhiteKnight, f1)
	pos.castling.castlingRights = Kk

	expected := false
	got := pos.canCastle(White, kingsideCastle)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCannotCastleShortIfItsInCheck(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKing, e1)
	pos.AddPiece(WhiteRook, h1)
	pos.AddPiece(BlackRook, f4)
	pos.castling.castlingRights = Kk

	expected := false
	got := pos.canCastle(White, kingsideCastle)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCanCastleLong(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKing, e1)
	pos.AddPiece(WhiteRook, a1)
	pos.castling.castlingRights = Qq

	expected := true
	got := pos.canCastle(White, queensideCastle)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBlackKingCannotCastleLongIfPathBlocked(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(BlackKing, e8)
	pos.AddPiece(BlackRook, a8)
	pos.AddPiece(BlackKnight, b8)
	pos.castling.castlingRights = Qq

	expected := false
	got := pos.canCastle(Black, queensideCastle)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKingCannotCastleLongIfItsInCheck(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(BlackKing, e8)
	pos.AddPiece(BlackRook, a8)
	pos.AddPiece(WhiteRook, c6)
	pos.castling.castlingRights = Qq

	expected := false
	got := pos.canCastle(Black, queensideCastle)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGenKingMoves(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKing, e1)
	pos.AddPiece(WhiteRook, h1)
	pos.AddPiece(WhiteRook, a1)
	pos.castling.castlingRights = KQ
	kingBB := bitboardFromCoordinates("e1")
	ml := NewMoveList(100)
	pd := pos.generatePositionData()

	expected := 5 // Castles moves are treated separately
	genMovesFromTargets(&kingBB, kingMoves(&kingBB, pos, White), &ml, &pd)
	got := len(ml)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGenCastles(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKing, e1)
	pos.AddPiece(WhiteRook, h1)
	pos.AddPiece(WhiteRook, a1)
	pos.castling.castlingRights = KQ
	ml := NewMoveList(100)

	expected := 2 // Castles moves are treated separately

	genCastleMoves(pos, &ml)
	got := len(ml)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

// Rook tests
func TestRookAttacksOnEmptyBoard(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(BlackRook, e4)
	rookBB := bitboardFromCoordinates("e4")

	expected := bitboardFromCoordinates("e1", "e2", "e3", "e5", "e6", "e7", "e8",
		"a4", "b4", "c4", "d4", "f4", "g4", "h4")
	got := rookAttacks(Bsf(rookBB), pos.pieces[White]|pos.pieces[Black])

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookAttacksWithBlockedSquares(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(BlackRook, e4)
	pos.AddPiece(WhiteKnight, c4) // Knight blocking on c4
	rookBB := bitboardFromCoordinates("e4")

	expected := bitboardFromCoordinates("e1", "e2", "e3", "e5", "e6", "e7", "e8",
		"c4", "d4", "f4", "g4", "h4")
	got := rookAttacks(Bsf(rookBB), pos.pieces[White]|pos.pieces[Black])

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookAttacksWithAllSquaresBlocked(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(BlackRook, b3)
	pos.AddPiece(WhiteKnight, b4) // Knight blocking on b4
	pos.AddPiece(WhiteKnight, b2) // Knight blocking on b5
	pos.AddPiece(WhiteKnight, a3) // Knight blocking on a3
	pos.AddPiece(WhiteKnight, c3) // Knight blocking on c3
	rookBB := bitboardFromCoordinates("b3")

	expected := bitboardFromCoordinates("b4", "b2", "a3", "c3")
	got := rookAttacks(Bsf(rookBB), pos.pieces[White]|pos.pieces[Black])

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMovesWithCaptures(t *testing.T) {
	pos := NewPosition()
	pos.Turn = Black
	pos.AddPiece(BlackRook, a8)
	pos.AddPiece(WhiteKnight, a4) // Can move(capture) white knight on a4
	pos.AddPiece(WhiteKnight, c8) // Can move(capture) knight on c8
	rookBB := bitboardFromCoordinates("a8")
	pd := pos.generatePositionData()

	expected := bitboardFromCoordinates("a7", "a6", "a5", "a4", "b8", "c8")
	got := rookMoves(&rookBB, &pd)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMovesWithBlockingPieces(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteRook, g2)
	pos.AddPiece(BlackKnight, g4) // Can move(capture) white knight on g4
	pos.AddPiece(WhiteKnight, f2) // Cannot move to f2, because its blocked by Knight
	rookBB := bitboardFromCoordinates("g2")
	pd := pos.generatePositionData()

	expected := bitboardFromCoordinates("g1", "h2", "g3", "g4")
	got := rookMoves(&rookBB, &pd)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMoveWhenCanBlockCheck(t *testing.T) {
	pos := NewPosition()
	pos.Turn = Black
	pos.AddPiece(BlackRook, c6) // Black king is in check, so only legal move is Re6 blocking the check
	pos.AddPiece(WhiteRook, e1)
	pos.AddPiece(BlackKing, e8)
	rookBB := bitboardFromCoordinates("c6")
	pd := pos.generatePositionData()

	expected := bitboardFromCoordinates("e6")
	got := rookMoves(&rookBB, &pd)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMovesWhenKingInDoubleCheck(t *testing.T) {
	pos := NewPosition()
	pos.Turn = Black
	pos.AddPiece(BlackRook, c6)
	pos.AddPiece(WhiteRook, e1) // Gives check to the black king on e8
	pos.AddPiece(WhiteRook, a8) // Gives check to the black king on e8
	pos.AddPiece(BlackKing, e8)
	rookBB := bitboardFromCoordinates("c6")
	pd := pos.generatePositionData()

	expected := bitboardFromCoordinates()
	got := rookMoves(&rookBB, &pd)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookMovesWhenTheRookIsPinned(t *testing.T) {
	pos := NewPosition()
	pos.Turn = Black
	pos.AddPiece(BlackKing, e8)
	pos.AddPiece(BlackRook, e4)
	pos.AddPiece(WhiteRook, e2)
	pos.AddPiece(WhiteRook, e3)

	// The rook can only move along the e file, becuase it's pinned if moves
	// along the 4 rank, the king will be in check!
	rookBB := bitboardFromCoordinates("e4")
	pd := pos.generatePositionData()

	expected := bitboardFromCoordinates("e3", "e5", "e6", "e7")
	got := rookMoves(&rookBB, &pd)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}

func TestGenTargetMovesForRook(t *testing.T) {
	pos := NewPosition()
	pos.Turn = Black
	pos.AddPiece(BlackRook, a8)
	pos.AddPiece(WhiteKnight, a4) // Can move(capture) white knight on a4
	pos.AddPiece(WhiteKnight, c8) // Can move(capture) knight on c8
	rookBB := bitboardFromCoordinates("a8")
	ml := NewMoveList(100)
	pd := pos.generatePositionData()

	genMovesFromTargets(&rookBB, rookMoves(&rookBB, &pd), &ml, &pd)
	expectedSquares := []string{"a7", "a6", "a5", "a4", "b8", "c8"}

	expected := len(expectedSquares)
	got := len(ml)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expectedSquares, got)
	}

}

// Bishop tests
func TestBishopAttacksOnEmptyBoard(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteBishop, h1)
	bishopBB := bitboardFromCoordinates("h1")

	expected := bitboardFromCoordinates("g2", "f3", "e4", "d5", "c6", "b7", "a8")
	got := bishopAttacks(Bsf(bishopBB), pos.pieces[White]|pos.pieces[Black])

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopAttacksWithBlockedSquares(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteBishop, e3)
	pos.AddPiece(BlackRook, g5)
	bishopBB := bitboardFromCoordinates("e3")

	expected := bitboardFromCoordinates("f2", "g1", "d4", "c5", "b6", "a7", "f4", "g5", "d2", "c1")
	got := bishopAttacks(Bsf(bishopBB), pos.pieces[White]|pos.pieces[Black])

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopMovesWithCaptures(t *testing.T) {
	pos := NewPosition()
	pos.Turn = Black
	pos.AddPiece(BlackBishop, c4)
	pos.AddPiece(WhiteRook, f7)   // Can move(capture) white rook on f7
	pos.AddPiece(WhiteKnight, d3) // Can move(capture) knight on d3
	bishopBB := bitboardFromCoordinates("c4")
	pd := pos.generatePositionData()

	expected := bitboardFromCoordinates("a2", "b3", "d3", "d5", "e6", "f7", "b5", "a6")
	got := bishopMoves(&bishopBB, &pd)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopMovesWithBlockingPieces(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteBishop, g6)
	pos.AddPiece(WhiteKnight, e8) // Cannot move, blocked by same color knight
	pos.AddPiece(WhiteRook, f5)   // Cannot move to f5, because its blocked by Rook
	bishopBB := bitboardFromCoordinates("g6")
	pd := pos.generatePositionData()

	expected := bitboardFromCoordinates("h7", "h5", "f7")
	got := bishopMoves(&bishopBB, &pd)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopMoveWhenCanBlockCheck(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKing, d1) // White king is in check, only legal move is Bd2
	pos.AddPiece(BlackRook, d8) // And also Bxd8 by capturing the Rook which is checking the king
	pos.AddPiece(WhiteBishop, g5)
	bishopBB := bitboardFromCoordinates("g5")
	pd := pos.generatePositionData()

	expected := bitboardFromCoordinates("d2", "d8")
	got := bishopMoves(&bishopBB, &pd)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopMovesWhenPinnedAndInCheck(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKing, d1)
	pos.AddPiece(BlackRook, h1) // Gives check to the white king on d1
	pos.AddPiece(BlackRook, d8) // Gives check to the white king on d1 (by xrays) -> pins the bishop
	pos.AddPiece(WhiteBishop, d4)
	bishopBB := bitboardFromCoordinates("d4")
	pd := pos.generatePositionData()

	expected := bitboardFromCoordinates()
	got := bishopMoves(&bishopBB, &pd)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishpMovesWhenTheBishopIsPinned(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKing, c4)
	pos.AddPiece(BlackBishop, g8)
	pos.AddPiece(WhiteBishop, d5)

	bishopBB := bitboardFromCoordinates("d5")
	pd := pos.generatePositionData()

	// Can only move along the g8 c4 diagonal because of the pin
	expected := bitboardFromCoordinates("e6", "f7", "g8")
	got := bishopMoves(&bishopBB, &pd)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGenTargetMovesForBishop(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteBishop, g6)
	pos.AddPiece(WhiteKnight, e8) // Cannot move, blocked by same color knight
	pos.AddPiece(WhiteRook, f5)   // Cannot move to f5, because its blocked by Rook
	bishopBB := bitboardFromCoordinates("g6")
	ml := NewMoveList(100)
	pd := pos.generatePositionData()

	expectedSquares := []string{"h7", "h5", "f7"}

	expected := len(expectedSquares)
	genMovesFromTargets(&bishopBB, bishopMoves(&bishopBB, &pd), &ml, &pd)
	got := len(ml)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

// Knight tests
func TestKnightAttacks(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(BlackKnight, e4)
	knightBB := bitboardFromCoordinates("e4")

	expected := bitboardFromCoordinates("d6", "f6", "d2", "f2", "g5", "g3", "c5", "c3")
	got := knightAttacksTable[Bsf(knightBB)]

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKnightMovesWhenBlockedBySameColorPieces(t *testing.T) {
	pos := NewPosition()
	pos.Turn = Black
	pos.AddPiece(BlackKnight, e4)
	pos.AddPiece(BlackRook, d6)
	pos.AddPiece(BlackRook, f6)
	pos.AddPiece(BlackKing, d2)
	pos.AddPiece(BlackBishop, f2)
	knightBB := bitboardFromCoordinates("e4")
	pd := pos.generatePositionData()

	expected := bitboardFromCoordinates("g5", "g3", "c5", "c3")
	got := knightMoves(&knightBB, &pd)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKnightMovesWithCaptures(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKnight, b1)
	pos.AddPiece(BlackBishop, c3)
	pos.AddPiece(WhiteRook, a3)
	pos.AddPiece(WhiteRook, d2) // Blocks Knight move
	knightBB := bitboardFromCoordinates("b1")
	pd := pos.generatePositionData()

	expected := bitboardFromCoordinates("c3") // The Knight can only capture the bishop. "a3" and "d2" are blocked by the rook, so it cannot move there
	got := knightMoves(&knightBB, &pd)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestKnightMovesWhenPinned(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKnight, e4)
	pos.AddPiece(BlackRook, e8)
	pos.AddPiece(WhiteKing, e1)
	knightBB := bitboardFromCoordinates("e4")
	pd := pos.generatePositionData()

	expected := Bitboard(0) // The Knight is pinned, it cannot move at all
	got := knightMoves(&knightBB, &pd)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}

func TestGenTargetMovesForKnight(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteKnight, b1)
	pos.AddPiece(BlackBishop, c3)
	pos.AddPiece(WhiteRook, a3)
	pos.AddPiece(WhiteRook, d2) // Blocks Knight move
	knightBB := bitboardFromCoordinates("b1")
	pd := pos.generatePositionData()
	ml := NewMoveList(100)

	expected := []Move{*encodeMove(1, 18, capture)} // The Knight can only capture the bishop. "a3" and "d2" are blocked by the rook, so it cannot move there
	genMovesFromTargets(&knightBB, knightMoves(&knightBB, &pd), &ml, &pd)
	got := ml

	if got[0] != expected[0] {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

// Pawn tests
func TestPawnAttacks(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhitePawn, e2)
	pawnBB := bitboardFromCoordinates("e2")

	expectedSquares := []string{"d3", "f3"}

	expected := bitboardFromCoordinates(expectedSquares...)
	got := pawnAttacks(&pawnBB, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnAttacksOnEdgeFiles(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhitePawn, h2)
	// pawn, _ := pos.PieceAt("h2")
	pawnBB := bitboardFromCoordinates("h2")

	expectedSquares := []string{"g3"}

	expected := bitboardFromCoordinates(expectedSquares...)
	got := pawnAttacks(&pawnBB, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnMovesOnEmptyBoard(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhitePawn, e2)
	pawnBB := bitboardFromCoordinates("e2")
	pd := pos.generatePositionData()

	expectedSquares := []string{"e3", "e4"}

	expected := bitboardFromCoordinates(expectedSquares...)
	got := pawnMoves(&pawnBB, &pd, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnMovesWithCapturesFrom7thRank(t *testing.T) {
	pos := NewPosition()
	pos.Turn = Black
	pos.AddPiece(BlackPawn, b7)
	pos.AddPiece(WhiteBishop, a6)
	pos.AddPiece(BlackKnight, c6)
	pawnBB := bitboardFromCoordinates("b7")
	pd := pos.generatePositionData()

	// Can capture white bishop on a6 and is blocked by black knight on c6
	// Can also move to b6 and b7
	expectedSquares := []string{"a6", "b6", "b5"}

	expected := bitboardFromCoordinates(expectedSquares...)
	got := pawnMoves(&pawnBB, &pd, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnCanBlockACheckOnFirstMove(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhitePawn, f2)
	pos.AddPiece(BlackRook, h4)
	pos.AddPiece(WhiteKing, c4)
	pawnBB := bitboardFromCoordinates("f2")
	pd := pos.generatePositionData()

	// The only legal move of the pawn is to block the check on f4
	expectedSquares := []string{"f4"}

	expected := bitboardFromCoordinates(expectedSquares...)
	got := pawnMoves(&pawnBB, &pd, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnCanOnlyMoveInThePinnedDirection(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhitePawn, f2)
	pos.AddPiece(BlackBishop, e3)
	pos.AddPiece(WhiteKing, g1)
	pawnBB := bitboardFromCoordinates("f2")
	pd := pos.generatePositionData()

	// The only legal move of the pawn is to capture the bishop on e3
	expectedSquares := []string{"e3"}

	expected := bitboardFromCoordinates(expectedSquares...)
	got := pawnMoves(&pawnBB, &pd, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnPinnedAndInCheck(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhitePawn, f2)
	pos.AddPiece(BlackBishop, e3)
	pos.AddPiece(BlackRook, g8)
	pos.AddPiece(WhiteKing, g1)
	pawnBB := bitboardFromCoordinates("f2")
	pd := pos.generatePositionData()

	expectedSquares := []string{}

	expected := bitboardFromCoordinates(expectedSquares...)
	got := pawnMoves(&pawnBB, &pd, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBlackPawnInA4Moves(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("rnbqkbnr/1ppppppp/8/8/p7/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1")
	pawnBB := bitboardFromCoordinates("a4")
	pd := pos.generatePositionData()

	expected := bitboardFromCoordinates("a3")
	got := pawnMoves(&pawnBB, &pd, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnIsNotPinnedIfCapturesThePinnedPiece(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("r1bqkbnr/7p/2p1p1p1/p1pp1p1Q/P4P2/3PP3/1PPBN1PP/RN3RK1 b kq - 1 9")
	pawnBB := bitboardFromCoordinates("g6")
	pd := pos.generatePositionData()

	expected := bitboardFromCoordinates("h5")
	got := pawnMoves(&pawnBB, &pd, Black)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnsMoves(t *testing.T) {
	pos := NewPosition()
	pawnBB := bitboardFromCoordinates("e2")
	pd := pos.generatePositionData()
	ml := NewMoveList(100)

	expected := 2
	genPawnMovesFromTarget(&pawnBB, pawnMoves(&pawnBB, &pd, White), White, &ml, &pd)
	got := len(ml)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnsMovesPromo(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("8/7P/2k5/8/8/8/8/4K3 w - - 0 1")
	pawnBB := bitboardFromCoordinates("h7")
	pd := pos.generatePositionData()
	ml := NewMoveList(100)

	expected := 4
	genPawnMovesFromTarget(&pawnBB, pawnMoves(&pawnBB, &pd, White), White, &ml, &pd)
	got := len(ml)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestGenEpPawnCaptures(t *testing.T) {
	ml := NewMoveList(100)

	pos := NewPosition()
	pos.LoadFromFenString("4r3/8/8/R7/3Pp2k/8/8/4K3 b - d3 0 1")

	genEnPassantCaptures(pos, Black, &ml)

	expected := 1
	got := len(ml)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMultiplePawnAttacks(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhitePawn, e2)
	pos.AddPiece(WhitePawn, f2)
	pos.AddPiece(WhitePawn, g2)
	pos.AddPiece(WhitePawn, h2)
	pawns := ^pos.EmptySquares()

	expected := bitboardFromCoordinates("d3", "e3", "f3", "g3", "h3")
	got := pawnAttacks(&pawns, White)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
