package aconcagua

import "testing"

func TestResetPVLine(t *testing.T) {
	pvLine := NewPvLine(100)
	pvLine.insert(*encodeMove(0, 8, quiet), &pvLine)

	pvLine.reset()

	expected := 0
	got := len(pvLine)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestInsert(t *testing.T) {
	pvLine := NewPvLine(100)
	move := encodeMove(0, 8, quiet)
	pvLine.insert(*move, &pvLine)

	expected := *move
	got := pvLine[0]

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestString(t *testing.T) {
	pvLine := NewPvLine(100)
	pvLine.insert(*encodeMove(0, 8, quiet), &pvLine)

	expected := "a1a2"
	got := pvLine.String()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
