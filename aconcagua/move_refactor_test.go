package aconcagua

import "testing"

// Tests that diferent moves produces the correct update on the board

func TestMoveFlagEncode(t *testing.T) {
	move := encodeMove(0, 8, capture)

	got := move.flag()
	expected := capture

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveCoordinates(t *testing.T) {
	move := encodeMove(0, 8, capture)

	got := move.from()
	expected := 0 // A1 square

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveString(t *testing.T) {
	move := encodeMove(0, 8, knightPromotion)

	got := move.String()
	expected := "a1a2n"

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
