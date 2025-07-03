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

// Tests for new move list
func TestNewMoveList(t *testing.T) {
	ml := NewMoveList(10)

	expected := 10
	got := cap(ml)

	if expected != got {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveListAdd(t *testing.T) {
	ml := NewMoveList(10)
	ml.add(*encodeMove(8, 16, quiet)) // a2a3

	expected := "a2a3"
	got := ml[0].String()

	if expected != got {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveListPick(t *testing.T) {
	ml := NewMoveList(10)
	ml.add(*encodeMove(8, 16, quiet)) // a2a3

	expected := "a2a3"
	got := ml.pickFirst().String()

	if expected != got {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

	if len(ml) != 0 {
		t.Errorf("Expected lenght: %v, got: %v", 0, len(ml))
	}
}

func TestMoveListPickWithNoMoves(t *testing.T) {
	ml := NewMoveList(10)

	expected := NoMove
	got := *ml.pickFirst()

	if expected != got {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveListSort(t *testing.T) {
	ml := NewMoveList(3)
	move1 := encodeMove(0, 0, quiet)
	move2 := encodeMove(0, 8, capture)
	move3 := encodeMove(0, 16, quiet)

	ml.add(*move1)
	ml.add(*move2)
	ml.add(*move3)

	expected := []Move{*move2, *move1, *move3}
	ml.sort([]int{2, 3, 1})

	if ml[0] != expected[0] && ml[1] != expected[1] && ml[2] != expected[2] && ml[3] != expected[3] {
		t.Errorf("Expected: %v, got: %v", expected, ml)
	}
}
