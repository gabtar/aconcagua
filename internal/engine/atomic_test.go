package engine

import "testing"

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

func TestIsExplosion(t *testing.T) {
	testCases := []struct {
		move     *Move
		expected bool
	}{
		{encodeMove(0, 8, capture), true},
		{encodeMove(0, 8, quiet), false},
		{encodeMove(0, 8, epCapture), true},
		{encodeMove(0, 8, queenCapturePromotion), true},
		{encodeMove(0, 8, kingsideCastle), false},
	}

	for _, tc := range testCases {
		got := isExplosion(tc.move)
		if got != tc.expected {
			t.Errorf("Expected: %v, got: %v", tc.expected, got)
		}
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
	expected := pos.Hash

	at := NewAtomicPosition(*pos)
	move := encodeMove(c4, b5, capture)
	at.MakeMove(move)
	at.UnmakeMove(move)

	got := at.pos.Hash

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestEnPassantCaptureExplosions(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("6k1/1q6/3r1p2/2pP2p1/8/6P1/5PK1/8 w - c6 0 1")
	expected := pos.Hash

	at := NewAtomicPosition(*pos)
	move := encodeMove(d5, c6, epCapture)
	at.MakeMove(move)
	at.UnmakeMove(move)

	got := at.pos.Hash
	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestQueenCapturePromotion(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("5b2/1q2P3/2pR3p/2n3p1/6P1/4k2K/QN6/2r5 w - - 0 1")
	expected := pos.Hash

	at := NewAtomicPosition(*pos)
	move := encodeMove(f7, g8, queenCapturePromotion)

	at.MakeMove(move)
	at.UnmakeMove(move)

	got := at.pos.Hash
	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}

}

func TestIsLegal(t *testing.T) {
	testCases := []struct {
		name     string
		fen      string
		move     *Move
		expected bool
	}{
		{"King captures Knight", "8/2k5/2p5/8/4n3/5K2/1N6/8 w - - 0 1", encodeMove(f3, e4, capture), false},
		{"Capture explodes own king", "8/2k5/2pR3p/r5p1/4n1P1/7K/1N6/8 b - - 0 1", encodeMove(e4, d6, capture), false},
		{"Explosion far away from king", "5b2/1q2P3/2pR3p/2n3p1/6P1/4k2K/QN6/2r5 w - - 0 1", encodeMove(f7, g8, queenCapturePromotion), true},
		{"Ep capture creates a discovered check", "b4k2/8/8/2pP4/8/5K2/8/8 w - c6 0 1", encodeMove(d5, c6, epCapture), false},
		{"Deliver checkmate to enemy king", "rnb3k1/p4Npp/Bpp2p1n/q2p4/P5P1/2P5/5P1P/R1BQR1K1 w - - 0 14", encodeMove(e1, e8, quiet), true},
	}

	for _, tc := range testCases {
		pos := NewPosition()
		pos.LoadFromFenString(tc.fen)
		at := NewAtomicPosition(*pos)
		got := at.IsLegal(*tc.move)
		if got != tc.expected {
			t.Errorf("Expected: %v, got: %v", tc.expected, got)
		}
	}
}
