package board

// Bishop models a bishop piece in chess
type Pawn struct {
	color  rune
	square Bitboard
}

// -------------
// PAWN ♙
// -------------
// Attacks returns all squares that a Pawn attacks in a chess board
func (p *Pawn) Attacks(pos *Position) (attacks Bitboard) {
	notInHFile := p.square & ^(p.square & files[7])
	notInAFile := p.square & ^(p.square & files[0])

	if p.color == WHITE {
		attacks = notInAFile<<7 | notInHFile<<9
	} else {
		attacks = notInAFile>>9 | notInHFile>>7
	}
	return
}

// Moves returns a bitboard with the legal squares the Pawn can move to in a chess position
func (p *Pawn) Moves(pos *Position) (moves Bitboard) {
	posibleCaptures := p.Attacks(pos) & pos.Pieces(opponentSide(p.color))
	posiblesMoves := Bitboard(0)

	if p.color == WHITE {
		singleMove := p.square << 8 & pos.EmptySquares()
		firstPawnMoveAvailable := (p.square & ranks[1]) << 16 & (singleMove << 8) & pos.EmptySquares()
		posiblesMoves = singleMove | firstPawnMoveAvailable
	} else {
		singleMove := p.square >> 8 & pos.EmptySquares()
		firstPawnMoveAvailable := (p.square & ranks[6]) >> 16 & (singleMove >> 8) & pos.EmptySquares()
		posiblesMoves = singleMove | firstPawnMoveAvailable
	}

	moves = (posibleCaptures | posiblesMoves) &
		pinRestrictedDirection(p.square, p.color, pos) &
		checkRestrictedMoves(p.square, p.color, pos)
	return
}

// Square returns the bitboard with the position of the piece
func (p *Pawn) Square() Bitboard {
	return p.square
}

// Color returns the color(side) of the piece
func (p *Pawn) Color() rune {
	return p.color
}

// Returns if the piece is an sliding piece(bishops, rooks, queens)
func (p *Pawn) IsSliding() bool {
	return false
}

// role Returns the role of the piece in the board
func (p *Pawn) role() int {
	if p.color == WHITE {
		return WHITE_PAWN
	} else {
		return BLACK_PAWN
	}
}

// validMoves returns an slice of the valid moves for the Pawn in the position
func (p *Pawn) validMoves(pos *Position) (moves []Move) {
	destinationsBB := p.Moves(pos)
	opponentPieces := pos.Pieces(opponentSide(p.color))
	piece := WHITE_PAWN
	doublePushFrom := ranks[1]
	doublePushTo := ranks[3]
	queeningRank := ranks[7]
	promotions := []int{WHITE_KNIGHT, WHITE_BISHOP, WHITE_ROOK, WHITE_QUEEN}
	if p.color == BLACK {
		queeningRank = ranks[0]
		doublePushFrom = ranks[6]
		doublePushTo = ranks[4]
		piece = BLACK_PAWN
		promotions = []int{BLACK_KNIGHT, BLACK_BISHOP, BLACK_ROOK, BLACK_QUEEN}
	}

	for destinationsBB > 0 {
		destSq := Bitboard(0b1 << Bsf(destinationsBB))

		switch {
		case (opponentPieces & destSq) > 0:
			moves = append(moves, Move{
				from:     squareReference[Bsf(p.square)],
				to:       squareReference[Bsf(destinationsBB)],
				piece:    piece,
				moveType: CAPTURE,
			})
		case (destSq&doublePushTo) > 0 && (p.square&doublePushFrom) > 0:
			moves = append(moves, Move{
				from:     squareReference[Bsf(p.square)],
				to:       squareReference[Bsf(destinationsBB)],
				piece:    piece,
				moveType: PAWN_DOUBLE_PUSH,
			})
		case (destSq & queeningRank) > 0:
			for _, promotedRole := range promotions {
				moves = append(moves, Move{
					from:       squareReference[Bsf(p.square)],
					to:         squareReference[Bsf(destinationsBB)],
					piece:      piece,
					promotedTo: promotedRole,
					moveType:   PROMOTION,
				})
			}
		default:
			moves = append(moves, Move{
				from:     squareReference[Bsf(p.square)],
				to:       squareReference[Bsf(destinationsBB)],
				piece:    piece,
				moveType: NORMAL,
			})
		}
		destinationsBB ^= destSq
	}
	return
}

// pawnMoves returns a bitboard with the squares the pawn can move from the position passed
// TODO: refactor...
func getPawnMoves(p *Bitboard, pos *Position, side rune) (moves []move) {
	destinationsBB := pawnMoves(p, pos, side)
	opponentPieces := pos.Pieces(opponentSide(side))
	piece := WHITE_PAWN
	doublePushFrom := ranks[1]
	doublePushTo := ranks[3]
	queeningRank := ranks[7]
	promotions := []int{WHITE_KNIGHT, WHITE_BISHOP, WHITE_ROOK, WHITE_QUEEN}
	if side == BLACK {
		queeningRank = ranks[0]
		doublePushFrom = ranks[6]
		doublePushTo = ranks[4]
		piece = BLACK_PAWN
		promotions = []int{BLACK_KNIGHT, BLACK_BISHOP, BLACK_ROOK, BLACK_QUEEN}
	}

	for destinationsBB > 0 {
		destSq := destinationsBB.nextOne()

		switch {
		case (opponentPieces & destSq) > 0:
			moves = append(moves, MoveEncode(Bsf(*p), Bsf(destSq), piece, 0, CAPTURE))
		case (destSq&doublePushTo) > 0 && (*p&doublePushFrom) > 0:
			moves = append(moves, MoveEncode(Bsf(*p), Bsf(destSq), piece, 0, PAWN_DOUBLE_PUSH))
		case (destSq & queeningRank) > 0:
			for _, promotedRole := range promotions {
				moves = append(moves, MoveEncode(Bsf(*p), Bsf(destSq), piece, promotedRole, PROMOTION))
			}
		default:
			moves = append(moves, MoveEncode(Bsf(*p), Bsf(destSq), piece, 0, NORMAL))
		}
	}
	return
}

func pawnMoves(p *Bitboard, pos *Position, side rune) (moves Bitboard) {
	posibleCaptures := pawnAttacks(p, pos, side) & pos.Pieces(opponentSide(side))
	posiblesMoves := Bitboard(0)

	if side == WHITE {
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
func pawnAttacks(p *Bitboard, pos *Position, side rune) (attacks Bitboard) {
	notInHFile := *p & ^(*p & files[7])
	notInAFile := *p & ^(*p & files[0])

	if side == WHITE {
		attacks = notInAFile<<7 | notInHFile<<9
	} else {
		attacks = notInAFile>>9 | notInHFile>>7
	}
	return
}
