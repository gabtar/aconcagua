package aconcagua

import "testing"

func TestRookRelevantMask(t *testing.T) {
	expectedSquares := []string{"e2", "e3", "e5", "e6", "e7",
		"b4", "c4", "d4", "f4", "g4"}

	expected := bitboardFromCoordinates(expectedSquares...)
	got := rookMask(28)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopRelevantMask(t *testing.T) {
	expectedSquares := []string{"c2", "d3", "f5", "g6", "b7", "c6", "d5", "f3", "g2"}

	expected := bitboardFromCoordinates(expectedSquares...)
	got := bishopMask(28)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookAttacksWithBlockers(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteRook, c4)
	pos.AddPiece(WhiteBishop, b4)

	expectedSquares := []string{"c1", "c2", "c3", "c5", "c6", "c7", "c8",
		"b4", "d4", "e4", "f4", "g4", "h4"}

	expected := bitboardFromCoordinates(expectedSquares...)
	got := rooksAttacksWithBlockers(26, pos.pieces[White])

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopAttacksWithBlockers(t *testing.T) {
	expectedSquares := []string{"a1", "c1", "a3", "c3", "d4", "e5", "f6"}

	expected := bitboardFromCoordinates(expectedSquares...)
	got := bishopAttacksWithBlockers(9, Bitboard(1<<45)) // from b2 square, with blocker in f6

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestRookAttacksSquares(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(BlackRook, f8)
	pos.AddPiece(WhiteBishop, f6)
	pos.AddPiece(BlackKing, e8)

	expectedSquares := []string{"h8", "g8", "e8", "f7", "f6"}

	expected := bitboardFromCoordinates(expectedSquares...)
	got := rookAttacks(Bsf(bitboardFromCoordinates("f8")), pos.pieces[White]|pos.pieces[Black])

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBishopAttacksSquares(t *testing.T) {
	pos := NewPosition()
	pos.AddPiece(WhiteBishop, c5)
	pos.AddPiece(BlackKing, d4)
	pos.AddPiece(WhiteRook, f8)

	expectedSquares := []string{"a3", "b4", "a7", "b6", "d6", "e7", "f8", "d4"}

	expected := bitboardFromCoordinates(expectedSquares...)
	got := bishopAttacks(Bsf(bitboardFromCoordinates("c5")), pos.pieces[White]|pos.pieces[Black])

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
