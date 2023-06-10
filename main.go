package main

import ( // "flag"
	// "fmt"
	// "github.com/gabtar/aconcagua/board"
	// "fmt"

	"github.com/gabtar/aconcagua/engine"
	// "github.com/gabtar/aconcagua/search"
)

func main() {
	// Create channels for reading/writing from stdin/stdout
	input := make(chan string)
	output := make(chan string)

	// Read commands from stdin (GUI app)
	go engine.ReadStdin(input)
	// sendBack commands to gui -> print to stdout
	go engine.WriteStdout(output)

	// Parse uci commands in the engine
	engine.Uci(input, output)
}
