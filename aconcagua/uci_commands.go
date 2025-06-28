package aconcagua

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// MaxSearchDepth is the maximum depth of the search
const MaxSearchDepth = 50

// UciCommand handles a uci command instruction from the gui
type UciCommand func(en *Engine, stdout chan string, params ...string)

// uciCommand takes care of uci command from gui
func uciCommand(en *Engine, stdout chan string, params ...string) {
	stdout <- "id aconcagua"
	stdout <- "author gabtar"
	stdout <- ""
	stdout <- "option name BookPath type string default <empty>"
	stdout <- "option name UseBook type button"
	stdout <- "uciok"
}

// uciNewGameCommand starts a new game
func uciNewGameCommand(en *Engine, stdout chan string, params ...string) {
	en.pos = *InitialPosition()
}

// positionCommand sets up the current position
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
				}
			}
		}
		en.pos.positionHistory.clear()
	}
}

// isReadyCommand responses to the gui whenever the engine is ready
func isReadyCommand(en *Engine, stdout chan string, params ...string) {
	stdout <- "readyok"
}

// goCommand starts looking for the best move in the current position
func goCommand(en *Engine, stdout chan string, params ...string) {
	if en.options.useOpeningBook {
		polyglotEntry := en.openingBook.pickRandomOpeningVariation(PolyglotHashFromPosition(&en.pos))
		polyglotMove := PolyglotMove(polyglotEntry.Move)

		if polyglotMove != PolyglotMove(0) {
			stdout <- "bestmove " + polyglotMove.String()
			return
		}
	}
	depth := MaxSearchDepth

	depthIndex := findParam(params, "depth")
	wtime := findParam(params, "wtime")
	btime := findParam(params, "btime")
	movetime := findParam(params, "movetime")

	searchStrategy, clock := timeStrategy(params, depth, wtime, btime, movetime)
	en.timeControl.init(searchStrategy, int(en.pos.Turn), en.pos.fullmoveNumber, clock)

	// Set default depth if not passed
	if depthIndex != -1 {
		depth, _ = strconv.Atoi(params[depthIndex+1])
	}

	go func() {
		score, bestMove := en.search.root(&en.pos, depth, stdout)

		absScore := abs(score)
		isMate := absScore >= MateScore-depth
		if isMate {
			mateIn := (MateScore - absScore + 1) / 2 // NOTE: in full moves, not ply!
			stdout <- "info score mate " + strconv.Itoa((score/absScore)*mateIn)
		} else {
			stdout <- "info score cp " + strconv.Itoa(score)
		}

		stdout <- "bestmove " + bestMove
	}()
}

// timeStrategy returns the search strategy and the clock for a search
func timeStrategy(params []string, depth int, wtime int, btime int, movetime int) (int, Clock) {
	if movetime != -1 {
		movetime, _ = strconv.Atoi(params[movetime+1])
		return MoveTimeStrategy, Clock{0, 0, 0, 0, movetime}
	}
	if wtime != -1 && btime != -1 {
		wtime, _ = strconv.Atoi(params[wtime+1])
		btime, _ = strconv.Atoi(params[btime+1])
		return TimeLeftStrategy, Clock{wtime, btime, 0, 0, 0}
	}
	if depth != -1 {
		return DepthStrategy, Clock{0, 0, 0, 0, 0}
	}
	return InfiniteStrategy, Clock{0, 0, 0, 0, 0}
}

// setOptionCommand sets an option on the engine
func setOptionCommand(en *Engine, stdout chan string, params ...string) {
	// TODO: for now just to setup opening book. Later will handle more options. eg hash table size, etc

	// sample usage: setoption name bookpath value <bookfilename>
	if strings.ToLower(params[1]) == "bookpath" {
		err := en.openingBook.Load(params[3])

		if err != nil {
			stdout <- "option name BookPath value " + err.Error()
			return
		}

		stdout <- "option name BookPath set entries found " + strconv.Itoa(int(en.openingBook.size))
	}

	// sample usage: setoption name usebook
	if strings.ToLower(params[1]) == "usebook" {
		en.options.useOpeningBook = !en.options.useOpeningBook
		stdout <- "option name UseBook value " + strconv.FormatBool(en.options.useOpeningBook)
	}

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
	en.search.timeControl.stop = true
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
	for perft := range strings.SplitSeq(en.pos.Divide(depth), ",") {
		stdout <- perft
	}
}
