package aconcagua

// -------------
// PAWN ♙
// -------------

// pawnMoves returns a Bitboard with the squares a pawn can move to in the passed position
func pawnMoves(p *Bitboard, pos *Position, side Color) (moves Bitboard) {
	posibleCaptures := pawnAttacks(p, side) & pos.Pieces(side.Opponent())
	posiblesMoves := Bitboard(0)

	if side == White {
		singleMove := *p << 8 & pos.EmptySquares()
		firstPawnMoveAvailable := (*p & ranks[1]) << 16 & (singleMove << 8) & pos.EmptySquares()
		posiblesMoves = singleMove | firstPawnMoveAvailable
	} else {
		singleMove := *p >> 8 & pos.EmptySquares()
		firstPawnMoveAvailable := (*p & ranks[6]) >> 16 & (singleMove >> 8) & pos.EmptySquares()
		posiblesMoves = singleMove | firstPawnMoveAvailable
	}

	moves = (posibleCaptures | posiblesMoves) &
		pinRestrictedDirection(p, side, pos) &
		checkRestrictedMoves(*p, side, pos)

	return
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

// genPawnMoves generates the pawn moves in the move list
func genPawnMoves(from *Bitboard, pos *Position, side Color, ml *moveList) {
	toSquares := pawnMoves(from, pos, side)

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := pawnMoveFlag(from, &toSquare, pos, side)

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

// pawnMoveFlag returns the move flag for the pawn move
func pawnMoveFlag(from *Bitboard, to *Bitboard, pos *Position, side Color) uint16 {
	fromSq := Bsf(*from)
	toSq := Bsf(*to)
	opponentPieces := pos.Pieces(side.Opponent())
	promotion := lastRank(side) & *to

	switch {
	case promotion > 0 && opponentPieces&*to > 0:
		return knightCapturePromotion
	case promotion > 0:
		return knightPromotion
	case toSq-fromSq == 16 || fromSq-toSq == 16:
		return doublePawnPush
	case opponentPieces&*to > 0:
		return capture
	default:
		return quiet
	}
}

// genEpPawnCaptures generates the enPassant captures on the move list
func genEpPawnCaptures(pos *Position, side Color, ml *moveList) {
	if pos.enPassantTarget == 0 {
		return
	}
	epShift := pos.enPassantTarget >> 8
	if side == Black {
		epShift = epShift << 16
	}

	notInHFile := epShift & ^(epShift & files[7])
	notInAFile := epShift & ^(epShift & files[0])

	from := pos.getBitboards(side)[5] & (notInAFile>>1 | notInHFile<<1)

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

// lastRank returns the rank of the last rank for the side passed
func lastRank(side Color) (rank Bitboard) {
	if side == White {
		rank = ranks[7]
	} else {
		rank = ranks[0]
	}
	return
}
