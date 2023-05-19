package board

import "fmt"

// Type of move
const(
  NORMAL = 1    // Normal push to an empty sqaure
  PAWN_DOUBLE_PUSH = 2
  CAPTURE = 3
  EN_PASSANT = 4
  CASTLE = 5
  PROMOTION = 6
)

// Represents a move of chess
type Move struct {
  from string
  to string
  piece int
  promotedTo int
  moveType int
}

// String returns the move string
func (m Move) String() string {
  return fmt.Sprintf(
    "%s -> %s",
    m.from, 
    m.to)
}
