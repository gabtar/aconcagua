package aconcagua

import "testing"

func TestAdd(t *testing.T) {
	pv := newPrincipalVariation()
	branchPv := newPrincipalVariation()

	m0 := Move(0)
	m1 := Move(1)

	pv.insert(m0, branchPv)
	pv.insert(m1, pv)

	expected := 2
	got := len(*pv)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
