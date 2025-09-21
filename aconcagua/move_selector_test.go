package aconcagua

import "testing"

func TestMoveSelectorHasNextMove(t *testing.T) {
	pos := EmptyPosition()
	hashMove := encodeMove(0, 0, quiet)
	killers := Killer{NoMove, NoMove}
	ms := NewMoveSelector(pos, hashMove, &killers[0], &killers[1], &HistoryMovesTable{})

	got := ms.nextMove()

	if got != NoMove {
		t.Errorf("Expected: %v, got: %v", NoMove, got)
	}
}

func TestMoveSelectorNotHasNextMove(t *testing.T) {
	pos := EmptyPosition()
	hashMove := NoMove
	killers := Killer{NoMove, NoMove}
	ms := NewMoveSelector(pos, &hashMove, &killers[0], &killers[1], &HistoryMovesTable{})
	ms.stage = EndStage

	expected := NoMove
	got := ms.nextMove()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveSelectorCreatesCaptures(t *testing.T) {
	pos := NewPositionFromFen("1b4k1/5pp1/3r3p/4P3/5PN1/3RK3/8/8 w - - 0 1") // Only 3 captures
	ms := NewMoveSelector(pos, nil, nil, nil, nil)
	move := NoMove
	ms.hashMove = &move

	expected := *encodeMove(36, 43, capture) // Best capture. Pawn takes rook
	got := ms.nextMove()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveSelectorCreatesNonCaptures(t *testing.T) {
	pos := NewPositionFromFen("1b4k1/5pp1/3r3p/4P3/5PN1/3RK3/8/8 w - - 0 1") // Only 3 captures
	noMove := NoMove
	ms := NewMoveSelector(pos, &noMove, &noMove, &noMove, &HistoryMovesTable{})
	ms.stage = FirstKillerStage // NOTE: Non captures are generated in killers stage to validate legaliy of killers
	move := NoMove
	ms.hashMove = &move

	ms.pd = ms.pos.generatePositionData()
	ms.nextMove()

	expected := 16 // 17 - 1 (the one picked)
	got := len(ms.nonCaptures)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveSelectorGetsAllMoves(t *testing.T) {
	pos := NewPositionFromFen("1b4k1/5pp1/3r3p/4P3/5PN1/3RK3/8/8 w - - 0 1") // 3 captures + 17 non capt
	hashMove := NoMove

	ml := NewMoveList(20)
	pd := pos.generatePositionData()
	pos.generateNonCaptures(&ml, &pd)
	killers := Killer{ml[0], ml[5]}

	ms := NewMoveSelector(pos, &hashMove, &killers[0], &killers[1], &HistoryMovesTable{})
	for ms.nextMove() != NoMove {
	}

	expected := 20
	got := ms.moveNumber

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
