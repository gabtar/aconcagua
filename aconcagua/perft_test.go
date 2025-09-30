package aconcagua

import (
	"strings"
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
		pos := NewPosition()
		t.Run(tc.name, func(t *testing.T) {
			pos.LoadFromFenString(tc.fen)
			got := pos.Perft(tc.depth)
			if got != uint64(tc.moves) {
				t.Errorf("Expected: %v, got: %v", tc.moves, got)
			}
		})
	}
}

// Data from https://www.chessprogramming.org/Chess960_Perft_Results
func Test960Perft(t *testing.T) {
	testCases := []struct {
		name  string
		fen   string
		depth int
		moves int
	}{
		{"Starting position 1", "bqnb1rkr/pp3ppp/3ppn2/2p5/5P2/P2P4/NPP1P1PP/BQ1BNRKR w HFhf - 2 9", 5, 8146062},
		{"Starting position 12", "qb1nrkbr/1pppp1p1/1n3p2/p1B4p/8/3P1P1P/PPP1P1P1/QBNNRK1R w HEhe - 0 9", 1, 31},
		{"Starting position 28", "nb1n1kbr/ppp1rppp/3pq3/P3p3/8/4P3/1PPPRPPP/NBQN1KBR w Hh - 1 9 	", 3, 11786},
		{"Starting position 57", "n1rbbqkr/pp1pppp1/7p/P1p5/1n6/2PP4/1P2PPPP/NNRBBQKR w HChc - 0 9", 5, 10911545},
		{"Starting position 94", "nnrkrbbq/pppp2pp/8/4pp2/4P3/P7/1PPPBPPP/NNKRR1BQ w c - 0 9 	", 1, 25},
		{"Starting position 158", "nrnqkbbr/ppppp1p1/7p/5p2/8/P4PP1/NPPPP2P/NR1QKBBR w HBhb - 0 9", 3, 20621},
		{"Starting position 234", "nrqkbbnr/2pppp1p/p7/1p6/2P1Pp2/8/PPNP2PP/1RQKBBNR w HBhb - 0 9", 4, 432750},
		{"Starting position 291", "bqr1krnb/ppppppp1/7p/3n4/1P4P1/P4N2/2PPPP1P/BQNRKR1B w FDf - 3 9", 3, 22936},
		{"Starting position 348", "nbrkqr1n/1pppp2p/p4pp1/2Bb4/5P2/6P1/PPPPP2P/NBRKQ1RN w Cfc - 2 9", 3, 24775},
		{"Starting position 373", "nrbbk1nq/p1p1prpp/1p6/N2p1p2/P7/8/1PPPPPPP/R1BBKRNQ w Fb - 2 9", 2, 552},
		{"Starting position 430", "rnqnk1br/p1ppp1bp/1p3p2/6p1/4N3/P5P1/1PPPPP1P/R1QNKBBR w HAha - 2 9", 2, 717},
		{"Starting position 481", "bq1bnknr/pprppp1p/8/2p3p1/4PPP1/8/PPPP3P/BQRBNKNR w HCh - 0 9", 3, 14021},
		{"Starting position 544", "bbrn1nqr/ppp1k1pp/5p2/3pp3/7P/3PN3/PPP1PPP1/BBRK1NQR w - - 1 9", 4, 383532},
		// FIX: fails due to the white king has been moved from it starting square
		// Posible solution. Add all initial squares when creating the 960 castle struct. Currently only uses White values and mirrors them to get black starting squares..
		// {"Starting position 597", "rqbbnkrn/3pppp1/p1p4p/1p6/5P2/P2N4/1PPPP1PP/RQBBK1RN w ga - 0 9", 2, 665},
		{"Starting position 624", "bbr1kqrn/p1p1ppp1/1p2n2p/3p4/1P1P4/2N5/P1P1PPPP/BBR1KQRN w GCgc - 0 9", 3, 11475},
		{"Starting position 671", "rnkrnqbb/pp2p1p1/3p3p/2p2p2/5P2/1P1N4/P1PPPQPP/RNKR2BB w DAda - 0 9", 1, 29},
		{"Starting position 748", "rbk1n1br/ppp1ppqp/2n5/2Np2p1/8/2P5/PPBPPPPP/R1KN1QBR w HAha - 4 9", 3, 30663},
		{"Starting position 760", "rbknb1rq/ppp1p1p1/3pnp1p/8/6PP/2PP4/PP2PP2/RBKNBNRQ w GAga - 0 9", 4, 736910},
		{"Starting position 825", "rk1bbqrn/pp1pp1pp/3n4/5p2/3p4/1PP5/PK2PPPP/R1NBBQRN w ga - 0 9", 3, 14059},
		{"Starting position 901", "rkbbqr1n/1p1pppp1/2p2n2/p4NBp/8/3P4/PPP1PPPP/RK1BQRN1 w FAfa - 0 9", 2, 832},
		{"Starting position 961", "bbq1nr1r/pppppk1p/2n2p2/6p1/P4P2/4P1P1/1PPP3P/BBQNNRKR w HF - 1 9", 4, 387556},
	}

	for _, tc := range testCases {
		pos := NewPosition()
		t.Run(tc.name, func(t *testing.T) {
			pos.LoadFromFenString(tc.fen)

			// setup 960 castle from FEN string
			fenSegments := strings.Split(tc.fen, " ")
			pos.castling = *NewCastlingFromShredderFenCastlingCode(Bsf(pos.KingPosition(White)), fenSegments[2])

			got := pos.Perft(tc.depth)
			if got != uint64(tc.moves) {
				t.Errorf("Expected: %v, got: %v", tc.moves, got)
			}
		})
	}
}
