package board

// Bishop models a bishop piece in chess
type Bishop struct {
	color  rune
	square Bitboard
}

// -------------
// BISHOP ♗
// -------------
// Attacks returns all squares that a Bishop attacks in a chess board
func (b *Bishop) Attacks(pos *Position) (attacks Bitboard) {
	blockers := ^pos.EmptySquares()

	for _, direction := range []uint64{NORTHEAST, SOUTHEAST, SOUTHWEST, NORTHWEST} {
		attacks |= raysAttacks[direction][Bsf(b.square)]
		blockersInDirection := blockers & raysAttacks[direction][Bsf(b.square)]
		nearestBlocker := Bitboard(0)

		switch direction {
		case NORTHEAST, NORTHWEST:
			nearestBlocker = BitboardFromIndex(Bsf(blockersInDirection))
		case SOUTHEAST, SOUTHWEST:
			nearestBlocker = BitboardFromIndex(63 - Bsr(blockersInDirection))
		}

		if nearestBlocker > 0 {
			attacks &= ^raysAttacks[direction][Bsf(nearestBlocker)]
		}
	}
	return
}

// Moves returns a bitboard with the legal squares the Bishop can move to in a chess position
func (b *Bishop) Moves(pos *Position) (moves Bitboard) {
	moves = b.Attacks(pos) & ^pos.Pieces(b.color) &
		pinRestrictedDirection(b.square, b.color, pos) &
		checkRestrictedMoves(b.square, b.color, pos)
	return
}

// Square returns the bitboard with the position of the piece
func (b *Bishop) Square() Bitboard {
	return b.square
}

// Color returns the color(side) of the piece
func (b *Bishop) Color() rune {
	return b.color
}

// Returns if the piece is an sliding piece(bishops, rooks, queens)
func (b *Bishop) IsSliding() bool {
	return true
}

// role Returns the role of the piece in the board
func (b *Bishop) role() int {
	if b.color == WHITE {
		return WHITE_BISHOP
	} else {
		return BLACK_BISHOP
	}
}

// validMoves returns an slice of the valid moves for the Bishop in the position
func (b *Bishop) validMoves(pos *Position) (moves []Move) {
	destinationsBB := b.Moves(pos)
	opponentPieces := pos.Pieces(opponentSide(b.color))
	piece := WHITE_BISHOP
	if b.color == BLACK {
		piece = BLACK_BISHOP
	}

	for destinationsBB > 0 {
		square := Bitboard(0b1 << Bsf(destinationsBB))
		if opponentPieces&square > 0 {
			moves = append(moves, Move{
				from:     squareMap[Bsf(b.square)],
				to:       squareMap[Bsf(destinationsBB)],
				piece:    piece,
				moveType: CAPTURE,
			})
		} else {
			moves = append(moves, Move{
				from:     squareMap[Bsf(b.square)],
				to:       squareMap[Bsf(destinationsBB)],
				piece:    piece,
				moveType: NORMAL,
			})
		}
		destinationsBB ^= Bitboard(square)
	}
	return
}
