package board

// Queen models a queen piece in chess
type Queen struct {
	color  rune
	square Bitboard
}

// -------------
// QUEEN ♕
// -------------
// Attacks returns all squares that a Queen attacks in a chess board
func (q *Queen) Attacks(pos *Position) (attacks Bitboard) {
	blockers := ^pos.EmptySquares()

	for _, direction := range []uint64{NORTH, NORTHEAST, EAST, SOUTHEAST, SOUTH, SOUTHWEST, WEST, NORTHWEST} {
		attacks |= raysAttacks[direction][Bsf(q.square)]
		blockersInDirection := blockers & raysAttacks[direction][Bsf(q.square)]
		nearestBlocker := Bitboard(0)

		switch direction {
		case NORTH, EAST, NORTHEAST, NORTHWEST:
			nearestBlocker = BitboardFromIndex(Bsf(blockersInDirection))
		case SOUTH, WEST, SOUTHEAST, SOUTHWEST:
			nearestBlocker = BitboardFromIndex(63 - Bsr(blockersInDirection))
		}

		// Need this becuase if its zero, LeadingZeros returns the length of uint64 and goes out of bounds
		if nearestBlocker > 0 {
			attacks &= ^raysAttacks[direction][Bsf(nearestBlocker)]
		}
	}
	return
}

// Moves returns a bitboard with the legal squares the Rook can move to in a chess position
func (q *Queen) Moves(pos *Position) (moves Bitboard) {
	moves = q.Attacks(pos) & ^pos.Pieces(q.color) &
		pinRestrictedDirection(q.square, q.color, pos) &
		checkRestrictedMoves(q.square, q.color, pos)
	return
}

// Square returns the bitboard with the position of the piece
func (q *Queen) Square() Bitboard {
	return q.square
}

// Color returns the color(side) of the piece
func (q *Queen) Color() rune {
	return q.color
}

// Returns if the piece is an sliding piece(bishops, rooks, queens)
func (q *Queen) IsSliding() bool {
	return true
}

// role Returns the role of the piece in the board
func (q *Queen) role() int {
	if q.color == WHITE {
		return WHITE_QUEEN
	} else {
		return BLACK_QUEEN
	}
}

// validMoves returns an slice of the valid moves for the Queen in the position
func (q *Queen) validMoves(pos *Position) (moves []Move) {
	destinationsBB := q.Moves(pos)
	opponentPieces := pos.Pieces(opponentSide(q.color))
	piece := WHITE_QUEEN
	if q.color == BLACK {
		piece = BLACK_QUEEN
	}

	for destinationsBB > 0 {
		square := Bitboard(0b1 << Bsf(destinationsBB))
		if opponentPieces&square > 0 {
			moves = append(moves, Move{
				from:     squareReference[Bsf(q.square)],
				to:       squareReference[Bsf(destinationsBB)],
				piece:    piece,
				moveType: CAPTURE,
			})
		} else {
			moves = append(moves, Move{
				from:     squareReference[Bsf(q.square)],
				to:       squareReference[Bsf(destinationsBB)],
				piece:    piece,
				moveType: NORMAL,
			})
		}
		destinationsBB ^= Bitboard(square)
	}
	return
}
