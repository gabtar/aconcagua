package main

import (
	"github.com/gabtar/aconcagua/engine"
)

func main() {
	input := make(chan string)
	output := make(chan string)

	go engine.ReadStdin(input)
	go engine.WriteStdout(output)

	engine.Uci(input, output)
}
