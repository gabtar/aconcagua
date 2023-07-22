package board

// -------------
// PAWN ♙
// -------------

// pawnMoves returns a bitboard with the squares the pawn can move from the position passed
// TODO: refactor...
func getPawnMoves(p *Bitboard, pos *Position, side Color) (moves []Move) {
	destinationsBB := pawnMoves(p, pos, side)
	opponentPieces := pos.Pieces(side.opponent())
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
	// TODO: Refactor/improve....
	oldEpTarget := 0
	if pos.enPassantTarget > 0 {
		oldEpTarget = Bsf(pos.enPassantTarget)
	}

	for destinationsBB > 0 {
		destSq := destinationsBB.nextOne()

		switch {
		case (opponentPieces & destSq) > 0:
			capturedPiece, _ := pos.PieceAt(destSq.ToStringSlice()[0])
			moves = append(moves, MoveEncode(Bsf(*p), Bsf(destSq), int(piece), 0, CAPTURE, int(capturedPiece), oldEpTarget))
		case (destSq&doublePushTo) > 0 && (*p&doublePushFrom) > 0:
			moves = append(moves, MoveEncode(Bsf(*p), Bsf(destSq), int(piece), 0, PAWN_DOUBLE_PUSH, 0, oldEpTarget))
		case (destSq & queeningRank) > 0:
			for _, promotedRole := range promotions {
				moves = append(moves, MoveEncode(Bsf(*p), Bsf(destSq), int(piece), int(promotedRole), PROMOTION, 0, oldEpTarget))
			}
		default:
			moves = append(moves, MoveEncode(Bsf(*p), Bsf(destSq), int(piece), 0, NORMAL, 0, oldEpTarget))
		}
	}
	return
}

// pawnMoves returns a Bitboard with the squares a pawn can move to in the passed position
func pawnMoves(p *Bitboard, pos *Position, side Color) (moves Bitboard) {
	posibleCaptures := pawnAttacks(p, pos, side) & pos.Pieces(side.opponent())
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
		pinRestrictedDirection(*p, side, pos) &
		checkRestrictedMoves(*p, side, pos)
	return
}

// pawnAttacks returns a bitboard with the squares the pawn attacks from the position passed
func pawnAttacks(p *Bitboard, pos *Position, side Color) (attacks Bitboard) {
	notInHFile := *p & ^(*p & files[7])
	notInAFile := *p & ^(*p & files[0])

	if side == White {
		attacks = notInAFile<<7 | notInHFile<<9
	} else {
		attacks = notInAFile>>9 | notInHFile>>7
	}
	return
}
