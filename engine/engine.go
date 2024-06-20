// engine package handles uci commands
package engine

import (
	"strings"

	"github.com/gabtar/aconcagua/aconcagua"
	"github.com/gabtar/aconcagua/search"
)

var engine Engine = Engine{pos: *aconcagua.InitialPosition()}

type Engine struct {
	pos aconcagua.Position
	pv  search.PrincipalVariation
}

func NewEngine() *Engine {
	return &Engine{
		pos: *aconcagua.EmptyPosition(),
		pv:  search.PrincipalVariation{},
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
