package aconcagua

import "testing"

func TestMoveSelectorHasNextMove(t *testing.T) {
	pos := EmptyPosition()
	hashMove := encodeMove(0, 0, quiet)
	killers := Killer{NoMove, NoMove}
	ms := NewMoveGenerator(pos, hashMove, &killers[0], &killers[1], &HistoryMoves{})

	got := ms.nextMove()

	if got != NoMove {
		t.Errorf("Expected: %v, got: %v", NoMove, got)
	}
}

func TestMoveSelectorNotHasNextMove(t *testing.T) {
	pos := EmptyPosition()
	hashMove := NoMove
	killers := Killer{NoMove, NoMove}
	ms := NewMoveGenerator(pos, &hashMove, &killers[0], &killers[1], &HistoryMoves{})
	ms.stage = EndStage

	expected := NoMove
	got := ms.nextMove()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveSelectorCreatesCaptures(t *testing.T) {
	pos := From("1b4k1/5pp1/3r3p/4P3/5PN1/3RK3/8/8 w - - 0 1") // Only 3 captures
	ms := NewMoveGenerator(pos, nil, nil, nil, nil)
	move := NoMove
	ms.hashMove = &move

	expected := *encodeMove(36, 43, capture) // Best capture. Pawn takes rook
	got := ms.nextMove()

	// fmt.Println(got.String())

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveSelectorCreatesNonCaptures(t *testing.T) {
	pos := From("1b4k1/5pp1/3r3p/4P3/5PN1/3RK3/8/8 w - - 0 1") // Only 3 captures
	noMove := NoMove
	ms := NewMoveGenerator(pos, &noMove, &noMove, &noMove, &HistoryMoves{})
	ms.stage = FirstKillerStage // NOTE: Non captures are generated in killers stage to validate legaliy of killers
	move := NoMove
	ms.hashMove = &move

	ms.nextMove()

	expected := 16 // 17 - 1 (the one picked)
	got := len(ms.nonCaptures)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestMoveSelectorGetsAllMoves(t *testing.T) {
	pos := From("1b4k1/5pp1/3r3p/4P3/5PN1/3RK3/8/8 w - - 0 1") // 3 captures + 17 non capt
	hashMove := NoMove

	ml := NewMoveList(20)
	pos.generateNonCaptures(&ml)
	killers := Killer{ml[0], ml[5]}

	ms := NewMoveGenerator(pos, &hashMove, &killers[0], &killers[1], &HistoryMoves{})
	for ms.nextMove() != NoMove {
	}

	expected := 20
	got := ms.moveNumber

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
