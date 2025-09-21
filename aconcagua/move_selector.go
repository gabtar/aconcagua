package aconcagua

const (
	// Move Generation Stages flags
	HashMoveStage = iota
	GenerateCapturesStage
	CapturesStage
	FirstKillerStage
	SecondKillerStage
	// TODO: Counter move heruistic???
	GenerateNonCapturesStage
	NonCapturesStage
	BadCapturesStage
	EndStage
)

type MoveSelector struct {
	stage                              int
	moveNumber                         int // the move count selected so far
	pos                                *Position
	pd                                 PositionData
	hashMove                           *Move
	killer1, killer2                   *Move
	historyMoves                       *HistoryMovesTable
	captures, nonCaptures, badCaptures MoveList
}

// NewMoveSelector returns a new move generator
func NewMoveSelector(pos *Position, hashMove *Move, killer1 *Move, killer2 *Move, historyMoves *HistoryMovesTable) *MoveSelector {
	return &MoveSelector{
		stage:        HashMoveStage,
		pos:          pos,
		hashMove:     hashMove,
		killer1:      killer1,
		killer2:      killer2,
		moveNumber:   -1, // NOTE: initialize with -1 to make the first move selected to have moveNumber = 0
		captures:     NewMoveList(30),
		badCaptures:  NewMoveList(30),
		nonCaptures:  NewMoveList(100),
		historyMoves: historyMoves,
	}
}

// nextMove return the nextMove move of the position
func (ms *MoveSelector) nextMove() (move Move) {
	ms.moveNumber++
	switch ms.stage {
	case HashMoveStage:
		ms.stage = GenerateCapturesStage
		if *ms.hashMove != NoMove {
			return *ms.hashMove
		}
		fallthrough
	case GenerateCapturesStage:
		ms.pd = ms.pos.generatePositionData()
		ms.stage = CapturesStage
		ms.pos.generateCaptures(&ms.captures, &ms.pd)
		scores := make([]int, len(ms.captures))
		for i := range len(ms.captures) {
			scores[i] = ms.pos.see(ms.captures[i].from(), ms.captures[i].to())
		}
		ms.captures.sort(scores)

		// BadCaptures
		for i := range scores {
			if scores[i] < 0 {
				ms.badCaptures = ms.captures[i:]
				ms.captures = ms.captures[:i]
				break
			}
		}

		fallthrough
	case CapturesStage:
		move = *ms.captures.pickFirst()
		if move != NoMove && move == *ms.hashMove {
			move = *ms.captures.pickFirst()
		}
		if move != NoMove {
			return move
		}
		ms.stage = FirstKillerStage
		fallthrough
	case FirstKillerStage:
		ms.stage = SecondKillerStage
		move = *ms.killer1
		// NOTE: we need to validate legality of killers for this position, because the may be for the same ply, but of another branch of the tree!!
		ms.pos.generateNonCaptures(&ms.nonCaptures, &ms.pd)
		if move != NoMove && move != *ms.hashMove && isLegalKiller(move, &ms.nonCaptures) {
			return move
		}
		fallthrough
	case SecondKillerStage:
		ms.stage = GenerateNonCapturesStage
		move = *ms.killer2
		if move != NoMove && move != *ms.hashMove && isLegalKiller(move, &ms.nonCaptures) {
			return move
		}
		fallthrough
	case GenerateNonCapturesStage:
		ms.stage = NonCapturesStage
		scores := make([]int, len(ms.nonCaptures))
		for i := range len(ms.nonCaptures) {
			scores[i] = ms.historyMoves[ms.pos.Turn][ms.nonCaptures[i].from()][ms.nonCaptures[i].to()]
		}
		ms.nonCaptures.sort(scores)
		fallthrough
	case NonCapturesStage:
		move = *ms.nonCaptures.pickFirst()

		if move != NoMove && move == *ms.hashMove {
			move = *ms.nonCaptures.pickFirst()
		}

		if move != NoMove && move == *ms.killer1 {
			move = *ms.nonCaptures.pickFirst()
		}

		if move != NoMove && move == *ms.killer2 {
			move = *ms.nonCaptures.pickFirst()
		}

		if move != NoMove {
			return move
		}
		ms.stage = BadCapturesStage
		fallthrough
	case BadCapturesStage:
		move = *ms.badCaptures.pickFirst()
		if move != NoMove && move == *ms.hashMove {
			move = *ms.badCaptures.pickFirst()
		}
		if move != NoMove {
			return move
		}
		ms.stage = EndStage
	case EndStage:
		return NoMove
	}
	return
}

// isLegalKiller returns if the move is legal in the current position
func isLegalKiller(move Move, ml *MoveList) bool {
	// Killer moves are always quiet moves, so we can just pass the non captures list to check if killer exits
	for i := range len(*ml) {
		if move == (*ml)[i] {
			return true
		}
	}
	return false
}
