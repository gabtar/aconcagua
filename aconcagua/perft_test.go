package aconcagua

import (
	"testing"
)

// Perft tests
// Data from https://gist.github.com/peterellisjones/8c46c28141c162d1d8a0f0badbc9cff9

func TestStandardPerft(t *testing.T) {
	testCases := []struct {
		name  string
		fen   string
		depth int
		moves int
	}{
		{"Perft 1", "r6r/1b2k1bq/8/8/7B/8/8/R3K2R b KQ - 3 2", 1, 8},
		{"Perft 2", "8/8/8/2k5/2pP4/8/B7/4K3 b - d3 0 3", 1, 8},
		{"Perft 3", "r1bqkbnr/pppppppp/n7/8/8/P7/1PPPPPPP/RNBQKBNR w KQkq - 2 2", 1, 19},
		{"Perft 4", "r3k2r/p1pp1pb1/bn2Qnp1/2qPN3/1p2P3/2N5/PPPBBPPP/R3K2R b KQkq - 3 2", 1, 5},
		{"Perft 5", "2kr3r/p1ppqpb1/bn2Qnp1/3PN3/1p2P3/2N5/PPPBBPPP/R3K2R b KQ - 3 2", 1, 44},
		{"Perft 6", "rnb2k1r/pp1Pbppp/2p5/q7/2B5/8/PPPQNnPP/RNB1K2R w KQ - 3 9", 1, 39},
		{"Perft 7", "2r5/3pk3/8/2P5/8/2K5/8/8 w - - 5 4", 1, 9},
		{"Perft 8", "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 3, 62379},
		{"Perft 9", "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10", 3, 89890},
		{"Perft 10", "3k4/3p4/8/K1P4r/8/8/8/8 b - - 0 1", 6, 1134888},
		{"Perft 11", "8/8/4k3/8/2p5/8/B2P2K1/8 w - - 0 1", 6, 1015133},
		{"Perft 12", "8/8/1k6/2b5/2pP4/8/5K2/8 b - d3 0 1", 6, 1440467},
		{"Perft 13", "5k2/8/8/8/8/8/8/4K2R w K - 0 1", 6, 661072},
		{"Perft 14", "3k4/8/8/8/8/8/8/R3K3 w Q - 0 1", 6, 803711},
		{"Perft 15", "r3k2r/1b4bq/8/8/8/8/7B/R3K2R w KQkq - 0 1", 4, 1274206},
		{"Perft 16", "r3k2r/8/3Q4/8/8/5q2/8/R3K2R b KQkq - 0 1", 4, 1720476},
		{"Perft 17", "2K2r2/4P3/8/8/8/8/8/3k4 w - - 0 1", 6, 3821001},
		{"Perft 18", "8/8/1P2K3/8/2n5/1q6/8/5k2 b - - 0 1", 5, 1004658},
		{"Perft 19", "4k3/1P6/8/8/8/8/K7/8 w - - 0 1", 6, 217342},
		{"Perft 20", "8/P1k5/K7/8/8/8/8/8 w - - 0 1", 6, 92683},
		{"Perft 21", "K1k5/8/P7/8/8/8/8/8 w - - 0 1", 6, 2217},
		{"Perft 22", "8/k1P5/8/1K6/8/8/8/8 w - - 0 1", 7, 567584},
		{"Perft 23", "8/8/2k5/5q2/5n2/8/5K2/8 b - - 0 1", 4, 23527},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pos := From(tc.fen)
			got := pos.Perft(tc.depth)
			if got != uint64(tc.moves) {
				t.Errorf("Expected: %v, got: %v", tc.moves, got)
			}
		})
	}
}

// TODO: test perft for chess960 to test make and unmake move. May need also to update the move encoding for castling in chess960
// Data from https://www.chessprogramming.org/Chess960_Perft_Results
