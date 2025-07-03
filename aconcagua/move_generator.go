package aconcagua

// ------------------------------------------------------------------
// PIECE ATTACKS GENERATION
// ------------------------------------------------------------------

// kingAttacks returns a bitboard with the squares the king attacks from the passed bitboard
func kingAttacks(k *Bitboard) (attacks Bitboard) {
	notInHFile := *k & ^(*k & files[7])
	notInAFile := *k & ^(*k & files[0])

	attacks = notInAFile<<7 | *k<<8 | notInHFile<<9 |
		notInHFile<<1 | notInAFile>>1 | notInHFile>>7 |
		*k>>8 | notInAFile>>9
	return
}

// queenAttacks returns a Bitboard with all the squares a queen is attacking
func queenAttacks(q *Bitboard, blocks Bitboard) (attacks Bitboard) {
	attacks = rookAttacks(Bsf(*q), blocks) | bishopAttacks(Bsf(*q), blocks)
	return
}

// rookAttacks returns a bitboard with the attack mask of a rook from the square passed taking into account the blockers
func rookAttacks(square int, blocks Bitboard) Bitboard {
	blocks &= rooksMaskTable[square]
	magicIndex := (blocks * rookMagics[square]) >> (64 - rooksMaskTable[square].count())
	return rookAttacksTable[square][magicIndex]
}

// bishopAttacks returns a bitboard with the attack mask of a bishop from the square passed taking into account the blockers
func bishopAttacks(square int, blocks Bitboard) Bitboard {
	blocks &= bishopMaskTable[square]
	magicIndex := (blocks * bishopMagics[square]) >> (64 - bishopMaskTable[square].count())
	return bishopAttacksTable[square][magicIndex]
}

// pawnAttacks returns a bitboard with the squares the pawn attacks from the position passed
func pawnAttacks(p *Bitboard, side Color) (attacks Bitboard) {
	notInHFile := *p & ^(*p & files[7])
	notInAFile := *p & ^(*p & files[0])

	if side == White {
		attacks = notInAFile<<7 | notInHFile<<9
	} else {
		attacks = notInAFile>>9 | notInHFile>>7
	}
	return
}

// ------------------------------------------------------------------
// PIECE MOVES GENERATION (BITBOARD)
// ------------------------------------------------------------------

// kingMoves returns a bitboard with the legal moves of the king from the bitboard passed
func kingMoves(k *Bitboard, pos *Position, side Color) (moves Bitboard) {
	withoutKing := *pos
	withoutKing.RemovePiece(*k)
	moves = kingAttacks(k) & ^withoutKing.AttackedSquares(side.Opponent()) & ^pos.Pieces(side)
	return
}

// rookMoves returns a bitboard with the legal moves of the rook from the bitboard passed
func rookMoves(r *Bitboard, pd *PositionData) (moves Bitboard) {
	return rookAttacks(Bsf(*r), pd.allies|pd.enemies) &
		^pd.allies &
		pd.checkRestrictedSquares &
		pinRestrictedSquares(*r, pd.kingPosition, pd.pinnedPieces)
}

// bishopMoves returns a bitboard with the legal moves of the bishop from the bitboard passed
func bishopMoves(b *Bitboard, pd *PositionData) (moves Bitboard) {
	return bishopAttacks(Bsf(*b), pd.allies|pd.enemies) &
		^pd.allies &
		pd.checkRestrictedSquares &
		// checkRestrictedSquares(pd.kingPosition, pd.checkingSliders, pd.checkingNonSliders) &
		pinRestrictedSquares(*b, pd.kingPosition, pd.pinnedPieces)
}

// knightMoves returns a bitboard with the legal moves of the knight from the bitboard passed
func knightMoves(k *Bitboard, pd *PositionData) (moves Bitboard) {
	// If the knight is pinned it can move at all
	if *k&pd.pinnedPieces > 0 {
		return Bitboard(0)
	}
	return knightAttacksTable[Bsf(*k)] & ^pd.allies & pd.checkRestrictedSquares
}

// pawnMoves returns a Bitboard with the squares a pawn can move to in the passed position
func pawnMoves(p *Bitboard, pd *PositionData, side Color) (moves Bitboard) {
	posibleCaptures := pawnAttacks(p, side) & pd.enemies
	posiblesMoves := Bitboard(0)
	emptySquares := ^(pd.allies | pd.enemies)

	if side == White {
		singleMove := *p << 8 & emptySquares
		firstPawnMoveAvailable := (*p & ranks[1]) << 16 & (singleMove << 8) & emptySquares
		posiblesMoves = singleMove | firstPawnMoveAvailable
	} else {
		singleMove := *p >> 8 & emptySquares
		firstPawnMoveAvailable := (*p & ranks[6]) >> 16 & (singleMove >> 8) & emptySquares
		posiblesMoves = singleMove | firstPawnMoveAvailable
	}

	moves = (posibleCaptures | posiblesMoves) & pd.checkRestrictedSquares &
		pinRestrictedSquares(*p, pd.kingPosition, pd.pinnedPieces)
	return
}

// ------------------------------------------------------------------
// PIECE MOVES GENERATION (MOVE LIST)
// ------------------------------------------------------------------

// genTargetMoves generates the moves from the square passed to the targets passed into the move list
func genTargetMoves(from *Bitboard, targets Bitboard, ml *moveList, pd *PositionData) {
	for targets > 0 {
		toSquare := targets.NextBit()
		flag := quiet
		if toSquare&pd.enemies > 0 {
			flag = capture
		}
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
}

// genMovesFromTargets generates the moves from the square passed to the targets passed in a MoveList
func genMovesFromTargets(from *Bitboard, targets Bitboard, ml *MoveList, pd *PositionData) {
	for targets > 0 {
		toSquare := targets.NextBit()
		flag := quiet
		if toSquare&pd.enemies > 0 {
			flag = capture
		}
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), uint16(flag)))
	}
}

// genCastleMoves generates the castles moves availabes in the move list
func genCastleMoves(from *Bitboard, pos *Position, ml *moveList) {
	if canCastleShort(from, pos, pos.Turn) {
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(*from<<2)), kingsideCastle))
	}
	if canCastleLong(from, pos, pos.Turn) {
		ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(*from>>2)), queensideCastle))
	}
}

// genPawnMoves generates the pawn moves in the move list
func genPawnMoves(from *Bitboard, side Color, ml *moveList, pd *PositionData) {
	toSquares := pawnMoves(from, pd, side)

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flags := pawnMoveFlag(from, &toSquare, pd, side)

		for i := range flags {
			ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), flags[i]))
		}
	}
}

// genPawnCaptures generates the pawn captures in the move list
func genPawnCaptures(from *Bitboard, side Color, ml *moveList, pd *PositionData) {
	toSquares := pawnMoves(from, pd, side) & pd.enemies

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flags := pawnMoveFlag(from, &toSquare, pd, side)

		for i := range flags {
			ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), flags[i]))
		}
	}
}

// genPawnCapturesMoves generates the pawn captures in the move list
func genPawnCapturesMoves(from *Bitboard, side Color, ml *MoveList, pd *PositionData) {
	toSquares := pawnMoves(from, pd, side) & pd.enemies

	for toSquares > 0 {
		toSquare := toSquares.NextBit()
		flags := pawnMoveFlag(from, &toSquare, pd, side)

		for i := range flags {
			ml.add(*encodeMove(uint16(Bsf(*from)), uint16(Bsf(toSquare)), flags[i]))
		}
	}
}

// genEpPawnCaptures generates the enPassant captures on the move list
func genEpPawnCaptures(pos *Position, side Color, ml *moveList) {
	if pos.enPassantTarget == 0 {
		return
	}
	from := potentialEpCapturers(pos, side)

	for from > 0 {
		fromBB := from.NextBit()
		move := *encodeMove(uint16(Bsf(fromBB)), uint16(Bsf(pos.enPassantTarget)), epCapture)

		pos.MakeMove(&move)
		if !pos.Check(side) {
			ml.add(move)
		}
		pos.UnmakeMove(&move)
	}
}

// ------------------------------------------------------------------
// LEGAL MOVE VALIDATION FUNCTIONS
// ------------------------------------------------------------------

// checkRestrictedSquares returns a bitboard with the squares that are allowed to move when in check
func checkRestrictedSquares(king Bitboard, checkingSliders Bitboard, checkingNonSliders Bitboard) (allowedSquares Bitboard) {
	checkingPieces := checkingSliders | checkingNonSliders
	if checkingPieces.count() == 0 {
		return AllSquares
	}

	if checkingPieces == checkingSliders && checkingPieces.count() == 1 {
		return getRayPath(&checkingPieces, &king) | checkingPieces
	}

	if checkingPieces.count() == 1 {
		return checkingPieces
	}

	return
}

// pinRestrictedSquares returns a bitboard with the squares allowed to move when the piece is pinned
func pinRestrictedSquares(piece Bitboard, king Bitboard, pinnedPieces Bitboard) (restrictedSquares Bitboard) {
	if pinnedPieces&piece > 0 {
		direction := directions[Bsf(piece)][Bsf(king)]
		return raysDirection(king, direction)
	}
	return AllSquares
}

// generateCaptures generates all captures in the position and stores them in the move list
func (pos *Position) generateCaptures(ml *MoveList) {
	pd := pos.generatePositionData()
	bitboards := pos.getBitboards(pos.Turn)

	for piece, bb := range bitboards {
		for bb > 0 {
			pieceBB := bb.NextBit()
			switch piece {
			case 0: // King
				genMovesFromTargets(&pieceBB, kingMoves(&pieceBB, pos, pos.Turn)&pd.enemies, ml, &pd)
			case 1: // Queen
				genMovesFromTargets(&pieceBB, (rookMoves(&pieceBB, &pd)|bishopMoves(&pieceBB, &pd))&pd.enemies, ml, &pd)
			case 2: // Rook
				genMovesFromTargets(&pieceBB, rookMoves(&pieceBB, &pd)&pd.enemies, ml, &pd)
			case 3: // Bishop
				genMovesFromTargets(&pieceBB, bishopMoves(&pieceBB, &pd)&pd.enemies, ml, &pd)
			case 4: // Knight
				genMovesFromTargets(&pieceBB, knightMoves(&pieceBB, &pd)&pd.enemies, ml, &pd)
			case 5: // Pawn
				genPawnCapturesMoves(&pieceBB, pos.Turn, ml, &pd)
			}
		}
	}
	// TODO: add en passant captures if any w/ the new move list
}

// generateNonCaptures generates all non captures in the position and stores them in the move list
func (pos *Position) generateNonCaptures(from Bitboard, targets Bitboard, ml *moveList) {
	// TODO: implement
}

// Stagged Move Generator
// Order for move generation
// 1 transposition table/hash move
// 2 Generate all captures
// 3 Pick captures by static exchange evaluation score
// 4 Pick killer moves if any
// 5 Generate all Non Captures moves
// 6 Pick non captures until no moves left

const (
	// Move Generation Stages flags
	HashMoveStage = iota
	GenerateCapturesStage
	CapturesStage // TODO: Maybe later separate into 'good' captures(see >= 0) and bad captures (see < 0)
	FirstKillerStage
	SecondKillerStage
	GenerateNonCapturesStage
	NonCapturesStage
	EndStage
	// TODO: use counter move/any other move ordering types/strategies
)

type MoveGenerator struct {
	// TODO: maybe i should add the positionData reference directly here for the current position to generate intermediate moves
	stage                 int
	moveNumber            int // the last move count generated for this position
	pos                   *Position
	hashMove              *Move
	killers               *Killer
	captures, nonCaptures MoveList
}

// NewMoveGenerator returns a new move generator
func NewMoveGenerator(pos *Position, hashMove *Move, killers *Killer) *MoveGenerator {
	return &MoveGenerator{
		stage:       HashMoveStage,
		pos:         pos,
		hashMove:    hashMove,
		killers:     killers,
		moveNumber:  0,
		captures:    NewMoveList(50), // TODO: check if 50 and 100 default allocated capacity is enought for most cases
		nonCaptures: NewMoveList(100),
	}
}

// nextMove return the nextMove move of the position
func (mg *MoveGenerator) nextMove() (move Move) {
	mg.moveNumber++
	switch mg.stage {
	case HashMoveStage:
		mg.stage = GenerateCapturesStage
		if *mg.hashMove != NoMove {
			return *mg.hashMove
		}
		fallthrough
	case GenerateCapturesStage:
		mg.stage = CapturesStage
		mg.pos.generateCaptures(&mg.captures)
		scores := make([]int, len(mg.captures))
		for i := range len(mg.captures) {
			scores[i] = mg.pos.see(mg.captures[i].from(), mg.captures[i].to())
		}
		mg.captures.sort(scores)
		fallthrough
	case CapturesStage:
		move = *mg.captures.pickFirst()
		if move == *mg.hashMove {
			move = *mg.captures.pickFirst()
		}
		if move != NoMove {
			return
		}
		mg.stage = FirstKillerStage
		fallthrough
	case FirstKillerStage:
		mg.stage = SecondKillerStage
		if *mg.hashMove != (*mg.killers)[0] {
			return (*mg.killers)[0]
		}
		fallthrough
	case SecondKillerStage:
		mg.stage = GenerateNonCapturesStage
		if *mg.hashMove != (*mg.killers)[1] {
			return (*mg.killers)[1]
		}
		fallthrough
	case GenerateNonCapturesStage:
		fallthrough
	case NonCapturesStage:
		return NoMove
	case EndStage:
		return NoMove
	}
	return
}
