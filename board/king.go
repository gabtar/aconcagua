package board

// -------------
// KING ♔
// -------------

// getKingMoves returns a move slice with all the legal moves of a king from the bitboard passed
func getKingMoves(b *Bitboard, pos *Position, side Color) (moves []Move) {
	movesBB := kingMoves(b, pos, side)
	pieces := ^pos.EmptySquares()
	from := Bsf(*b)
	piece := WhiteKing
	if side == Black {
		piece = BlackKing
	}

	for movesBB > 0 {
		to := movesBB.nextOne()
		moveType := NORMAL
		if to&pieces > 0 {
			moveType = CAPTURE
		}
		moves = append(moves, MoveEncode(from, Bsf(to), int(piece), 0, moveType))
	}
	return
}

// kingMoves returns a bitboard with the legal moves of the king from the bitboard passed
func kingMoves(k *Bitboard, pos *Position, side Color) (moves Bitboard) {
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
