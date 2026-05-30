package engine

import "testing"

func TestAdd(t *testing.T) {
	ph := NewPositionHistory()
	ph.add(positionBefore(0), KQkq, 0)

	if ph.moveCount != 1 {
		t.Errorf("Expected: %v, got: %v", 1, ph.moveCount)
	}
}

func TestPop(t *testing.T) {
	ph := NewPositionHistory()
	ph.add(positionBefore(0), KQkq, 0)
	_, _ = ph.pop()

	if ph.moveCount != 0 {
		t.Errorf("Expected: %v, got: %v", 0, ph.moveCount)
	}
}
