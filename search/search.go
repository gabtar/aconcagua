package search

import (
	"math"
	"sort"

	"github.com/gabtar/aconcagua/board"
	"github.com/gabtar/aconcagua/evaluation"
)

// negamax returns the score of the best posible move by the evaluation function
// for a fixed depth
func negamax(pos board.Position, depth int, alpha int, beta int, pV *PrincipalVariation) (score int) {
	alphaOrig := alpha

	if ttEntry, exists := tt.table[pos.Zobrist]; exists && tt.table[pos.Zobrist].depth >= depth {
		if ttEntry.flag == EXACT {
			return ttEntry.score
		} else if ttEntry.flag == LOWERBOUND {
			alpha = max(alpha, ttEntry.score)
		} else if ttEntry.flag == UPPERBOUND {
			beta = min(beta, ttEntry.score)
		}

		if alpha >= beta {
			return ttEntry.score
		}
	}

	if depth == 0 || pos.Checkmate(board.White) || pos.Checkmate(board.Black) {
		// TODO: extract to function? Color modifier for evaluation

		color := 1
		if pos.ToMove() == board.Black {
			color = -1
		}

		score = color * evaluation.Evaluate(pos)
		return
	}

	score = math.MinInt

	for _, move := range sortMoves(pos.LegalMoves(pos.ToMove()), pV, pV.maxDepth-depth) {
		pos.MakeMove(&move)
		newScore := -negamax(pos, depth-1, -beta, -alpha, pV)
		pos.UnmakeMove(move)

		if newScore > score {
			score = newScore
			pV.moves[pV.maxDepth-depth] = move
			alpha = max(newScore, alpha)
		}

		if alpha >= beta {
			return
		}
	}

	if score <= alphaOrig {
		tt.save(pos.Zobrist, depth, score, UPPERBOUND)
	} else if score >= beta {
		tt.save(pos.Zobrist, depth, score, LOWERBOUND)
	} else {
		tt.save(pos.Zobrist, depth, score, EXACT)
	}

	return
}

// max returns the maximum value of the 2 integers passed
func max(a int, b int) int {
	if a >= b {
		return a
	}
	return b
}

// min returns the minimum value of the 2 integers passed
func min(a int, b int) int {
	if b <= a {
		return b
	}
	return a
}

// BestMove returns the best move sequence in the position (for the current side)
// with its score evaluation.
func BestMove(pos *board.Position, maxDepth int) (bestMoveScore int, bestMove []board.Move) {

	min := math.MinInt + 1 // NOTE: +1 Needed to avoid overflow when invert in negamax
	max := math.MaxInt
	tt = newTranspositionTable()
	pV := newPrincipalVariation(maxDepth)

	// TODO: In each depth i should 'emit' on channel the best variation found...
	for d := 1; d <= maxDepth; d++ {
		pV.maxDepth = d
		bestMoveScore = negamax(*pos, d, min, max, pV)
	}

	bestMove = pV.moves

	return
}

// sortMoves sorts an slice of moves acording to the principalVariation
func sortMoves(legalMoves []board.Move, pV *PrincipalVariation, depth int) (sortedMoves []board.Move) {
	sort.Slice(legalMoves, func(i, j int) bool {

		if pV.moves[depth] == legalMoves[i] {
			return true
		}

		// rest of the moves sort acording to type (captures, promotion...)
		// TODO: improve move ordering...
		return legalMoves[i].MoveType() > legalMoves[j].MoveType()
	})
	return legalMoves
}
