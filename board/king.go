package board

// King models a king piece in chess
type King struct {
	color  byte
	square Bitboard
}

// -------------
// KING ♔
// -------------
// Attacks returns all squares that a King attacks in a chess board
func (k *King) Attacks(pos *Position) (attacks Bitboard) {
	//  Bitwise displacements for all possible King attacks
	//   ------------------
	//   | <<9 | <<8 | <<7 |
	//   ------------------
	//   | <<1 |  K  | >>1 |
	//   ------------------
	//   | >>7 | >>8 | >>9 |
	//   ------------------
	notInHFile := k.square & ^(k.square & files[7])
	notInAFile := k.square & ^(k.square & files[0])

	attacks = notInAFile<<7 | k.square<<8 | notInHFile<<9 |
		notInHFile<<1 | notInAFile>>1 | notInHFile>>7 |
		k.square>>8 | notInAFile>>9
	return
}

// Moves returns a bitboard with the legal squares the King can move to
func (k *King) Moves(pos *Position) (moves Bitboard) {
	// King can only move to an empty square or capture an opponent piece not defended by
	// opposite side
  withoutKing := pos.RemovePiece(k.square)
	moves = k.Attacks(pos) & ^withoutKing.AttackedSquares(opponentSide(k.color)) & ^pos.Pieces(k.color)
	return
}

// Square returns the bitboard with the position of the piece
func (k *King) Square() Bitboard {
	return k.square
}

// Color returns the color(side) of the piece
func (k *King) Color() byte {
	return k.color
}

// Returns if the piece is an sliding piece(bishops, rooks, queens)
func (k *King) IsSliding() bool {
	return false
}
