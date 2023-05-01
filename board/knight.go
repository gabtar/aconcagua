package board

// Knight models a knight piece in chess
type Knight struct {
	color  rune
	square Bitboard
}

// -------------
// KNIGHT ♘
// -------------
// Attacks returns all squares that a Knight attacks in a chess board
func (k *Knight) Attacks(pos *Position) (attacks Bitboard) {
	//  Bitwise displacements for all possible Knight attacks
	//   -------------------------------
	//   |     |<<15 |     |<<17 |     |
	//   -------------------------------
	//   |<<10 |     |     |     | <<6 |
	//   -------------------------------
	//   |     |     |  K  |     |     |
	//   -------------------------------
	//   |>>6  |     |     |     | >>10|
	//   -------------------------------
	//   |     |>>15 |     |>>17 |     |
	//   -------------------------------

	// Removes moves when in corner squares
	notInHFile := k.square & ^(k.square & files[7])
	notInAFile := k.square & ^(k.square & files[0])
	notInABFiles := k.square & ^(k.square & (files[0] | files[1]))
	notInGHFiles := k.square & ^(k.square & (files[7] | files[6]))

	attacks = notInAFile<<15 | notInHFile<<17 | notInGHFiles<<10 |
		notInABFiles<<6 | notInHFile>>15 | notInAFile>>17 |
		notInABFiles>>10 | notInGHFiles>>6
	return
}

// Moves returns a bitboard with the legal squares the Knight can move to in a chess position
func (k *Knight) Moves(pos *Position) (moves Bitboard) {
  // If the knight is pinned it can move at all
  if isPinned(k.square, k.color, pos) {
    return Bitboard(0)
  }

	moves = k.Attacks(pos) & ^pos.Pieces(k.color) &
          checkRestrictedMoves(k.square, k.color, pos)
  return
}

// Square returns the bitboard with the position of the piece
func (k *Knight) Square() Bitboard {
	return k.square
}

// Color returns the color(side) of the piece
func (k *Knight) Color() rune {
	return k.color
}

// Returns if the piece is an sliding piece(bishops, rooks, queens)
// TODO refactor to a function instead of part of the interface?
func (k *Knight) IsSliding() bool {
	return false
}

// role Returns the role of the piece in the board
func (k *Knight) role() int {
  if k.color == WHITE {
    return WHITE_KNIGHT
  } else {
    return BLACK_KNIGHT
  }
}
