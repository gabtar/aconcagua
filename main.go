package main

import ( // "flag"
	// "fmt"
	// "github.com/gabtar/aconcagua/board"
	"github.com/gabtar/aconcagua/engine"
	// "github.com/gabtar/aconcagua/search"
)

func main() {

	// fenPtr := flag.String("fen", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "A valid fen string")
	// depthPtr := flag.Int("depth", 3, "The numbers of moves to look ahead in the position")
	// flag.Parse()
	//
	// // Load position
	// pos := board.From(*fenPtr)
	//
	// // Start minmax in engine
	// bestMoveScore, bestMoveTrace := search.BestMove(pos, *depthPtr)
	//
	// // Print best move trace
	// fmt.Println("Score: ", bestMoveScore)
	// for i, move := range bestMoveTrace {
	// 	if i%2 == 0 {
	// 		fmt.Print(i / 2)
	// 		fmt.Print(": ")
	// 	}
	// 	fmt.Print(move)
	// 	fmt.Print(", ")
	// }

  // Runs the engine uci linstening mode
  engine.Start()
}
