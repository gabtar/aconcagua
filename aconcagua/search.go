package aconcagua

import (
	"fmt"
	"math"
	"time"
)

// Constants to use in the search
const (
	MateScore = 100000
	MinInt    = math.MinInt + 2 // NOTE: use +2 to avoid overflow when inverting due to negamax
	MaxInt    = math.MaxInt - 2
)

// isCheckmateOrStealmate validates if the current position is checkmated or stealmated
func isCheckmateOrStealmate(pos *Position, ml *moveList, ply int) (int, bool) {
	found := ml.length == 0
	if found {
		if !pos.Check(pos.Turn) {
			return 0, found
		}
		return -MateScore + ply, found
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

	for i := range ml.length {
		if pvMove, exists := s.pv.moveAt(ply); exists && pvMove.String() == ml.moves[i].String() {
			scores[i] = 200
			continue
		}

		if ml.moves[i].flag() == capture {
			victim := pos.PieceAt(squareReference[ml.moves[i].from()])
			aggresor := pos.PieceAt(squareReference[ml.moves[i].to()])

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

// Search is the main struct for the search
type Search struct {
	nodes              int
	currentDepth       int
	maxDepth           int
	pv                 *PV
	killers            [100]Killer
	transpositionTable *TranspositionTable
	timeControl        *TimeControl
}

// init initializes the Search struct
func (s *Search) init(depth int) {
	s.nodes = 0
	s.currentDepth = 0
	s.maxDepth = depth
	s.pv = newPV()
	s.killers = [100]Killer{}
	s.transpositionTable = NewTranspositionTable(DefaultTableSizeInMb)
	s.pv = newPV()
}

// reset sets the new iteration parameters in the NewSearch
func (s *Search) reset(currentDepth int) {
	s.currentDepth = currentDepth
	s.nodes = 0
}

// Killer is a list of quiet moves that produces a beta cutoff
type Killer [2]Move

// add adds a non capture move to the killer list
func (k *Killer) add(move Move) {
	if move.flag() != capture {
		k[1] = k[0]
		k[0] = move
	}
}

// root is the entry point of the search
func (s *Search) root(pos *Position, maxDepth int, stdout chan string) (bestMoveScore int) {
	alpha := MinInt
	beta := MaxInt
	s.init(maxDepth)

	for d := 1; d <= maxDepth; d++ {
		s.reset(d)

		lastPv := *(s.pv)
		lastScore := bestMoveScore

		bestMoveScore = s.negamax(pos, d, alpha, beta, s.pv, true)

		// If stop by time, set last iteration score and best move/pv
		if s.timeControl.stop {
			bestMoveScore = lastScore
			s.pv = &lastPv
			break
		}

		if bestMoveScore <= alpha || bestMoveScore >= beta {
			alpha = MinInt
			beta = MaxInt
			d--
			continue
		}

		alpha = bestMoveScore - 45
		beta = bestMoveScore + 45

		depthTime := time.Since(s.timeControl.iterationStartTime)
		s.timeControl.iterationStartTime = time.Now()
		stdout <- fmt.Sprintf("info depth %d nodes %d time %v pv %v", d, s.nodes, depthTime.Milliseconds(), s.pv)
	}

	return
}

// negamax returns the score of the best posible move by the evaluation function
// for a fixed depth
func (s *Search) negamax(pos *Position, depth int, alpha int, beta int, pv *PV, nullMoveAllowed bool) int {
	if s.timeControl.stop {
		return 0
	}

	ttScore, exists := s.transpositionTable.probe(pos.Hash, depth, alpha, beta)
	if exists {
		return ttScore
	}

	flag := FlagAlpha
	foundPv := false
	check := pos.Check(pos.Turn)
	ply := s.currentDepth - depth
	s.nodes++
	branchPv := newPV()
	futilityPruningAllowed := false

	moves := pos.LegalMoves()
	moves.sort(scoreMoves(pos, moves, s, ply))

	score, checkOrStealmateFound := isCheckmateOrStealmate(pos, moves, ply)
	if checkOrStealmateFound {
		return score
	}

	if depth == 0 {
		return quiescent(pos, s, alpha, beta)
	}

	// Null Move Pruning
	if depth >= 4 && !check && nullMoveAllowed {
		ep := pos.makeNullMove()
		sc := -s.negamax(pos, depth-4, -beta, -beta+1, branchPv, false)
		pos.unmakeNullMove(ep)

		if sc >= beta {
			return beta
		}
	}

	// Futility pruning flag
	if depth <= 3 && alpha > -MateScore && beta < MateScore {
		futilityMargin := []int{0, 280, 500, 900}
		sc := Eval(pos)
		futilityPruningAllowed = sc+futilityMargin[depth] <= alpha
	}

	for moveNumber := range moves.length {
		pos.MakeMove(&moves.moves[moveNumber])
		newScore := MinInt
		isQuietMove := moves.moves[moveNumber].flag() == quiet && !check

		// Futility Pruning
		if futilityPruningAllowed && isQuietMove && !foundPv && moveNumber > 0 {
			pos.UnmakeMove(&moves.moves[moveNumber])
			continue
		}

		// Late Moves Reduction
		applyLMR := depth >= 3 && !foundPv && isQuietMove && moveNumber >= 4
		if applyLMR {
			reduction := 1
			newScore = -s.negamax(pos, depth-1-reduction, -alpha-1, -alpha, branchPv, false)

			if newScore <= alpha {
				pos.UnmakeMove(&moves.moves[moveNumber])
				continue
			}
		}

		// Principal Variation Search
		if foundPv {
			newScore = -s.negamax(pos, depth-1, -alpha-1, -alpha, branchPv, true)
			if newScore > alpha && newScore < beta {
				newScore = -s.negamax(pos, depth-1, -beta, -alpha, branchPv, true)
			}
		} else {
			newScore = -s.negamax(pos, depth-1, -beta, -alpha, branchPv, true)
		}

		pos.UnmakeMove(&moves.moves[moveNumber])

		if newScore >= beta {
			s.transpositionTable.store(pos.Hash, depth, FlagBeta, beta)
			s.killers[ply].add(moves.moves[moveNumber])
			return beta
		}

		if newScore > alpha {
			flag = FlagExact

			alpha = newScore
			pv.insert(moves.moves[moveNumber], branchPv)
			foundPv = true
		}
	}

	s.transpositionTable.store(pos.Hash, depth, flag, alpha)
	return alpha
}
