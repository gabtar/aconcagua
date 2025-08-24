package aconcagua

import (
	"strconv"
	"strings"
)

// Constants for pieces
const (
	// Pieces
	WhiteKing = iota
	WhiteQueen
	WhiteRook
	WhiteBishop
	WhiteKnight
	WhitePawn
	BlackKing
	BlackQueen
	BlackRook
	BlackBishop
	BlackKnight
	BlackPawn
	NoPiece
	// Piece role/type
	King   = 0
	Queen  = 1
	Rook   = 2
	Bishop = 3
	Knight = 4
	Pawn   = 5
	// Colors
	White = 0
	Black = 1
)

// squareReference maps an integer(the index of the array) with the
// corresponding square by using the notation of Little-Endian Rank-File Mapping
var squareReference = []string{
	"a1", "b1", "c1", "d1", "e1", "f1", "g1", "h1",
	"a2", "b2", "c2", "d2", "e2", "f2", "g2", "h2",
	"a3", "b3", "c3", "d3", "e3", "f3", "g3", "h3",
	"a4", "b4", "c4", "d4", "e4", "f4", "g4", "h4",
	"a5", "b5", "c5", "d5", "e5", "f5", "g5", "h5",
	"a6", "b6", "c6", "d6", "e6", "f6", "g6", "h6",
	"a7", "b7", "c7", "d7", "e7", "f7", "g7", "h7",
	"a8", "b8", "c8", "d8", "e8", "f8", "g8", "h8",
}

// pieceRole returns the role/type of the piece passed
func pieceRole(piece int) int {
	return piece % 6
}

// piece returns the piece of the role and color passed
func pieceColor(role int, color Color) int {
	return role + int(color)*6
}

// isSliding returns a the passed Piece is an sliding piece(Queen, Rook or Bishop)
func isSliding(piece int) bool {
	return (piece >= WhiteQueen && piece <= WhiteBishop) ||
		(piece >= BlackQueen && piece <= BlackBishop)
}

// Color is an int referencing the white or black player in chess
type Color int

// Opponent returns the Opponent color to the actual color
func (c Color) Opponent() Color {
	if c == White {
		return Black
	}
	return White
}

// Position contains all information about a chess position
type Position struct {
	// Bitboards piece order -> King, Queen, Rook, Bishop, Knight, Pawn (first white, second black)
	Bitboards       [12]Bitboard
	pieces          [2]Bitboard // black and white pieces 'map' on the board
	Turn            Color
	Hash            uint64
	PawnHash        uint64
	castling        castling
	enPassantTarget Bitboard
	halfmoveClock   int
	fullmoveNumber  int
	positionHistory PositionHistory
	evaluation      Evaluation
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

// material returns the total material of the position
func (pos *Position) material(side Color) int {
	pieceValue := [6]int{0, 900, 500, 300, 300, 100}
	material := 0

	for piece, bitboard := range pos.getBitboards(side) {
		material += pieceValue[pieceRole(piece)] * bitboard.count()
	}
	return material
}

// kingAndPawnsOnlyEndgame returns if the position is a king and pawns only endgame
func (pos *Position) kingAndPawnsOnlyEndgame() bool {
	whiteKingAndPawns := pos.Bitboards[WhiteKing] | pos.Bitboards[WhitePawn]
	blackKingAndPawns := pos.Bitboards[BlackKing] | pos.Bitboards[BlackPawn]

	return pos.pieces[White] == whiteKingAndPawns && pos.pieces[Black] == blackKingAndPawns
}

// PositionData contains relevant data for legal move validations of a position
type PositionData struct {
	kingPosition           Bitboard
	checkRestrictedSquares Bitboard
	pinnedPieces           Bitboard
	allies                 Bitboard
	enemies                Bitboard
}

// generatePositionData returns the position data for the current position
func (pos *Position) generatePositionData() PositionData {
	checkingPieces, checkingSliders := pos.CheckingPieces(pos.Turn)
	checkRestrictedSquares := checkRestrictedSquares(pos.KingPosition(pos.Turn), checkingSliders, checkingPieces&^checkingSliders)

	return PositionData{
		kingPosition:           pos.KingPosition(pos.Turn),
		checkRestrictedSquares: checkRestrictedSquares,
		pinnedPieces:           pos.pinnedPieces(pos.Turn),
		allies:                 pos.pieces[pos.Turn],
		enemies:                pos.pieces[pos.Turn.Opponent()],
	}
}

// PieceAt returns a Piece at the given square coordinate in the Position or error
func (pos *Position) PieceAt(square string) (piece int) {
	bitboardSquare := bitboardFromCoordinates(square)

	for index, bitboard := range pos.Bitboards {
		if bitboard&bitboardSquare > 0 {
			piece = index
			return
		}
	}

	piece = NoPiece
	return
}

// AddPiece adds to the position the Piece passed in the square passed
func (pos *Position) AddPiece(piece int, square string) {
	if piece == NoPiece {
		return
	}

	bitboardSquare := bitboardFromCoordinates(square)
	pos.Hash = pos.Hash ^ zobristHashKeys.getPieceSquareKey(piece, Bsf(bitboardSquare))
	if pieceRole(piece) == Pawn {
		pos.PawnHash = pos.PawnHash ^ zobristHashKeys.getPieceSquareKey(piece, Bsf(bitboardSquare))
	}
	pos.Bitboards[piece] |= bitboardSquare
	pos.pieces[int(piece/6)] |= bitboardSquare
}

// EmptySquares returns a Bitboard with the empty sqaures of the position
func (pos *Position) EmptySquares() (emptySquares Bitboard) {
	return ^pos.pieces[White] ^ pos.pieces[Black]
}

// AttackedSquares returns a bitboard with all squares attacked by the passed side
func (pos *Position) AttackedSquares(side Color) (attackedSquares Bitboard) {
	startingPiece := startingPieceNumber(side)
	bitboards := pos.getBitboards(side)

	for i, bb := range bitboards[0:5] {
		piece := startingPiece + i
		sq := bb.NextBit()

		for sq > 0 {
			attackedSquares |= Attacks(piece, sq, ^pos.EmptySquares())
			sq = bb.NextBit()
		}
	}
	attackedSquares |= pawnAttacks(&bitboards[5], side)

	return
}

// startingPieceNumber return an integer with the number of the starting bitboard piece for the color passed
func startingPieceNumber(side Color) int {
	startingBitboard := 0
	if side != White {
		startingBitboard = 6
	}
	return startingBitboard
}

// CheckingPieces returns two Bitboards, one with all the checking pieces and one with only the checking sliders pieces
func (pos *Position) CheckingPieces(side Color) (checkingPieces Bitboard, checkingSliders Bitboard) {
	if !pos.Check(side) {
		return
	}

	kingSq := pos.KingPosition(side)
	startingPieceNumber := startingPieceNumber(side.Opponent())

	for i, bitboard := range pos.getBitboards(side.Opponent()) {
		piece := startingPieceNumber + i

		for bitboard > 0 {
			sq := bitboard.NextBit()

			if kingSq&Attacks(piece, sq, ^pos.EmptySquares()) > 0 {
				if isSliding(piece) {
					checkingSliders |= sq
				}
				checkingPieces |= sq
			}
		}

	}
	return
}

// Check returns if the side passed is in check
func (pos *Position) Check(side Color) (inCheck bool) {
	kingPos := pos.KingPosition(side)

	if (kingPos & pos.AttackedSquares(side.Opponent())) > 0 {
		inCheck = true
	}
	return
}

// pinnedPieces returns a bitboard with the pieces pinned in the position for the side passed
func (pos *Position) pinnedPieces(side Color) (pinned Bitboard) {
	king := pos.KingPosition(side)
	opponentDiagonalAttackers := pos.Bitboards[int(side.Opponent())*6+int(WhiteQueen)] | pos.Bitboards[int(side.Opponent())*6+int(WhiteBishop)]
	opponentOrthogonalAttackers := pos.Bitboards[int(side.Opponent())*6+int(WhiteQueen)] | pos.Bitboards[int(side.Opponent())*6+int(WhiteRook)]
	ownPieces := pos.pieces[side] &^ king
	opponentPieces := pos.pieces[side.Opponent()]

	if king == 0 {
		return
	}

	for opponentDiagonalAttackers > 0 {
		opponent := opponentDiagonalAttackers.NextBit()
		opponentDiagonalRays := bishopAttacks(Bsf(opponent), Bitboard(0))
		kingToOpponentPath := getRayPath(&opponent, &king)
		intersectionRay := opponentDiagonalRays & kingToOpponentPath
		piecesBetween := intersectionRay & ownPieces
		oponentPiecesBetween := intersectionRay & opponentPieces

		if opponentDiagonalRays&king > 0 && piecesBetween.count() == 1 && oponentPiecesBetween == 0 {
			pinned |= piecesBetween
		}
	}

	for opponentOrthogonalAttackers > 0 {
		opponent := opponentOrthogonalAttackers.NextBit()
		opponentOrthogonalRays := rookAttacks(Bsf(opponent), Bitboard(0))
		kingToOpponentPath := getRayPath(&opponent, &king)
		intersectionRay := opponentOrthogonalRays & kingToOpponentPath
		piecesBetween := intersectionRay & ownPieces
		oponentPiecesBetween := intersectionRay & opponentPieces

		if opponentOrthogonalRays&king > 0 && piecesBetween.count() == 1 && oponentPiecesBetween == 0 {
			pinned |= piecesBetween
		}
	}

	return
}

// KingPosition returns the bitboard of the passed side king
func (pos *Position) KingPosition(side Color) (king Bitboard) {
	king = pos.Bitboards[pieceColor(King, side)]
	return
}

// RemovePiece returns a new position without the piece passed
func (pos *Position) RemovePiece(piece int, sq Bitboard) {
	if piece == NoPiece {
		return
	}

	pos.Bitboards[piece] &= ^sq
	pos.pieces[piece/6] &= ^sq
	pos.Hash = pos.Hash ^ zobristHashKeys.getPieceSquareKey(piece, Bsf(sq))
	if pieceRole(piece) == Pawn {
		pos.PawnHash = pos.PawnHash ^ zobristHashKeys.getPieceSquareKey(piece, Bsf(sq))
	}
}

func (pos *Position) getBitboards(side Color) (bitboards []Bitboard) {
	if side == White {
		bitboards = pos.Bitboards[WhiteKing:BlackKing]
	} else {
		bitboards = pos.Bitboards[BlackKing:NoPiece]
	}
	return
}

// isDraw returns if the current position is a draw by repetition, 50 move rule or insuficient material
func (pos *Position) isDraw() bool {
	return pos.positionHistory.repetitionCount(pos.Hash) >= 2 ||
		pos.halfmoveClock >= 100 ||
		pos.insuficientMaterial()
}

// insuficientMaterial returns if the current position is a draw by insuficient material
func (pos *Position) insuficientMaterial() bool {
	if pos.Bitboards[WhitePawn] > 0 || pos.Bitboards[BlackPawn] > 0 {
		return false
	}
	if pos.Bitboards[WhiteQueen] > 0 || pos.Bitboards[BlackQueen] > 0 {
		return false
	}
	if pos.Bitboards[WhiteRook] > 0 || pos.Bitboards[BlackRook] > 0 {
		return false
	}
	if pos.Bitboards[WhiteBishop].count() > 1 || pos.Bitboards[BlackBishop].count() > 1 {
		return false
	}
	if pos.Bitboards[WhiteKnight].count() > 1 || pos.Bitboards[BlackKnight].count() > 1 {
		return false
	}
	if pos.Bitboards[WhiteBishop].count() == pos.Bitboards[BlackBishop].count() && pos.Bitboards[WhiteBishop] > 0 {
		return false
	}

	return true
}

// MakeMove executes a chess move, updating the board state
func (pos *Position) MakeMove(move *Move) {
	pieceToMove := pos.PieceAt(squareReference[move.from()])
	pieceCaptured := pos.getCapturedPiece(move)

	pos.positionHistory.add(
		encodePositionBefore(
			uint16(pieceToMove),
			uint16(pieceCaptured),
			uint16(Bsf(pos.enPassantTarget)),
			uint16(pos.halfmoveClock),
		),
		pos.castling.castlingRights,
		pos.Hash,
	)

	pos.updateCastleRights(pos.castling.updateCastleRights(move.from(), move.to()))
	pos.RemovePiece(pieceToMove, bitboardFromIndex(move.from()))

	pos.updateMoveState()
	pos.handleSpecialMoveTypes(move, &pieceToMove, &pieceCaptured)

	pos.toggleSide()

	toSq := getMoveDestinationSquare(move, pos)
	pos.AddPiece(pieceToMove, squareReference[toSq])
}

func (pos *Position) updateCastleRights(next castlingRights) {
	pos.Hash = pos.Hash ^ zobristHashKeys.getCastleKey(pos.castling.castlingRights)
	pos.castling.castlingRights = next
	pos.Hash = pos.Hash ^ zobristHashKeys.getCastleKey(pos.castling.castlingRights)
}

// toggleSide toggles the side of the position
func (pos *Position) toggleSide() {
	pos.Turn = pos.Turn.Opponent()
	pos.Hash = pos.Hash ^ zobristHashKeys.getSideKey()
}

// updateMoveState resets and updates move-related state
func (pos *Position) updateMoveState() {
	if pos.enPassantTarget > 0 { // Undo en passant key if set
		pos.Hash = pos.Hash ^ zobristHashKeys.getEpKey(Bsf(pos.enPassantTarget))
	}
	pos.enPassantTarget = 0
	pos.halfmoveClock++

	if pos.Turn == Black {
		pos.fullmoveNumber++
	}
}

// handleSpecialMoveTypes processes different types of chess moves
func (pos *Position) handleSpecialMoveTypes(move *Move, pieceToMove *int, pieceCaptured *int) {
	switch flag := move.flag(); flag {
	case quiet:
		pos.handleQuietMove(*pieceToMove)
	case doublePawnPush:
		pos.handleDoublePawnPush(*move, *pieceToMove)
	case capture:
		pos.handleCapture(*move, pieceCaptured)
	case knightPromotion, bishopPromotion, rookPromotion, queenPromotion,
		knightCapturePromotion, bishopCapturePromotion, rookCapturePromotion, queenCapturePromotion:
		pos.handlePromotion(*move, flag, pieceToMove, pieceCaptured)
	case kingsideCastle, queensideCastle:
		pos.updateRookPositionOnCaslte(flag, true)
	case epCapture:
		pos.handleEnPassantCapture(*move, *pieceToMove)
	}
}

// handleQuietMove resets halfmove clock for pawn moves
func (pos *Position) handleQuietMove(pieceToMove int) {
	if pieceToMove == WhitePawn || pieceToMove == BlackPawn {
		pos.halfmoveClock = 0
	}
}

// handleDoublePawnPush sets en passant target square
func (pos *Position) handleDoublePawnPush(move Move, pieceToMove int) {
	if pieceToMove == WhitePawn {
		pos.enPassantTarget = bitboardFromIndex(move.to()) >> 8
	} else {
		pos.enPassantTarget = bitboardFromIndex(move.to()) << 8
	}
	pos.Hash = pos.Hash ^ zobristHashKeys.getEpKey(Bsf(pos.enPassantTarget))
}

// handleCapture removes captured piece and resets halfmove clock
func (pos *Position) handleCapture(move Move, pieceCaptured *int) {
	pos.RemovePiece(*pieceCaptured, bitboardFromIndex(move.to()))
	pos.halfmoveClock = 0
}

// handlePromotion processes pawn promotion
func (pos *Position) handlePromotion(move Move, flag int, pieceToMove *int, pieceCaptured *int) {
	pos.RemovePiece(*pieceCaptured, bitboardFromIndex(move.to()))
	*pieceToMove = getPromotedToPiece(flag, pos.Turn)

	pos.halfmoveClock = 0
}

// handleEnPassantCapture removes the captured pawn in en passant
func (pos *Position) handleEnPassantCapture(move Move, pieceToMove int) {
	if pieceToMove == WhitePawn {
		pos.RemovePiece(BlackPawn, bitboardFromIndex(move.to())>>8)
	} else {
		pos.RemovePiece(WhitePawn, bitboardFromIndex(move.to())<<8)
	}
	pos.halfmoveClock = 0
}

// getPromotedToPiece returns the piece promoted to based on the flag passed
func getPromotedToPiece(flag int, side Color) (piece int) {
	switch flag {
	case knightPromotion, knightCapturePromotion:
		return pieceColor(Knight, side)
	case bishopPromotion, bishopCapturePromotion:
		return pieceColor(Bishop, side)
	case rookPromotion, rookCapturePromotion:
		return pieceColor(Rook, side)
	case queenCapturePromotion, queenPromotion:
		return pieceColor(Queen, side)
	}
	return NoPiece
}

// getCapturedPiece determines the piece captured in the move
func (pos *Position) getCapturedPiece(move *Move) int {
	if move.flag() == epCapture {
		return pieceColor(Pawn, pos.Turn.Opponent())
	}
	pieceCaptured := pos.PieceAt(squareReference[move.to()])
	return pieceCaptured
}

// getMoveDestinationSquare returns the destination square of the move
func getMoveDestinationSquare(move *Move, pos *Position) int {
	if move.flag() == queensideCastle || move.flag() == kingsideCastle {
		return pos.castling.kingsEndSquare[pos.Turn.Opponent()][move.flag()-kingsideCastle]
	}
	return move.to()
}

// UnmakeMove undoes a the passed move in the postion
func (pos *Position) UnmakeMove(move *Move) {
	prevState, castle := pos.positionHistory.pop()

	toSq := getMoveDestinationSquare(move, pos)

	pieceToRemove := prevState.pieceMoved()
	if move.flag() >= knightPromotion {
		pieceToRemove = getPromotedToPiece(move.flag(), pos.Turn.Opponent())
	}
	pos.RemovePiece(pieceToRemove, bitboardFromIndex(toSq))

	if pos.Turn == White {
		pos.fullmoveNumber--
	}

	if pos.enPassantTarget > 0 {
		pos.Hash = pos.Hash ^ zobristHashKeys.getEpKey(Bsf(pos.enPassantTarget))
	}
	if prevState.epTarget() > 0 {
		pos.enPassantTarget = bitboardFromIndex(prevState.epTarget())
		pos.Hash = pos.Hash ^ zobristHashKeys.getEpKey(prevState.epTarget())
	}
	pos.halfmoveClock = prevState.rule50()

	pos.updateCastleRights(castle)

	pieceToAdd := prevState.pieceMoved()

	switch flag := move.flag(); flag {
	case capture, knightCapturePromotion, bishopCapturePromotion, rookCapturePromotion, queenCapturePromotion:
		pos.AddPiece(prevState.pieceCaptured(), squareReference[move.to()])
	case queensideCastle, kingsideCastle:
		pos.updateRookPositionOnCaslte(flag, false)
	case epCapture:
		colorModifier := 1 - int(pos.Turn)*2
		restoreSq := move.to() + 8*colorModifier
		pos.AddPiece(pieceColor(Pawn, pos.Turn), squareReference[restoreSq])
	case doublePawnPush:
		pos.Hash = pos.Hash ^ zobristHashKeys.getEpKey(prevState.epTarget())
		pos.enPassantTarget = 0
	}

	pos.toggleSide()

	fromSq := move.from()
	if move.flag() == queensideCastle || move.flag() == kingsideCastle {
		fromSq = pos.castling.kingsStartSquare[pos.Turn]
	}
	pos.AddPiece(pieceToAdd, squareReference[fromSq])
}

// updateRookPositionOnCaslte updates the rook square after making or unmaking a castle move on the position
func (pos *Position) updateRookPositionOnCaslte(castle int, makeMove bool) {
	castleType := castle - kingsideCastle
	rookFrom := pos.castling.rooksStartSquare[pos.Turn][castleType]
	rookTo := pos.castling.rooksEndSquare[pos.Turn][castleType]
	rookToMove := pieceColor(Rook, pos.Turn)
	if !makeMove {
		rookFrom = pos.castling.rooksEndSquare[pos.Turn.Opponent()][castleType]
		rookTo = pos.castling.rooksStartSquare[pos.Turn.Opponent()][castleType]
		rookToMove = pieceColor(Rook, pos.Turn.Opponent())
	}

	pos.RemovePiece(rookToMove, bitboardFromIndex(rookFrom))
	pos.AddPiece(rookToMove, squareReference[rookTo])
}

// makeNullMove performs a null move
func (pos *Position) makeNullMove() Bitboard {
	pos.toggleSide()
	if pos.enPassantTarget > 0 {
		pos.Hash = pos.Hash ^ zobristHashKeys.getEpKey(Bsf(pos.enPassantTarget))
	}
	ep := pos.enPassantTarget
	pos.enPassantTarget = 0 // NOTE: IMPORTANT! - If not done search goes inestable and outputs random moves eg. promotions....
	return ep
}

// unmakeNullMove restores the position after a null move
func (pos *Position) unmakeNullMove(ep Bitboard) {
	pos.toggleSide()
	if ep > 0 {
		pos.Hash = pos.Hash ^ zobristHashKeys.getEpKey(Bsf(ep))
	}
	pos.enPassantTarget = ep
}

// ToFen serializes the position as a fen string
func (pos *Position) ToFen() (fen string) {
	squares := toRuneArray(pos)

	for rank := 7; rank >= 0; rank-- {
		blankSquares := 0
		currentFenSqNumber := 8*(rank+1) - 8 // Fen string is read from a8 -> h8, a7 -> h7, ..., so i have to reverse the order

		for file := range 8 {
			piece := squares[currentFenSqNumber+file]
			if piece == rune(0) {
				blankSquares++
				continue
			}
			if blankSquares > 0 {
				fen += string(rune(blankSquares + 48))
			}
			fen += string(piece)
			blankSquares = 0
		}
		if blankSquares > 0 {
			fen += string(rune(blankSquares + 48))
		}
		if rank != 0 {
			fen += "/"
		}
	}

	if pos.Turn == White {
		fen += " w"
	} else {
		fen += " b"
	}

	fen += " " + pos.castling.castlingRights.toFen()
	if pos.enPassantTarget > 0 {
		fen += " " + squareReference[Bsf(pos.enPassantTarget)]
	} else {
		fen += " " + "-"
	}
	fen += " " + strconv.Itoa(pos.halfmoveClock)
	fen += " " + strconv.Itoa(pos.fullmoveNumber)
	return
}

// String prints the Position to the terminal from white's view perspective
func (pos *Position) String() string {
	position := ""
	board := toRuneArray(pos)

	currentSq := 63
	position += "\n    -------------------------------\n"
	for rank := 7; rank >= 0; rank-- {
		position += " " + strconv.Itoa(rank+1)
		for file := 7; file >= 0; file-- {
			piece := board[currentSq-file]
			if piece == rune(0) { // Default rune char
				piece = ' '
			}
			position += " | " + string(piece)
		}
		position += " |\n    -------------------------------\n"
		currentSq -= 8
	}
	position += "     a   b   c   d   e   f   g   h \n\n"
	position += "Fen: " + pos.ToFen() + "\n"
	position += "Key: " + strconv.FormatUint(pos.Hash, 10)
	return position
}

// toRuneArray returns an array of 64 runes with the position of the pieces
// in the board using Little endian rank-file mapping
func toRuneArray(pos *Position) [64]rune {
	squares := [64]rune{}
	pieceSymbol := [12]rune{'K', 'Q', 'R', 'B', 'N', 'P', 'k', 'q', 'r', 'b', 'n', 'p'}
	for pieceType, bitboard := range pos.Bitboards {
		for i := range len(squares) {
			if bitboard&(0b1<<i) > 0 {
				squares[i] = pieceSymbol[pieceType]
			}
		}
	}
	return squares
}

// Utility functions

// NewPositionFromFen creates a new Position struct from a fen string
func NewPositionFromFen(fen string) (pos *Position) {
	// TODO: validate fen string
	// maybe use a chain of responsibility while parsing each segment??
	pieceReference := map[string]int{
		"k": BlackKing, "q": BlackQueen, "r": BlackRook, "b": BlackBishop, "n": BlackKnight, "p": BlackPawn,
		"K": WhiteKing, "Q": WhiteQueen, "R": WhiteRook, "B": WhiteBishop, "N": WhiteKnight, "P": WhitePawn,
	}

	pos = EmptyPosition()
	elements := strings.Split(fen, " ")

	// NOTE: Order is reversed to match the square mapping in bitboards
	currentSquare := 56
	for rank := range strings.SplitSeq(elements[0], "/") {
		for piece := range strings.SplitSeq(rank, "") {
			switch piece {
			case "k", "q", "r", "b", "n", "p", "K", "Q", "R", "B", "N", "P":
				pos.Bitboards[pieceReference[piece]] |= (0b1 << currentSquare)
				pos.pieces[pieceReference[piece]/6] |= (0b1 << currentSquare)
				currentSquare++
			default:
				currentSquare += int(piece[0]) - 48 // Updates square number
			}
		}
		currentSquare -= 16
	}
	if elements[1][0] == 'w' {
		pos.Turn = White
	} else {
		pos.Turn = Black
	}

	pos.castling = *NewCastlingFromFen(fen, false)
	if elements[3] != "-" {
		pos.enPassantTarget = bitboardFromCoordinates(elements[3])
	}
	pos.halfmoveClock, _ = strconv.Atoi(elements[4]) // TODO: handle errors
	pos.fullmoveNumber, _ = strconv.Atoi(elements[5])

	pos.Hash = zobristHashKeys.fullZobristHash(pos)
	pos.PawnHash = zobristHashKeys.pawnHash(pos)
	return
}

// InitialPosition is a factory that returns an initial postion board
func InitialPosition() (pos *Position) {
	pos = NewPositionFromFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	return
}

// EmptyPosition returns an empty Position struct
func EmptyPosition() (pos *Position) {
	return &Position{
		positionHistory: *NewPositionHistory(),
		castling:        *NewCastling(4, 7, 0),
		evaluation:      *NewEvaluation(),
	}
}
