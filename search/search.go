package search

import (
	"math"

	"github.com/gabtar/aconcagua/board"
	"github.com/gabtar/aconcagua/evaluation"
)

// negamaxAlphaBeta returns the bestMove score and the move sequence
func minmax(pos board.Position, depth int, alpha float64, beta float64, moveTrace []board.Move) (score float64, moves []board.Move) {
	if depth == 0 || pos.Checkmate(board.WHITE) || pos.Checkmate(board.BLACK) {
		score = evaluation.Evaluate(pos)
		moves = moveTrace
		return
	}

	if pos.ToMove() == board.WHITE {
		score = math.Inf(-1)

		for _, move := range pos.LegalMoves(pos.ToMove()) {
			newPos := pos.MakeMove(&move)
			newMoves := append(moveTrace, move)
			newScore, newMoveTrace := minmax(newPos, depth-1, alpha, beta, newMoves)
			if newScore > score {
				score = newScore
				moves = newMoveTrace
			}
      if score > alpha {
        alpha = score
      }
      if beta <= alpha {
        break
      }
		}
	} else {
		score = math.Inf(1)

		for _, move := range pos.LegalMoves(pos.ToMove()) {
			newPos := pos.MakeMove(&move)
			newMoves := append(moveTrace, move)
			newScore, newMoveTrace := minmax(newPos, depth-1, alpha, beta, newMoves)
			if newScore < score {
				score = newScore
				moves = newMoveTrace
			}
      if score < beta {
        beta = score
      }
      if beta <= alpha {
        break
      }
		}
	}
	return
}

// BestMove returns the best move sequence in the position (for the current side)
// with its score evaluation.
func BestMove(pos *board.Position, depth int) (bestMoveScore float64, bestMoves []board.Move) {
	bestMoveScore, bestMoves = minmax(*pos, depth, math.Inf(-1), math.Inf(1), bestMoves)

	return
}
