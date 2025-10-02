package aconcagua

import "testing"

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

func TestPositionBeforePieceMoved(t *testing.T) {
	positionBefore := encodePositionBefore(1, 0, 0, 0)

	expected := 1
	got := positionBefore.pieceMoved()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPositionBeforePieceCaptured(t *testing.T) {
	positionBefore := encodePositionBefore(1, 2, 0, 0)

	expected := 2
	got := positionBefore.pieceCaptured()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPositionBeforeEpTarget(t *testing.T) {
	positionBefore := encodePositionBefore(1, 2, 56, 0)

	expected := 56
	got := positionBefore.epTarget()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPositionBeforeRule50(t *testing.T) {
	positionBefore := encodePositionBefore(1, 2, 8, 50)

	expected := 50
	got := positionBefore.rule50()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
