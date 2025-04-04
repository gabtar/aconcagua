package aconcagua

import (
	"fmt"
	"math"
	"time"
)

// MateScore defines a base score for detecting checkmates
const MateScore = 100000

// HistoryMoves stores moves score during search by piece moved/target square
type HistoryMoves [12][64]int

// validateCheckmateOrStealmate validates if the current position is checkmated or stealmated
func validateCheckmateOrStealmate(pos *Position, moves *[]Move, depth *int) (int, bool) {
	found := len(*moves) == 0
	if found {
		if !pos.Check(pos.Turn) {
			return 0, found
		}
		return -MateScore - *depth, found
	}

	return 0, false
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

// mvvLvaScore (Most Valuable Victim - Least Valuable Aggressor) for scoring captures
var mvvLvaScore = [6][6]int{
	{0, 0, 0, 0, 0, 0},             // Victim K, agressor K, Q, R, B, N, P - No point to capture a king!
	{100, 101, 102, 103, 104, 105}, // Victim Q, agressor K, Q, R, B, N, P
	{90, 91, 92, 93, 94, 95},       // Victim R, agressor K, Q, R, B, N, P
	{80, 81, 82, 83, 84, 85},       // Victim B, agressor K, Q, R, B, N, P
	{70, 71, 72, 73, 74, 75},       // Victim N, agressor K, Q, R, B, N, P
	{60, 61, 62, 63, 64, 65},       // Victim P, agressor K, Q, R, B, N, P
}

// scoreMoves
func scoreMoves(pos *Position, ml *moveList, s *NewSearch, ply int) []int {
	scores := make([]int, ml.length)

	for i := 0; i < ml.length; i++ {
		// TODO: extract to an score function

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
			// 1st Killer
			if s.killers[ply][0] == ml.moves[i] {
				scores[i] = 50
				continue
			}
			// 2nd Killer
			if s.killers[ply][1] == ml.moves[i] {
				scores[i] = 40
				continue
			}

		}

		// TODO: history moves table
	}

	return scores
}

type NewSearch struct {
	nodes        int
	currentDepth int
	maxDepth     int
	pv           *PV
	killers      [100]Killer
	time         time.Time
	totalTime    time.Time
	stop         bool
	// TODO: History moves...
}

// init initializes the NewSearch struct
func (s *NewSearch) init(depth int) {
	s.nodes = 0
	s.currentDepth = 0
	s.maxDepth = depth
	s.pv = newPV()
	s.killers = [100]Killer{}
	s.time = time.Now()
	s.totalTime = time.Now()
	s.stop = false
}

// reset sets the new iteration parameters in the NewSearch
func (s *NewSearch) reset(currentDepth int) {
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
func root(pos *Position, s *NewSearch, maxDepth int, stdout chan string) (bestMoveScore int) {
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

		bestMoveScore = newNegamax(pos, s, d, alpha, beta, s.pv)

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

// newNegamax returns the score of the best posible move by the evaluation function
// for a fixed depth
func newNegamax(pos *Position, s *NewSearch, depth int, alpha int, beta int, pv *PV) int {
	if s.stop {
		return alpha
	}

	foundPv := false
	ply := s.currentDepth - depth
	s.nodes++
	branchPv := newPV()

	// result, checkOrStealmateFound := validateCheckmateOrStealmate(pos, &moves, &depth)
	// if checkOrStealmateFound {
	// 	return result
	// }

	if depth == 0 {
		// TODO: quiescent
		return Eval(pos)
	}

	moves := pos.LegalMoves()
	moves.sort(scoreMoves(pos, moves, s, ply))

	for i := 0; i < moves.length; i++ {
		pos.MakeMove(&moves.moves[i])
		newScore := math.MinInt + 1

		if foundPv {
			newScore = -newNegamax(pos, s, depth-1, -beta, -alpha, branchPv)
			if newScore > alpha && newScore < beta {
				newScore = -newNegamax(pos, s, depth-1, -beta, -newScore, branchPv)
			}
		} else {
			newScore = -newNegamax(pos, s, depth-1, -beta, -alpha, branchPv)
		}

		pos.UnmakeMove(&moves.moves[i])

		if newScore >= beta {
			s.killers[ply].add(moves.moves[i])
			return beta
		}

		if newScore > alpha {
			alpha = newScore
			pv.insert(moves.moves[i], branchPv)
			foundPv = true
		}
	}

	return alpha
}
