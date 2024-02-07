package engine

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gabtar/aconcagua/board"
	"github.com/gabtar/aconcagua/search"
)

type UciCommand func(en *Engine, stdout chan string, params ...string)

var uciCommands map[string]UciCommand = map[string]UciCommand{
	"uci":      uciCommand,
	"isready":  isReadyCommand,
	"position": positionCommand,
	"go":       goCommand,

	"d":      printBoardCommand,
	"perft":  perftCommand,
	"divide": divideCommand,
}

func (en *Engine) execute(command string, stdout chan string, params ...string) {
	comm, exists := uciCommands[command]

	if !exists {
		stdout <- "invalid command"
		return
	}

	comm(en, stdout, params...)
}

// uciCommand takes care of uciok command from gui
func uciCommand(en *Engine, stdout chan string, params ...string) {
	// TODO: add more data/options of the engine. Checkout the stockfish's uciok implementation
	stdout <- "id aconcagua"
	stdout <- "author gabtar"
	stdout <- "uciok"
}

// positionCommand implements the position uci command
func positionCommand(en *Engine, stdout chan string, params ...string) {
	if params[0] == "startpos" {
		engine.pos = *board.From("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	} else if params[0] == "fen" {
		fen := strings.Join(params[1:], " ")
		en.pos = *board.From(fen)
	} else {
		stdout <- "invalid command"
		return
	}
	movesIndex := findParam(params, "moves")

	if movesIndex != -1 {
		for _, move := range params[movesIndex:] {
			for _, legalMove := range engine.pos.LegalMoves(engine.pos.ToMove()) {
				if legalMove.ToUci() == move {
					engine.pos.MakeMove(&legalMove)
				}
			}
		}
	}
}

// isReadyCommand responses to the gui whenever the engine is ready
func isReadyCommand(en *Engine, stdout chan string, params ...string) {
	stdout <- "readyok"
}

// goCommand starts looking for the best move in the current position
func goCommand(en *Engine, stdout chan string, params ...string) {
	// TODO: implement remaining 'go' uci options

	depth := 5

	dIndex := findParam(params, "depth")
	if dIndex != -1 {
		depth, _ = strconv.Atoi(params[dIndex+1])
	}

	// TODO: use search with a goroutine
	score, bestMove := search.Search(&engine.pos, depth, stdout)
	stdout <- "info score cp " + strconv.Itoa(score)
	stdout <- "bestmove " + bestMove[0].ToUci()
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

// findParam returns the index if the passed params slice contains the searched param string or -1 if not
func findParam(params []string, param string) int {
	for i, p := range params {
		if p == param {
			return i
		}
	}
	return -1
}

// NOTE: Non uci commands, for debug only!

// printBoardCommand prtints the board on the terminal
func printBoardCommand(en *Engine, stdout chan string, params ...string) {
	en.pos.Print()
	stdout <- "Fen: " + engine.pos.ToFen()
}

// perftCommand returns the number of moves up to the passed depth for the current position
func perftCommand(en *Engine, stdout chan string, params ...string) {
	depth, err := strconv.Atoi(params[0])
	if err != nil {
		stdout <- "invalid command"
		return
	}
	stdout <- "nodes " + strconv.FormatUint(engine.pos.Perft(depth), 10)
}

// divide returns the number of moves up to the depth passed for each move of the current position
func divideCommand(en *Engine, stdout chan string, params ...string) {
	depth, err := strconv.Atoi(params[0])
	if err != nil {
		stdout <- "invalid command"
		return
	}
	for _, perft := range strings.Split(engine.pos.Divide(depth), ",") {
		stdout <- perft
	}
}
