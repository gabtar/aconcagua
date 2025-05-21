package aconcagua

// pinnedPieces returns a bitboard with the pieces pinned in the position for the side passed
func (pos *Position) pinnedPieces(side Color) (pinned Bitboard) {
	king := pos.KingPosition(side)
	opponentDiagonalAttackers := pos.Bitboards[int(side.Opponent())*6+int(WhiteQueen)] | pos.Bitboards[int(side.Opponent())*6+int(WhiteBishop)]
	opponentOrthogonalAttackers := pos.Bitboards[int(side.Opponent())*6+int(WhiteQueen)] | pos.Bitboards[int(side.Opponent())*6+int(WhiteRook)]
	ownPieces := pos.Pieces(side) &^ king
	opponentPieces := pos.Pieces(side.Opponent())

	if king == 0 {
		return
	}

	for opponentDiagonalAttackers > 0 {
		opponent := opponentDiagonalAttackers.NextBit()
		opponentDiagonalRays := bishopMagicAttacks(Bsf(opponent), Bitboard(0))
		kingToOpponentPath := getRayPath(&opponent, &king)
		intersectionRay := opponentDiagonalRays & kingToOpponentPath
		piecesBetween := intersectionRay & ownPieces
		oponentPiecesBetween := intersectionRay & opponentPieces

		if opponentDiagonalRays&king > 0 && piecesBetween.count() == 1 && oponentPiecesBetween == 0 {
			pinned |= piecesBetween
		}
	}

	for opponentOrthogonalAttackers > 0 {
		opponent := opponentOrthogonalAttackers.NextBit()
		opponentOrthogonalRays := rookMagicAttacks(Bsf(opponent), Bitboard(0))
		kingToOpponentPath := getRayPath(&opponent, &king)
		intersectionRay := opponentOrthogonalRays & kingToOpponentPath
		piecesBetween := intersectionRay & ownPieces
		oponentPiecesBetween := intersectionRay & opponentPieces

		if opponentOrthogonalRays&king > 0 && piecesBetween.count() == 1 && oponentPiecesBetween == 0 {
			pinned |= piecesBetween
		}
	}

	return
}

// isPinned returns if the passed piece is pinned in the passed position
func isPinned(piece *Bitboard, side Color, pos *Position) bool {
	// needed to be passed instead position
	// 1 King of the side being evaluated
	// 1 Bitboard with own Pieces
	// 1 bitboard with enemies pieces

	return pos.pinnedPieces(side)&*piece > 0
}

// raysDirection returns the rays along the direction passed that intersects the
// piece in the square passed
func raysDirection(square Bitboard, direction uint64) Bitboard {
	oppositeDirections := [8]uint64{SOUTH, SOUTHWEST, WEST, NORTHWEST, NORTH, NORTHEAST, EAST, SOUTHEAST}

	return rayAttacks[direction][Bsf(square)] | square |
		rayAttacks[oppositeDirections[direction]][Bsf(square)]
}

// getRayPath returns a Bitboard with the path between 2 bitboards pieces
// (not including the 2 pieces)
func getRayPath(from *Bitboard, to *Bitboard) (rayPath Bitboard) {
	fromSq := Bsf(*from)
	toSq := Bsf(*to)

	fromDirection := directions[fromSq][toSq]
	toDirection := directions[toSq][fromSq]

	if fromDirection == INVALID || toDirection == INVALID {
		return
	}

	return rayAttacks[fromDirection][fromSq] &
		rayAttacks[toDirection][toSq]
}

// pinRestrictedDirection returns a bitboard with the restricted direction of moves
func pinRestrictedDirection(piece *Bitboard, side Color, pos *Position) (restrictedDirection Bitboard) {
	restrictedDirection = AllSquares // No initial restrictions
	kingBB := pos.KingPosition(side)

	if isPinned(piece, side, pos) {
		direction := directions[Bsf(*piece)][Bsf(kingBB)]
		allowedMovesDirection := raysDirection(kingBB, direction)
		restrictedDirection = allowedMovesDirection
	}
	return
}

func checkRestrictedMoves(side Color, pos *Position) (allowedSquares Bitboard) {
	checkingPieces := pos.CheckingPieces(side, false)

	switch {
	case checkingPieces.count() == 0:
		allowedSquares = AllSquares
	case checkingPieces.count() == 1:
		piece := pos.PieceAt(squareReference[Bsf(checkingPieces)])

		if isSliding(piece) {
			king := pos.KingPosition(side)
			allowedSquares |= getRayPath(&checkingPieces, &king)
		}
		allowedSquares |= checkingPieces
	}
	// If there are more than 2 CheckingPieces it cant move at all (default allowedSquares value = 0)
	return
}

// isSliding returns a the passed Piece is an sliding piece(Queen, Rook or Bishop)
func isSliding(piece Piece) bool {
	return (piece >= WhiteQueen && piece <= WhiteBishop) ||
		(piece >= BlackQueen && piece <= BlackBishop)
}
