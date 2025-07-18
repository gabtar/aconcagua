package aconcagua

import (
	"strings"
)

// Engine represents a chess engine
type Engine struct {
	pos         Position
	search      Search
	timeControl *TimeControl
	openingBook PolyglotBook
	options     Options
}

// Options are various engine options that can be set
type Options struct {
	useOpeningBook bool
}

// NewEngine returns a new Engine instance
func NewEngine() *Engine {
	tc := TimeControl{}

	return &Engine{
		pos:         *InitialPosition(),
		timeControl: &tc,
		search: Search{
			nodes:        0,
			currentDepth: 0,
			maxDepth:     0,
			PVTable:      NewPVTable(MaxSearchDepth),
			killers:      [100]Killer{},
			timeControl:  &tc,
		},
		openingBook: PolyglotBook{},
		options: Options{
			useOpeningBook: false,
		},
	}
}

// StartUci initializes the Uci Protocol loop in the engine
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
		en.execute(commands[0], stdout, commands[1:]...)
	}
}

// execute executes a command
func (en *Engine) execute(command string, stdout chan string, params ...string) {
	var uciCommands = map[string]UciCommand{
		"uci":        uciCommand,
		"ucinewgame": uciNewGameCommand,
		"isready":    isReadyCommand,
		"position":   positionCommand,
		"go":         goCommand,
		"stop":       stopCommand,
		"setoption":  setOptionCommand,

		// utility/debug commands
		"d":      printBoardCommand,
		"perft":  perftCommand,
		"divide": divideCommand,
	}

	comm, exists := uciCommands[command]

	if !exists {
		stdout <- "invalid command"
		return
	}

	comm(en, stdout, params...)
}
