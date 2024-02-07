package search

import (
	"testing"

	"github.com/gabtar/aconcagua/board"
)

func TestAdd(t *testing.T) {
	pv := newPrincipalVariation(2)

	m0 := board.Move(0)
	m1 := board.Move(1)

	pv.add(m0, 2)
	pv.add(m1, 1)
	pv.add(m1, 1)

	expected := 2
	got := len(pv.moves)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestClear(t *testing.T) {
	pv := newPrincipalVariation(5)

	pv.clear()

	expected := 0
	got := len(pv.moves)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
