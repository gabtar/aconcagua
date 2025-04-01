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

// chessMove represents an encoded chess move on the board
// the move is represented by 16 bits of information
// first 6 bits for the from square (1-64)
// second 6 bits for the to square (1-64)
// last 4 bits for the move flag/type
type chessMove uint16

// encodeMove returns a reference to an encoded chess move with the values passed
func encodeMove(from uint16, to uint16, flag uint16) *chessMove {
	mv := chessMove(from | to<<6 | flag<<12)
	return &mv
}

// from returns the number of the origin square of the move in Little-Endian Rank-File Mapping notation
func (m *chessMove) from() int {
	return int(*m & 0b111111)
}

// to returns the number of the origin square of the move in Little-Endian Rank-File Mapping notation
func (m *chessMove) to() int {
	return int((*m & (0b111111 << 6)) >> 6)
}

// flag returns the flag corresponding to the type of the move
func (m *chessMove) flag() int {
	return int((*m & (0b1111 << 12)) >> 12)
}

// String returns the long algebraic notation of the move used in uci protocol
func (m *chessMove) String() (move string) {
	move += squareReference[m.from()]
	move += squareReference[m.to()]
	flag := m.flag()

	if flag > 5 { // NOTE: >5 all are promotions
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

// TODO: move ordering idea
// have 2 separated list for moves an scores
// and then sort all of them at the same time
// moves []
// scores []

// store board state to undo move
// type of piece moved -> 4 bit
// type of piece captured if any -> 4 bit
// epTarget -> 6 bit (0-64)
// rule50 -> 6 bit
// castles... -> separate castle struct

type positionBefore uint32

// encodePositionBefore returns a reference to an encoded board state before the move
func encodePositionBefore(pieceMoved uint16, pieceCaptured uint16, epTarget uint16, rule50 uint16) positionBefore {
	rule50uint32 := positionBefore(uint32(rule50) << 15)
	bb := positionBefore(pieceMoved | pieceCaptured<<4 | epTarget<<8)
	bb = bb | rule50uint32
	return bb
}

// pieceMoved returns the type of the piece pieceMoved
func (bb *positionBefore) pieceMoved() int {
	return int(*bb & 0b1111)
}

// pieceCaptured returns the type of the piece pieceCaptured
func (bb *positionBefore) pieceCaptured() int {
	return int((*bb & (0b1111 << 4)) >> 4)
}

// epTarget returns the en passant target square
func (bb *positionBefore) epTarget() int {
	return int((*bb & (0b111111 << 8)) >> 8)
}

// rule50 returns the rule50 counter before the move
func (bb *positionBefore) rule50() int {
	return int((*bb & (0b111111 << 15)) >> 15)
}
