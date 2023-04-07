// Pieces movements in a chess board
package main

// King models a king piece in chess
type King struct {
	color  byte
	square Bitboard
}

// Knight models a knight piece in chess
type Knight struct {
	color  byte
	square Bitboard
}

// Piece is the interface that has all methods for chess pieces
type Piece interface {
  attacks(pos *Position) Bitboard
  moves(pos *Position) Bitboard // Legal moves -> attacks - occupied by same color piece - pinned in direction
}

// attacks returns all squares that a King attacks in a chess board
func (k *King) attacks(pos *Position) (attacks Bitboard) {
  //  Bitwise displacements for all possible King attacks
  //   ------------------
  //   | <<9 | <<8 | <<7 |
  //   ------------------
  //   | <<1 |  K  | >>1 |
  //   ------------------
  //   | >>7 | >>8 | >>9 |
  //   ------------------
  notInHFile := k.square & ^(k.square & hFile)
  notInAFile := k.square & ^(k.square & aFile)

  attacks = notInAFile << 7 | k.square << 8 | notInHFile << 9 |
            notInHFile << 1 | notInAFile >> 1 | notInHFile >> 7 |
            k.square >> 8 | notInAFile >> 9
	return
}

// moves returns a bitboard with the legal squares the King can move to
func (k *King) moves(pos *Position) (moves Bitboard) {
  // King can only move to an empty square or capture an opponent piece not defended by 
  // opposite side
  moves = k.attacks(pos) & ^pos.attackedSquares(opponentSide(k.color)) & ^pos.pieces(k.color)
  return
}

// attacks returns all squares that a Knight attacks in a chess board
func (k *Knight) attacks(pos *Position) (attacks Bitboard) {
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
  notInHFile := k.square & ^(k.square & hFile)
  notInAFile := k.square & ^(k.square & aFile)
  notInABFiles := k.square & ^(k.square & (aFile | aFile << 1))
  notInGHFiles := k.square & ^(k.square & (hFile | hFile >> 1))

  attacks = notInAFile << 15 | notInHFile << 17 | notInGHFiles << 10 |
            notInABFiles << 6 | notInHFile >> 15 | notInAFile >> 17 | 
            notInABFiles >> 10 | notInGHFiles >> 6
  return
}

// attacks returns all squares that a Knight attacks in a chess board
func (k *Knight) moves(pos *Position) (moves Bitboard) {
  // TODO check if the knight is pinned -> the move will result in check
  moves = k.attacks(pos) & ^pos.pieces(k.color)
  return
}

// Helper functions

// opponentSide returns the opposite color of the passed
func opponentSide(color byte) byte {
  if color == BLACK {
    return WHITE
  }
  return BLACK
}
