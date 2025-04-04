package aconcagua

import "testing"

func TestAdd(t *testing.T) {
	pv := newPV()
	branchPv := newPV()

	m0 := encodeMove(0, 0, quiet)
	m1 := encodeMove(1, 0, capture)

	pv.insert(*m0, branchPv)
	pv.insert(*m1, pv)

	expected := 2
	got := len(*pv)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
