// Pieces movements in a chess board
package main

import (
	"math/bits"
)

// King models a king piece in chess
type King struct {
	color  byte
	square Bitboard
}

// Rook models a rook piece in chess
type Rook struct {
  color byte
  square Bitboard
}

// Knight models a knight piece in chess
type Knight struct {
	color  byte
	square Bitboard
}

// Piece is the interface that has all methods for chess pieces
type Piece interface {
  Attacks(pos *Position) Bitboard
  Moves(pos *Position) Bitboard // Legal moves -> attacks - occupied by same color piece - pinned in direction
  Square() Bitboard
  Color() byte
  IsSliding() bool
}

// rays contains all rays for a given square in all possible 8 directions
// https://www.chessprogramming.org/Classical_Approach
const NORTH uint64 = 0
const SOUTH uint64 = 1
const WEST uint64 = 2
const EAST uint64 = 3

// rays contains all rays for a given square in all possible 8 directions
// useful with calculating attacks/moves on sliding pieces(Rook, Bishop, Queens)
// https://www.chessprogramming.org/Classical_Approach
var raysAttacks [4][64]Bitboard = [4][64]Bitboard{
  NORTH: {0x101010101010100, 0x202020202020200, 0x404040404040400, 0x808080808080800,
          0x1010101010101000, 0x2020202020202000, 0x4040404040404000, 0x8080808080808000,
          0x101010101010000, 0x202020202020000, 0x404040404040000, 0x808080808080000,
          0x1010101010100000, 0x2020202020200000, 0x4040404040400000, 0x8080808080800000,
          0x101010101000000, 0x202020202000000, 0x404040404000000, 0x808080808000000,
          0x1010101010000000, 0x2020202020000000, 0x4040404040000000, 0x8080808080000000,
          0x101010100000000, 0x202020200000000, 0x404040400000000, 0x808080800000000,
          0x1010101000000000, 0x2020202000000000, 0x4040404000000000, 0x8080808000000000,
          0x101010000000000, 0x202020000000000, 0x404040000000000, 0x808080000000000,
          0x1010100000000000, 0x2020200000000000, 0x4040400000000000, 0x8080800000000000,
          0x101000000000000, 0x202000000000000, 0x404000000000000, 0x808000000000000,
          0x1010000000000000, 0x2020000000000000, 0x4040000000000000, 0x8080000000000000,
          0x100000000000000, 0x200000000000000, 0x400000000000000, 0x800000000000000,
          0x1000000000000000, 0x2000000000000000, 0x4000000000000000, 0x8000000000000000,
          0x000000000000000, 0x000000000000000, 0x000000000000000, 0x000000000000000,
          0x0000000000000000, 0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
  },
  EAST: {0, 0x1, 0x3, 0x7, 0xf, 0x1f, 0x3f, 0x7f,
         0, 0x100, 0x300, 0x700, 0xf00, 0x1f00, 0x3f00, 0x7f00,
         0, 0x10000, 0x30000, 0x70000, 0xf0000, 0x1f0000, 0x3f0000, 0x7f0000,
         0, 0x1000000, 0x3000000, 0x7000000, 0xf000000, 0x1f000000, 0x3f000000, 0x7f000000,
         0, 0x100000000, 0x300000000, 0x700000000, 0xf00000000, 0x1f00000000, 0x3f00000000, 0x7f00000000,
         0, 0x10000000000, 0x30000000000, 0x70000000000, 0xf0000000000, 0x1f0000000000, 0x3f0000000000, 0x7f0000000000,
         0, 0x1000000000000, 0x3000000000000, 0x7000000000000, 0xf000000000000, 0x1f000000000000, 0x3f000000000000, 0x7f000000000000,
         0, 0x100000000000000, 0x300000000000000, 0x700000000000000, 0xf00000000000000, 0x1f00000000000000, 0x3f00000000000000, 0x7f00000000000000,
  },
  SOUTH: {0, 0, 0, 0, 0, 0, 0, 0,
          0x1, 0x2, 0x4, 0x8, 0x10, 0x20, 0x40, 0x80,
          0x101, 0x202, 0x404, 0x808, 0x1010, 0x2020, 0x4040, 0x8080,
          0x10101, 0x20202, 0x40404, 0x80808, 0x101010, 0x202020, 0x404040, 0x808080,
          0x1010101, 0x2020202, 0x4040404, 0x8080808, 0x10101010, 0x20202020, 0x40404040, 0x80808080,
          0x101010101, 0x202020202, 0x404040404, 0x808080808, 0x1010101010, 0x2020202020, 0x4040404040, 0x8080808080,
          0x10101010101, 0x20202020202, 0x40404040404, 0x80808080808, 0x101010101010, 0x202020202020, 0x404040404040, 0x808080808080,
          0x1010101010101, 0x2020202020202, 0x4040404040404, 0x8080808080808, 0x10101010101010, 0x20202020202020, 0x40404040404040, 0x80808080808080,
  },
  WEST: {0xfe, 0xfc, 0xf8, 0xf0, 0xe0, 0xc0, 0x80, 0,
         0xfe << 8, 0xfc << 8, 0xf8 << 8, 0xf0 << 8, 0xe0 << 8, 0xc0 << 8, 0x80 << 8, 0,
         0xfe << 16, 0xfc << 16, 0xf8 << 16, 0xf0 << 16, 0xe0 << 16, 0xc0 << 16, 0x80 << 16, 0,
         0xfe << 24, 0xfc << 24, 0xf8 << 24, 0xf0 << 24, 0xe0 << 24, 0xc0 << 24, 0x80 << 24, 0,
         0xfe << 32, 0xfc << 32, 0xf8 << 32, 0xf0 << 32, 0xe0 << 32, 0xc0 << 32, 0x80 << 32, 0,
         0xfe << 40, 0xfc << 40, 0xf8 << 40, 0xf0 << 40, 0xe0 << 40, 0xc0 << 40, 0x80 << 40, 0,
         0xfe << 48, 0xfc << 48, 0xf8 << 48, 0xf0 << 48, 0xe0 << 48, 0xc0 << 48, 0x80 << 48, 0,
         0xfe << 56, 0xfc << 56, 0xf8 << 56, 0xf0 << 56, 0xe0 << 56, 0xc0 << 56, 0x80 << 56, 0,
  },
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

  attacks = notInAFile << 7 | k.square << 8 | notInHFile << 9 |
            notInHFile << 1 | notInAFile >> 1 | notInHFile >> 7 |
            k.square >> 8 | notInAFile >> 9
	return
}

// Moves returns a bitboard with the legal squares the King can move to
func (k *King) Moves(pos *Position) (moves Bitboard) {
  // King can only move to an empty square or capture an opponent piece not defended by 
  // opposite side
  moves = k.Attacks(pos) & ^pos.attackedSquares(opponentSide(k.color)) & ^pos.pieces(k.color)
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

  attacks = notInAFile << 15 | notInHFile << 17 | notInGHFiles << 10 |
            notInABFiles << 6 | notInHFile >> 15 | notInAFile >> 17 | 
            notInABFiles >> 10 | notInGHFiles >> 6
  return
}

// Moves returns a bitboard with the legal squares the Knight can move to in a chess position
func (k *Knight) Moves(pos *Position) (moves Bitboard) {
  // TODO check if the knight is pinned -> the move will result in check
  moves = k.Attacks(pos) & ^pos.pieces(k.color)
  return
}

// Square returns the bitboard with the position of the piece
func (k *Knight) Square() Bitboard {
  return k.square
}

// Color returns the color(side) of the piece
func (k *Knight) Color() byte {
  return k.color
}

// Returns if the piece is an sliding piece(bishops, rooks, queens)
func (k *Knight) IsSliding() bool {
  return false
}

// -------------
// ROOK ♖
// -------------
// Attacks returns all squares that a Rook attacks in a chess board
func (r *Rook) Attacks(pos *Position) (attacks Bitboard) {
  // Uses the raysAttacks array for each direction. -> Classical Approach
  // In each direction computes the nearest blocker piece(any color) and gets
  // the raysAttacks from that piece to the 'end' of the board in that direction.
  // By negating/inverting the rays(bitboard) to the full column/row (bitboard) 
  // we get the squares the rook can reach -> attack(if black pieces)/defend(if
  // it's a same color piece)

  square := r.square.ToStringSlice()[0]

  // Need to find in four directions the 'nearest' piece and attack up to that square
  pieces := ^pos.emptySquares()
  attacks |= files[int(square[0]) - 97] | ranks[int(square[1]) - 49]

  // Posibles Blockers bitboard for rook files and ranks
  blockers := (attacks & pieces)

  // North attacks -> up to nearest north piece
  northBlockers := blockers & raysAttacks[NORTH][bits.TrailingZeros64(uint64(r.square))]
  nearestNorthBlocker := Bitboard(0b1 << bits.TrailingZeros64(uint64(northBlockers)))
  // Need to do this because when no blockers TrailingZeros64 returns 64 and goes out of index/null pointer error
  if nearestNorthBlocker > 0 {
    attacks &= ^raysAttacks[NORTH][bits.TrailingZeros64(uint64(nearestNorthBlocker))]
  }

  // East attacks -> up to nearest east piece
  eastBlockers := blockers & raysAttacks[EAST][bits.TrailingZeros64(uint64(r.square))]
  nearestEastBlocker := Bitboard(0b1 << bits.TrailingZeros64(uint64(eastBlockers)))
  if nearestEastBlocker > 0 {
    attacks &= ^raysAttacks[EAST][bits.TrailingZeros64(uint64(nearestEastBlocker))]
  }

  // South attacks -> up to the neaest south piece
  southBlockers := blockers & raysAttacks[SOUTH][bits.TrailingZeros64(uint64(r.square))]
  nearestSouthBlocker := Bitboard(0b1 << bits.TrailingZeros64(uint64(southBlockers)))
  if nearestSouthBlocker > 0 {
    attacks &= ^raysAttacks[SOUTH][bits.TrailingZeros64(uint64(nearestSouthBlocker))]
  }

  // West attacks -> up to the neaest west piece
  westBlockers := blockers & raysAttacks[WEST][bits.TrailingZeros64(uint64(r.square))]
  nearestWestBlocker := Bitboard(0b1 << bits.TrailingZeros64(uint64(westBlockers)))
  if nearestWestBlocker > 0 {
    attacks &= ^raysAttacks[WEST][bits.TrailingZeros64(uint64(nearestWestBlocker))]
  }

  return attacks & ^r.square
}

// Moves returns a bitboard with the legal squares the Rook can move to in a chess position
func (r *Rook) Moves(pos *Position) (moves Bitboard) {
  // TODO, i need to validate when the piece is pinned
  // For now its only the attacks and removes the blocked squares by same color pieces
  attacks := r.Attacks(pos) & ^pos.pieces(r.color)

  if pos.check(r.color) {
    // TODO, if king(same color) is in check can only moves if the rook blocks the check
    // Check -> See if the rook can capture the opponent piece(on single check)
    //       -> See if there is a move that can block the checking piece attack(only for slider pieces)
    // Supose for now that there is only one 'single check' 
    // When double check -> the rook cannot do nothing

    // Obtener la pieza que está dando jaque
    // ver si la puedo capturar
    // si es a distancia ver tambien si hay alguna jugada que permita blockear el rayo
    checkingPieces := pos.checkingPieces(r.color)

    if len(checkingPieces) == 1 {
      checker := checkingPieces[0]
      // Check if i can obstruct the path when sliding piece
      if checker.IsSliding() {
        // Calculate the direction from piece to king first
        direction := getDirection(checker.Square(), pos.KingPosition(r.color))
        // Get the ray from piece to king
        // check if there is one move that can block in that direction
        moves |= raysAttacks[direction][bits.TrailingZeros64(uint64(checker.Square()))] & attacks
      }
      // Check if i can capture the piece
      moves |= checker.Square() & attacks
      return
    } else {
      // No se puede mover la pieza
      return Bitboard(0)
    }
  }
  return attacks
}

// Square returns the bitboard with the position of the piece
func (r *Rook) Square() Bitboard {
  return r.square
}

// Color returns the color(side) of the piece
func (r *Rook) Color() byte {
  return r.color
}

// Returns if the piece is an sliding piece(bishops, rooks, queens)
func (r *Rook) IsSliding() bool {
  return true
}

// Helper functions

// opponentSide returns the opposite color of the passed
func opponentSide(color byte) byte {
  if color == BLACK {
    return WHITE
  }
  return BLACK
}

// getDirection returns the direction between 2 bitboards containing only 1 piece each one
func getDirection(piece1 Bitboard, piece2 Bitboard) (dir uint64) {
  // TODO add bishop rays direction
  // Check displacement between bitboards?
  //   ------------------
  //   | <<9 | <<8 | <<7 |
  //   ------------------
  //   | <<1 |  P  | >>1 |
  //   ------------------
  //   | >>7 | >>8 | >>9 |
  //   ------------------
  filePiece1 := bits.TrailingZeros64(uint64(piece1)) / 8
  filePiece2 := bits.TrailingZeros64(uint64(piece2)) / 8
  rankPiece1 := bits.TrailingZeros64(uint64(piece1)) % 8
  rankPiece2 := bits.TrailingZeros64(uint64(piece2)) % 8
  fileDiff := filePiece1 - filePiece2
  rankDiff := rankPiece1 - rankPiece2

  switch {
  case fileDiff == 1 && rankDiff == 0:
    dir = SOUTH
  case fileDiff == -1 && rankDiff == 0:
    dir = NORTH
  case fileDiff == 0 && rankDiff == 1:
    dir = EAST
  case fileDiff == 0 && rankDiff == -1:
    dir = WEST
  }
  return
} 
