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
func isCheckmateOrStealmate(isCheck bool, ml *moveList, ply int) (int, bool) {
	if ml.length == 0 {
		if !isCheck {
			return 0, true
		}
		return -MateScore + ply, true
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
func scoreMoves(pos *Position, ml *moveList, s *Search, ply int, ttMove Move) []int {
	scores := make([]int, ml.length)

	for i := range ml.length {
		if ttMove != NoMove && ml.moves[i] == ttMove {
			scores[i] = 250
			continue
		}

		flag := ml.moves[i].flag()
		if flag == capture || flag >= knightCapturePromotion {
			aggresor := pieceRole(pos.PieceAt(squareReference[ml.moves[i].from()]))
			victim := pieceRole(pos.PieceAt(squareReference[ml.moves[i].to()]))
			scores[i] = mvvLvaScore[victim][aggresor]
		} else if flag == epCapture {
			scores[i] = mvvLvaScore[Pawn][Pawn]
			continue
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
	PVTable            PVTable
	killers            [100]Killer
	transpositionTable *TranspositionTable
	timeControl        *TimeControl
}

// init initializes the Search struct
func (s *Search) init(depth int) {
	s.nodes = 0
	s.currentDepth = 0
	s.maxDepth = depth
	s.killers = [100]Killer{}
	s.transpositionTable = NewTranspositionTable(DefaultTableSizeInMb)
}

// reset sets the new iteration parameters in the NewSearch
func (s *Search) reset(currentDepth int) {
	s.currentDepth = currentDepth
	s.nodes = 0
	s.PVTable = NewPVTable(currentDepth + 1)
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
func (s *Search) root(pos *Position, maxDepth int, stdout chan string) (bestMoveScore int, bestMove string) {
	alpha := MinInt
	beta := MaxInt
	s.init(maxDepth)

	for d := 1; d <= maxDepth; d++ {
		s.reset(d)

		lastScore := bestMoveScore
		bestMoveScore = s.negamax(pos, d, alpha, beta, true)

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
		stdout <- fmt.Sprintf("info depth %d nodes %d time %v pv %v", d, s.nodes, depthTime.Milliseconds(), s.PVTable[0].String())
		bestMove = s.PVTable[0].moves[0].String()
	}
	stdout <- fmt.Sprintf("info tt stored %d tried %d found %d pruned %d", s.transpositionTable.stored, s.transpositionTable.tried, s.transpositionTable.found, s.transpositionTable.pruned)

	return
}

// negamax returns the score of the best posible move by the evaluation function
// for a fixed depth
func (s *Search) negamax(pos *Position, depth int, alpha int, beta int, nullMoveAllowed bool) int {
	s.nodes++
	if s.timeControl.stop {
		return 0
	}

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
	ply := s.currentDepth - depth

	// Null Move Pruning
	if depth >= 4 && !isCheck && nullMoveAllowed && !pvNode {
		ep := pos.makeNullMove()
		sc := -s.negamax(pos, depth-4, -beta, -beta+1, false)
		pos.unmakeNullMove(ep)

		if sc >= beta {
			return beta
		}
	}

	// Futility pruning flag
	futilityPruningAllowed := false
	if depth <= 3 && alpha > -MateScore && beta < MateScore {
		futilityMargin := []int{0, 280, 500, 900}
		sc := Eval(pos)
		futilityPruningAllowed = sc+futilityMargin[depth] <= alpha
	}

	moves := pos.LegalMoves()
	moves.sort(scoreMoves(pos, moves, s, ply, ttMove))

	newScore := MinInt
	for moveNumber := range moves.length {
		pos.MakeMove(&moves.moves[moveNumber])
		s.PVTable.reset(ply + 1)
		moveFlag := moves.moves[moveNumber].flag()

		// Futility Pruning
		if futilityPruningAllowed && moveFlag == quiet && !isCheck && !pvNode && moveNumber > 0 {
			pos.UnmakeMove(&moves.moves[moveNumber])
			continue
		}

		// Principal Variation Search
		// Full search at first attempt of this subtree
		if moveNumber == 0 {
			newScore = -s.negamax(pos, depth-1, -beta, -alpha, true)
		} else {
			// Try first a quick, reduced search, with lmr and a null window(-alpha-1, -alpha)
			lmrFactor := lmrReductionFactor(depth, moveNumber, moveFlag, isCheck, pvNode)
			newScore = -s.negamax(pos, depth-1-lmrFactor, -alpha-1, -alpha, true)

			// If an improvement was found, we need to search again with a full window and depth
			if newScore > alpha {
				newScore = -s.negamax(pos, depth-1, -beta, -alpha, true)
			}
		}
		pos.UnmakeMove(&moves.moves[moveNumber])

		if newScore >= beta {
			s.transpositionTable.store(pos.Hash, depth, FlagBeta, beta, moves.moves[moveNumber])
			s.killers[ply].add(moves.moves[moveNumber])
			return beta
		}

		if newScore > alpha {
			flag = FlagExact

			alpha = newScore
			s.PVTable[ply].prepend(moves.moves[moveNumber], &s.PVTable[ply+1])
		}
	}

	score, checkmateOrStealmateFound := isCheckmateOrStealmate(isCheck, moves, ply)
	if checkmateOrStealmateFound {
		return score
	}

	s.transpositionTable.store(pos.Hash, depth, flag, alpha, NoMove)
	return alpha
}

// lrmReductionFactor returns a number to reduce the depth on search based on the conditions passed
func lmrReductionFactor(depth int, moveNumber int, moveFlag int, isCheck, foundPv bool) int {
	if isCheck || foundPv || depth < 3 || moveNumber < 4 || moveFlag == capture || moveFlag >= knightPromotion {
		return 0
	}
	return int(0.5 + math.Log(float64(depth))*math.Log(float64(moveNumber))/2.0)
}
