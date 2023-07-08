package board

// King models a king piece in chess
type King struct {
	color  rune
	square Bitboard
}

// -------------
// KING ♔
// -------------
// Attacks returns all squares that a King attacks in a chess board
func (k *King) Attacks(pos *Position) (attacks Bitboard) {
	notInHFile := k.square & ^(k.square & files[7])
	notInAFile := k.square & ^(k.square & files[0])

	attacks = notInAFile<<7 | k.square<<8 | notInHFile<<9 |
		notInHFile<<1 | notInAFile>>1 | notInHFile>>7 |
		k.square>>8 | notInAFile>>9
	return
}

// Moves returns a bitboard with the legal squares the King can move to
func (k *King) Moves(pos *Position) (moves Bitboard) {
	// King can only move to an empty square or capture an opponent piece not defended by
	// opposite side
	withoutKing := pos.RemovePiece(k.square)
	moves = k.Attacks(pos) & ^withoutKing.AttackedSquares(opponentSide(k.color)) & ^pos.Pieces(k.color)
	return
}

// Square returns the bitboard with the position of the piece
func (k *King) Square() Bitboard {
	return k.square
}

// Color returns the color(side) of the piece
func (k *King) Color() rune {
	return k.color
}

// Returns if the piece is an sliding piece(bishops, rooks, queens)
func (k *King) IsSliding() bool {
	return false
}

// role Returns the role of the piece in the board
func (k *King) role() int {
	if k.color == WHITE {
		return WHITE_KING
	} else {
		return BLACK_KING
	}
}

// validMoves returns an slice of the valid moves for the King in the position
func (k *King) validMoves(pos *Position) (moves []Move) {
	destinationsBB := k.Moves(pos)
	opponentPieces := pos.Pieces(opponentSide(k.color))
	piece := WHITE_KING
	if k.color == BLACK {
		piece = BLACK_KING
	}

	for destinationsBB > 0 {
		square := Bitboard(0b1 << Bsf(destinationsBB))
		if opponentPieces&square > 0 {
			moves = append(moves, Move{
				from:     squareReference[Bsf(k.square)],
				to:       squareReference[Bsf(destinationsBB)],
				piece:    piece,
				moveType: CAPTURE,
			})
		} else {
			moves = append(moves, Move{
				from:     squareReference[Bsf(k.square)],
				to:       squareReference[Bsf(destinationsBB)],
				piece:    piece,
				moveType: NORMAL,
			})
		}
		destinationsBB ^= Bitboard(square)
	}
	return
}

// getKingMoves returns a move slice with all the legal moves of a king from the bitboard passed
func getKingMoves(b *Bitboard, pos *Position, side rune) (moves []move) {
	movesBB := kingMoves(b, pos, side)
	pieces := ^pos.EmptySquares()
	from := Bsf(*b)
	piece := WHITE_KING
	if side == BLACK {
		piece = BLACK_KING
	}

	for movesBB > 0 {
		to := movesBB.nextOne()
		moveType := NORMAL
		if to&pieces > 0 {
			moveType = CAPTURE
		}
		moves = append(moves, MoveEncode(from, Bsf(to), piece, 0, moveType))
	}
	return
}

// kingMoves returns a bitboard with the legal moves of the king from the bitboard passed
func kingMoves(k *Bitboard, pos *Position, side rune) (moves Bitboard) {
	withoutKing := pos.RemovePiece(*k)
	moves = kingAttacks(k, pos) & ^withoutKing.AttackedSquares(opponentSide(side)) & ^pos.Pieces(side)
	return
}

// kingAttacks returns a bitboard with the squares the king attacks from the passed bitboard
func kingAttacks(k *Bitboard, pos *Position) (attacks Bitboard) {
	notInHFile := *k & ^(*k & files[7])
	notInAFile := *k & ^(*k & files[0])

	attacks = notInAFile<<7 | *k<<8 | notInHFile<<9 |
		notInHFile<<1 | notInAFile>>1 | notInHFile>>7 |
		*k>>8 | notInAFile>>9
	return
}
