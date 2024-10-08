package aconcagua

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// MateScore defines a base score for detecting checkmates
const MateScore = 100000

// HistoryMoves stores moves score during search by piece moved/target square
type HistoryMoves [12][64]int

// KillerMoves holds the killers moves(quiet/non capture) found during search that produces a beta cutoff
type KillerMoves struct {
	moves [2]Move
}

// adds a killer move to the struct
func (km *KillerMoves) add(move Move) {
	km.moves[1] = km.moves[0]
	km.moves[0] = move
}

// SearchState contains all the stats for the current search
type SearchState struct {
	nodes        int
	currentDepth int
	maxDepth     int
	pv           *PrincipalVariation
	killers      [100]KillerMoves // up to 100 depth for now
	history      HistoryMoves
	time         time.Time
	totalTime    time.Time
	stop         bool // flag to stop current search
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
	s.currentDepth = currentDepth
	s.nodes = 0
}

// negamax returns the score of the best posible move by the evaluation function
// for a fixed depth
func negamax(pos Position, ss *SearchState, depth int, alpha int, beta int, pv *PrincipalVariation) int {
	// TODO:!!!!! improve
	if ss.stop {
		return alpha
	}

	alphaOrig := alpha
	foundPv := false
	ply := ss.currentDepth - depth
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

	moves := pos.LegalMoves(pos.Turn)
	sortMoves(moves, ss, ply)

	// checkmate/stealmate detection
	// TODO: extract to stealmate/checkmate found function
	if len(moves) == 0 {
		// Stealmate
		if !pos.Check(pos.Turn) {
			return 0 // Stealmate is draw...
		}
		return -MateScore - depth
	}

	if depth == 0 {
		return quiescent(&pos, ss, alpha, beta)
	}

	for _, move := range moves {
		pos.MakeMove(&move)
		newScore := math.MinInt + 1

		// PVS Principal variation search
		if foundPv {
			newScore = -negamax(pos, ss, depth-1, -alpha-1, -alpha, branchPv)
			if newScore > alpha && newScore < beta {
				newScore = -negamax(pos, ss, depth-1, -beta, -alpha, branchPv)
			}
		} else {
			newScore = -negamax(pos, ss, depth-1, -beta, -alpha, branchPv)
		}

		pos.UnmakeMove(move)

		// beta cutoff
		if newScore >= beta {
			ss.killers[ply].add(move)
			return beta
		}

		// new best move found
		if newScore > alpha {
			ss.history[move.Piece()][move.To()] += depth

			alpha = newScore
			pv.insert(move, branchPv)
			foundPv = true
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
func quiescent(pos *Position, ss *SearchState, alpha int, beta int) int {
	// if pos.Check(pos.Turn) {
	// 	return negamax(*pos, ss, 1, alpha, beta, ss.pv)
	// }

	score := Eval(pos)
	if score >= beta {
		return beta
	}
	if score > alpha {
		alpha = score
	}

	// TODO: need a function to generate captures only...
	moves := pos.LegalMoves(pos.Turn)
	sortMoves(moves, ss, 0)

	for _, move := range moves {
		if move.MoveType() != Capture {
			continue
		}

		pos.MakeMove(&move)
		score = -quiescent(pos, ss, -beta, -alpha)
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
func sideModifier(color Color) int {
	if color == Black {
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
func Search(pos *Position, ss *SearchState, maxDepth int, stdout chan string) (bestMoveScore int, bestMove []Move) {

	alpha := math.MinInt + 1 // NOTE: +1 Needed to avoid overflow when inverting alpha and beta in negamax
	beta := math.MaxInt
	tt = newTranspositionTable()
	ss.init(maxDepth)

	for d := 1; d <= maxDepth; d++ {
		// stop search if stop command has been sended
		if ss.stop {
			break
		}

		ss.reset(d)

		bestMoveScore = negamax(*pos, ss, d, alpha, beta, ss.pv)
		bestMove = *ss.pv

		// Set aspiration window
		if bestMoveScore <= alpha || bestMoveScore >= beta {
			alpha = math.MinInt + 1
			beta = math.MaxInt
			d--
			continue
		}
		// TODO: try diferent window values...
		alpha = bestMoveScore - 50
		beta = bestMoveScore + 50

		depthTime := time.Since(ss.time)
		ss.time = time.Now()

		stdout <- fmt.Sprintf("info depth %d nodes %d time %v pv %v", d, ss.nodes, depthTime.Milliseconds(), ss.pv)
	}

	return
}

// sortMoves sorts an slice of moves acording to the principalVariation
func sortMoves(moves []Move, ss *SearchState, ply int) []Move {
	// socre moves
	// MVV_VLA score
	for idx := range moves {
		setMoveScore(&moves[idx], ss, ply)
	}

	// sort moves
	sort.Slice(moves, func(i, j int) bool {
		// PV pvMove from previous iteration always goes first

		return moves[i].Score() > moves[j].Score()
	})

	return moves
}

// mvvLvaScore (Most Valuable Victim - Least Valuable Aggressor) for scoring captures
var mvvLvaScore = [6][6]int{
	{0, 0, 0, 0, 0, 0},             // Victim K, agressor K, Q, R, B, N, P - No point to capture a king!
	{100, 101, 102, 103, 104, 105}, // Victim Q, agressor K, Q, R, B, N, P
	{90, 91, 92, 93, 94, 95},       // Victim R, agressor K, Q, R, B, N, P
	{80, 81, 82, 83, 84, 85},       // Victim B, agressor K, Q, R, B, N, P
	{70, 71, 72, 73, 74, 75},       // Victim N, agressor K, Q, R, B, N, P
	{60, 61, 62, 63, 64, 65},       // Victim P, agressor K, Q, R, B, N, P
}

// setMoveScore sets the score to a move according to MVV-LVA(Most Valuable Victim - Least Valuable Aggressor)
func setMoveScore(m *Move, ss *SearchState, ply int) {
	if pvMove, exists := ss.pv.moveAt(ply); exists && pvMove.ToUci() == m.ToUci() {
		m.SetScore(200)
		return
	}

	// MvvLva score
	if m.MoveType() == Capture {
		victim := m.CapturedPiece()
		aggresor := m.Piece()

		if victim > 5 {
			victim -= 6
		}
		if aggresor > 5 {
			aggresor -= 6
		}
		m.SetScore(mvvLvaScore[victim][aggresor])
	} else {

		// 1st Killer
		if ss.killers[ply].moves[0] == *m {
			m.SetScore(50)
			return
		}
		// 2nd Killer
		if ss.killers[ply].moves[1] == *m {
			m.SetScore(40)
			return
		}

		// score from history moves
		m.SetScore(ss.history[m.Piece()][m.To()])
	}
}
