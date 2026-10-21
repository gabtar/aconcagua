package engine

import (
	"fmt"
	"time"
)

// RunBench runs a benchmark on tests positions printing total nodes and nps
func RunBench() {
	depth := 12
	eng := NewEngine()
	var totalNodes uint64
	start := time.Now()

	for _, fen := range benchPositions {
		stdout := make(chan string, 100)
		eng.Search.clear()
		eng.Search.TranspositionTable.Clear()
		eng.Pos.LoadFromFenString(fen)
		searchStrategy, clock := TimeStrategy([]string{}, depth, -1, -1, -1, -1, -1, -1)
		eng.Search.TimeControl.Initialize(searchStrategy, int(eng.Pos.Turn), eng.Pos.FullMoveNumber, clock)
		go func() {
			for range stdout {
			}
		}()

		eng.Search.IterativeDeepening(&eng.Pos, depth, stdout)
		close(stdout)
		totalNodes += uint64(eng.Search.totalNodes)
	}

	elapsed := time.Since(start)
	nps := uint64(float64(totalNodes) / max(elapsed.Seconds(), 0.001))
	fmt.Printf("%d nodes %d nps\n", totalNodes, nps)
}

// TODO: add more bench positions
var benchPositions = []string{
	// Openings
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
	"r1bqkbnr/pppp1ppp/2n5/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R b KQkq - 3 3",
	"rnbqkb1r/pp2pppp/3p1n2/8/3NP3/2N5/PPP2PPP/R1BQKB1R b KQkq - 2 5",

	// Tactical middlegames (castling, promotions, pins)
	"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
	"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
	"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
	"r1bq1rk1/pp2ppbp/2np1np1/8/3NP3/2N1BP2/PPPQ2PP/R3KB1R w KQ - 0 9",

	// Quiet / closed middlegames
	"r2q1rk1/pp2bppp/2n1pn2/2pp4/3P1B2/2PBPN2/PP3PPP/RN1Q1RK1 w - - 0 9",

	// Endgames
	"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
	"4k3/8/8/8/8/8/4P3/4K3 w - - 0 1",
	"8/8/1k6/2b5/2pP4/8/5K2/8 b - d3 0 1",
	"6k1/5ppp/8/8/8/8/5PPP/3R2K1 w - - 0 1",
	"2r3k1/pp3ppp/4p3/3p4/3P4/2P1P3/PP3PPP/2R3K1 w - - 0 1",
	"8/pp3k2/2p2p2/8/3P4/P4K2/1P6/8 w - - 0 1",
	"8/5pk1/6p1/8/8/6P1/5PK1/3R4 w - - 0 1",
}
