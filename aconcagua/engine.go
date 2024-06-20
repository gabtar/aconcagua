package aconcagua

import (
	"strings"
)

var engine Engine = Engine{pos: *InitialPosition()}

type Engine struct {
	pos Position
	pv  PrincipalVariation
}

func NewEngine() *Engine {
	return &Engine{
		pos: *EmptyPosition(),
		pv:  PrincipalVariation{},
	}
}

// StartUci initializes the Uci Protocol in the engine
func (en *Engine) StartUci() {
	stdin := make(chan string)
	stdout := make(chan string)
	defer close(stdin)
	defer close(stdout)

	go readStdin(stdin)
	go writeStdout(stdout)

	for {
		command := <-stdin
		if command == "quit" {
			break
		}

		commands := strings.Split(strings.TrimSpace(command), " ")
		go engine.execute(commands[0], stdout, commands[1:]...)
	}

}
