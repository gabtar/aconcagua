package engine

import (
	"strings"
	"testing"
)

func TestZobristIncrementalUpdateOnMakeMove(t *testing.T) {
	zobristTestCases := []struct {
		name    string
		fromFen string
		move    string
	}{
		{"Double Pawn Push", "pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "e2e4"},
		{"Pawn Push", "pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "e2e3"},
		{"Kingside castle", "r4rk1/5ppp/1p2pq2/nP1p4/P7/1B1P1N2/3QPPPP/2R1K2R w KQ - 0 1", "e1h1"},
		{"Black Queenside castle", "r3k1nr/pppq1pbp/2npb1p1/4p3/4P3/2NPBN2/PPP1BPPP/R2Q1RK1 b kq - 0 1", "e8c8"},
		{"Capture", "r5k1/5pp1/4p2p/2Np4/1r1P4/pN2P1P1/Qq3PP1/R5K1 w - - 4 30", "a2b2"},
		{"En passant capture", "6k1/8/4p3/4Pp2/1pP2P2/8/8/6K1 b - c3 0 1", "b4c3"},
		{"Reset en passant capture sq", "6k1/8/4p3/4Pp2/1pP2P2/8/8/6K1 b - c3 0 1", "g8f7"},
		{"Knight promotion", "6k1/2P5/3Np3/4Pp2/4nP2/8/1p6/6K1 w - f6 0 1", "c7c8n"},
		{"Queen capture promotion", "6k1/2P5/3Np3/4Pp2/4nP2/8/1p6/2R3K1 b - - 0 1", "b2c1q"},
	}

	for _, tc := range zobristTestCases {
		pos := NewPosition()
		t.Run(tc.name, func(t *testing.T) {
			pos.LoadFromFenString(tc.fromFen)

			ml := NewMoveList()
			pd := pos.generatePositionData()
			pos.generateCaptures(ml, &pd)
			pos.generateNonCaptures(ml, &pd)

			for moveNumber := range ml.moves {
				if ml.moves[moveNumber].String() == tc.move {
					pos.MakeMove(&ml.moves[moveNumber])
					break
				}
			}

			expected := zobristHashKeys.fullZobristHash(pos)
			got := pos.Hash

			if got != expected {
				t.Errorf("%s Expected: %d, Got: %d", tc.name, expected, got)
			}

			// Zobrist pawn hash test
			expected = zobristHashKeys.pawnHash(pos)
			got = pos.PawnHash

			if got != expected {
				t.Errorf("%s Expected: %d, Got: %d", tc.name, expected, got)
			}
		})
	}
}

func TestZobristIncrementalUpdateOnUnmakeMove(t *testing.T) {
	zobristTestCases := []struct {
		name    string
		fromFen string
		move    string
	}{
		{"Double Pawn Push", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "e2e4 e7e5 b1c3 g8f6"},
		{"Pawn Push", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "e2e3 d7d6 c2c4 d5c4"},
		{"Kingside castle", "r4rk1/5ppp/1p2pq2/nP1p4/P7/1B1P1N2/3QPPPP/2R1K2R w KQ - 0 1", "e1h1 a5b3 d2c2 b3c1 f1c1 h7h5"},
		{"Black Queenside castle", "r3k1nr/pppq1pbp/2npb1p1/4p3/4P3/2NPBN2/PPP1BPPP/R2Q1RK1 b kq - 0 1", "e8c8 c3d5 e6d5 e4d5 c6d5 f3d4 e5d4"},
		{"Capture", "r5k1/5pp1/4p2p/2Np4/1r1P4/pN2P1P1/Qq3PP1/R5K1 w - - 4 30", "a2b2 a3b2 a1a8 g8h7 g3g4 b2a1r"},
		{"En passant capture", "6k1/8/4p3/4Pp2/1pP2P2/8/8/6K1 b - c3 0 1", "b4c3 g1f2 c3c2 f2e2 c2c1q"},
		{"Reset en passant capture sq", "6k1/8/4p3/4Pp2/1pP2P2/8/8/6K1 b - c3 0 1", "g8f7 c4c5 f7e7 c4c3"},
		{"Knight promotion", "6k1/2P5/3Np3/4Pp2/4nP2/8/1p6/6K1 w - f6 0 1", "c7c8n b2b1q g1h2 e4d6"},
		{"Queen capture promotion", "6k1/2P5/3Np3/4Pp2/4nP2/8/1p6/2R3K1 b - - 0 1", "b2c1q g1h2 e4g3 h2g3"},
	}

	for _, tc := range zobristTestCases {
		pos := NewPosition()
		t.Run(tc.name, func(t *testing.T) {
			pos.LoadFromFenString(tc.fromFen)
			expected := zobristHashKeys.fullZobristHash(pos)

			// Make the moves and keep track on a pv line to be able to unmake them
			movesMade := pvLine{}
			for move := range strings.SplitSeq(tc.move, " ") {
				ml := NewMoveList()
				pd := pos.generatePositionData()
				pos.generateCaptures(ml, &pd)
				pos.generateNonCaptures(ml, &pd)

				for moveNumber := range ml.moves {
					if ml.moves[moveNumber].String() == move {
						pos.MakeMove(&ml.moves[moveNumber])
						movesMade.insert(ml.moves[moveNumber], &movesMade)
						break
					}
				}
			}

			// Unmake the moves
			for move := range movesMade {
				pos.UnmakeMove(&movesMade[move])
			}

			got := pos.Hash // after make and unmake moves

			if got != expected {
				t.Errorf("%s Expected: %d, Got: %d", tc.name, expected, got)
			}

			// Zobrist pawn hash test
			expected = zobristHashKeys.pawnHash(pos)
			got = pos.PawnHash

			if got != expected {
				t.Errorf("%s Expected: %d, Got: %d", tc.name, expected, got)
			}
		})
	}
}

func TestPawnHash(t *testing.T) {
	pos := NewPosition()
	pos2 := NewPosition()

	pos.LoadFromFenString("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	pos2.LoadFromFenString("8/pppppppp/8/8/8/8/PPPPPPPP/8 w - - 0 1")

	pawnHash := zobristHashKeys.pawnHash(pos)
	pawnHash2 := zobristHashKeys.pawnHash(pos2)

	if pawnHash != pawnHash2 {
		t.Errorf("Expected: %v, Got: %v", pawnHash, pawnHash2)
	}
}
