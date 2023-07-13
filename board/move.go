package board

// Type of move
const (
	NORMAL = iota
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

type Move uint32

// from returns the number of the origin square of the move
func (m *Move) from() int {
	return int(*m & 0b111111)
}

// to returns the number of the destination square of the move
func (m *Move) to() int {
	return int((*m & (0b111111 << 6)) >> 6)
}

// piece returns the piece which is being moved
func (m *Move) piece() int {
	return int((*m & (0b1111 << 12)) >> 12)
}

// promotedTo returns the piece which is going to be replaced by the pawn
func (m *Move) promotedTo() int {
	return int((*m & (0b1111 << 16)) >> 16)
}

// moveType returns the type of move
func (m *Move) moveType() int {
	return int((*m & (0b1111 << 20)) >> 20)
}

func (m *Move) ToUci() (uciString string) {
	uciString += squareReference[m.from()]
	uciString += squareReference[m.to()]
	if m.moveType() == PROMOTION {
		promotedTo := Piece(m.promotedTo())
		switch promotedTo {
		case WhiteQueen, BlackQueen:
			uciString += "q"
		case WhiteRook, BlackRook:
			uciString += "r"
		case WhiteBishop, BlackBishop:
			uciString += "b"
		case WhiteKnight, BlackKnight:
			uciString += "n"
		}
	}
	return
}

// MoveEncode returns a move with the specified values
func MoveEncode(from int, to int, piece int, promotedTo int, moveType int) (mov Move) {
	mov |= Move(from)
	mov |= Move(to << 6)
	mov |= Move(piece << 12)
	mov |= Move(promotedTo << 16)
	mov |= Move(moveType << 20)
	return
}
