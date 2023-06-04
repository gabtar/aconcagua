package search

import (
	"math"

	"github.com/gabtar/aconcagua/board"
	"github.com/gabtar/aconcagua/evaluation"
)

// minmax returns the bestMove score and the move sequence
func minmax(pos board.Position, depth int, alpha int, beta int, moveTrace []board.Move) (score int, moves []board.Move) {
	if depth == 0 || pos.Checkmate(board.WHITE) || pos.Checkmate(board.BLACK) {
		score = evaluation.Evaluate(pos)
		moves = moveTrace
		return
	}

	if pos.ToMove() == board.WHITE {
		score = math.MinInt

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
		score = math.MaxInt

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
func BestMove(pos *board.Position, depth int) (bestMoveScore int, bestMoves []board.Move) {
	bestMoveScore, bestMoves = minmax(*pos, depth, math.MinInt, math.MaxInt, bestMoves)

	return
}
