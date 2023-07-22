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
// 4 bits to describe the type of the piece (2^4) (12 types now) - Not necessary needed
// 4 bits to describe the promotedToPiece (2^4) (12 types now)
// 3 bits to describe the type of move (2^3) (6 types) - NOTE: using 4 bits now
// NOTE: For unmaking moves purposes only
// 4 bits to describe the type of the piece captured (2^4) (12 types)
// 6 bits to describe the old en passant target square
// Total := 34 bits, fits in a 64 int

type Move uint64

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

// MoveType returns the type of move
func (m *Move) MoveType() int {
	return int((*m & (0b1111 << 20)) >> 20)
}

// capturedPiece returns the type of the piece captured in the move (for unmake move purposes only)
func (m *Move) capturedPiece() int {
	return int((*m & (0b1111 << 24)) >> 24)
}

// oldEpTarget stores the old en passant square of the position before making the move
func (m *Move) oldEpTarget() int {
	return int((*m & (0b111111 << 28)) >> 28)
}

func (m *Move) ToUci() (uciString string) {
	uciString += squareReference[m.from()]
	uciString += squareReference[m.to()]
	if m.MoveType() == PROMOTION {
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
// TODO: try to use instead a builder pattern for creating moves. It may be faster if i use a pointer to a list of moves preloaded structs in memomory
func MoveEncode(from int, to int, piece int, promotedTo int, moveType int, capturedPiece int, oldEpTarget int) (mov Move) {
	mov |= Move(from)
	mov |= Move(to << 6)
	mov |= Move(piece << 12)
	mov |= Move(promotedTo << 16)
	mov |= Move(moveType << 20)
	mov |= Move(capturedPiece << 24)
	mov |= Move(oldEpTarget << 28)
	return
}
