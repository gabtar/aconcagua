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
		{"Give check to enemy king", "1k6/8/8/1bpP1Q2/8/8/6K1/8 w - c6 0 1", encodeMove(f5, f8, quiet), true},
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

func TestAtomicPerft(t *testing.T) {
	testCases := []struct {
		name  string
		fen   string
		depth int
		nodes int
	}{
		{"Perft 1", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 1, 20},
		{"Perft 2", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 3, 8902},
		{"Perft 3", "r4k2/1p5p/5pp1/PN1p4/4p3/4P3/1PPP1PPP/R1B1K2R b KQ - 1 13", 1, 18},
		{"Perft 4", "r2qkbnr/pp2p1pp/2n2p2/1B1p2N1/8/4P3/PPPP1PPP/R1B1K2R w KQkq - 0 8", 2, 805},
		{"Perft 5", "7k/6np/q1N5/1b6/4K3/8/8/R7 b - - 0 1", 1, 24},
		{"Perft 6", "1q5k/5ppp/8/8/8/8/5PPP/3R2K1 w - - 0 1", 2, 508},
		{"Perft 7", "8/8/6P1/3k1K2/8/8/8/8 w - - 0 1", 1, 8},
		{"Perft 7 - depth 3", "8/8/6P1/3k1K2/8/8/8/8 w - - 0 1", 3, 542},
		{"Perft 8", "1k6/8/8/1bpP1Q2/8/8/6K1/8 w - c6 0 1", 1, 30},
		{"Perft - Black Checkmated", "8/2N4p/1p1k2p1/3R1p2/8/8/1PP2PPP/7R w K - 0 23", 1, 0},
	}

	for _, tc := range testCases {
		pos := NewPosition()
		pos.LoadFromFenString(tc.fen)
		at := NewAtomicPosition(*pos)
		got := PerftVariant(tc.depth, at)
		if got != uint64(tc.nodes) {
			t.Errorf("Expected: %v, got: %v", tc.nodes, got)
		}
	}
}

// BUG: Tests againt fairly stockfish output
func TestLegalMovesDifference(t *testing.T) {
	legalMoves := []string{
		"d5d6", "d5c6", "f5b1", "f5f1", "f5c2", "f5f2", "f5d3", "f5f3", "f5h3", "f5e4",
		"f5f4", "f5g4", "f5e5", "f5g5", "f5h5", "f5e6", "f5f6", "f5g6", "f5d7", "f5f7",
		"f5h7", "f5c8", "f5f8", "g2h3", "g2g1", "g2h1", "g2f2", "g2h2", "g2f3", "g2g3",
	}

	pos := NewPosition()
	pos.LoadFromFenString("1k6/8/8/1bpP1Q2/8/8/6K1/8 w - c6 0 1")
	at := NewAtomicPosition(*pos)

	ml := NewMoveList()
	pd := at.pos.generatePositionData()
	at.GenerateCaptures(ml, &pd)
	at.GenerateNonCaptures(ml, &pd)

	// Find the one is not generating
	for _, move := range legalMoves {
		found := false
		for i := range ml.length {
			if ml.moves[i].String() == move {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Move %s not generated", move)
		}
	}
}
