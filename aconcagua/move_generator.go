package aconcagua

// ------------------------------------------------------------------
// PIECE ATTACKS GENERATION
// ------------------------------------------------------------------

// kingAttacks returns a bitboard with the squares the king attacks from the passed bitboard
func kingAttacks(k *Bitboard) (attacks Bitboard) {
	notInHFile := *k & ^(*k & files[7])
	notInAFile := *k & ^(*k & files[0])

	attacks = notInAFile<<7 | *k<<8 | notInHFile<<9 |
		notInHFile<<1 | notInAFile>>1 | notInHFile>>7 |
		*k>>8 | notInAFile>>9
	return
}

// queenAttacks returns a Bitboard with all the squares a queen is attacking
func queenAttacks(q *Bitboard, blocks Bitboard) (attacks Bitboard) {
	attacks = rookAttacks(Bsf(*q), blocks) | bishopAttacks(Bsf(*q), blocks)
	return
}

// rookAttacks returns a bitboard with the attack mask of a rook from the square passed taking into account the blockers
func rookAttacks(square int, blocks Bitboard) Bitboard {
	blocks &= rooksMaskTable[square]
	magicIndex := (blocks * rookMagics[square]) >> (64 - rooksMaskTable[square].count())
	return rookAttacksTable[square][magicIndex]
}

// bishopAttacks returns a bitboard with the attack mask of a bishop from the square passed taking into account the blockers
func bishopAttacks(square int, blocks Bitboard) Bitboard {
	blocks &= bishopMaskTable[square]
	magicIndex := (blocks * bishopMagics[square]) >> (64 - bishopMaskTable[square].count())
	return bishopAttacksTable[square][magicIndex]
}

// pawnAttacks returns a bitboard with the squares the pawn attacks from the position passed
func pawnAttacks(p *Bitboard, side Color) (attacks Bitboard) {
	notInHFile := *p & ^(*p & files[7])
	notInAFile := *p & ^(*p & files[0])

	if side == White {
		attacks = notInAFile<<7 | notInHFile<<9
	} else {
		attacks = notInAFile>>9 | notInHFile>>7
	}
	return
}

// ------------------------------------------------------------------
// PIECE MOVES GENERATION (BITBOARD)
// ------------------------------------------------------------------

// kingMoves returns a bitboard with the legal moves of the king from the bitboard passed
func kingMoves(k *Bitboard, pos *Position, side Color) (moves Bitboard) {
	withoutKing := *pos
	withoutKing.RemovePiece(*k)
	moves = kingAttacks(k) & ^withoutKing.AttackedSquares(side.Opponent()) & ^pos.Pieces(side)
	return
}

// rookMoves returns a bitboard with the legal moves of the rook from the bitboard passed
func rookMoves(r *Bitboard, pd *PositionData) (moves Bitboard) {
	return rookAttacks(Bsf(*r), pd.allies|pd.enemies) &
		^pd.allies &
		pd.checkRestrictedSquares &
		pinRestrictedSquares(*r, pd.kingPosition, pd.pinnedPieces)
}

// bishopMoves returns a bitboard with the legal moves of the bishop from the bitboard passed
func bishopMoves(b *Bitboard, pd *PositionData) (moves Bitboard) {
	return bishopAttacks(Bsf(*b), pd.allies|pd.enemies) &
		^pd.allies &
		pd.checkRestrictedSquares &
		// checkRestrictedSquares(pd.kingPosition, pd.checkingSliders, pd.checkingNonSliders) &
		pinRestrictedSquares(*b, pd.kingPosition, pd.pinnedPieces)
}

// knightMoves returns a bitboard with the legal moves of the knight from the bitboard passed
func knightMoves(k *Bitboard, pd *PositionData) (moves Bitboard) {
	// If the knight is pinned it can move at all
	if *k&pd.pinnedPieces > 0 {
		return Bitboard(0)
	}
	return knightAttacksTable[Bsf(*k)] & ^pd.allies & pd.checkRestrictedSquares
}

// pawnMoves returns a Bitboard with the squares a pawn can move to in the passed position
func pawnMoves(p *Bitboard, pd *PositionData, side Color) (moves Bitboard) {
	posibleCaptures := pawnAttacks(p, side) & pd.enemies
	posiblesMoves := Bitboard(0)
	emptySquares := ^(pd.allies | pd.enemies)

	if side == White {
		singleMove := *p << 8 & emptySquares
		firstPawnMoveAvailable := (*p & ranks[1]) << 16 & (singleMove << 8) & emptySquares
		posiblesMoves = singleMove | firstPawnMoveAvailable
	} else {
		singleMove := *p >> 8 & emptySquares
		firstPawnMoveAvailable := (*p & ranks[6]) >> 16 & (singleMove >> 8) & emptySquares
		posiblesMoves = singleMove | firstPawnMoveAvailable
	}

	moves = (posibleCaptures | posiblesMoves) & pd.checkRestrictedSquares &
		pinRestrictedSquares(*p, pd.kingPosition, pd.pinnedPieces)
	return
}

// ------------------------------------------------------------------
// PIECE MOVES GENERATION (MOVE LIST)
// ------------------------------------------------------------------

// genKingMoves generates the king moves in the move list
func genKingMoves(from *Bitboard, pos *Position, side Color, ml *moveList) {
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

// genQueenMoves generates the queen moves in the move list
func genQueenMoves(from *Bitboard, ml *moveList, pd *PositionData) {
	toSquares := bishopMoves(from, pd) | rookMoves(from, pd)

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := quiet

		if toSquare&pd.enemies > 0 {
			flag = capture
		}

		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
}

// genRookMoves generates the rook moves in the move list
func genRookMoves(from *Bitboard, ml *moveList, pd *PositionData) {
	toSquares := rookMoves(from, pd)

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := quiet

		if toSquare&pd.enemies > 0 {
			flag = capture
		}

		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
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

// genKnightMoves generates the knight moves in the move list
func genKnightMoves(from *Bitboard, pos *Position, side Color, ml *moveList, pd *PositionData) {
	toSquares := knightMoves(from, pd)

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := quiet

		if toSquare&pd.enemies > 0 {
			flag = capture
		}

		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
}

// genPawnMoves generates the pawn moves in the move list
func genPawnMoves(from *Bitboard, side Color, ml *moveList, pd *PositionData) {
	toSquares := pawnMoves(from, pd, side)

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := pawnMoveFlag(from, &toSquare, pd, side)

		if flag == knightPromotion || flag == knightCapturePromotion {
			addPawnPromotions(ml, from, toSquare, flag)
		} else {
			ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), flag))
		}
	}
}

// addPawnPromotions add the 4 different promotions to the move list
func addPawnPromotions(ml *moveList, from *Bitboard, to Bitboard, flag uint16) {
	promotionTypes := []uint16{
		knightPromotion, bishopPromotion, rookPromotion, queenPromotion,
	}

	capturePromotionTypes := []uint16{
		knightCapturePromotion, bishopCapturePromotion, rookCapturePromotion, queenCapturePromotion,
	}

	if flag == knightPromotion {
		for _, promotionFlag := range promotionTypes {
			ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(to)), promotionFlag))
		}
	}

	if flag == knightCapturePromotion {
		for _, capturePromotionFlag := range capturePromotionTypes {
			ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(to)), capturePromotionFlag))
		}
	}
}

// genEpPawnCaptures generates the enPassant captures on the move list
func genEpPawnCaptures(pos *Position, side Color, ml *moveList) {
	if pos.enPassantTarget == 0 {
		return
	}
	from := potentialEpCapturers(pos, side)

	for from > 0 {
		fromBB := from.NextBit()

		move := *encodeMove(uint16(Bsf(fromBB)), uint16(Bsf(pos.enPassantTarget)), epCapture)

		pos.MakeMove(&move)
		if !pos.Check(side) {
			ml.add(move)
		}
		pos.UnmakeMove(&move)

	}
}

// ------------------------------------------------------------------
// SPECIAL PAWN MOVES GENERATION
// ------------------------------------------------------------------

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
