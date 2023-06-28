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
		attacks |= raysAttacks[direction][Bsf(r.square)]
		blockersInDirection := blockers & raysAttacks[direction][Bsf(r.square)]
		nearestBlocker := Bitboard(0)

		switch direction {
		case NORTH, EAST:
			nearestBlocker = BitboardFromIndex(Bsf(blockersInDirection))
		case SOUTH, WEST:
			nearestBlocker = BitboardFromIndex(63 - Bsr(blockersInDirection))
		}

		if nearestBlocker > 0 {
			attacks &= ^raysAttacks[direction][Bsf(nearestBlocker)]
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

// validMoves returns an slice of the valid moves for the Rook in the position
func (r *Rook) validMoves(pos *Position) (moves []Move) {
	destinationsBB := r.Moves(pos)
	opponentPieces := pos.Pieces(opponentSide(r.color))
	piece := WHITE_ROOK
	if r.color == BLACK {
		piece = BLACK_ROOK
	}

	for destinationsBB > 0 {
		square := Bitboard(0b1 << Bsf(destinationsBB))
		if opponentPieces&square > 0 {
			moves = append(moves, Move{
				from:     squareReference[Bsf(r.square)],
				to:       squareReference[Bsf(destinationsBB)],
				piece:    piece,
				moveType: CAPTURE,
			})
		} else {
			moves = append(moves, Move{
				from:     squareReference[Bsf(r.square)],
				to:       squareReference[Bsf(destinationsBB)],
				piece:    piece,
				moveType: NORMAL,
			})
		}
		destinationsBB ^= Bitboard(square)
	}
	return
}
