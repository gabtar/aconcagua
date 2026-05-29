package engine

import (
	"fmt"
	"strings"
	"testing"
)

func TestAdd(t *testing.T) {
	ph := NewPositionHistory()
	ph.add(positionBefore(0), KQkq, 0)

	if ph.moveCount != 1 {
		t.Errorf("Expected: %v, got: %v", 1, ph.moveCount)
	}
}

func TestPop(t *testing.T) {
	ph := NewPositionHistory()
	ph.add(positionBefore(0), KQkq, 0)
	_, _ = ph.pop()

	if ph.moveCount != 0 {
		t.Errorf("Expected: %v, got: %v", 0, ph.moveCount)
	}
}

func TestRepetitionCount(t *testing.T) {
	ph := NewPositionHistory()
	halfmoveClock := 2
	hash := uint64(5)
	ph.add(positionBefore(0), KQkq, hash)
	ph.add(positionBefore(0), KQkq, 0)

	if ph.isRepetition(5, halfmoveClock) != true {
		t.Errorf("Expected: %v, got: %v", 2, ph.isRepetition(5, halfmoveClock))
	}
}

// Threefold repetition bugs...
// Make some test in positions that will produce repetitions
// Recreate the ones failing on fastchess and see why its not detecting as a repetition properly...
func TestRepetitionBug(t *testing.T) {
	// BUG #1: Halfmove clock not reset after double pawn pushes...
	// BUG #2: If ep square is not reachable by enemy pawns, there  no ep avaialbe. So do not hash the zobrist ep square key.
	// Because it will also count as repetition, regardless a double push created a ep square. Apparently, if is reacheable but
	// not legal because puts own king on check, we have to hash the key. Need to test this first!!!
	// Warning; PV continues after threefold repetition - move a4a5 from Aconcagua DEV
	// Info; info depth 13 seldepth 21 score cp -60 nodes 22290 nps 3026546 hashfull 5 time 25 pv d7e6 c2d3 e6f7 g2g4 a4a3 d3e4 a3a4 e4d3 a4a5 d3e4 a5a4 e4d3 a4a5 d3d4 a5b5 d4e4
	// Info; info depth 13 seldepth 21 score cp -60 nodes 22290 nps 3026546 hashfull 5 time 25 pv d7e6 c2d3 e6f7 g2g4 a4a3 d3e4 a3a4 e4d3 a4a5 d3e4 a5a4 (e4d3 - this is the draw by repetition) a4a5 d3d4 a5b5 d4e4
	// Position; fen r1bqkb1r/pppp1ppp/2n2n2/4P3/4P3/5N2/PPP2PPP/RNBQKB1R b KQkq - 0 4
	// Moves; f6g4 f1e2 f8c5 e1g1 g4e5 b1c3 e5f3 e2f3 e8g8 c3d5 c6e5 c1e3 d7d6 f3e2 c5e3 d5e3 f8e8 f2f4 e5d7 d1d4 d7c5 e4e5 c8d7 a1d1 a7a5 e5d6 c7d6 e2f3 a8a6 e3c4 c5e6 d4d2 d7c6 f4f5 e6g5 f3c6 a6c6 c4d6 d8b6 g1h1 e8d8 d2g5 c6d6 d1d6 b6d6 g5h4 f7f6 h4e4 d6c6 e4c6 b7c6 h1g1 d8d2 f1c1 g8f7 b2b3 f7e7 c1e1 e7d6 e1e6 d6d7 e6e3 d2c2 e3g3 g7g5 f5g6 h7g6 g3g6 c2c1 g1f2 c1c2 f2f1 c2c1 f1e2 c1c2 e2d3 c2a2 h2h4 a2a3 d3c2 a5a4 b3a4 a3a4 h4h5

	pos := NewPosition()
	pos.LoadFromFenString("r1bqkb1r/pppp1ppp/2n2n2/4P3/4P3/5N2/PPP2PPP/RNBQKB1R b KQkq - 0 4")
	movesString := "f6g4 f1e2 f8c5 e1g1 g4e5 b1c3 e5f3 e2f3 e8g8 c3d5 c6e5 c1e3 d7d6 f3e2 c5e3 d5e3 f8e8 f2f4 e5d7 d1d4 d7c5 e4e5 c8d7 a1d1 a7a5 e5d6 c7d6 e2f3 a8a6 e3c4 c5e6 d4d2 d7c6 f4f5 e6g5 f3c6 a6c6 c4d6 d8b6 g1h1 e8d8 d2g5 c6d6 d1d6 b6d6 g5h4 f7f6 h4e4 d6c6 e4c6 b7c6 h1g1 d8d2 f1c1 g8f7 b2b3 f7e7 c1e1 e7d6 e1e6 d6d7 e6e3 d2c2 e3g3 g7g5 f5g6 h7g6 g3g6 c2c1 g1f2 c1c2 f2f1 c2c1 f1e2 c1c2 e2d3 c2a2 h2h4 a2a3 d3c2 a5a4 b3a4 a3a4 h4h5 d7e6 c2d3 e6f7 g2g4 a4a3 d3e4 a3a4 e4d3 a4a5 d3e4 a5a4 e4d3"

	moves := strings.Split(movesString, " ")
	pos.LoadMoves(moves...)

	fmt.Println("Final position: ", pos.ToFen())
	fmt.Println(pos.String())
	fmt.Println(pos.positionHistory.previousPosition)
	fmt.Println("Is Draw by repetition: ", pos.positionHistory.isRepetition(pos.Hash, pos.halfmoveClock))
	// BUG: probably a half move clock improper reset/update. Perhaps i need to check again the full logic on make/unmaking moves...
	// Also the hash passed when calling the isRepetition, is wrong!!!!

	t.Fail()
}
