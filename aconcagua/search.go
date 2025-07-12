package aconcagua

import (
	"fmt"
	"math"
	"time"
)

// Constants to use in the search
const (
	MateScore = 100000
	MinInt    = math.MinInt32
	MaxInt    = math.MaxInt32
)

// isCheckmateOrStealmate validates if the current position is checkmated or stealmated
func isCheckmateOrStealmate(isCheck bool, moves int, ply int) (int, bool) {
	if moves == 0 {
		if !isCheck {
			return 0, true
		}
		return -MateScore + ply, true
	}
	return 0, false
}

// Search is the main struct for the search
type Search struct {
	nodes              int
	currentDepth       int
	maxDepth           int
	pvLine             pvLine
	killers            [MaxSearchDepth]Killer
	historyMoves       HistoryMoves
	transpositionTable *TranspositionTable
	timeControl        *TimeControl
}

// init initializes the Search struct
func (s *Search) init(depth int) {
	s.nodes = 0
	s.currentDepth = 0
	s.maxDepth = depth
	s.killers = [MaxSearchDepth]Killer{}
	s.historyMoves = HistoryMoves{}
	s.transpositionTable = NewTranspositionTable(DefaultTableSizeInMb)
}

// reset sets the new iteration parameters in the NewSearch
func (s *Search) reset(currentDepth int) {
	s.currentDepth = currentDepth
	s.nodes = 0
	s.pvLine = NewPvLine(MaxSearchDepth)
}

// HistoryMoves is a table for holding the history of moves
type HistoryMoves [2][64][64]int

// update updates the history of moves
func (hm *HistoryMoves) update(depth int, from int, to int, side Color) {
	hm[side][from][to] += depth * depth
}

// Killer is a list of quiet moves that produces a beta cutoff
type Killer [2]Move

// add adds a non capture move to the killer list
func (k *Killer) add(move Move) {
	if move.flag() == quiet {
		k[1] = k[0]
		k[0] = move
	}
}

// root is the entry point of the search
func (s *Search) root(pos *Position, maxDepth int, stdout chan string) (bestMoveScore int, bestMove string) {
	alpha := MinInt
	beta := MaxInt
	s.init(maxDepth)

	for d := 1; d <= maxDepth; d++ {
		s.reset(d)

		lastScore := bestMoveScore
		bestMoveScore = s.negamax(pos, d, 0, alpha, beta, &s.pvLine, true)

		// If stop by time, set last iteration score and best move/pv
		if s.timeControl.stop {
			bestMoveScore = lastScore
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
		stdout <- fmt.Sprintf("info depth %d nodes %d time %v pv %v", d, s.nodes, depthTime.Milliseconds(), s.pvLine.String())
		bestMove = s.pvLine[0].String()
	}

	return
}

// negamax returns the score of the best posible move by the evaluation function
// for a fixed depth
func (s *Search) negamax(pos *Position, depth int, ply int, alpha int, beta int, pvLine *pvLine, nullMoveAllowed bool) int {
	s.nodes++
	if s.timeControl.stop {
		return 0
	}

	// TODO: check repetition by 3fold, 50move rule and insufficient material
	if pos.positionHistory.repetitionCount(pos.Hash) >= 2 {
		return 0
	}

	if depth == 0 {
		return quiescent(pos, s, alpha, beta)
	}

	pvNode := beta-alpha > 1
	ttScore, ttMove, exists := s.transpositionTable.probe(pos.Hash, depth, alpha, beta)
	if exists && !pvNode {
		return ttScore
	}

	flag := FlagAlpha
	isCheck := pos.Check(pos.Turn)
	branchPv := NewPvLine(depth)
	pvLine.reset()

	// Null Move Pruning
	if depth >= 4 && !isCheck && nullMoveAllowed && !pvNode {
		ep := pos.makeNullMove()
		sc := -s.negamax(pos, depth-4, ply+1, -beta, -beta+1, &branchPv, false)
		pos.unmakeNullMove(ep)

		if sc >= beta {
			return beta
		}
	}

	// Futility pruning flag
	futilityPruningAllowed := false
	if depth <= 3 && alpha > -MateScore && beta < MateScore {
		futilityMargin := []int{0, 300, 500, 900}
		sc := Eval(pos)
		futilityPruningAllowed = sc+futilityMargin[depth] <= alpha
	}

	newScore := MinInt
	mg := NewMoveGenerator(pos, &ttMove, &s.killers[ply][0], &s.killers[ply][1], &s.historyMoves)

	for move := mg.nextMove(); move != NoMove; move = mg.nextMove() {
		pos.MakeMove(&move)
		branchPv.reset()
		moveFlag := move.flag()

		// Futility Pruning
		if futilityPruningAllowed && moveFlag == quiet && !isCheck && !pvNode && mg.moveNumber > 0 {
			pos.UnmakeMove(&move)
			continue
		}

		extension := 0
		if isCheck {
			extension = 1
		}

		// Principal Variation Search
		// Full search at first attempt of this subtree
		if mg.moveNumber == 0 {
			newScore = -s.negamax(pos, depth-1+extension, ply+1, -beta, -alpha, &branchPv, true)
		} else {
			// Try first a quick, reduced search, with lmr and a null window(-alpha-1, -alpha)
			reduction := lmrReductionFactor(depth, mg.moveNumber, moveFlag, isCheck, pvNode)
			newScore = -s.negamax(pos, depth-1-reduction+extension, ply+1, -alpha-1, -alpha, &branchPv, true)

			// If an improvement was found, we need to search again with a full window and depth
			if newScore > alpha {
				newScore = -s.negamax(pos, depth-1+extension, ply+1, -beta, -alpha, &branchPv, true)
			}
		}
		pos.UnmakeMove(&move)

		if newScore >= beta {
			s.transpositionTable.store(pos.Hash, depth, FlagBeta, beta, move)
			s.killers[ply].add(move)

			if flag == quiet {
				s.historyMoves.update(depth, move.from(), move.to(), pos.Turn)
			}
			return beta
		}

		if newScore > alpha {
			flag = FlagExact

			alpha = newScore
			pvLine.insert(move, &branchPv)
		}
	}

	score, checkmateOrStealmateFound := isCheckmateOrStealmate(isCheck, mg.moveNumber, ply)
	if checkmateOrStealmateFound {
		return score
	}

	s.transpositionTable.store(pos.Hash, depth, flag, alpha, NoMove)
	return alpha
}

// lrmReductionFactor returns a number to reduce the depth on search based on the conditions passed
func lmrReductionFactor(depth int, moveNumber int, moveFlag int, isCheck, pvNode bool) int {
	if isCheck || pvNode || depth < 3 || moveNumber < 4 || moveFlag >= capture {
		return 0
	}

	return int(0.5 + math.Log(float64(depth))*math.Log(float64(moveNumber))/2.0)
}
