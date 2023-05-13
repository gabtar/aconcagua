package board

// Rook models a rook piece in chess
type Rook struct {
	color  rune
	square Bitboard
}

// -------------
// ROOK ♖
// -------------
// Attacks returns all squares that a Rook attacks in a chess board
func (r *Rook) Attacks(pos *Position) (attacks Bitboard) {
	blockers := ^pos.EmptySquares()

	for _, direction := range []uint64{NORTH, EAST, SOUTH, WEST} {
		attacks |= raysAttacks[direction][bsf(r.square)]
		blockersInDirection := blockers & raysAttacks[direction][bsf(r.square)]
		nearestBlocker := Bitboard(0)

		switch direction {
		case NORTH, EAST:
			nearestBlocker = bitboardFromIndex(bsf(blockersInDirection))
		case SOUTH, WEST:
			nearestBlocker = bitboardFromIndex(63 - bsr(blockersInDirection))
		}

		if nearestBlocker > 0 {
			attacks &= ^raysAttacks[direction][bsf(nearestBlocker)]
		}
	}
	return
}

// Moves returns a bitboard with the legal squares the Rook can move to in a chess position
func (r *Rook) Moves(pos *Position) (moves Bitboard) {
	moves = r.Attacks(pos) & ^pos.Pieces(r.color) & 
          pinRestrictedDirection(r.square, r.color, pos) &
          checkRestrictedMoves(r.square, r.color, pos)
	return
}

// Square returns the bitboard with the position of the piece
func (r *Rook) Square() Bitboard {
	return r.square
}

// Color returns the color(side) of the piece
func (r *Rook) Color() rune {
	return r.color
}

// Returns if the piece is an sliding piece(bishops, rooks, queens)
func (r *Rook) IsSliding() bool {
	return true
}

// role Returns the role of the piece in the board
func (r *Rook) role() int {
  if r.color == WHITE {
    return WHITE_ROOK
  } else {
    return BLACK_ROOK
  }
}
