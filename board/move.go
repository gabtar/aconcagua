package board

import "fmt"

// Type of move
const (
	NORMAL           = 1 // Normal push to an empty sqaure
	PAWN_DOUBLE_PUSH = 2
	CASTLE           = 3
	EN_PASSANT       = 4
	PROMOTION        = 5
	CAPTURE          = 6
)

// Represents a move of chess
type Move struct {
	from       string
	to         string
	piece      int
	promotedTo int
	moveType   int
}

// String returns the move string
func (m Move) String() string {
	return fmt.Sprintf(
		"%s -> %s",
		m.from,
		m.to)
}

func (m Move) ToUci() (uciString string) {
	uciString += m.from
	uciString += m.to
	if m.moveType == PROMOTION {
		switch m.promotedTo {
		case WHITE_QUEEN, BLACK_QUEEN:
			uciString += "q"
		case WHITE_ROOK, BLACK_ROOK:
			uciString += "r"
		case WHITE_BISHOP, BLACK_BISHOP:
			uciString += "b"
		case WHITE_KNIGHT, BLACK_KNIGHT:
			uciString += "n"
		}
	}
	return
}
