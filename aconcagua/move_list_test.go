package aconcagua

import (
	"testing"
)

func TestAddMove(t *testing.T) {
	ml := newMoveList()
	move := encodeMove(0, 0, quiet)

	ml.add(*move)

	expected := 1
	got := ml.length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestSortMoves(t *testing.T) {
	ml := newMoveList()

	scores := []int{40, 200, 50, 120}
	ml.add(*encodeMove(0, 0, quiet))          // index 0
	ml.add(*encodeMove(0, 0, capture))        // index 1
	ml.add(*encodeMove(0, 0, epCapture))      // index 2
	ml.add(*encodeMove(0, 0, kingsideCastle)) // index 3

	ml.sort(scores)

	expected := []Move{*encodeMove(0, 0, capture), *encodeMove(0, 0, kingsideCastle), *encodeMove(0, 0, epCapture), *encodeMove(0, 0, quiet)}
	got := ml.moves

	if got[0] != expected[0] {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
