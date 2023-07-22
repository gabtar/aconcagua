package board

// -------------
// KING ♔
// -------------

// getKingMoves returns a move slice with all the legal moves of a king from the bitboard passed
func getKingMoves(b *Bitboard, pos *Position, side Color) (moves []Move) {
	movesBB := kingMoves(b, pos, side)
	pieces := ^pos.EmptySquares()
	from := Bsf(*b)
	piece := pieceOfColor[King][side]
	oldEpTarget := 0
	if pos.enPassantTarget > 0 {
		oldEpTarget = Bsf(pos.enPassantTarget)
	}

	for movesBB > 0 {
		to := movesBB.nextOne()
		moveType := NORMAL
		capturedPiece := Piece(0)
		if to&pieces > 0 {
			moveType = CAPTURE
			capturedPiece, _ = pos.PieceAt(to.ToStringSlice()[0])
		}
		moves = append(moves, MoveEncode(from, Bsf(to), int(piece), 0, moveType, int(capturedPiece), oldEpTarget))
	}
	return
}

// kingMoves returns a bitboard with the legal moves of the king from the bitboard passed
func kingMoves(k *Bitboard, pos *Position, side Color) (moves Bitboard) {
	withoutKing := pos.RemovePiece(*k)
	moves = kingAttacks(k, pos) & ^withoutKing.AttackedSquares(side.opponent()) & ^pos.Pieces(side)
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
