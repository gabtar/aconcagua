package aconcagua

// -------------
// KNIGHT ♘
// -------------

// knightAttackSquares is an array representing the squares a knight attacks from the square(int) passed
var knightAttackSquares = [64]Bitboard{
	0x20400,
	0x50800,
	0xA1100,
	0x142200,
	0x284400,
	0x508800,
	0xA01000,
	0x402000,
	0x2040004,
	0x5080008,
	0xA110011,
	0x14220022,
	0x28440044,
	0x50880088,
	0xA0100010,
	0x40200020,
	0x204000402,
	0x508000805,
	0xA1100110A,
	0x1422002214,
	0x2844004428,
	0x5088008850,
	0xA0100010A0,
	0x4020002040,
	0x20400040200,
	0x50800080500,
	0xA1100110A00,
	0x142200221400,
	0x284400442800,
	0x508800885000,
	0xA0100010A000,
	0x402000204000,
	0x2040004020000,
	0x5080008050000,
	0xA1100110A0000,
	0x14220022140000,
	0x28440044280000,
	0x50880088500000,
	0xA0100010A00000,
	0x40200020400000,
	0x204000402000000,
	0x508000805000000,
	0xA1100110A000000,
	0x1422002214000000,
	0x2844004428000000,
	0x5088008850000000,
	0xA0100010A0000000,
	0x4020002040000000,
	0x400040200000000,
	0x800080500000000,
	0x1100110A00000000,
	0x2200221400000000,
	0x4400442800000000,
	0x8800885000000000,
	0x100010A000000000,
	0x2000204000000000,
	0x4020000000000,
	0x8050000000000,
	0x110A0000000000,
	0x22140000000000,
	0x44280000000000,
	0x88500000000000,
	0x10A00000000000,
	0x20400000000000,
}

// getKnightMoves returns a move slice of all posible moves of the knight passed
func getKnightMoves(b *Bitboard, pos *Position, side Color) (moves []Move) {
	movesBB := knightMoves(b, pos, side)
	pieces := ^pos.EmptySquares()
	from := Bsf(*b)
	piece := pieceOfColor[Knight][side]

	for movesBB > 0 {
		to := movesBB.NextBit()
		move := newMove().
			setFromSq(from).
			setToSq(Bsf(to)).
			setPiece(piece).
			setMoveType(Normal).
			setEpTargetBefore(pos.enPassantTarget).
			setRule50Before(pos.halfmoveClock).
			setCastleRightsBefore(pos.castlingRights)

		if to&pieces > 0 {
			capturedPiece, _ := pos.PieceAt(squareReference[Bsf(to)])
			move.setMoveType(Capture).setCapturedPiece(capturedPiece)
		}
		moves = append(moves, *move)
	}
	return
}

// knightMoves returns a bitboard with the legal moves of the knight from the bitboard passed
func knightMoves(k *Bitboard, pos *Position, side Color) (moves Bitboard) {
	// If the knight is pinned it can move at all
	if isPinned(k, side, pos) {
		return Bitboard(0)
	}
	moves = knightAttackSquares[Bsf(*k)] & ^pos.Pieces(side) &
		checkRestrictedMoves(*k, side, pos)
	return
}

// newKnightMoves returns a moves array with the knight moves in chessMove format
func newKnightMoves(from *Bitboard, pos *Position, side Color, ml *moveList) {
	toSquares := knightMoves(from, pos, side)
	opponentPieces := pos.Pieces(side.Opponent())

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := quiet

		if toSquare&opponentPieces > 0 {
			flag = capture
		}

		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
}
