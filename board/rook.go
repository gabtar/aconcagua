package board

import "math/bits"

// Rook models a rook piece in chess
type Rook struct {
	color  byte
	square Bitboard
}


// -------------
// ROOK ♖
// -------------
// Attacks returns all squares that a Rook attacks in a chess board
func (r *Rook) Attacks(pos *Position) (attacks Bitboard) {
  blockers := ^pos.EmptySquares()

  for _, direction := range []uint64{NORTH, EAST, SOUTH, WEST} {
    attacks |= raysAttacks[direction][bits.TrailingZeros64(uint64(r.square))]
    blockersInDirection := blockers & raysAttacks[direction][bits.TrailingZeros64(uint64(r.square))]
    nearestBlocker := Bitboard(0)

    switch direction {
    case NORTH, EAST:
	    nearestBlocker = Bitboard(0b1 << bits.TrailingZeros64(uint64(blockersInDirection)))
    case SOUTH, WEST:
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
func (r *Rook) Moves(pos *Position) (moves Bitboard) {
	posiblesMoves := r.Attacks(pos) & ^pos.Pieces(r.color)
	moves |= posiblesMoves

  // If Rook is pinned only allow moves along the pinned direction
  if isPinned(r.square, r.color, pos) && !pos.Check(r.color) {
    kingSq := pos.KingPosition(r.color).ToStringSlice()[0]
    rookSq := r.square.ToStringSlice()[0]

    rookFileRank := files[int(rookSq[0])-97] | ranks[int(rookSq[1])-49]
    kingFileRank := files[int(kingSq[0])-97] | ranks[int(kingSq[1])-49]

    moves &= rookFileRank & kingFileRank
  }

	if pos.Check(r.color) {
		checkingPieces := pos.CheckingPieces(r.color)

		if len(checkingPieces) == 1 {
			checker := checkingPieces[0]
			moves &= checker.Square() & posiblesMoves // Check if can capture the checker

			// Check also if i can block the path to the king when it's a sliding piece
			if checker.IsSliding() {
				direction := getDirection(checker.Square(), pos.KingPosition(r.color))
				moves |= raysAttacks[direction][bits.TrailingZeros64(uint64(checker.Square()))] & posiblesMoves
			}
		} else {
			// Double check -> cannot avoid check by capture/blocking
			moves = Bitboard(0)
		}
	}

	return
}

// Square returns the bitboard with the position of the piece
func (r *Rook) Square() Bitboard {
	return r.square
}

// Color returns the color(side) of the piece
func (r *Rook) Color() byte {
	return r.color
}

// Returns if the piece is an sliding piece(bishops, rooks, queens)
func (r *Rook) IsSliding() bool {
	return true
}

