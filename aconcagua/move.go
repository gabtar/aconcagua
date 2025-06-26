package aconcagua

// constants representing the type of move(flags)
const (
	quiet = iota
	doublePawnPush
	kingsideCastle
	queensideCastle
	capture
	epCapture
	// Promotions
	knightPromotion
	bishopPromotion
	rookPromotion
	queenPromotion
	knightCapturePromotion
	bishopCapturePromotion
	rookCapturePromotion
	queenCapturePromotion
)

// NoMove is an empty move
const NoMove = Move(0)

// Move represents an encoded chess move on the board
// the move is represented by 16 bits of information
// first 6 bits for the from square (1-64)
// second 6 bits for the to square (1-64)
// last 4 bits for the move flag/type
type Move uint16

// encodeMove returns a reference to an encoded chess move with the values passed
func encodeMove(from uint16, to uint16, flag uint16) *Move {
	mv := Move(from | to<<6 | flag<<12)
	return &mv
}

// from returns the number of the origin square of the move in Little-Endian Rank-File Mapping notation
func (m *Move) from() int {
	return int(*m & 0b111111)
}

// to returns the number of the origin square of the move in Little-Endian Rank-File Mapping notation
func (m *Move) to() int {
	return int((*m & (0b111111 << 6)) >> 6)
}

// flag returns the flag corresponding to the type of the move
func (m *Move) flag() int {
	return int((*m & (0b1111 << 12)) >> 12)
}

// String returns the long algebraic notation of the move used in uci protocol
func (m *Move) String() (move string) {
	move += squareReference[m.from()]
	move += squareReference[m.to()]
	flag := m.flag()

	if flag >= knightPromotion {
		switch flag {
		case knightPromotion, knightCapturePromotion:
			move += "n"
		case bishopPromotion, bishopCapturePromotion:
			move += "b"
		case rookPromotion, rookCapturePromotion:
			move += "r"
		case queenCapturePromotion, queenPromotion:
			move += "q"
		}
	}
	return
}

// positionBefore is an encoded board state before the making a move
type positionBefore uint32

// encodePositionBefore returns a reference to an encoded board state before the move
func encodePositionBefore(pieceMoved uint16, pieceCaptured uint16, epTarget uint16, rule50 uint16) positionBefore {
	rule50uint32 := positionBefore(uint32(rule50) << 15)
	bb := positionBefore(pieceMoved | pieceCaptured<<4 | epTarget<<8)
	bb = bb | rule50uint32
	return bb
}

// pieceMoved returns the type of the piece pieceMoved
func (pb *positionBefore) pieceMoved() int {
	return int(*pb & 0b1111)
}

// pieceCaptured returns the type of the piece pieceCaptured
func (pb *positionBefore) pieceCaptured() int {
	return int((*pb & (0b1111 << 4)) >> 4)
}

// epTarget returns the en passant target square
func (pb *positionBefore) epTarget() int {
	return int((*pb & (0b111111 << 8)) >> 8)
}

// rule50 returns the rule50 counter before the move
func (pb *positionBefore) rule50() int {
	return int((*pb & (0b111111 << 15)) >> 15)
}
