package engine

import "testing"

func TestAdd(t *testing.T) {
	ph := NewPositionHistory()
	ph.add(positionBefore(0), KQkq, 0)

	if ph.currentIndex != 1 {
		t.Errorf("Expected: %v, got: %v", 1, ph.currentIndex)
	}
}

func TestPop(t *testing.T) {
	ph := NewPositionHistory()
	ph.add(positionBefore(0), KQkq, 0)
	_, _ = ph.pop()

	if ph.currentIndex != 0 {
		t.Errorf("Expected: %v, got: %v", 0, ph.currentIndex)
	}
}

func TestRepetitionCount(t *testing.T) {
	ph := NewPositionHistory()
	ph.add(positionBefore(0), KQkq, 5)
	ph.add(positionBefore(0), KQkq, 5)

	if ph.repetitionCount(5) != 2 {
		t.Errorf("Expected: %v, got: %v", 2, ph.repetitionCount(0))
	}
}
