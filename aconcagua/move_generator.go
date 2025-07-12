package aconcagua

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

// pawnMoveFlag returns the move flag for the pawn move
func pawnMoveFlag(from *Bitboard, to *Bitboard, pd *PositionData, side Color) []uint16 {
	fromSq := Bsf(*from)
	toSq := Bsf(*to)
	promotion := lastRank(side) & *to

	switch {
	case promotion > 0 && pd.enemies&*to > 0:
		return []uint16{knightCapturePromotion, bishopCapturePromotion, rookCapturePromotion, queenCapturePromotion}
	case promotion > 0:
		return []uint16{knightPromotion, bishopPromotion, rookPromotion, queenPromotion}
	case toSq-fromSq == 16 || fromSq-toSq == 16:
		return []uint16{doublePawnPush}
	case pd.enemies&*to > 0:
		return []uint16{capture}
	default:
		return []uint16{quiet}
	}
}

// potentialEpCapturers returns a bitboard with the potential pawn that can caputure enPassant
func potentialEpCapturers(pos *Position, side Color) (epCaptures Bitboard) {
	epShift := pos.enPassantTarget >> 8
	if side == Black {
		epShift = epShift << 16
	}
	notInHFile := epShift & ^(epShift & files[7])
	notInAFile := epShift & ^(epShift & files[0])

	epCaptures |= pos.getBitboards(side)[Pawn] & (notInAFile>>1 | notInHFile<<1)
	return
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

// ------------------------------------------------------------------
// PIECE MOVES GENERATION (MOVE LIST)
// ------------------------------------------------------------------

// genMovesFromTargets generates the moves from the square passed to the targets passed in a MoveList
func genMovesFromTargets(from *Bitboard, targets Bitboard, ml *MoveList, pd *PositionData) {
	for targets > 0 {
		toSquare := targets.NextBit()
		flag := quiet
		if toSquare&pd.enemies > 0 {
			flag = capture
		}
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
}

// genCastleMoves generates the castles moves availabes in the move list
func genCastleMoves(from *Bitboard, pos *Position, ml *MoveList) {
	if canCastleShort(from, pos, pos.Turn) {
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(*from<<2)), kingsideCastle))
	}
	if canCastleLong(from, pos, pos.Turn) {
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(*from>>2)), queensideCastle))
	}
}

// genPawnMovesFromTarget generates the pawn moves in the move list
func genPawnMovesFromTarget(from *Bitboard, targets Bitboard, side Color, ml *MoveList, pd *PositionData) {
	for targets > 0 {
		toSquare := targets.NextBit()
		flags := pawnMoveFlag(from, &toSquare, pd, side)

		for i := range flags {
			ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), flags[i]))
		}
	}
}

// genPawnCapturesMoves generates the pawn captures in the move list
func genPawnCapturesMoves(from *Bitboard, side Color, ml *MoveList, pd *PositionData) {
	toSquares := pawnMoves(from, pd, side) & pd.enemies

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flags := pawnMoveFlag(from, &toSquare, pd, side)

		for i := range flags {
			ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), flags[i]))
		}
	}
}

// genEnPassantCaptures generates the enPassant captures on the move list
func genEnPassantCaptures(pos *Position, side Color, ml *MoveList) {
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

// generateCaptures generates all captures in the position and stores them in the move list
func (pos *Position) generateCaptures(ml *MoveList) {
	pd := pos.generatePositionData()
	bitboards := pos.getBitboards(pos.Turn)

	for piece, bb := range bitboards {
		for bb > 0 {
			pieceBB := bb.NextBit()
			switch piece {
			case King:
				genMovesFromTargets(&pieceBB, kingMoves(&pieceBB, pos, pos.Turn)&pd.enemies, ml, &pd)
			case Queen:
				genMovesFromTargets(&pieceBB, (rookMoves(&pieceBB, &pd)|bishopMoves(&pieceBB, &pd))&pd.enemies, ml, &pd)
			case Rook:
				genMovesFromTargets(&pieceBB, rookMoves(&pieceBB, &pd)&pd.enemies, ml, &pd)
			case Bishop:
				genMovesFromTargets(&pieceBB, bishopMoves(&pieceBB, &pd)&pd.enemies, ml, &pd)
			case Knight:
				genMovesFromTargets(&pieceBB, knightMoves(&pieceBB, &pd)&pd.enemies, ml, &pd)
			case Pawn:
				genPawnCapturesMoves(&pieceBB, pos.Turn, ml, &pd)
			}
		}
	}
	// TODO: add en passant captures here??. Need to fix see first(because 'to' square is an empty square)
}

// generateNonCaptures generates all non captures in the position and stores them in the move list
func (pos *Position) generateNonCaptures(ml *MoveList) {
	pd := pos.generatePositionData()
	bitboards := pos.getBitboards(pos.Turn)

	for piece, bb := range bitboards {
		for bb > 0 {
			pieceBB := bb.NextBit()
			switch piece {
			case King:
				genMovesFromTargets(&pieceBB, kingMoves(&pieceBB, pos, pos.Turn)&^pd.enemies, ml, &pd)
				genCastleMoves(&pieceBB, pos, ml)
			case Queen:
				genMovesFromTargets(&pieceBB, (rookMoves(&pieceBB, &pd)|bishopMoves(&pieceBB, &pd))&^pd.enemies, ml, &pd)
			case Rook:
				genMovesFromTargets(&pieceBB, rookMoves(&pieceBB, &pd)&^pd.enemies, ml, &pd)
			case Bishop:
				genMovesFromTargets(&pieceBB, bishopMoves(&pieceBB, &pd)&^pd.enemies, ml, &pd)
			case Knight:
				genMovesFromTargets(&pieceBB, knightMoves(&pieceBB, &pd)&^pd.enemies, ml, &pd)
			case Pawn:
				genPawnMovesFromTarget(&pieceBB, pawnMoves(&pieceBB, &pd, pos.Turn)&^pd.enemies, pos.Turn, ml, &pd)
			}
		}
	}
	genEnPassantCaptures(pos, pos.Turn, ml)
}
