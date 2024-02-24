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
	pv:           newPrincipalVariation(),
	time:         time.Now(),
	totalTime:    time.Now(),
}

// init clears previous search state
func (s *SearchState) init(depth int) {
	s.nodes = 0
	s.currentDepth = 0
	s.maxDepth = depth
	s.pv = newPrincipalVariation()
	s.time = time.Now()
	s.totalTime = time.Now()
}

// reset sets the new iteration parameters in the SearchState
func (s *SearchState) reset(currentDepth int) {
	ss.currentDepth = currentDepth
	ss.nodes = 0
}

// negamax returns the score of the best posible move by the evaluation function
// for a fixed depth
func negamax(pos board.Position, depth int, alpha int, beta int, pv *PrincipalVariation) int {
	alphaOrig := alpha
	ss.nodes++
	branchPv := newPrincipalVariation()

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
		pv.clear() // NOTE: needed to clear when mate is found!!
		// return sideModifier(pos.Turn) * evaluation.Evaluate(&pos)
		return quiescent(&pos, alpha, beta)
	}

	moves := pos.LegalMoves(pos.Turn)
	sortMoves(moves, ss.pv, ss.currentDepth-depth)

	for _, move := range moves {
		pos.MakeMove(&move)
		newScore := -negamax(pos, depth-1, -beta, -alpha, branchPv)
		pos.UnmakeMove(move)

		if newScore >= beta {
			return beta
		}

		if newScore > alpha {
			alpha = newScore
			pv.insert(move, branchPv)
		}
	}

	if alpha <= alphaOrig {
		tt.save(pos.Hash, depth, alpha, UPPERBOUND)
	} else if alpha >= beta {
		tt.save(pos.Hash, depth, alpha, LOWERBOUND)
	} else {
		tt.save(pos.Hash, depth, alpha, EXACT)
	}

	return alpha
}

// quiescent performs a quiescent search (evaluates the position, while being careful to avoid overlooking extremely obvious tactical conditions)
func quiescent(pos *board.Position, alpha int, beta int) int {
	score := evaluation.Eval(pos)
	if score >= beta {
		return beta
	}
	if score > alpha {
		alpha = score
	}

	// TODO: need a function to generate captures only...
	moves := pos.LegalMoves(pos.Turn)
	sortMoves(moves, ss.pv, 0)

	for _, move := range moves {
		// skip non capture
		if move.MoveType() != board.Capture {
			continue
		}

		pos.MakeMove(&move)
		score = -quiescent(pos, -beta, -alpha)
		pos.UnmakeMove(move)

		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha
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

		bestMoveScore = negamax(*pos, d, min, max, ss.pv)
		bestMove = *ss.pv

		depthTime := time.Since(ss.time)
		ss.time = time.Now()

		stdout <- fmt.Sprintf("info depth %d nodes %d time %v pv %v", d, ss.nodes, depthTime.Milliseconds(), ss.pv)
	}

	return
}

// sortMoves sorts an slice of moves acording to the principalVariation
func sortMoves(moves []board.Move, pv *PrincipalVariation, ply int) []board.Move {
	sort.Slice(moves, func(i, j int) bool {
		// PV pvMove from previous iteration always goes first
		if pvMove, exists := pv.moveAt(ply); exists && pvMove.ToUci() == moves[i].ToUci() {
			return true
		}

		return moves[i].Score() > moves[j].Score()
	})

	return moves
}
