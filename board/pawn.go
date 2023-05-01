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
    attacks = notInAFile << 7 | notInHFile << 9
  } else {
    attacks = notInAFile >> 7 | notInHFile >> 9
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
