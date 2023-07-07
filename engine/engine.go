// engine package handles uci commands
package engine

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gabtar/aconcagua/board"
	"github.com/gabtar/aconcagua/search"
)

// readStdin reads strings from standard input
// (from GUI to engine)
func ReadStdin(input chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input <- scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "info string error reading standard input:", err)
			os.Exit(1)
		}
	}
	close(input)
}

// WriteStdin writes strings to standard output
// (form engine to GUI)
func WriteStdout(output <-chan string) {
	for cmd := range output {
		fmt.Println(cmd)
	}
}

// Uci recives a command string and performs the requested actions in the in the engine
func Uci(cmd chan string, output chan string) {

	// TODO: this should be an engine struct with search options, current best move, depth, time, etc
	pos := board.EmptyPosition()

	for {
		command := <-cmd
		commands := strings.Split(strings.TrimSpace(command), " ")

		switch commands[0] {
		case "quit":
			// TODO: stop engine, and output move if any
			return
		case "uci":
			output <- "id aconcagua"
			output <- "uciok"
		case "isready":
			// TODO: check when engine is already calculating
			output <- "readyok"
		case "position":
			if commands[1] == "startpos" {
				pos = board.From("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
			} else if commands[1] == "fen" {
				fen := strings.Join(commands[2:], " ")
				pos = board.From(fen)
			}
			hasMoves, _ := regexp.MatchString("moves", command)

			if hasMoves {
				// Match de moves para adelante
				re := regexp.MustCompile("(?:moves )(.+)")
				movesList := re.FindString(command)
				moves := strings.Split(movesList, " ")
				for _, move := range moves {
					// Find the move that matches the uci command and perform on the position
					for _, legalMove := range pos.LegalMoves(pos.ToMove()) {
						if legalMove.ToUci() == move {
							// TODO: check other ways to dereference
							pos = func(p board.Position) *board.Position { return &p }(pos.MakeMove(&legalMove))
						}
					}
				}
			}
		case "go":
			// TODO: set flag on engine that its calculating
			// TODO: parse depth, time, other paramterts
			score, bestMove := search.BestMove(pos, 4)
			// FIX: Negamax relative score based on the side, i should invert sign if black to move...
			output <- "info score cp " + strconv.Itoa(score)
			output <- "bestmove " + bestMove.ToUci()

			// TODO:  ----------- only for testing internal state via uci command line ----------------
		case "printboard":
			pos.Print()
		default:
			output <- "invalid command"
		}
	}
}
