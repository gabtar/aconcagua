// Pieces movements in a chess board
package main

import "math/bits"

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
  attacks(pos *Position) Bitboard
  moves(pos *Position) Bitboard // Legal moves -> attacks - occupied by same color piece - pinned in direction
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
}

// -------------
// KING ♔
// -------------
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
  notInHFile := k.square & ^(k.square & files[7])
  notInAFile := k.square & ^(k.square & files[0])

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

// -------------
// KNIGHT ♘
// -------------
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
  notInHFile := k.square & ^(k.square & files[7])
  notInAFile := k.square & ^(k.square & files[0])
  notInABFiles := k.square & ^(k.square & (files[0] | files[1]))
  notInGHFiles := k.square & ^(k.square & (files[7] | files[6]))

  attacks = notInAFile << 15 | notInHFile << 17 | notInGHFiles << 10 |
            notInABFiles << 6 | notInHFile >> 15 | notInAFile >> 17 | 
            notInABFiles >> 10 | notInGHFiles >> 6
  return
}

// moves returns a bitboard with the legal squares the Knight can move to in a chess position
func (k *Knight) moves(pos *Position) (moves Bitboard) {
  // TODO check if the knight is pinned -> the move will result in check
  moves = k.attacks(pos) & ^pos.pieces(k.color)
  return
}

// -------------
// ROOK ♖
// -------------
// attacks returns all squares that a Rook attacks in a chess board
func (r *Rook) attacks(pos *Position) (attacks Bitboard) {
  //  Bitwise displacements for all possible King attacks
  //   ------------------
  //   |     | <<8 |     |
  //   ------------------
  //   | <<1 |  R  | >>1 |
  //   ------------------
  //   |     | >>8 |     |
  //   ------------------

  // TODO, need to check blocked if squares are blocked
  // Get rank and file
  // go up to oponent piece in that file/rank
  // or go up to piece -1 if piece if of same color
  square := r.square.ToStringSlice()[0]

  // Need to find in four directions the 'nearest' piece and attack up to that square
  pieces := ^pos.emptySquares()
  attacks |= files[int(square[0]) - 97] | ranks[int(square[1]) - 49]

  // Posibles Blockers bitboard for rook files and ranks
  blockers := (attacks & pieces & ^r.square)
  fileBlockers := blockers & files[int(square[0]) - 97]
  // rankBlockers := blockers & ranks[int(square[1]) - 49]


  // TODO!!!!!!!
  // Better way of finding on each direction:
  // TrailingZeros64 and LeadingZeros64 can be used to find the distance to the nearest blocker
  // For NORTH/EAST -> TrailingZeros64 and for SOUTH/WEST LeadingZeros64
  // Then get the rays for that blocker in same direction
  // And by negating it(the rayAttack) can be used to get the squares up to the rook can go
  // nortBlk := blockers & raysAttacks[NORTH][bits.TrailingZeros64(uint64(r.square))]
  // nearestNorthBlocker := Bitboard( 0b1 << bits.TrailingZeros64(uint64(nortBlk)))
  // nearestNorthBlocker.Print()

  // Need to split into simple bitboards
  // Map to square bitboard

  // TODO For each direction find the nearest blocker and check color? to find if it can move there
  // FOR RIGHT
  // for _, square := range blockers.ToStringSlice() {
  // }

  // Need to initialize at max value, so the diference will always be smaller 
  var nearestNorthBlocker, nearestSouthBlocker Bitboard = 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF
  // Nearest Horizontal blocker
  // Need to check the nearest if is > or < of r.square and store in a variable result
  for _, sq := range fileBlockers.ToStringSlice() {
    bitboard := sqaureToBitboard([]string{sq})
    // Remap each square to a bitboard
    // Check if higher / lower than r.square
    // store in a variable the nearest of all in each direction
    if bitboard > r.square {
      // Its upstairs the in the file
      if (bitboard - r.square) < (nearestNorthBlocker - r.square) {
        nearestNorthBlocker = bitboard 
      }
    } else {
      // Its downstairs in the file
      if (r.square - bitboard) < (r.square - nearestSouthBlocker) {
        nearestSouthBlocker = bitboard 
      }
    }
  }
  // TODO do the same for right and left blockers
  raysAttacks[NORTH][bits.TrailingZeros64(uint64(nearestNorthBlocker))].Print()

  attacks &= (nearestSouthBlocker | nearestNorthBlocker)
  return
}

// moves returns a bitboard with the legal squares the Rook can move to in a chess position
func (r *Rook) moves(pos *Position) (moves Bitboard) {
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
