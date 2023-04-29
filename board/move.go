package board

// Type of move
const(
  NORMAL = 1    // Normal push to an empty sqaure
  CAPTURE = 2
  EN_PASSANT = 3
  CASTLE = 4
  PROMOTION = 5
)

// Represents a move of chess
type Move struct {
  from string
  to string
  piece int
  promotedTo int
  moveType int
}
