package board

import "fmt"

// Type of move
const (
	NORMAL           = iota
	PAWN_DOUBLE_PUSH
	CASTLE
	EN_PASSANT
	PROMOTION
	CAPTURE
)

// move stores all information related to a chess move
// The encoding is as follows:
// We use the bits notation to describe the move
// 6 bits to describe the from square (2^6) - 0 to 64
// 6 bits to describe the destination square (2^6) - 0 to 64
// 4 bits to describe the type of the piece (2^4) (12 types now)
// 4 bits to describe the promotedToPiece (2^4) (12 types now)
// 3 bits to describe the type of move (2^3) (6 types)
// Total := 23 bits, fits in a 32 int

type move uint32

// from returns the number of the origin square of the move
func (m *move) from() int {
	return int(*m & 0b111111)
}

// to returns the number of the destination square of the move
func (m *move) to() int {
	return int((*m & (0b111111 << 6 )) >> 6)
}

// piece returns the piece which is being moved
func (m *move) piece() int {
	return int((*m & (0b1111 << 12 )) >> 12)
}

// promotedTo returns the piece which is going to be replaced by the pawn
func (m *move) promotedTo() int {
	return int((*m & (0b1111 << 16 )) >> 16)
}

// moveType returns the type of move
func (m *move) moveType() int {
	return int((*m & (0b1111 << 20 )) >> 20)
}

func (m *move) ToUci() (uciString string) {
	uciString += squareReference[m.from()]
	uciString += squareReference[m.to()]
	if m.moveType() == PROMOTION {
		switch m.promotedTo() {
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

// MoveEncode returns a move with the specified values
func MoveEncode(from int, to int, piece int, promotedTo int, moveType int) (mov move){
	mov |= move(from)
	mov |= move(to << 6)
	mov |= move(piece << 12)
	mov |= move(promotedTo << 18)
	mov |= move(moveType << 22)
	return
}

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
