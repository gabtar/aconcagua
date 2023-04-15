package board

import (
	"errors"
)

const WHITE byte = 0
const BLACK byte = 1

// References for pieces role/color for bitboards in position struct
const WHITE_KING int = 0
const WHITE_QUEEN int = 1
const WHITE_ROOK int = 2
const WHITE_BISHOP int = 3
const WHITE_KNIGHT int = 4
const WHITE_PAWN int = 5
const BLACK_KING int = 6
const BLACK_QUEEN int = 7
const BLACK_ROOK int = 8
const BLACK_BISHOP int = 9
const BLACK_KNIGHT int = 10
const BLACK_PAWN int = 11

// Maps squares to uint64 (the index of the array is the bit position on a bitboard that represents that square)
var squareMap = []string{"a1", "b1", "c1", "d1", "e1", "f1", "g1", "h1",
                         "a2", "b2", "c2", "d2", "e2", "f2", "g2", "h2",
                         "a3", "b3", "c3", "d3", "e3", "f3", "g3", "h3",
                         "a4", "b4", "c4", "d4", "e4", "f4", "g4", "h4",
                         "a5", "b5", "c5", "d5", "e5", "f5", "g5", "h5",
                         "a6", "b6", "c6", "d6", "e6", "f6", "g6", "h6",
                         "a7", "b7", "c7", "d7", "e7", "f7", "g7", "h7",
                         "a8", "b8", "c8", "d8", "e8", "f8", "g8", "h8"}

// Position contains all information about a chess position
type Position struct {
	// bitboards piece order -> King, Queen, Rook, Bishop, Knight, Pawn (first white, second black)
	bitboards [12]Bitboard
	turn      byte
}

type IPosition interface {
	PieceAt(square string) (Piece, error)
  AddPiece(role int, square string)
  EmptySquares() Bitboard
  AttackedSquares(side byte) Bitboard
  CheckingPieces(side byte) []Piece
  Pieces(side byte) Bitboard
  Check(side byte) bool
  KingPosition(side byte) Bitboard
}

// pieceAt returns a Piece at the given square coordinate in the Position
func (pos *Position) PieceAt(square string) (piece Piece, e error) {
  bitboardSquare := sqaureToBitboard([]string{square})

	for role, bitboard := range pos.bitboards {
		if bitboard & bitboardSquare > 1 {
      piece = makePiece(role, bitboardSquare)
		}
	}
  if piece == nil {
    return piece, errors.New("No piece")
  }
	return piece, nil
}

// addPiece adds a new Piece to the Position
func (pos *Position) AddPiece(role int, square string) {
  bitboardSquare := sqaureToBitboard([]string{square})
  pos.bitboards[role] |= bitboardSquare

}

// emptySquares returns a Bitboard with the empty sqaures of the position
func (pos *Position) EmptySquares() (emptySquares Bitboard) {
  // Set all as empty
  emptySquares = 0xFFFFFFFFFFFFFFFF

	for _, bitboard := range pos.bitboards {
    emptySquares &= ^bitboard
	}
  return
}

// attackedSquares returns a bitboard with all squares attacked by the passed side
func (pos *Position) AttackedSquares(side byte) (attackedSquares Bitboard) {
  startingBitboard := 0
  if side != WHITE {
    startingBitboard = 6
  }

  for currentBitboard := startingBitboard; currentBitboard - startingBitboard < 6 ; currentBitboard++ {
    for _, square := range pos.bitboards[currentBitboard].ToStringSlice() {
      piece, err := pos.PieceAt(square)
      // TODO remove later. I need this check for now because not all pieces are yet implemented
      if err != nil {
        continue
      }
      attackedSquares |= piece.Attacks(pos)
    }
  }

  return
}

// checkingPieces returns an slice of Piece{} that are checking the passed side king
func (pos *Position) CheckingPieces(side byte) (pieces []Piece) {
  if !pos.Check(side) {
    return
  }

  kingSq := pos.bitboards[WHITE_KING]
  if side != WHITE {
    kingSq = pos.bitboards[BLACK_KING]
  }
  // iterate over all opponent pieces an add the ones that are attacking the king
  for _, sq := range pos.Pieces(opponentSide(side)).ToStringSlice() {
    piece, _ := pos.PieceAt(sq)

    if (kingSq & piece.Attacks(pos)) > 0 {
      pieces = append(pieces, piece)
    }
  }
  return
}

// pieces returns a Bitboard with the pieces of the color pased
func (pos *Position) Pieces(side byte) (pieces Bitboard) {
  startingBitboard := 0
  if side != WHITE {
    startingBitboard = 6
  }
  for currentBitboard := startingBitboard; currentBitboard - startingBitboard < 6 ; currentBitboard++ {
    pieces |= pos.bitboards[currentBitboard]
  }
  return
}

// check returns if the side passed is in check
func (pos *Position) Check(side byte) (inCheck bool) {
  king := WHITE_KING
  if side == BLACK {
    king = BLACK_KING
  }
  kingPos := pos.bitboards[king]

  if (kingPos & pos.AttackedSquares(opponentSide(side))) > 0 {
    inCheck = true
  }
  return
}

func (pos *Position) KingPosition(side byte) (king Bitboard) {
  if side == WHITE {
    king = pos.bitboards[WHITE_KING]
  } else {
    king = pos.bitboards[BLACK_KING]
  }
  return
}

// Remove Piece returns a new position without the piece passed
func (pos Position) RemovePiece(piece Bitboard) Position {
  newPos := pos

  // Iterate over all bitboards to find the piece and remove from it
	for role, bitboard := range newPos.bitboards {
		if bitboard & piece > 0 {
      // Found the piece, remove from bitboard
      newPos.bitboards[role] &= ^piece
		}
	}

  return newPos
}

// Utility functions

// makePiece is a factory function that returns a Piece based on the role and square passed
func makePiece(role int, square Bitboard) (piece Piece) {
  switch role {
  case WHITE_KING:
    piece = &King{color: WHITE, square: square}
  case WHITE_ROOK:
    piece = &Rook{color: WHITE, square: square}
  case WHITE_BISHOP:
    piece = &Bishop{color: WHITE, square: square}
  case WHITE_KNIGHT:
    piece = &Knight{color: WHITE, square: square}
  case BLACK_KING:
    piece = &King{color: BLACK, square: square}
  case BLACK_ROOK:
    piece = &Rook{color: BLACK, square: square}
  case BLACK_KNIGHT:
    piece = &Knight{color: BLACK, square: square}
  }
  return
}

// squareToBitboard returns a bitboard containing the position of the squares coordinates passed
func sqaureToBitboard(coordinates []string) (bitboard Bitboard) {
  for _, coordinate := range coordinates {
    fileNumber := int(coordinate[0]) - 96
    rankNumber := int(coordinate[1]) - 48
    squareNumber :=  (fileNumber - 1) + 8*(rankNumber - 1)

    // displaces 1 bit to the coordinate passed
    bitboard |= 0b1 << squareNumber
  }
	return
}

// InitialPosition is a factory that returns an initial postion board
func InitialPosition() (pos *Position) {
  var initialBiboards = [12]Bitboard{WHITE_KING: 0b10000,
                                     WHITE_QUEEN: 0b1000,
                                     WHITE_ROOK: 0b10000001,
                                     WHITE_BISHOP: 0b100100,
                                     WHITE_KNIGHT: 0b1000010,
                                     WHITE_PAWN: 0b11111111 << 8,
                                     BLACK_KING: 0b1 << 60,
                                     BLACK_QUEEN: 0b1 << 59,
                                     BLACK_ROOK: 0b10000001 << 56,
                                     BLACK_BISHOP: 0b100100 << 56,
                                     BLACK_KNIGHT: 0b1000010 << 56,
                                     BLACK_PAWN: 0b11111111 << 48,
                                   }

  pos = &Position{bitboards: initialBiboards, turn: WHITE}
	return
}

// EmptyPosition returns an empty Position struct
func EmptyPosition() (pos *Position) {
  pos = &Position{turn: WHITE}
	return
}
