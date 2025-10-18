package aconcagua

import (
	"fmt"
	"math"
	"time"
)

// Constants to use in the search
const (
	MateScore                     = 100000
	MinInt                        = math.MinInt32
	MaxInt                        = math.MaxInt32
	EndgameMaterialThreshold      = 1600
	ReverseFutitlityPruningMargin = 200
	LateMovePruningMoveNumber     = 20
	AspirationWindowSize          = 45
)

// Search is the main struct for the search
type Search struct {
	nodes              int
	pvLine             pvLine
	killers            KillersTable
	historyMoves       HistoryMovesTable
	transpositionTable TranspositionTable
	timeControl        TimeControl
}

// NewSearch returns a pointer to a new Search struct
func NewSearch() *Search {
	return &Search{
		nodes:              0,
		pvLine:             NewPvLine(MaxSearchDepth),
		killers:            KillersTable{},
		historyMoves:       HistoryMovesTable{},
		transpositionTable: *NewTranspositionTable(DefaultTableSizeInMb),
		timeControl:        TimeControl{},
	}
}

// clear clears the search
func (s *Search) clear() {
	s.nodes = 0
	s.killers.clear()
	s.historyMoves.clear()
	s.transpositionTable.clear()
}

// reset sets the new iteration parameters in the NewSearch
func (s *Search) reset() {
	s.nodes = 0
	s.pvLine.reset()
}

// HistoryMovesTable is a table for holding the history of moves
type HistoryMovesTable [2][64][64]int

// update updates the history of moves
func (hm *HistoryMovesTable) update(depth int, from int, to int, side Color) {
	hm[side][from][to] += depth * depth
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

// add adds a non capture move to the killer list
func (k *Killer) add(move Move) {
	if move.flag() < capture {
		k[1] = k[0]
		k[0] = move
	}
}

// KillersTable is a table of killers moves entries
type KillersTable [MaxSearchDepth]Killer

// clear clears the killers table
func (kt *KillersTable) clear() {
	for i := range kt {
		kt[i][0] = NoMove
		kt[i][1] = NoMove
	}
}

// root is the entry point of the search
func (s *Search) root(pos *Position, maxDepth int, stdout chan string) (bestMoveScore int, bestMove string) {
	alpha := MinInt
	beta := MaxInt
	s.clear()
	pos.eval.pawnHashTable.clear()

	// Ensure to return a move to the GUI
	bestMove = setDefaultMove(pos)

	for d := 1; d <= maxDepth; d++ {
		s.reset()

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

		alpha = bestMoveScore - AspirationWindowSize
		beta = bestMoveScore + AspirationWindowSize

		depthTime := time.Since(s.timeControl.iterationStartTime)
		s.timeControl.iterationStartTime = time.Now()
		nps := int(float64(s.nodes) / depthTime.Seconds())
		stdout <- fmt.Sprintf("info depth %d score %s nodes %d nps %d hashfull %d time %v pv %v", d, convertScore(bestMoveScore, d), s.nodes, nps, s.transpositionTable.hashfull(), depthTime.Milliseconds(), s.pvLine.String())
		bestMove = s.pvLine[0].String()
	}

	return
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
	if s.timeControl.stop {
		return 0
	}

	if pos.isDraw() {
		return 0
	}

	isCheck := pos.Check(pos.Turn)
	if depth <= 0 && !isCheck {
		return Quiescent(pos, s, alpha, beta)
	}

	pvNode := beta-alpha > 1
	ttScore, ttMove, exists := s.transpositionTable.probe(pos.Hash, depth, alpha, beta)
	if exists && !pvNode {
		return ttScore
	}

	flag := FlagAlpha
	branchPv := NewPvLine(depth)
	pvLine.reset()

	// Reverse Futility Pruning / Static Null Move pruning
	if depth <= 4 && !isCheck && !pvNode {
		sc := pos.Evaluate()
		margin := ReverseFutitlityPruningMargin * depth
		if sc-margin >= beta {
			return beta
		}
	}

	// Null Move Pruning
	if depth >= 4 && !isCheck && nullMoveAllowed && !pvNode && pos.canNullMove() {
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
		sc := pos.Evaluate()
		futilityPruningAllowed = sc+futilityMargin[depth] <= alpha
	}

	// Internal Iterative Deepening
	if depth > 5 && pvNode && ttMove == NoMove {
		s.negamax(pos, depth/2, ply+1, alpha, beta, &branchPv, true)
		if len(branchPv) > 0 {
			ttMove = branchPv[0]
		}
		branchPv.reset()
	}

	newScore := MinInt
	bestMove := NoMove
	mg := NewMoveGenerator(pos, &ttMove, &s.killers[ply][0], &s.killers[ply][1], &s.historyMoves)

	for move := mg.nextMove(); move != NoMove; move = mg.nextMove() {
		pos.MakeMove(&move)
		branchPv.reset()
		moveFlag := move.flag()

		// Late Move Pruning
		if depth <= 3 && mg.stage >= NonCapturesStage && mg.moveNumber > LateMovePruningMoveNumber && !isCheck && !pvNode && !pos.Check(pos.Turn) {
			pos.UnmakeMove(&move)
			continue
		}

		// Futility Pruning
		if futilityPruningAllowed && moveFlag < capture && !isCheck && !pvNode && mg.moveNumber > 0 {
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
			if flag < capture {
				s.historyMoves.update(depth, move.from(), move.to(), pos.Turn)
			}

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
		return score
	}

	s.transpositionTable.store(pos.Hash, depth, flag, alpha, bestMove)
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
func lmrReductionFactor(depth int, moveNumber int, moveFlag int, isCheck, pvNode bool) int {
	if isCheck || pvNode || depth < 3 || moveNumber < 4 || moveFlag >= capture {
		return 0
	}

	return int(0.5 + math.Log(float64(depth))*math.Log(float64(moveNumber))/2.0)
}
