package uci

import (
	"strconv"
	"strings"

	"github.com/gabtar/aconcagua/internal/engine"
)

// UciCommand defines the interface for all UCI commands.
type UciCommand interface {
	Execute(en *engine.Engine, stdout chan string, params ...string)
}

// UciCommandStruct represents the "uci" command.
type UciCommandStruct struct{}

// Execute handles the "uci" command logic.
func (c *UciCommandStruct) Execute(en *engine.Engine, stdout chan string, params ...string) {
	stdout <- "id name Aconcagua 3.4.0"
	stdout <- "id author Gabriel Tarifa"
	stdout <- ""
	stdout <- "option name BookPath type string default <empty>"
	stdout <- "option name UseBook type check default false"
	stdout <- "option name UCI_Chess960 type check default false"
	stdout <- "uciok"
}

// UciNewGameCommandStruct represents the "ucinewgame" command.
type UciNewGameCommandStruct struct{}

// Execute handles the "ucinewgame" command logic.
func (c *UciNewGameCommandStruct) Execute(en *engine.Engine, stdout chan string, params ...string) {
	en.Pos.LoadFromFenString(engine.StartingFenString)
}

// UciPositionCommandStruct represents the "position" command.
type UciPositionCommandStruct struct{}

// Execute handles the "position" command logic
func (c *UciPositionCommandStruct) Execute(en *engine.Engine, stdout chan string, params ...string) {
	switch params[0] {
	case "startpos":
		en.Pos.LoadFromFenString(engine.StartingFenString)
	case "fen":
		fen := strings.Join(params[1:], " ")
		en.Pos.LoadFromFenString(fen)

		if en.Options.Chess960 {
			en.Pos.SetUp960Castles(fen)
		}
	default:
		stdout <- "invalid command"
		return
	}
	movesIndex := findParam(params, "moves")

	if movesIndex != -1 {
		en.Pos.LoadMoves(params[movesIndex+1:]...)
	}
}

// UciIsReadyCommandStruct represents the "isready" command.
type UciIsReadyCommandStruct struct{}

// Execute handles the "isready" command logic.
func (c *UciIsReadyCommandStruct) Execute(en *engine.Engine, stdout chan string, params ...string) {
	stdout <- "readyok"
}

// UciGoCommandStruct represents the "go" command.
type UciGoCommandStruct struct{}

// Execute handles the "go" command logic
func (c *UciGoCommandStruct) Execute(en *engine.Engine, stdout chan string, params ...string) {
	if en.Options.UseOpeningBook {
		polyglotEntry := en.OpeningBook.PickRandomOpeningVariation(engine.PolyglotHashFromPosition(&en.Pos))
		polyglotMove := engine.PolyglotMove(polyglotEntry.Move)

		if polyglotMove != engine.PolyglotMove(0) {
			stdout <- "bestmove " + polyglotMove.String()
			return
		}
	}
	depth := engine.MaxSearchDepth

	depthIndex := findParam(params, "depth")
	wtime := findParam(params, "wtime")
	btime := findParam(params, "btime")
	winc := findParam(params, "winc")
	binc := findParam(params, "binc")

	movetime := findParam(params, "movetime")

	searchStrategy, clock := engine.TimeStrategy(params, depth, wtime, btime, winc, binc, movetime)
	en.Search.TimeControl.Initialize(searchStrategy, int(en.Pos.Turn), en.Pos.FullMoveNumber, clock)

	// Set default depth if not passed
	if depthIndex != -1 {
		depth, _ = strconv.Atoi(params[depthIndex+1])
	}

	go func() {
		_, bestMove := en.Search.IterativeDeepening(&en.Pos, depth, stdout)
		stdout <- "bestmove " + bestMove
	}()
}

// TODO: for now just to setup opening book. Later will handle more options. eg hash table size, etc
// also use an error handling when parsing input
// Use specific options?? UciSetChess960CommandStruct ....

// UciSetOptionCommandStruct represents the "setoption" command.
type UciSetOptionCommandStruct struct{}

// Execute handles the "setoption" command logic.
func (c *UciSetOptionCommandStruct) Execute(en *engine.Engine, stdout chan string, params ...string) {
	// sample usage: setoption name bookpath value <bookfilename>
	if strings.ToLower(params[1]) == "bookpath" {
		entries, err := en.OpeningBook.Load(params[3])

		if err != nil {
			stdout <- "option name BookPath value " + err.Error()
			return
		}

		stdout <- "option name BookPath set entries found " + strconv.Itoa(entries)
	}

	// sample usage: setoption name usebook
	if strings.ToLower(params[1]) == "usebook" {
		useBook := params[3] == "true"
		en.Options.UseOpeningBook = useBook
		stdout <- "option name UseBook value " + strconv.FormatBool(en.Options.UseOpeningBook)
	}

	// sample usage: setoption name uci_chess960 value true
	if strings.ToLower(params[1]) == "uci_chess960" {
		if len(params) < 3 {
			stdout <- "invalid command"
			return
		}

		en.Options.Chess960 = params[3] == "true"
		stdout <- "option name UCI_Chess960 value " + strconv.FormatBool(en.Options.Chess960)
	}

}

// UciStopCommandStruct represents the "stop" command.
type UciStopCommandStruct struct{}

// Execute handles the "stop" command logic.
func (c *UciStopCommandStruct) Execute(en *engine.Engine, stdout chan string, params ...string) {
	en.Search.Stop()
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

// PrintBoardCommandStruct prints the board on the terminal
type PrintBoardCommandStruct struct{}

func (c *PrintBoardCommandStruct) Execute(en *engine.Engine, stdout chan string, params ...string) {
	stdout <- en.Pos.String()
}

// PerftCommandStruct returns the number of moves up to the passed depth for the current position
type PerftCommandStruct struct{}

func (c *PerftCommandStruct) Execute(en *engine.Engine, stdout chan string, params ...string) {
	depth, err := strconv.Atoi(params[0])
	if err != nil {
		stdout <- "invalid command"
		return
	}
	stdout <- "nodes " + strconv.FormatUint(en.Pos.Perft(depth), 10)
}

// DivideCommandStruct returns the number of moves up to the depth passed for each move of the current position
type DivideCommandStruct struct{}

func (c *DivideCommandStruct) Execute(en *engine.Engine, stdout chan string, params ...string) {
	depth, err := strconv.Atoi(params[0])
	if err != nil {
		stdout <- "invalid command"
		return
	}
	for perft := range strings.SplitSeq(en.Pos.Divide(depth), ",") {
		stdout <- perft
	}
}
