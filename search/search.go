package search

import (
	"math"

	"github.com/gabtar/aconcagua/board"
	"github.com/gabtar/aconcagua/evaluation"
)

// minmax searchs the best move in the position using the minmax algorithm
func minmax(pos board.Position, depth int, white bool, moveTrace []board.Move) (score float64, moves []board.Move) {
	if depth == 0 {
		score = evaluation.Evaluate(pos)
    moves = moveTrace
		return
	}

	if white {
		score = math.Inf(-1)

		for _, move := range pos.LegalMoves(board.WHITE) {
			newPos := pos.MakeMove(&move)
      newMoves := append(moveTrace, move)
      newScore, newMoveTrace := minmax(newPos, depth - 1, false, newMoves)
      if newScore > score {
        score = newScore
        moves = newMoveTrace
      }
		}
	} else {
		score = math.Inf(1)

		for _, move := range pos.LegalMoves(board.BLACK) {
			newPos := pos.MakeMove(&move)
      newMoves := append(moveTrace, move)
      newScore, newMoveTrace := minmax(newPos, depth - 1, true, newMoves)
      if newScore < score {
        score = newScore
        moves = newMoveTrace
      }
		}
	}
	return
}

// BestMove returns the best move sequence in the position (for the current side)
// with its score evaluation.
func BestMove(pos *board.Position, depth int) (bestMoveScore float64, bestMoves []board.Move) {

	if pos.ToMove() == board.WHITE {
		bestMoveScore, bestMoves = minmax(*pos, depth, true, bestMoves)
	} else {
		bestMoveScore, bestMoves = minmax(*pos, depth, false, bestMoves)
	}
	return
}
