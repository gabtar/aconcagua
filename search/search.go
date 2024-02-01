package search

import (
	"fmt"
	"math"
	"sort"

	"github.com/gabtar/aconcagua/board"
	"github.com/gabtar/aconcagua/evaluation"
)

var nodes int = 0

// negamax returns the score of the best posible move by the evaluation function
// for a fixed depth
func negamax(pos board.Position, depth int, alpha int, beta int, pV *PrincipalVariation) (score int) {
	alphaOrig := alpha
	nodes++

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
		return sideModifier(pos.ToMove()) * evaluation.Evaluate(&pos)
	}

	score = math.MinInt

	moves := pos.LegalMoves(pos.ToMove())
	sortMoves(moves, pV, pV.maxDepth-depth)

	for _, move := range moves {
		pos.MakeMove(&move)
		newScore := -negamax(pos, depth-1, -beta, -alpha, pV)
		pos.UnmakeMove(move)

		if newScore > score {
			score = newScore
			pV.add(move, depth)
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

// sideModifier returns a multiplier factor for the evaluation score based on the current color turn(side to move) passed
func sideModifier(color board.Color) int {
	if color == board.Black {
		return -1
	}
	return 1
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
func BestMove(pos *board.Position, maxDepth int, stdout chan string) (bestMoveScore int, bestMove []board.Move) {

	min := math.MinInt + 1 // NOTE: +1 Needed to avoid overflow when invert in negamax
	max := math.MaxInt
	pV := newPrincipalVariation(maxDepth)
	tt = newTranspositionTable()

	for d := 1; d <= maxDepth; d++ {
		nodes = 0
		pV.maxDepth = d
		bestMoveScore = negamax(*pos, d, min, max, pV)
		stdout <- fmt.Sprintf("info depth %d nodes %d", d, nodes)
	}

	bestMove = pV.moves

	return
}

// sortMoves sorts an slice of moves acording to the principalVariation
func sortMoves(legalMoves []board.Move, pV *PrincipalVariation, depth int) []board.Move {
	sort.Slice(legalMoves, func(i, j int) bool {
		// PV move from previous iteration always goes first
		if pV.moves[depth] == legalMoves[i] {
			return true
		}

		return legalMoves[i].Score() > legalMoves[j].Score()
	})

	return legalMoves
}
