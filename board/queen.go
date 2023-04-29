package board

import "math/bits"

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
		attacks |= raysAttacks[direction][bits.TrailingZeros64(uint64(q.square))]
		blockersInDirection := blockers & raysAttacks[direction][bits.TrailingZeros64(uint64(q.square))]
		nearestBlocker := Bitboard(0)

		switch direction {
		case NORTH, EAST, NORTHEAST, NORTHWEST:
			nearestBlocker = Bitboard(0b1 << bits.TrailingZeros64(uint64(blockersInDirection)))
		case SOUTH, WEST, SOUTHEAST, SOUTHWEST:
			nearestBlocker = Bitboard((0x1 << 63) >> bits.LeadingZeros64(uint64(blockersInDirection)))
		}

		// Need this becuase if its zero, LeadingZeros returns the length of uint64 and goes out of bounds
		if nearestBlocker > 0 {
			attacks &= ^raysAttacks[direction][bits.TrailingZeros64(uint64(nearestBlocker))]
		}
	}
	return
}

// Moves returns a bitboard with the legal squares the Rook can move to in a chess position
func (q *Queen) Moves(pos *Position) (moves Bitboard) {
	posiblesMoves := q.Attacks(pos) & ^pos.Pieces(q.color)
	moves |= posiblesMoves
	kingBB := pos.KingPosition(q.color) // King 'bitboard position
	// If Queen is pinned only allow moves along the pinned direction
	if isPinned(q.square, q.color, pos) && !pos.Check(q.color) {
		direction := getDirection(kingBB, q.square)

		// Need to move along the king-rook direction because of the pin
		allowedMovesDirection := raysDirection(kingBB, direction)
		moves &= allowedMovesDirection
	}

	if pos.Check(q.color) {
    checkingPieces := pos.CheckingPieces(q.color)

    if len(checkingPieces) == 1 {
			checker := checkingPieces[0]
      checkerKingPath := Bitboard(0)

			if checker.IsSliding() {
        checkerKingPath = getRayPath(checker.Square(), kingBB)
			}
      // Check if can capture the checker or block the path
			moves &= (checker.Square() | checkerKingPath) & posiblesMoves
		} else {
			// Double check -> cannot avoid check by capture/blocking
			moves = Bitboard(0)
		}
	}

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
