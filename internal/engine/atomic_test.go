package engine

import (
	"fmt"
	"testing"
)

func TestEncodeExplodedPiece(t *testing.T) {
	piece := WhiteQueen
	square := e4

	ep := encodeExplodedPiece(square, piece)

	expectedSquare, expectedPiece := square, piece
	gotSquare, gotPiece := ep.decode()

	if gotSquare != expectedSquare {
		t.Errorf("Expected: %v, got: %v", expectedSquare, gotSquare)
	}
	if gotPiece != expectedPiece {
		t.Errorf("Expected: %v, got: %v", expectedPiece, gotPiece)
	}
}

func TestAddExplosionHistory(t *testing.T) {
	piece := WhiteQueen
	square := e4

	ep := encodeExplodedPiece(square, piece)
	e := Explosion{}
	e.add(ep)

	eh := NewExplosionHistory()
	eh.add(piece, square)
	eh.increment()

	got := eh.pop()

	if got.count != 1 {
		t.Errorf("Expected: %v, got: %v", 1, got.count)
	}
	if got.explodedPieces[0] != ep {
		t.Errorf("Expected: %v, got: %v", ep, got.explodedPieces[0])
	}
}

// Tests for MakeMove
// TODO: Need more tests cases

func TestCaptureExplodesSourrundingPieces(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("k7/2q1p3/2pr4/2b5/4N3/4K3/8/4R3 w - - 0 1")
	at := NewAtomicPosition(*pos)

	move := encodeMove(e4, d6, capture) // Knight captures a rook and explodes the bishop and queen
	expected := at.pos.ToFen()

	at.MakeMove(move)
	at.UnmakeMove(move)

	got := at.pos.ToFen()

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestPawnCaptureRemovesPawnsAndExplodedPieces(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("6k1/5ppp/1B6/np6/2P5/6P1/5P1P/6K1 w - - 0 1")

	at := NewAtomicPosition(*pos)
	move := encodeMove(c4, b5, capture)
	at.MakeMove(move)

	fmt.Println(at.pos.String()) // TODO: should also remove the pawn at b5 when its a pawn capture!!!

	t.Fail()
}
