package aconcagua

import (
	"fmt"
	"math"
	"time"
)

// MateScore defines a base score for detecting checkmates
const MateScore = 100000

// isCheckmateOrStealmate validates if the current position is checkmated or stealmated
func isCheckmateOrStealmate(pos *Position, ml *moveList, depth *int) (int, bool) {
	found := ml.length == 0
	if found {
		if !pos.Check(pos.Turn) {
			return 0, found
		}
		return -MateScore - *depth, found
	}

	return 0, false
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

// scoreMoves assigns a score to each move
func scoreMoves(pos *Position, ml *moveList, s *Search, ply int) []int {
	scores := make([]int, ml.length)

	for i := 0; i < ml.length; i++ {
		if pvMove, exists := s.pv.moveAt(ply); exists && pvMove.String() == ml.moves[i].String() {
			scores[i] = 200
			continue
		}

		if ml.moves[i].flag() == capture {
			victim, _ := pos.PieceAt(squareReference[ml.moves[i].from()])
			aggresor, _ := pos.PieceAt(squareReference[ml.moves[i].to()])

			if victim > 5 {
				victim -= 6
			}
			if aggresor > 5 {
				aggresor -= 6
			}
			scores[i] = mvvLvaScore[victim][aggresor]
		} else {
			if s.killers[ply][0] == ml.moves[i] {
				scores[i] = 50
				continue
			}
			if s.killers[ply][1] == ml.moves[i] {
				scores[i] = 40
				continue
			}
		}
	}

	return scores
}

type Search struct {
	nodes              int
	currentDepth       int
	maxDepth           int
	pv                 *PV
	killers            [100]Killer
	transpositionTable *TranspositionTable
	time               time.Time
	totalTime          time.Time
	stop               bool
}

// init initializes the Search struct
func (s *Search) init(depth int) {
	s.nodes = 0
	s.currentDepth = 0
	s.maxDepth = depth
	s.pv = newPV()
	s.killers = [100]Killer{}
	s.transpositionTable = NewTranspositionTable(64)
	s.time = time.Now()
	s.totalTime = time.Now()
	s.stop = false
}

// reset sets the new iteration parameters in the NewSearch
func (s *Search) reset(currentDepth int) {
	s.currentDepth = currentDepth
	s.nodes = 0
}

// new killer
type Killer [2]Move

func (k *Killer) add(move Move) {
	k[1] = k[0]
	k[0] = move
}

// root is the entry point of the search
func root(pos *Position, s *Search, maxDepth int, stdout chan string) (bestMoveScore int) {
	// NOTE: not sure if returning the bestMove is necessary. Its contained in the NewSearch struct

	alpha := math.MinInt + 1 // NOTE: +1 Needed to avoid overflow when inverting alpha and beta in negamax
	beta := math.MaxInt
	s.init(maxDepth)

	for d := 1; d <= maxDepth; d++ {
		// stop search if stop command has been sended
		if s.stop {
			break
		}
		s.reset(d)

		bestMoveScore = negamax(pos, s, d, alpha, beta, s.pv)

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

		depthTime := time.Since(s.time)
		s.time = time.Now()

		stdout <- fmt.Sprintf("info depth %d nodes %d time %v pv %v", d, s.nodes, depthTime.Milliseconds(), s.pv)
	}

	return
}

// negamax returns the score of the best posible move by the evaluation function
// for a fixed depth
func negamax(pos *Position, s *Search, depth int, alpha int, beta int, pv *PV) int {
	if s.stop {
		return alpha
	}

	// TODO: Probe hash with alpha/beta
	// move ordering w/ tt score....
	value, exists := s.transpositionTable.probe(pos.Hash, depth, alpha, beta)
	if exists {
		return value
	}

	flag := FlagUpperBound
	foundPv := false
	ply := s.currentDepth - depth
	s.nodes++
	branchPv := newPV()

	moves := pos.LegalMoves()
	moves.sort(scoreMoves(pos, moves, s, ply))

	score, checkOrStealmateFound := isCheckmateOrStealmate(pos, moves, &depth)
	if checkOrStealmateFound {
		return score
	}

	if depth == 0 {
		s.transpositionTable.store(pos.Hash, depth, FlagExact, Eval(pos))
		return quiescent(pos, s, alpha, beta)
	}

	for i := 0; i < moves.length; i++ {
		pos.MakeMove(&moves.moves[i])
		newScore := math.MinInt + 1

		if foundPv {
			newScore = -negamax(pos, s, depth-1, -beta, -alpha, branchPv)
			if newScore > alpha && newScore < beta {
				newScore = -negamax(pos, s, depth-1, -beta, -newScore, branchPv)
			}
		} else {
			newScore = -negamax(pos, s, depth-1, -beta, -alpha, branchPv)
		}

		pos.UnmakeMove(&moves.moves[i])

		if newScore >= beta {
			s.transpositionTable.store(pos.Hash, depth, FlagUpperBound, newScore)
			s.killers[ply].add(moves.moves[i])
			return beta
		}

		if newScore > alpha {
			flag = FlagExact

			alpha = newScore
			pv.insert(moves.moves[i], branchPv)
			foundPv = true
		}
	}

	s.transpositionTable.store(pos.Hash, depth, flag, alpha)
	return alpha
}

// quiescent s an evaluation function that takes into account some dynamic possibilities
func quiescent(pos *Position, s *Search, alpha int, beta int) int {
	if s.stop {
		return alpha
	}
	score := Eval(pos)

	if score >= beta {
		return beta
	}
	if score > alpha {
		alpha = score
	}

	ml := pos.LegalMoves()
	ml.capturesOnly()

	for i := 0; i < ml.length; i++ {
		pos.MakeMove(&ml.moves[i])
		score = -quiescent(pos, s, -beta, -alpha)
		pos.UnmakeMove(&ml.moves[i])
		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha
}
