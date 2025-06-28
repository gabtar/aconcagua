package aconcagua

import "testing"

func TestResetPVLine(t *testing.T) {
	pvLine := PVLine{}

	pvLine.length = 100
	pvLine.reset()

	expected := 0
	got := pvLine.length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPrepend(t *testing.T) {
	pvLine := PVLine{
		moves:  []Move{NoMove},
		length: 1,
	}
	move := encodeMove(0, 8, quiet)

	pvLine.prepend(*move, &pvLine)

	expected := 2
	got := pvLine.length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

	expectedMove := *move
	gotMove := pvLine.moves[0]

	if gotMove != expectedMove {
		t.Errorf("Expected: %v, got: %v", expectedMove, gotMove)
	}
}

func TestString(t *testing.T) {
	pvLine := PVLine{
		moves:  []Move{*encodeMove(0, 8, quiet)},
		length: 1,
	}

	expected := "a1a2"
	got := pvLine.String()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestNewPVTable(t *testing.T) {
	pvTable := NewPVTable(1)

	expected := 1
	got := len(pvTable)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestResetPVTable(t *testing.T) {
	pvTable := NewPVTable(1)

	pvTable[0].moves = []Move{*encodeMove(0, 8, quiet)}
	pvTable[0].length = 1

	pvTable.reset(0)

	expected := 0
	got := pvTable[0].length

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
