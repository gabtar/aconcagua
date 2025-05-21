package aconcagua

// ------------------------------------------------------------------
// LEGAL MOVE VALIDATION FUNCTIONS
// ------------------------------------------------------------------

// checkRestrictedSquares returns a bitboard with the squares that are allowed to move when in check
func checkRestrictedSquares(king Bitboard, checkingSliders Bitboard, checkingNonSliders Bitboard) (allowedSquares Bitboard) {
	checkingPieces := checkingSliders | checkingNonSliders
	if checkingPieces.count() == 0 {
		return AllSquares
	}

	if checkingPieces == checkingSliders && checkingPieces.count() == 1 {
		return getRayPath(&checkingPieces, &king) | checkingPieces
	}

	if checkingPieces.count() == 1 {
		return checkingPieces
	}

	return
}

// pinRestrictedSquares returns a bitboard with the squares allowed to move when the piece is pinned
func pinRestrictedSquares(piece Bitboard, king Bitboard, pinnedPieces Bitboard) (restrictedSquares Bitboard) {
	if pinnedPieces&piece > 0 {
		direction := directions[Bsf(piece)][Bsf(king)]
		return raysDirection(king, direction)
	}
	return AllSquares
}

// ------------------------------------------------------------------
// PIECE ATTACKS GENERATION
// ------------------------------------------------------------------

// ------------------------------------------------------------------
// PIECE MOVES GENERATION
// ------------------------------------------------------------------

// ------------------------------------------------------------------
// SPECIAL PAWN MOVES GENERATION
// ------------------------------------------------------------------
