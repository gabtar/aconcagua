package engine

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// Constants to use in the search
const (
	MaxSearchDepth                = 50
	MateScore                     = 20000
	MinInt                        = math.MinInt32
	MaxInt                        = math.MaxInt32
	EndgameMaterialThreshold      = 1600
	ReverseFutitlityPruningMargin = 130
	AspirationWindowSize          = 25
)

var (
	// LateMovePruningMoveNumber contains the move number to start pruning for each depth
	LateMovePruningMoveNumber = [5]int{0, 5, 10, 15, 20}

	// FutilityPruningMargin contains the margin for futility pruning for each depth
	FutilityPruningMargin = [5]int{0, 120, 180, 280, 420}

	// LateMoveReductionFactor contains the reduction factor for each depth and move number
	LateMoveReductionFactor = [MaxSearchDepth * 2][MaxLegalMoves * 2]int{}
)

func init() {
	for depth := range MaxSearchDepth * 2 {
		for moveNumber := range MaxLegalMoves * 2 {
			LateMoveReductionFactor[depth][moveNumber] = int(0.5 + math.Log(float64(depth))*math.Log(float64(moveNumber))/2.0)
		}
	}
}

// Search is the main struct for the search
type Search struct {
	nodes              int
	pvLine             pvLine
	seldepth           uint8
	killers            KillersTable
	historyMoves       HistoryMovesTable
	TranspositionTable TranspositionTable
	counterMovesTable  CounterMoveTable
	stack              Stack
	TimeControl        TimeControl
	Evaluation         Evaluation
}

// NewSearch returns a pointer to a new Search struct
func NewSearch() *Search {
	return &Search{
		nodes:              0,
		pvLine:             NewPvLine(MaxSearchDepth),
		killers:            KillersTable{},
		historyMoves:       HistoryMovesTable{},
		TranspositionTable: *NewTranspositionTable(DefaultTableSizeInMb),
		counterMovesTable:  CounterMoveTable{},
		stack:              Stack{},
		TimeControl:        TimeControl{},
		Evaluation:         *NewEvaluation(DefaultPawnHashTableSizeInMb),
	}
}

// clear clears the search
func (s *Search) clear() {
	s.nodes = 0
	s.killers.clear()
	s.historyMoves.clear()
	s.TranspositionTable.newSearch()
	s.counterMovesTable.clear()
	s.stack.clear()
	s.Evaluation.PawnCache.newSearch()
}

// reset sets the new iteration parameters in the NewSearch
func (s *Search) reset() {
	s.seldepth = 0
	s.nodes = 0
	s.pvLine.reset()
	s.stack.clear()
}

// Stop stops the search
func (s *Search) Stop() {
	s.TimeControl.stop = true
}

// HistoryMovesTable is a table for holding the history of moves
type HistoryMovesTable [2][64][64]int

// increment increase the history score of the move passed at depth
func (hm *HistoryMovesTable) increment(depth int, move *Move, side Color) {
	if move.flag() < capture {
		hm[side][move.from()][move.to()] += depth * depth
	}
}

// decrement decrements the score of history moves
func (hm *HistoryMovesTable) decrement(ml *MoveList, start int, side Color) {
	for i := range start {
		mv := ml.moves[ml.length+i]
		if mv.flag() < capture {
			hm[side][mv.from()][mv.to()]--
		}
	}
}

// clear clears the history of moves
func (hm *HistoryMovesTable) clear() {
	for i := range hm {
		for j := range hm[i] {
			for k := range hm[i][j] {
				hm[i][j][k] = 0
			}
		}
	}
}

// Killer is a list of quiet moves that produces a beta cutoff
type Killer [2]Move

// KillersTable is a table of killers moves entries
type KillersTable [MaxSearchDepth]Killer

// clear clears the killers table
func (kt *KillersTable) clear() {
	for i := range kt {
		kt[i][0] = NoMove
		kt[i][1] = NoMove
	}
}

// store stores a move in the killers table
func (kt *KillersTable) store(ply int, move Move) {
	if move.flag() < capture && ply < MaxSearchDepth && kt[ply][0] != move {
		kt[ply][1] = kt[ply][0]
		kt[ply][0] = move
	}
}

// get returns the killers at the given ply
func (kt *KillersTable) get(ply int) (Move, Move) {
	if ply >= MaxSearchDepth {
		return NoMove, NoMove
	}
	return kt[ply][0], kt[ply][1]
}

// Stack contains the history of moves played in the current branch of the search
type Stack struct {
	moves [MaxSearchDepth * 2]Move
}

// store stores the move in the stack at the given ply
func (s *Stack) store(move Move, ply int) {
	s.moves[ply] = move
}

// clear clears the stack
func (s *Stack) clear() {
	for i := range s.moves {
		s.moves[i] = NoMove
	}
}

// getPriorMove returns the previous move at the given ply
func (s *Stack) getPriorMove(ply int) Move {
	if ply == 0 {
		return NoMove
	}

	return s.moves[ply-1]
}

// CounterMoveTable is a table for storing the following move that might produce a beta cutoff
type CounterMoveTable [2][64][64]Move

// clear clears the history of moves
func (cm *CounterMoveTable) clear() {
	for i := range cm[White] {
		for j := range cm[White][i] {
			cm[White][i][j] = NoMove
			cm[Black][i][j] = NoMove
		}
	}
}

// store stores a move in the counter move table
func (cm *CounterMoveTable) store(priorMove, move Move, side Color) {
	if priorMove == NoMove {
		return
	}

	cm[side][priorMove.from()][priorMove.to()] = move
}

// get gets a move from the counter move table
func (cm *CounterMoveTable) get(priorMove Move, side Color) Move {
	if priorMove == NoMove {
		return NoMove
	}

	return cm[side][priorMove.from()][priorMove.to()]
}

// IterativeDeepening performs a progressive deepening search and returns the best move
func (s *Search) IterativeDeepening(pos *Position, maxDepth int, stdout chan string) (bestMoveScore int, bestMove string) {
	s.clear()

	// Ensure to return a move to the GUI
	bestMove = setDefaultMove(pos)

	for d := 1; d <= maxDepth; d++ {
		s.reset()

		lastScore := bestMoveScore
		bestMoveScore = s.aspirationSearch(pos, d, lastScore)

		if s.TimeControl.stop {
			bestMoveScore = lastScore
			break
		}

		depthTime := time.Since(s.TimeControl.iterationStartTime)
		s.TimeControl.iterationStartTime = time.Now()
		nps := int(float64(s.nodes) / depthTime.Seconds())
		stdout <- fmt.Sprintf("info depth %d seldepth %d score %s nodes %d nps %d hashfull %d time %v pv %v", d, s.seldepth, convertScore(bestMoveScore, d), s.nodes, nps, s.TranspositionTable.hashfull(), depthTime.Milliseconds(), s.pvLine.String())

		// TODO: handle out of bounds when indexing s.pvLine[0] ???
		bestMove = s.pvLine[0].String()
	}

	return
}

// aspirationSearch performs aspiration window search
func (s *Search) aspirationSearch(pos *Position, depth int, lastScore int) int {
	if depth < 4 {
		return s.negamax(pos, depth, 0, MinInt, MaxInt, &s.pvLine, true)
	}

	delta := AspirationWindowSize
	alpha := lastScore - delta
	beta := lastScore + delta

	for {
		score := s.negamax(pos, depth, 0, alpha, beta, &s.pvLine, true)

		// Score falls inside the aspiration window
		if score > alpha && score < beta {
			return score
		}

		// Adjust window size if fail
		if score <= alpha {
			alpha = max(alpha-delta, MinInt)
			delta *= 2
		} else if score >= beta {
			beta = min(beta+delta, MaxInt)
			delta *= 2
		}

		// If delta gets too big, make a full search
		if delta > 500 {
			alpha = MinInt
			beta = MaxInt
		}
	}
}

// setDefaultMove move returns the first legal move or the uci default null move (0000)
func setDefaultMove(pos *Position) string {
	pd := pos.generatePositionData()
	ml := NewMoveList()
	pos.generateCaptures(ml, &pd)
	pos.generateNonCaptures(ml, &pd)
	if ml.length > 0 {
		return ml.moves[0].String()
	}
	return "0000"
}

// negamax returns the score of the best posible move by the evaluation function for a fixed depth
func (s *Search) negamax(pos *Position, depth int, ply int, alpha int, beta int, pvLine *pvLine, nullMoveAllowed bool) int {
	s.nodes++
	pvLine.reset()

	if s.TimeControl.stop {
		return 0
	}

	rootNode := ply == 0
	// Avoid these on root nodes, because otherwise, we'll not be able to return a move
	if !rootNode {
		// If the position is either a draw by repetition, 50 move rule or insuficient material
		// stop inmediatly, to prevent redundant search
		if pos.isDraw() {
			return 0
		}

		// Mate Distance Pruning. Cut the trees and ajust alpha/beta bounds of lines where no shorter mate is possible
		alpha = max(alpha, -MateScore+ply)
		beta = min(beta, MateScore-ply-1)
		if alpha >= beta {
			return alpha
		}
	}

	isCheck := pos.Check(pos.Turn)
	if depth <= 0 && !isCheck {
		return Quiescent(pos, s, alpha, beta, ply)
	}

	pvNode := beta-alpha > 1

	// Transposition Table probe. If we already have searched this position with a sufficient depth
	// we will trust our previous evaluation and return the score
	ttScore, ttEval, ttMove, ttHit := s.TranspositionTable.probe(pos.Hash, depth, ply, alpha, beta)
	if ttHit && !pvNode {
		return ttScore
	}

	flag := FlagAlpha
	branchPv := NewPvLine(depth)
	staticEval := s.evaluate(pos, ttMove, ttEval)

	// Reverse Futility Pruning / Static Null Move pruning
	if depth <= 4 && !isCheck && !pvNode {
		margin := ReverseFutitlityPruningMargin * depth
		if staticEval-margin >= beta {
			return beta
		}
	}

	// Null Move Pruning. Gives a free shot to the opponent by passing the turn.
	// If we still exceed beta in the reduced search, we will trust our position is so good,
	// that it will also exceed beta if we search all moves.
	if depth >= 4 && !isCheck && nullMoveAllowed && !pvNode && pos.canNullMove() {
		ep := pos.makeNullMove()
		s.stack.store(NoMove, ply)
		sc := -s.negamax(pos, depth-4, ply+1, -beta, -beta+1, &branchPv, false)
		pos.unmakeNullMove(ep)

		if sc >= beta {
			return beta
		}
	}

	// Extended Futility Pruning
	// Discards potential moves near the horizon that are not likely to raise alpha (will not improve the position)
	futilityPruningAllowed := false
	if depth <= 4 && alpha > -MateScore && beta < MateScore {
		futilityPruningAllowed = staticEval+FutilityPruningMargin[depth] <= alpha
	}

	// Internal Iterative Deepening
	// If we dont have a move from the transposition table, make a reduced search to find a good
	// move to improve our move ordering
	if depth > 5 && pvNode && ttMove == NoMove {
		s.negamax(pos, depth/2, ply+1, alpha, beta, &branchPv, true)
		if len(branchPv) > 0 {
			ttMove = branchPv[0]
		}
		branchPv.reset()
	}

	newScore := MinInt
	bestMove := NoMove
	cm := s.counterMovesTable.get(s.stack.getPriorMove(ply), pos.Turn)
	k1, k2 := s.killers.get(ply)
	mg := NewMoveGenerator(pos, &ttMove, &k1, &k2, &cm, &s.historyMoves)

	for move := mg.nextMove(); move != NoMove; move = mg.nextMove() {
		branchPv.reset()
		moveFlag := move.flag()

		// Late Move Pruning
		// Prunes quiet moves that are likely not to be good, by assuming we have a good move ordering
		if depth <= 4 && mg.stage == NonCapturesStage && mg.moveNumber > LateMovePruningMoveNumber[depth] && !isCheck && !pvNode {
			continue
		}

		// Futility Pruning. Apply futility pruning if conditions are met
		if futilityPruningAllowed && moveFlag < capture && !isCheck && !pvNode && mg.moveNumber > 0 {
			continue
		}

		pos.MakeMove(&move)
		s.stack.store(move, ply)

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
			reduction := lmrReductionFactor(depth, mg.moveNumber, mg.stage, moveFlag, isCheck, pvNode)
			newScore = -s.negamax(pos, depth-1-reduction+extension, ply+1, -alpha-1, -alpha, &branchPv, true)

			// If an improvement was found, we need to search again with a full window and depth
			if newScore > alpha {
				newScore = -s.negamax(pos, depth-1+extension, ply+1, -beta, -alpha, &branchPv, true)
			}
		}
		pos.UnmakeMove(&move)

		if newScore >= beta {
			s.TranspositionTable.store(pos.Hash, depth, ply, FlagBeta, beta, staticEval, move)
			s.killers.store(ply, move)
			s.counterMovesTable.store(s.stack.getPriorMove(ply), move, pos.Turn)
			s.historyMoves.increment(depth, &move, pos.Turn)
			// Reduce history score for previous 'quiet' moves that did not produce the cutoff
			s.historyMoves.decrement(mg.moves, mg.moveNumber, pos.Turn)

			return beta
		}

		if newScore > alpha {
			flag = FlagExact
			bestMove = move

			alpha = newScore
			pvLine.insert(move, &branchPv)
		}
	}

	score, checkmateOrStealmateFound := isCheckmateOrStealmate(isCheck, mg.moveNumber, ply)
	if checkmateOrStealmateFound {
		s.TranspositionTable.store(pos.Hash, depth, ply, FlagExact, score, staticEval, NoMove)
		return score
	}

	s.TranspositionTable.store(pos.Hash, depth, ply, flag, alpha, staticEval, bestMove)
	return alpha
}

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

// evaluate returns the evaluation of the position
func (s *Search) evaluate(pos *Position, ttMove Move, ttEval int) int {
	if ttMove != NoMove {
		return ttEval
	}
	return s.Evaluation.Evaluate(pos)
}

// canNullMove returns if the current position allows a null move pruning
func (pos *Position) canNullMove() bool {
	if pos.material(pos.Turn) < EndgameMaterialThreshold {
		return false
	}

	if pos.kingAndPawnsOnlyEndgame() {
		return false
	}

	return true
}

// kingAndPawnsOnlyEndgame returns if the position is a king and pawns only endgame
func (pos *Position) kingAndPawnsOnlyEndgame() bool {
	whiteKingAndPawns := pos.Bitboards[WhiteKing] | pos.Bitboards[WhitePawn]
	blackKingAndPawns := pos.Bitboards[BlackKing] | pos.Bitboards[BlackPawn]

	return pos.pieces[White] == whiteKingAndPawns && pos.pieces[Black] == blackKingAndPawns
}

// material returns the total material of the position for the side passed
func (pos *Position) material(side Color) int {
	pieceValue := [6]int{0, 900, 500, 300, 300, 100}
	material := 0

	for piece, bitboard := range pos.getBitboards(side) {
		material += pieceValue[pieceRole(piece)] * bitboard.count()
	}
	return material
}

// lrmReductionFactor returns a number to reduce the depth on search based on the conditions passed
func lmrReductionFactor(depth, moveNumber, stage, moveFlag int, isCheck, pvNode bool) int {
	if isCheck || depth < 3 || moveNumber < 1 {
		return 0
	}
	reduction := LateMoveReductionFactor[depth][moveNumber]

	// Reduce less on pvNode
	if pvNode {
		reduction--
	}

	// Reduces less on captures/promotions
	if moveFlag >= capture {
		reduction--
	}

	// Reduce less on Killers and counter moves
	if stage < NonCapturesStage && stage > CapturesStage {
		reduction--
	}

	return max(reduction, 0)
}

// convertScore returns the proper score if it's mate or current centipawns score
func convertScore(score int, depth int) (result string) {
	absScore := abs(score)
	isMate := absScore >= MateScore-depth-5 // NOTE: 5 is used as a fail safe that helps when mate is found earlier/before the actual mate depth
	if isMate {
		mateIn := (MateScore - absScore + 1) / 2 // NOTE: in full moves, not ply!
		result = "mate " + strconv.Itoa((score/absScore)*mateIn)
	} else {
		result = "cp " + strconv.Itoa(score)
	}
	return
}

// abs returns the absolute value of the number passed
func abs(number int) int {
	if number < 0 {
		return -number
	}
	return number
}
