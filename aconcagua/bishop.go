package aconcagua

// -------------
// BISHOP ♗
// -------------

// bishopMoves returns a bitboard with the legal moves of the bishop from the bitboard passed
func bishopMoves(b *Bitboard, pd *PositionData) (moves Bitboard) {
	return bishopMagicAttacks(Bsf(*b), pd.allies|pd.enemies) &
		^pd.allies &
		pd.checkRestrictedSquares &
		// checkRestrictedSquares(pd.kingPosition, pd.checkingSliders, pd.checkingNonSliders) &
		pinRestrictedSquares(*b, pd.kingPosition, pd.pinnedPieces)
}

// bishopMagicAttacks returns a bitboard with the attack mask of a bishop from the square passed taking into account the blockers
func bishopMagicAttacks(square int, blocks Bitboard) Bitboard {
	blocks &= bishopMaskTable[square]
	magicIndex := (blocks * bishopMagics[square]) >> (64 - bishopMaskTable[square].count())
	return bishopAttacksTable[square][magicIndex]
}

// genBishopMoves generates the bishop moves in the move list
func genBishopMoves(from *Bitboard, ml *moveList, pd *PositionData) {
	toSquares := bishopMoves(from, pd)

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := quiet

		if toSquare&pd.enemies > 0 {
			flag = capture
		}

		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
}

// PositionData contains relevant data for legal move validations of a position
type PositionData struct {
	kingPosition           Bitboard
	checkRestrictedSquares Bitboard
	pinnedPieces           Bitboard
	allies                 Bitboard
	enemies                Bitboard
}

// generatePositionData returns the position data for the current position
func (pos *Position) generatePositionData() PositionData {
	checkingPieces, checkingSliders := pos.CheckingPieces(pos.Turn)
	checkRestrictedSquares := checkRestrictedSquares(pos.KingPosition(pos.Turn), checkingSliders, checkingPieces&^checkingSliders)

	return PositionData{
		kingPosition:           pos.KingPosition(pos.Turn),
		checkRestrictedSquares: checkRestrictedSquares,
		pinnedPieces:           pos.pinnedPieces(pos.Turn),
		allies:                 pos.Pieces(pos.Turn),
		enemies:                pos.Pieces(pos.Turn.Opponent()),
	}
}
