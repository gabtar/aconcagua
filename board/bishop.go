package board

import "math/bits"

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
		attacks |= raysAttacks[direction][bits.TrailingZeros64(uint64(b.square))]
		blockersInDirection := blockers & raysAttacks[direction][bits.TrailingZeros64(uint64(b.square))]
		nearestBlocker := Bitboard(0)

		switch direction {
		case NORTHEAST, NORTHWEST:
			nearestBlocker = Bitboard(0b1 << bits.TrailingZeros64(uint64(blockersInDirection)))
		case SOUTHEAST, SOUTHWEST:
			nearestBlocker = Bitboard((0x1 << 63) >> bits.LeadingZeros64(uint64(blockersInDirection)))
		}

		// Need this becuase if its zero, LeadingZeros returns the length of uint64 and goes out of bounds
		if nearestBlocker > 0 {
			attacks &= ^raysAttacks[direction][bits.TrailingZeros64(uint64(nearestBlocker))]
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
