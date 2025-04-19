package aconcagua

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type UciCommand func(en *Engine, stdout chan string, params ...string)

var uciCommands map[string]UciCommand = map[string]UciCommand{
	"uci":        uciCommand,
	"ucinewgame": uciNewGameCommand,
	"isready":    isReadyCommand,
	"position":   positionCommand,
	"go":         goCommand,
	"stop":       stopCommand,

	// utility/debug commands
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
	stdout <- "id aconcagua"
	stdout <- "author gabtar"
	stdout <- "uciok"
}

// uciNewGameCommand starts a new game
func uciNewGameCommand(en *Engine, stdout chan string, params ...string) {
	en.pos = *InitialPosition()
}

// positionCommand implements the position uci command
func positionCommand(en *Engine, stdout chan string, params ...string) {
	if params[0] == "startpos" {
		en.pos = *From("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	} else if params[0] == "fen" {
		fen := strings.Join(params[1:], " ")
		en.pos = *From(fen)
	} else {
		stdout <- "invalid command"
		return
	}
	movesIndex := findParam(params, "moves")

	if movesIndex != -1 {
		for _, move := range params[movesIndex:] {
			for _, legalMove := range en.pos.LegalMoves().moves {
				if legalMove.String() == move {
					en.pos.MakeMove(&legalMove)
					en.pos.positionHistory = *NewPositionHistory()
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
	score := 0
	depth := findParam(params, "depth")
	wtime := findParam(params, "wtime")
	btime := findParam(params, "btime")
	movetime := findParam(params, "movetime")

	if movetime != -1 {
		depth = 100

		go func() {
			moveTime, _ := strconv.Atoi(params[movetime+1])
			time.Sleep(time.Duration(moveTime) * time.Millisecond)
			en.search.stop = true
		}()

	} else if wtime != -1 && btime != -1 {
		wtime, _ = strconv.Atoi(params[wtime+1])
		btime, _ = strconv.Atoi(params[btime+1])

		engineTime := wtime
		if en.pos.Turn == Black {
			engineTime = btime
		}
		en.timeControl.timeLeftInMiliseconds = engineTime
		en.timeControl.setSearchTime(en.pos.fullmoveNumber / 2)
		depth = 100

		go func() {
			time.Sleep(time.Duration(en.timeControl.searchTimeInMiliseconds) * time.Millisecond)
			en.search.stop = true
		}()

	} else if depth != -1 {
		depth, _ = strconv.Atoi(params[depth+1])
	} else {
		depth = 8
	}

	go func() {
		score = root(&en.pos, &en.search, depth, stdout)
		pv := en.search.pv

		absScore := abs(score)
		if absScore >= MateScore {
			mateIn := ((depth - (absScore - MateScore)) + 1) / 2 // NOTE: in full moves, not ply!
			stdout <- "info score mate " + strconv.Itoa((score/absScore)*mateIn)
		} else {
			stdout <- "info score cp " + strconv.Itoa(score)
		}

		stdout <- "bestmove " + (*pv)[0].String()
	}()
}

// abs returns the absolute value of the number passed
func abs(number int) int {
	if number < 0 {
		return -number
	}
	return number
}

// stopCommand
func stopCommand(en *Engine, stout chan string, params ...string) {
	en.search.stop = true
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
		strings.TrimSpace(cmd)
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
	stdout <- en.pos.String()
}

// perftCommand returns the number of moves up to the passed depth for the current position
func perftCommand(en *Engine, stdout chan string, params ...string) {
	depth, err := strconv.Atoi(params[0])
	if err != nil {
		stdout <- "invalid command"
		return
	}
	stdout <- "nodes " + strconv.FormatUint(en.pos.Perft(depth), 10)
}

// divide returns the number of moves up to the depth passed for each move of the current position
func divideCommand(en *Engine, stdout chan string, params ...string) {
	depth, err := strconv.Atoi(params[0])
	if err != nil {
		stdout <- "invalid command"
		return
	}
	for _, perft := range strings.Split(en.pos.Divide(depth), ",") {
		stdout <- perft
	}
}
