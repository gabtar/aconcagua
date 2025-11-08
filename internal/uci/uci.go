package uci

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gabtar/aconcagua/internal/engine"
)

// UciProtocol represents the UCI protocol
type UciProtocol struct {
	engine *engine.Engine
}

// NewUciProtocol creates a new UCI protocol
func NewUciProtocol(en *engine.Engine) *UciProtocol {
	return &UciProtocol{
		engine: en,
	}
}

// Execute executes a command
func (uci *UciProtocol) Execute(command string, stdout chan string, params ...string) {
	var uciCommands = map[string]UciCommand{
		"uci":        &UciCommandStruct{},
		"ucinewgame": &UciNewGameCommandStruct{},
		"isready":    &UciIsReadyCommandStruct{},
		"position":   &UciPositionCommandStruct{},
		"go":         &UciGoCommandStruct{},
		"stop":       &UciStopCommandStruct{},
		"setoption":  &UciSetOptionCommandStruct{},

		// utility/debug commands
		"d":      &PrintBoardCommandStruct{},
		"perft":  &PerftCommandStruct{},
		"divide": &DivideCommandStruct{},
	}

	comm, exists := uciCommands[command]

	if !exists {
		stdout <- "invalid command"
		return
	}

	comm.Execute(uci.engine, stdout, params...)
}

// Start initializes the Uci Protocol loop in the engine
func (uci *UciProtocol) Start() {
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
		uci.Execute(commands[0], stdout, commands[1:]...)
	}
}

// readStdin reads strings from standard input (from GUI to engine)
func readStdin(input chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input <- scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "info string error reading standard input:", err)
			os.Exit(1)
		}
	}
}

// writeStdout writes strings to standard output (form engine to GUI)
func writeStdout(output <-chan string) {
	for cmd := range output {
		fmt.Println(cmd)
	}
}
