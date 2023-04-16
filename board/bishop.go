package board

import "math/bits"

// Bishop models a bishop piece in chess
type Bishop struct {
	color  byte
	square Bitboard
}

// -------------
// BISHOP ♖
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
	posiblesMoves := b.Attacks(pos) & ^pos.Pieces(b.color)
	moves |= posiblesMoves
	kingBB := pos.KingPosition(b.color)

	if isPinned(b.square, b.color, pos) && !pos.Check(b.color) {
		direction := getDirection(kingBB, b.square)

		allowedMovesDirection := raysDirection(kingBB, direction)
		moves &= allowedMovesDirection
	}

	if pos.Check(b.color) {
		checkingPieces := pos.CheckingPieces(b.color)

		if len(checkingPieces) == 1 {
			checker := checkingPieces[0]
			moves &= checker.Square() & posiblesMoves // Check if can capture the checker

			// Check also if i can block the path to the king when it's a sliding piece
			if checker.IsSliding() {
				direction := getDirection(checker.Square(), kingBB)
				moves |= raysDirection(kingBB, direction) & posiblesMoves
			}
		} else {
			// Double check -> cannot avoid check by capture/blocking
			moves = Bitboard(0)
		}
	}
	return
}

// Square returns the bitboard with the position of the piece
func (b *Bishop) Square() Bitboard {
	return b.square
}

// Color returns the color(side) of the piece
func (b *Bishop) Color() byte {
	return b.color
}

// Returns if the piece is an sliding piece(bishops, rooks, queens)
func (b *Bishop) IsSliding() bool {
	return true
}
