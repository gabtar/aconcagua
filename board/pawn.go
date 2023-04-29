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
	//  Bitwise displacements for all possible Pawns attacks
	//   ------------------
	//   | <<9 |     | <<7 |  White pawns
	//   ------------------
	//   |     |  P  |     |
	//   ------------------
	//   | >>7 |     | >>9 |  Black pawns
	//   ------------------
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
  kingBB := pos.KingPosition(p.color)

  if p.color == WHITE {
    posiblesMoves = p.square << 8 | (p.square & ranks[1]) << 16
  } else {
    posiblesMoves = p.square >> 8 | (p.square & ranks[6]) >> 16
  }
  moves = posibleCaptures | posiblesMoves

  // Restrict moves!
  // If the piece is pinned then can only move along the pinned direction
	if isPinned(p.square, p.color, pos) {
		direction := getDirection(kingBB, p.square)
		allowedMovesDirection := raysDirection(kingBB, direction)
		moves &= allowedMovesDirection
	}

  // If the king is in check can only capture the checking piece or block the check
	if pos.Check(p.color) {
		checkingPieces := pos.CheckingPieces(p.color)

		if len(checkingPieces) == 1 {
			checker := checkingPieces[0]
      checkerKingPath := Bitboard(0)

			if checker.IsSliding() {
        checkerKingPath = getRayPath(checker.Square(), kingBB)
			}
      // Check if can capture the checker or block the path
			moves &= (checker.Square() | checkerKingPath) & posiblesMoves
		} else {
			// Double check -> cannot avoid check by capture/blocking
			moves = Bitboard(0)
		}
	}
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
