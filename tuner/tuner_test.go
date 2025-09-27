package tuner

import (
	"testing"

	"github.com/gabtar/aconcagua/aconcagua"
)

func TestEvaluation(t *testing.T) {
	testCases := []struct {
		name string
		fen  string
	}{
		{"Eval 1", "8/8/2k5/5q2/5n2/8/5K2/8 b - - 0 1"},
		{"Eval 2", "5rk1/1bn2ppp/8/1P6/2P2P2/3NPQ2/3PK3/8 w - - 0 1"},
		{"Eval 3", "5rk1/8/8/8/8/8/1R2K3/8 w - - 0 1"},
		{"Eval 4", "2b1k3/7r/4p3/3pNN2/8/2p1R3/P4PPK/2q5 w - - 0 36"},
		{"Eval 5", "7Q/ppq2k2/3bpnr1/3p4/3P4/2P5/PP3P2/1RB1K2R w K - 1 22"},
		{"Eval 6", "6k1/5p2/2p5/1p3B1p/1P2P1P1/4q1P1/1R4PK/8 b - - 0 40"},
		{"Eval 7", "8/2p5/8/p1p1K3/P1P5/8/4k3/8 w - - 3 40"},
		{"Eval 8", "6k1/7p/2p2Pp1/p1pr2N1/8/3P3P/P5P1/3R3K w - - 1 38"},
		{"Eval 9", "n1q3k1/pp3pbp/3p2p1/3P4/2P5/P2p1PP1/1P2Q2P/R4RK1 w - - 0 23"},
		{"Eval 10", "1n1b2k1/5ppp/1p6/1B6/8/1P2P3/P4PPP/6K1 b - - 1 26"},
		{"Eval 11", "3k4/3p4/8/K1P4r/8/8/8/8 b - - 0 1"},
		{"Eval 12", "8/8/4k3/8/2p5/8/B2P2K1/8 w - - 0 1"},
		{"Eval 13", "8/8/1k6/2b5/2pP4/8/5K2/8 b - d3 0 1"},
		{"Eval 14", "3k4/8/8/8/8/8/8/R3K3 w Q - 0 1"},
		{"Eval 15", "r3k2r/1b4bq/8/8/8/8/7B/R3K2R w KQkq - 0 1"},
		{"Eval 16", "r3k2r/8/3Q4/8/8/5q2/8/R3K2R b KQkq - 0 1"},
		{"Eval 17", "2K2r2/4P3/8/8/8/8/8/3k4 w - - 0 1"},
		{"Eval 18", "8/8/1P2K3/8/2n5/1q6/8/5k2 b - - 0 1"},
		{"Eval 19", "4k3/1P6/8/8/8/8/K7/8 w - - 0 1"},
		{"Eval 20", "8/P1k5/K7/8/8/8/8/8 w - - 0 1"},
		{"Eval 21", "rnb1kbnr/ppp1ppp1/8/q5B1/8/2NPQN2/PPP2P1P/R3KB1q w Qkq - 0 0"},
		{"Eval 22", "2k5/8/8/8/8/4K3/8/8 w - - 0 1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pos := aconcagua.NewPositionFromFen(tc.fen)
			staticEval := pos.Evaluate(aconcagua.NewPawnHashTable(1))
			params := GetEvaluationParams()
			attr := generatePositionWeights(pos.ToFen())

			got := int(EvaluatePosition(params, attr))

			// Always return evaluation from white's perspective
			if pos.Turn == aconcagua.Black {
				got = -got
			}

			if got != staticEval {
				t.Errorf("Expected: %v, got: %v", staticEval, got)
			}
		})
	}
}
