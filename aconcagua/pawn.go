package aconcagua

// -------------
// PAWN ♙
// -------------

// pawnMoves returns a bitboard with the squares the pawn can move from the position passed
// TODO: refactor...
func getPawnMoves(p *Bitboard, pos *Position, side Color) (moves []Move) {
	destinationsBB := pawnMoves(p, pos, side)
	opponentPieces := pos.Pieces(side.Opponent())
	piece := pieceOfColor[Pawn][side]
	doublePushFrom := ranks[1]
	doublePushTo := ranks[3]
	queeningRank := ranks[7]
	promotions := []Piece{WhiteKnight, WhiteBishop, WhiteRook, WhiteQueen}
	if side == Black {
		queeningRank = ranks[0]
		doublePushFrom = ranks[6]
		doublePushTo = ranks[4]
		promotions = []Piece{BlackKnight, BlackBishop, BlackRook, BlackQueen}
	}

	for destinationsBB > 0 {
		destSq := destinationsBB.NextBit()
		move := newMove().
			setFromSq(Bsf(*p)).
			setToSq(Bsf(destSq)).
			setPiece(piece).
			setMoveType(Normal).
			setEpTargetBefore(pos.enPassantTarget).
			setRule50Before(pos.halfmoveClock).
			setCastleRightsBefore(pos.castlingRights)

		switch {
		case (destSq & queeningRank) > 0: // NOTE: This must be  first because of promotion captures..
			move.setMoveType(Promotion)
			for _, promotedRole := range promotions {
				if (destSq & opponentPieces) > 0 {
					capturedPiece, _ := pos.PieceAt(squareReference[Bsf(destSq)])
					move.setCapturedPiece(capturedPiece)
				}
				// FIX: temporary fix
				move2 := *move
				move2.setPromotedTo(promotedRole)
				moves = append(moves, move2)
			}
		case (pos.enPassantTarget > 0) && (pos.enPassantTarget&destSq) > 0:
			move.setMoveType(EnPassant)
			moves = append(moves, *move)
		case (opponentPieces & destSq) > 0:
			capturedPiece, _ := pos.PieceAt(squareReference[Bsf(destSq)])
			move.setMoveType(Capture).setCapturedPiece(capturedPiece)
			moves = append(moves, *move)
		case (destSq&doublePushTo) > 0 && (*p&doublePushFrom) > 0:
			move.setMoveType(PawnDoublePush)
			moves = append(moves, *move)
		default:
			moves = append(moves, *move)
		}
	}
	return
}

// pawnMoves returns a Bitboard with the squares a pawn can move to in the passed position
func pawnMoves(p *Bitboard, pos *Position, side Color) (moves Bitboard) {
	posibleCaptures := pawnAttacks(p, side) & pos.Pieces(side.Opponent())
	posibleEnPassant := pawnEnPassantCaptures(p, pos, side)
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

	moves = (posibleCaptures|posiblesMoves)&
		pinRestrictedDirection(*p, side, pos)&
		checkRestrictedMoves(*p, side, pos) |
		posibleEnPassant

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

// pawnEnPassantCaptures the bitboard with the squares the pawn can capture en passant
func pawnEnPassantCaptures(p *Bitboard, pos *Position, side Color) (enPassant Bitboard) {
	caughtPawn := pos.enPassantTarget >> 8
	if side == Black {
		caughtPawn = pos.enPassantTarget << 8
	}
	afterEp := *pos
	afterEp.RemovePiece(caughtPawn)
	afterEp.RemovePiece(*p)
	afterEp.Bitboards[pieceOfColor[Pawn][side]] |= pos.enPassantTarget

	if pos.enPassantTarget == 0 || afterEp.Check(side) {
		return
	}

	if pos.CheckingPieces(side) == caughtPawn && (pawnAttacks(p, side)&pos.enPassantTarget) > 0 {
		if side == White {
			enPassant |= caughtPawn << 8
		} else {
			enPassant |= caughtPawn >> 8
		}
	}

	enPassant |= pos.enPassantTarget &
		pawnAttacks(p, side) &
		pinRestrictedDirection(*p, side, pos) &
		checkRestrictedMoves(*p, side, pos)

	return
}

// newPawnMoves returns a moves array with the new pawn moves in chessMove format
func newPawnMoves(pos *Position, pawn *Bitboard, side Color) (moves []chessMove) {
	toSquares := pawnMoves(pawn, pos, side)

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flag := pawnMoveFlag(pawn, &toSquare, pos, side)

		if flag == knightPromotion || flag == knightCapturePromotion {
			moves = addPawnPromotions(moves, pawn, toSquare, flag)
		} else {
			moves = append(moves, encodeMove(uint16(Bsf(*pawn)), uint16(Bsf(toSquare)), flag))
		}
	}

	return moves
}

func addPawnPromotions(moves []chessMove, from *Bitboard, to Bitboard, flag uint16) []chessMove {
	promotionTypes := []uint16{
		knightPromotion, bishopPromotion, rookPromotion, queenPromotion,
	}

	capturePromotionTypes := []uint16{
		knightCapturePromotion, bishopCapturePromotion, rookCapturePromotion, queenCapturePromotion,
	}

	if flag == knightPromotion {
		for _, promotionFlag := range promotionTypes {
			moves = append(moves, encodeMove(uint16(*from), uint16(to), promotionFlag))
		}
	}

	if flag == knightCapturePromotion {
		for _, capturePromotionFlag := range capturePromotionTypes {
			moves = append(moves, encodeMove(uint16(*from), uint16(to), capturePromotionFlag))
		}
	}

	return moves
}

// pawnMoveFlag returns the move flag for the pawn move
func pawnMoveFlag(from *Bitboard, to *Bitboard, pos *Position, side Color) uint16 {
	availableEnPassant := pawnEnPassantCaptures(from, pos, side)
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
	case availableEnPassant&*to > 0:
		return EnPassant
	default:
		return quiet
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
