package board

import "math/bits"

// Bishop models a bishop piece in chess
type Bishop struct {
  color byte
  square Bitboard
}


// -------------
// BISHOP ♖
// -------------
// Attacks returns all squares that a Bishop attacks in a chess board
func (r *Bishop) Attacks(pos *Position) (attacks Bitboard) {
	// square := r.square.ToStringSlice()[0]

	pieces := ^pos.EmptySquares()
  // TODO should be diagoanl[x] | antidiagonal[x]
	attacks |= raysAttacks[NORTHWEST][bits.TrailingZeros64(uint64(r.square))]

	blockers := (attacks & pieces)

	northWestBlockers := blockers & raysAttacks[NORTHWEST][bits.TrailingZeros64(uint64(r.square))]
	nearestNorthWestBlocker := Bitboard(0b1 << bits.TrailingZeros64(uint64(northWestBlockers)))
	if nearestNorthWestBlocker > 0 {
		attacks &= ^raysAttacks[NORTHWEST][bits.TrailingZeros64(uint64(nearestNorthWestBlocker))]
	}
	//
	// eastBlockers := blockers & raysAttacks[EAST][bits.TrailingZeros64(uint64(r.square))]
	// nearestEastBlocker := Bitboard(0b1 << bits.TrailingZeros64(uint64(eastBlockers)))
	// if nearestEastBlocker > 0 {
	// 	attacks &= ^raysAttacks[EAST][bits.TrailingZeros64(uint64(nearestEastBlocker))]
	// }
	//
	// southBlockers := blockers & raysAttacks[SOUTH][bits.TrailingZeros64(uint64(r.square))]
	// nearestSouthBlocker := Bitboard((0x1 << 63) >> bits.LeadingZeros64(uint64(southBlockers)))
	// if nearestSouthBlocker > 0 {
	// 	attacks &= ^raysAttacks[SOUTH][bits.TrailingZeros64(uint64(nearestSouthBlocker))]
	// }
	//
	// westBlockers := blockers & raysAttacks[WEST][bits.TrailingZeros64(uint64(r.square))]
	// nearestWestBlocker := Bitboard((0x1 << 63) >> bits.LeadingZeros64(uint64(westBlockers)))
	// if nearestWestBlocker > 0 {
	// 	attacks &= ^raysAttacks[WEST][bits.TrailingZeros64(uint64(nearestWestBlocker))]
	// }

	return
}

// Moves returns a bitboard with the legal squares the Bishop can move to in a chess position
func (r *Bishop) Moves(pos *Position) (moves Bitboard) {
	return
}

// Square returns the bitboard with the position of the piece
func (r *Bishop) Square() Bitboard {
	return r.square
}

// Color returns the color(side) of the piece
func (r *Bishop) Color() byte {
	return r.color
}

// Returns if the piece is an sliding piece(bishops, rooks, queens)
func (r *Bishop) IsSliding() bool {
	return true
}

