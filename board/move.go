package board

// Type of move
const (
	NORMAL = iota
	CASTLE
	PAWN_DOUBLE_PUSH
	PROMOTION
	EN_PASSANT
	CAPTURE
)

// move stores all information related to a chess move
// The encoding is as follows:
// We use the bits notation to describe the move
// 6 bits to describe the from square (2^6) - 0 to 64
// 6 bits to describe the destination square (2^6) - 0 to 64
// 4 bits to describe the type of the piece (2^4) (12 types now) - Not necessary needed
// 4 bits to describe the promotedToPiece (2^4) (12 types now)
// 3 bits to describe the type of move (2^3) (6 types)

// 32 bits for describe the state of the board before "making" the move. For unmake move purposes only!

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

// Data for unmaking moves purposes
// --------------------------------

// capturedPiece 4 bits 0000
// epTargetBefore 6 bits 000000
// rule50Before 6 bits 000000
// castleRightsBefore 4 bits 0000

// capturedPiece
func (m *Move) capturedPiece() Piece {
	return Piece((*m & (0b1111 << 24)) >> 24)
}

// epTargetBefore
func (m *Move) epTargetBefore() Bitboard {
	// Check non zero...
	indexBB := int((*m & (0b111111 << 28)) >> 28)
	if indexBB == 0 {
		return Bitboard(0)
	}

	return bitboardFromIndex(indexBB)
}

// rule50Before
func (m *Move) rule50Before() int {
	return int((*m & (0b111111 << 34)) >> 34)
}

// castleRightsBefore
func (m *Move) castleRightsBefore() castling {
	return castling((*m & (0b1111 << 40)) >> 40)
}

// ToUci returns the move in UCI format (starting square string -> destinatnion square string)
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

// newMove returns a reference to a Move
func newMove() *Move {
	return new(Move)
}

// setFromSq sets the origin square in the Move
func (m *Move) setFromSq(from int) *Move {
	*m |= Move(from)
	return m
}

// setToSq sets the destination square in the Move
func (m *Move) setToSq(to int) *Move {
	*m |= Move(to << 6)
	return m
}

// setPiece sets the piece moved in the Move
func (m *Move) setPiece(piece Piece) *Move {
	*m |= Move(int(piece) << 12)
	return m
}

// setPromotedTo sets the piece which is going to be promoted to in the Move
func (m *Move) setPromotedTo(piece Piece) *Move {
	*m |= Move(int(piece) << 16)
	return m
}

// setMoveType sets the type of the move
func (m *Move) setMoveType(moveType int) *Move {
	*m |= Move(moveType << 20)
	return m
}

// setCapturedPiece sets the piece captured during the move
func (m *Move) setCapturedPiece(piece Piece) *Move {
	*m |= Move(int(piece) << 24)
	return m
}

// setEpTargetBefore sets the en passant square of the position before making the move
func (m *Move) setEpTargetBefore(epTarget Bitboard) *Move {
	// NOTE: need to do this because of Bsf(0) == 64.
	if epTarget == 0 {
		return m
	}
	*m |= Move(Bsf(epTarget) << 28)
	return m
}

// setRule50Before sets the halfmoveClock of the position before making the move
func (m *Move) setRule50Before(halfmoveClock int) *Move {
	*m |= Move(halfmoveClock << 34)
	return m
}

// setCastleRightsBefore sets the caslte rights of the position before making the move
func (m *Move) setCastleRightsBefore(castles castling) *Move {
	*m |= Move(int(castles) << 40)
	return m
}
