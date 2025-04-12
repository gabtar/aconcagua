package aconcagua

import (
	"strings"
	"time"
)

type Engine struct {
	pos         Position
	search      Search
	timeControl TimeControl
}

func NewEngine() *Engine {
	return &Engine{
		pos: *InitialPosition(),
		search: Search{
			nodes:        0,
			currentDepth: 0,
			maxDepth:     0,
			pv:           newPV(),
			killers:      [100]Killer{},
			time:         time.Now(),
			totalTime:    time.Now(),
			stop:         false,
		},
		timeControl: TimeControl{},
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
		go en.execute(commands[0], stdout, commands[1:]...)
	}

}
