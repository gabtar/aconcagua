package aconcagua

// -------------
// BISHOP ♗
// -------------

// bishopMoves returns a bitboard with the legal moves of the bishop from the bitboard passed
func bishopMoves(b *Bitboard, pos *Position, side Color) (moves Bitboard) {
	king := pos.KingPosition(side)
	checkingSliders := pos.CheckingPieces(side, true)
	checkingNonSliders := pos.CheckingPieces(side, false) &^ checkingSliders
	pinnedPieces := pos.pinnedPieces(side)

	return bishopMagicAttacks(Bsf(*b), pos.Pieces(White)|pos.Pieces(Black)) &
		^pos.Pieces(side) &
		checkRestrictedSquares(king, checkingSliders, checkingNonSliders) &
		pinRestrictedSquares(*b, king, pinnedPieces)
}

// bishopMagicAttacks returns a bitboard with the attack mask of a bishop from the square passed taking into account the blockers
func bishopMagicAttacks(square int, blocks Bitboard) Bitboard {
	blocks &= bishopMaskTable[square]
	magicIndex := (blocks * bishopMagics[square]) >> (64 - bishopMaskTable[square].count())
	return bishopAttacksTable[square][magicIndex]
}

// genBishopMoves generates the bishop moves in the move list
func genBishopMoves(from *Bitboard, pos *Position, side Color, ml *moveList) {
	toSquares := bishopMoves(from, pos, side)
	opponentPieces := pos.Pieces(side.Opponent())

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := quiet

		if toSquare&opponentPieces > 0 {
			flag = capture
		}

		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
}
