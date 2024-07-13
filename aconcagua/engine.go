package aconcagua

import (
	"strings"
	"time"
)

var engine Engine = Engine{pos: *InitialPosition()}

type Engine struct {
	pos         Position
	searching   bool // whenever the engine is looking for a position
	searchState SearchState
}

func NewEngine() *Engine {
	return &Engine{
		pos:       *EmptyPosition(),
		searching: false,
		searchState: SearchState{
			nodes:        0,
			currentDepth: 0,
			maxDepth:     0,
			pv:           newPrincipalVariation(),
			killers:      [100]KillerMoves{},
			history:      HistoryMoves{},
			time:         time.Now(),
			totalTime:    time.Now(),
			stop:         false,
		},
		// TODO: engine options
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
