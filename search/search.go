package search

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/gabtar/aconcagua/board"
	"github.com/gabtar/aconcagua/evaluation"
)

type SearchState struct {
	nodes        int
	currentDepth int
	maxDepth     int
	pv           *PrincipalVariation
	time         time.Time
	totalTime    time.Time
}

var ss SearchState = SearchState{
	nodes:        0,
	currentDepth: 0,
	maxDepth:     0,
	pv:           newPrincipalVariation(10),
	time:         time.Now(),
	totalTime:    time.Now(),
}

// init clears previous search state
func (s *SearchState) init(depth int) {
	s.nodes = 0
	s.currentDepth = 0
	s.maxDepth = depth
	s.pv = newPrincipalVariation(depth)
	s.time = time.Now()
	s.totalTime = time.Now()
}

// reset sets the new iteration parameters in the SearchState
func (s *SearchState) reset(currentDepth int) {
	ss.currentDepth = currentDepth
	ss.pv.maxDepth = currentDepth
	ss.nodes = 0
}

// negamax returns the score of the best posible move by the evaluation function
// for a fixed depth
func negamax(pos board.Position, depth int, alpha int, beta int) (score int) {
	ss.nodes++
	alphaOrig := alpha

	if ttEntry, exists := tt.table[pos.Hash]; exists && tt.table[pos.Hash].depth >= depth {
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
		return sideModifier(pos.Turn) * evaluation.Evaluate(&pos)
	}

	score = math.MinInt

	moves := pos.LegalMoves(pos.Turn)
	sortMoves(moves, ss.pv.moves[ss.currentDepth-depth])

	for _, move := range moves {
		pos.MakeMove(&move)
		newScore := -negamax(pos, depth-1, -beta, -alpha)
		pos.UnmakeMove(move)

		if newScore > score {
			score = newScore
			ss.pv.add(move, depth)
			alpha = max(newScore, alpha)
		}

		if alpha >= beta {
			return
		}
	}

	if score <= alphaOrig {
		tt.save(pos.Hash, depth, score, UPPERBOUND)
	} else if score >= beta {
		tt.save(pos.Hash, depth, score, LOWERBOUND)
	} else {
		tt.save(pos.Hash, depth, score, EXACT)
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

// Search returns the best move sequence in the position (for the current side)
// with its score evaluation.
func Search(pos *board.Position, maxDepth int, stdout chan string) (bestMoveScore int, bestMove []board.Move) {

	min := math.MinInt + 1 // NOTE: +1 Needed to avoid overflow when inverting alpha and beta in negamax
	max := math.MaxInt
	tt = newTranspositionTable()
	ss.init(maxDepth)

	for d := 1; d <= maxDepth; d++ {
		ss.reset(d)

		bestMoveScore = negamax(*pos, d, min, max)
		bestMove = ss.pv.moves

		depthTime := time.Since(ss.time)
		ss.time = time.Now()

		stdout <- fmt.Sprintf("info depth %d nodes %d time %v pv %v", d, ss.nodes, depthTime.Milliseconds(), ss.pv)
	}

	return
}

// sortMoves sorts an slice of moves acording to the principalVariation
func sortMoves(moves []board.Move, pvLastIteration board.Move) []board.Move {
	sort.Slice(moves, func(i, j int) bool {
		// PV move from previous iteration always goes first
		if pvLastIteration.ToUci() == moves[i].ToUci() {
			return true
		}

		return moves[i].Score() > moves[j].Score()
	})

	return moves
}
