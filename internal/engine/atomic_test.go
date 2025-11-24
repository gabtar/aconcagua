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

func TestCaptureExplodesSourrundingPieces2(t *testing.T) {
	pos := NewPosition()
	pos.LoadFromFenString("rnbqkb1r/1p2p1pp/p1p1Np1n/8/3pP3/2N4P/PPPP1PP1/R1BQKB1R w KQkq - 0 7")
	at := NewAtomicPosition(*pos)

	move := encodeMove(e6, g7, capture) // Knight captures a pawn and explodes 3 pieces (rook, bishop and knight)
	expected := "rnbqk3/1p2p2p/p1p2p2/8/3pP3/2N4P/PPPP1PP1/R1BQKB1R b KQq - 0 7"

	at.MakeMove(move)

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
		{"Explode removes the pinner piece", "3r2nr/1p1kp1Rp/3p1p2/4P3/1P6/7P/P1P5/2KR4 b - - 0 19", encodeMove(d6, e5, capture), false},
		{"In check, but atomic explosion produces checkmate", "r1R1k1nr/p2nB2p/4p2b/1p3p2/3P1Pp1/7P/PP4PK/4R3 b kq - 1 17", encodeMove(g4, h3, capture), true},
		{"In check, but atomic explosion removes checker", "r1b1k1nr/p2p2pp/4pp2/1p2n3/1bP5/3P4/PP2PPPP/RNBQKB1R w KQkq - 1 9", encodeMove(c4, b5, capture), true},
		{"Put the king in check by rook. Cannot recapture/explode, due to explosion removes both kings", "8/4B3/8/p7/P6p/4p1PP/3k4/4RK2 b - - 1 50", encodeMove(d2, e2, quiet), true},
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
		{"Perft 9", "1k6/8/8/1bpP4/5Q2/8/6K1/8 b - - 1 1", 1, 4},
		{"Perft 10", "rn1qkb1r/pp2p1pp/2p2n2/3p4/5P2/2N1P3/PPPPB1PP/R1BQK2R b KQkq - 2 7", 2, 864},
		{"Perft - Black Checkmated", "8/2N4p/1p1k2p1/3R1p2/8/8/1PP2PPP/7R w K - 0 23", 1, 0},
		{"Perft 11", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 4, 197326},
		{"Perft 12", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 5, 4864979},
		{"Perft 13", "r3r1k1/pp5p/n4pp1/1Q6/6PP/2P5/P7/2KR3R w - - 0 22", 3, 47272},
		{"Perft 14", "2kr2nr/1p2p1Rp/3p1p2/8/1P2P3/7P/P1P5/2KR4 b - - 3 18", 4, 192507},
		{"Perft 14 - d6d5 c1d2 e7e6", "2kr2nr/1p4Rp/4pp2/3p4/1P2P3/7P/P1PK4/3R4 w - - 0 20", 1, 33}, // Pined by explosion. If capture explodes produces a discovery check. And has a killer explosion. And has a killer explosion. And has a killer explosion...
		{"Perft 15", "rnbqkb1r/1p2p1pp/p1p1Np1n/3p4/4P3/2N4P/PPPP1PP1/R1BQKB1R b KQkq - 1 6", 3, 23192},
		{"Perft 16", "5k2/8/p4Q1p/2b1p1p1/4P3/2P3qP/P4PP1/3r1NK1 b - - 3 40", 5, 591262},
		{"Perft 17", "r3k1nr/p2nB2p/4p1pb/1p3p2/3P1P2/7P/PP4P1/2R1R1K1 b kq - 1 15", 5, 12119427},
		{"Perft 18", "r1b1kbnr/pp1p2pp/4pp2/4n3/8/2P5/PP1PPPPP/RNBQKB1R w KQkq - 1 7", 5, 8670840},
		{"Perft 19", "r1b1k2r/ppB1p3/n1p2pp1/3p4/3PPP1p/P6B/2P5/3QK2R b Kkq - 1 14", 4, 803459},
		{"Perft 20", "8/4B3/8/p6p/P7/4p1PP/3k4/4R1K1 b - - 1 49", 3, 688},
		{"Perft 20 h5h4", "8/4B3/8/p7/P6p/4p1PP/3k4/4R1K1 w - - 0 50", 2, 98},

		// Higher depth tests. To ensure perft is correct. Need to run with -timeout 30m flag to avoid 10m limit by default. Disabled by default.
		// Just random positions from lichess atomic games compared against Fairy Stockfish perft output
		// {"Perft 21", "5kr1/p7/2p1p2p/3pPp2/3Pb3/Q5PP/PPP5/R1B3K1 b - - 4 18", 7, 776815556},                   // PASS - 197.314s
		// {"Perft 22", "rnbq1knr/ppN3pp/2p2p2/3p4/1b5P/2P1P3/PP3PP1/R1BQKB1R b KQ - 0 9", 6, 1908167278},        // PASS - 556.814s
		// {"Perft 23", "2r2R2/p1P3B1/8/4p2p/7P/1P1P3k/P5bK/8 w - - 2 38", 7, 2109807698},                        // PASS - 402.215s
		// {"Perft 24", "5Bk1/pp3n2/6P1/5p2/4pP2/1PP5/P7/2K5 b - - 0 27", 8, 897474719},                          // PASS - 183.163s
		// {"Perft 25", "r1b2knr/2N2p2/n1p3pp/p2pp3/1b3N2/2P1PP2/PP1P2PP/R1B2RK1 w - - 1 14", 6, 572942491},      // PASS - 183.751s
		// {"Perft 26", "r4B1r/pp5p/2k2pp1/5b2/1PP5/8/P4PPP/2KR3R w - - 1 22", 6, 382358360},                     // PASS - 99.313s
		// {"Perft 27", "rnbqkb1r/pp1pp1pp/2p2p2/8/8/2N1P3/PPPP1PPP/R1BQKB1R b KQkq - 1 5", 6, 342465968},        // PASS - 112.222s
		// {"Perft 28", "r1b1kb1r/pppp1Np1/2n2n2/5p1p/1P1Pp2q/4PPP1/P1P4P/RNBQKB1R b KQkq - 0 8", 6, 1918375372}, // PASS - 593.412s
		// {"Perft 29", "r2qkb1r/p5p1/1p3p2/1P1pp1Pp/P1pPPP1P/2P5/5K2/RN4NR b kq - 0 15", 6, 278730039},          // PASS - 92.023s
		// {"Perft 30", "k3r3/1R6/P4pp1/3p2Pp/2pP3P/2P2N2/5K2/8 w - - 3 28", 7, 427909792},                       // PASS - 86.521s
		// {"Perft 31", "5r2/1p4p1/4N2p/P7/8/6P1/3p2kP/3K4 b - - 4 36", 7, 501841475},                            // PASS -  82.682s
		// {"Perft 32", "r1bqk1nr/1p3pp1/2p4p/P7/1b1pP3/2N2PPN/7P/R1BQKB1R b KQkq - 0 12", 6, 3605073073},        // PASS - 900.372s
		// {"Perft 33", "rnbqkbnr/ppppppp1/7p/8/4P3/7N/PPPP1PPP/RNBQKB1R b KQkq - 0 2", 6, 267397815},            // PASS - 95.796s
		// {"Perft 34", "rn1qkb1r/2p3p1/5p2/pp1Pp2p/2P3P1/N6P/PPbB1P2/R1Q1KB1R b KQkq - 2 12", 6, 1770446821},    // PASS - 568.027s
		// {"Perft 35", "r2qk2r/2p3p1/5p2/p2Pp2p/1b4P1/N3B2P/PP2KP2/R1Qb1B1R w kq - 5 16", 7, 2147276275},        // PASS - 642.148s
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

// func TestDivide(t *testing.T) {
// 	pos := NewPosition()
// 	pos.LoadFromFenString("2r2R2/p1P5/7B/7p/4p2P/1P1P3k/P5bK/8 w - - 0 39")
// 	at := NewAtomicPosition(*pos)
//
// 	for str := range strings.SplitSeq(DivideVariant(2, at), ",") {
// 		fmt.Println(str)
// 	}
//
// 	t.Fail()
// }

// BUG: Tests againt fairly stockfish output
func TestLegalMovesDifference(t *testing.T) {
	legalMoves := []string{
		"a2a3", "b3b4", "d3d4", "a2a4", "d3e4", "h6c1", "h6d2", "h6e3",
		"h6f4", "h6g5", "h6g7", "f8f1", "f8f2", "f8f3", "f8f4", "f8f5",
		"f8f6", "f8f7", "f8c8", "f8d8", "f8e8", "f8g8", "f8h8", "h2g1",
		"h2g3",
	}
	pos := NewPosition()
	pos.LoadFromFenString("2r2R2/p1P5/7B/7p/4p2P/1P1P3k/P5bK/8 w - - 0 39")
	at := NewAtomicPosition(*pos)

	ml := NewMoveList()
	pd := at.pos.generatePositionData()
	at.GenerateCaptures(ml, &pd)
	at.GenerateNonCaptures(ml, &pd)

	// for i := range ml.length {
	// 	fmt.Printf("Move generated: %s, isLegal: %t\n", ml.moves[i].String(), at.IsLegal(ml.moves[i]))
	// }
	// fmt.Println("Total moves: ", ml.length)

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
	// Check out if there is an extra move in the move list
	for i := range ml.length {
		exists := false
		for _, move := range legalMoves {
			if ml.moves[i].String() == move {
				exists = true
			}
		}
		if !exists {
			t.Errorf("Move %s not expected", ml.moves[i].String())
		}
	}
	// t.Fail()
}
