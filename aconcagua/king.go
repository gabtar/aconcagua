package aconcagua

// -------------
// KING ♔
// -------------

// getKingMoves returns a move slice with all the legal moves of a king from the bitboard passed
func getKingMoves(b *Bitboard, pos *Position, side Color) (moves []Move) {
	movesBB := kingMoves(b, pos, side)
	pieces := ^pos.EmptySquares()
	from := Bsf(*b)
	piece := pieceOfColor[King][side]

	for movesBB > 0 {
		to := movesBB.NextBit()
		move := newMove().
			setFromSq(from).
			setToSq(Bsf(to)).
			setPiece(piece).
			setMoveType(Normal).
			setEpTargetBefore(pos.enPassantTarget).
			setRule50Before(pos.halfmoveClock).
			setCastleRightsBefore(pos.castlingRights)

		if to&pieces > 0 {
			capturedPiece, _ := pos.PieceAt(squareReference[Bsf(to)])

			// NOTE: the captured piece will stay set but not used if move != CAPTURE
			move.setMoveType(Capture).setCapturedPiece(capturedPiece)
		}
		moves = append(moves, *move)
	}
	return
}

// kingMoves returns a bitboard with the legal moves of the king from the bitboard passed
func kingMoves(k *Bitboard, pos *Position, side Color) (moves Bitboard) {
	withoutKing := *pos
	withoutKing.RemovePiece(*k)
	moves = kingAttacks(k) & ^withoutKing.AttackedSquares(side.Opponent()) & ^pos.Pieces(side)
	return
}

// kingAttacks returns a bitboard with the squares the king attacks from the passed bitboard
func kingAttacks(k *Bitboard) (attacks Bitboard) {
	notInHFile := *k & ^(*k & files[7])
	notInAFile := *k & ^(*k & files[0])

	attacks = notInAFile<<7 | *k<<8 | notInHFile<<9 |
		notInHFile<<1 | notInAFile>>1 | notInHFile>>7 |
		*k>>8 | notInAFile>>9
	return
}

// newKingMoves returns a moves array with the king moves in chessMove format
func newKingMoves(from *Bitboard, pos *Position, side Color, ml *moveList) {
	toSquares := kingMoves(from, pos, side)
	opponentPieces := pos.Pieces(side.Opponent())

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := quiet

		if toSquare&opponentPieces > 0 {
			flag = capture
		}
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}

	if canCastleShort(from, pos, side) {
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(*from<<2)), kingsideCastle))
	}
	if canCastleLong(from, pos, side) {
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(*from>>2)), queensideCastle))
	}
}

// canCastleShort checks if the king can castle short
func canCastleShort(from *Bitboard, pos *Position, side Color) bool {
	if !pos.castlingRights.canCastle(K) && side == White {
		return false
	}
	if !pos.castlingRights.canCastle(k) && side == Black {
		return false
	}

	shortCastlePath := (files[5] | files[6]) & (*from<<2 | *from<<1)
	kingSquaresAttacked := pos.AttackedSquares(side.Opponent())&(shortCastlePath|*from) > 0
	kingSquaresClear := pos.EmptySquares()&shortCastlePath == shortCastlePath

	if !kingSquaresAttacked && kingSquaresClear {
		return true
	}

	return false
}

// canCastleLong checks if the king can castle long
func canCastleLong(from *Bitboard, pos *Position, side Color) bool {
	if !pos.castlingRights.canCastle(Q) && side == White {
		return false
	}
	if !pos.castlingRights.canCastle(q) && side == Black {
		return false
	}

	longCastlePath := (files[1] | files[2] | files[3]) & (*from>>3 | *from>>2 | *from>>1)
	kingPassSquares := *from>>2 | *from>>1 | *from
	kingSquaresAttacked := pos.AttackedSquares(side.Opponent())&(kingPassSquares) > 0
	kingSquaresClear := pos.EmptySquares()&longCastlePath == longCastlePath

	if !kingSquaresAttacked && kingSquaresClear {
		return true
	}

	return false
}
