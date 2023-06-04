// engine package handles uci commands
package engine

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/gabtar/aconcagua/board"
	"github.com/gabtar/aconcagua/search"
)

// UCI Protocol definition: https://www.wbec-ridderkerk.nl/html/UCIProtocol.html

// * all command strings the engine receives will end with '\n',
//   also all commands the GUI receives should end with '\n',
//   Note: '\n' can be 0x0d or 0x0a0d or any combination depending on your OS.
//   If you use Engine and GUI in the same OS this should be no problem if you communicate in text mode,
//   but be aware of this when for example running a Linux engine in a Windows GUI.
// * Before the engine is asked to search on a position, there will always be a position command
//   to tell the engine about the current position.

// Move format:
// ------------
//
// The move format is in long algebraic notation.
// A nullmove from the Engine to the GUI should be sent as 0000.
// Examples:  e2e4, e7e5, e1g1 (white short castling), e7e8q (for promotion)

// Start enables the uci protocol comunication with the chess engine for gui interfaces
func Start() {
  // TODO: use go routines/async comunication with the engine
  // uci protocol says: * the engine must always be able to process input from stdin, even while thinking.
	reader := bufio.NewReader(os.Stdin)

	// New Engine position
	pos := board.EmptyPosition()

	for {
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Invalid command:", command)
			continue
		}

		command = strings.TrimSpace(command)
		parts := strings.Split(command, " ")

		switch parts[0] {
		case "quit":
			// quit the program as soon as possible
			return
		case "uci":
			// tell engine to use the uci (universal chess interface),
			// this will be send once as a first command after program boot
			// to tell the engine to switch to uci mode.
			// After receiving the uci command the engine must identify itself with the "id" command
			// and sent the "option" commands to tell the GUI which engine settings the engine supports if any.
			// After that the engine should sent "uciok" to acknowledge the uci mode.
			// If no uciok is sent within a certain time period, the engine task will be killed by the GUI.

			fmt.Println("id aconcagua")
			fmt.Println("uciok")
		case "isready":
			// this is used to synchronize the engine with the GUI. When the GUI has sent a command or
			// multiple commands that can take some time to complete,
			// this command can be used to wait for the engine to be ready again or
			// to ping the engine to find out if it is still alive.
			// E.g. this should be sent after setting the path to the tablebases as this can take some time.
			// This command is also required once before the engine is asked to do any search
			// to wait for the engine to finish initializing.
			// This command must always be answered with "readyok" and can be sent also when the engine is calculating
			// in which case the engine should also immediately answer with "readyok" without stopping the search.
			fmt.Println("readyok")
		case "position":
			// * position [fen <fenstring> | startpos ]  moves <move1> .... <movei>
			// set up the position described in fenstring on the internal board and
			// play the moves on the internal chess board.
			// if the game was played  from the start position the string "startpos" will be sent
			// Note: no "new" command is needed. However, if this position is from a different game than
			// the last position sent to the engine, the GUI should have sent a "ucinewgame" inbetween.

			if parts[1] == "startpos" {
				pos = board.From("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
			} else if parts[1] == "fen" {
				fen := strings.Join(parts[2:], " ")
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
			//      start calculating on the current position set up with the "position" command.
			// There are a number of commands that can follow this command, all will be sent in the same string.
			// If one command is not sent its value should be interpreted as it would not influence the search.
			// * searchmoves <move1> .... <movei>
			// 	restrict search to this moves only
			// 	Example: After "position startpos" and "go infinite searchmoves e2e4 d2d4"
			// 	the engine should only search the two moves e2e4 and d2d4 in the initial position.
			// * ponder
			// 	start searching in pondering mode.
			// 	Do not exit the search in ponder mode, even if it's mate!
			// 	This means that the last move sent in in the position string is the ponder move.
			// 	The engine can do what it wants to do, but after a "ponderhit" command
			// 	it should execute the suggested move to ponder on. This means that the ponder move sent by
			// 	the GUI can be interpreted as a recommendation about which move to ponder. However, if the
			// 	engine decides to ponder on a different move, it should not display any mainlines as they are
			// 	likely to be misinterpreted by the GUI because the GUI expects the engine to ponder
			//    on the suggested move.
			// * wtime <x>
			// 	white has x msec left on the clock
			// * btime <x>
			// 	black has x msec left on the clock
			// * winc <x>
			// 	white increment per move in mseconds if x > 0
			// * binc <x>
			// 	black increment per move in mseconds if x > 0
			// * movestogo <x>
			//      there are x moves to the next time control,
			// 	this will only be sent if x > 0,
			// 	if you don't get this and get the wtime and btime it's sudden death
			// * depth <x>
			// 	search x plies only.
			// * nodes <x>
			//    search x nodes only,
			// * mate <x>
			// 	search for a mate in x moves
			// * movetime <x>
			// 	search exactly x mseconds
			// * infinite
			// 	search until the "stop" command. Do not exit the search without being told so in this mode!

			// TODO: fixed depth for now
			// Maybe i can make a depth based on the number of pieces in the board
			_, bestMove := search.BestMove(pos, 4)
			fmt.Println("bestmove", bestMove[0].ToUci())
		default:
			fmt.Println("Invalid command:", command)
		}
	}
}
