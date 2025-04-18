package aconcagua

import (
	"strconv"
	"strings"
)

// Piece is an int that references piece role/color for bitboards in position struct
type Piece int

const (
	WhiteKing Piece = iota
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
	// Constants for reference to piece type
	King
	Queen
	Rook
	Bishop
	Knight
	Pawn
)

// pieceColor references a Piece type to a Color type
var pieceColor = map[Piece]Color{
	WhiteKing:   White,
	WhiteQueen:  White,
	WhiteRook:   White,
	WhiteBishop: White,
	WhiteKnight: White,
	WhitePawn:   White,
	BlackKing:   Black,
	BlackQueen:  Black,
	BlackRook:   Black,
	BlackBishop: Black,
	BlackKnight: Black,
	BlackPawn:   Black,
}

// pieceOfColor returns a Piece of the color passed
var pieceOfColor = map[Piece]map[Color]Piece{
	King:   {White: WhiteKing, Black: BlackKing},
	Queen:  {White: WhiteQueen, Black: BlackQueen},
	Rook:   {White: WhiteRook, Black: BlackRook},
	Bishop: {White: WhiteBishop, Black: BlackBishop},
	Knight: {White: WhiteKnight, Black: BlackKnight},
	Pawn:   {White: WhitePawn, Black: BlackPawn},
}

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

// pieceReference is used for fen string conversion to internal Position struct
var pieceReference = map[string]Piece{
	"k": BlackKing,
	"q": BlackQueen,
	"r": BlackRook,
	"b": BlackBishop,
	"n": BlackKnight,
	"p": BlackPawn,
	"K": WhiteKing,
	"Q": WhiteQueen,
	"R": WhiteRook,
	"B": WhiteBishop,
	"N": WhiteKnight,
	"P": WhitePawn,
}

// Position contains all information about a chess position
type Position struct {
	// Bitboards piece order -> King, Queen, Rook, Bishop, Knight, Pawn (first white, second black)
	Bitboards       [12]Bitboard
	Turn            Color
	Hash            uint64
	castlingRights  castling
	enPassantTarget Bitboard
	halfmoveClock   int
	fullmoveNumber  int
	positionHistory PositionHistory
}

// pieceAt returns a Piece at the given square coordinate in the Position or error
func (pos *Position) PieceAt(square string) (piece Piece) {
	bitboardSquare := bitboardFromCoordinate(square)

	for index, bitboard := range pos.Bitboards {
		if bitboard&bitboardSquare > 0 {
			piece = Piece(index)
			return
		}
	}

	piece = NoPiece
	return
}

// addPiece adds to the position the Piece passed in the square passed
func (pos *Position) AddPiece(piece Piece, square string) {
	if piece == NoPiece {
		return
	}

	bitboardSquare := bitboardFromCoordinate(square)
	pos.Bitboards[piece] |= bitboardSquare
}

// emptySquares returns a Bitboard with the empty sqaures of the position
func (pos *Position) EmptySquares() (emptySquares Bitboard) {
	emptySquares = AllSquares

	for _, bitboard := range pos.Bitboards {
		emptySquares &= ^bitboard
	}
	return
}

// attackedSquares returns a bitboard with all squares attacked by the passed side
func (pos *Position) AttackedSquares(side Color) (attackedSquares Bitboard) {
	startingPiece := startingPieceNumber(side)

	for i, bitboard := range pos.getBitboards(side) {
		piece := Piece(startingPiece + i)
		sq := bitboard.NextBit()

		for sq > 0 {
			attackedSquares |= Attacks(piece, sq, pos)
			sq = bitboard.NextBit()
		}
	}

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

func Attacks(piece Piece, from Bitboard, pos *Position) (attacks Bitboard) {
	switch piece {
	case WhiteKing, BlackKing:
		attacks |= kingAttacks(&from)
	case WhiteQueen, BlackQueen:
		attacks |= queenAttacks(&from, pos)
	case WhiteRook, BlackRook:
		attacks |= rookAttacks(&from, pos)
	case WhiteBishop, BlackBishop:
		attacks |= bishopAttacks(&from, pos)
	case WhiteKnight, BlackKnight:
		attacks |= knightAttackSquares[Bsf(from)]
	case WhitePawn:
		attacks |= pawnAttacks(&from, White)
	case BlackPawn:
		attacks |= pawnAttacks(&from, Black)
	}
	return
}

// checkingPieces returns an Bitboard of squares that contain pieces attacking the king of the side passed
func (pos *Position) CheckingPieces(side Color) (kingAttackers Bitboard) {
	if !pos.Check(side) {
		return
	}

	kingSq := pos.KingPosition(side)

	startingPieceNumber := startingPieceNumber(side.Opponent())

	for i, bitboard := range pos.getBitboards(side.Opponent()) {
		piece := Piece(startingPieceNumber + i)

		for bitboard > 0 {
			sq := bitboard.NextBit()

			if kingSq&Attacks(piece, sq, pos) > 0 {
				kingAttackers |= sq
			}
		}

	}
	return
}

// Pieces returns a Bitboard with the pieces of the color pased
func (pos *Position) Pieces(side Color) (pieces Bitboard) {
	for _, bitboard := range pos.getBitboards(side) {
		pieces |= bitboard
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

// KingPosition returns the bitboard of the passed side king
func (pos *Position) KingPosition(side Color) (king Bitboard) {
	king = pos.Bitboards[pieceOfColor[King][side]]
	return
}

// Remove Piece returns a new position without the piece passed
func (pos *Position) RemovePiece(piece Bitboard) {
	for role, bitboard := range pos.Bitboards {
		if bitboard&piece > 0 {
			pos.Bitboards[role] &= ^piece
		}
	}
}

func (pos *Position) getBitboards(side Color) (bitboards []Bitboard) {
	if side == White {
		bitboards = pos.Bitboards[0:6]
	} else {
		bitboards = pos.Bitboards[6:12]
	}
	return
}

// LegalMoves returns an slice of Move of all legal moves for the side passed
func (pos *Position) LegalMoves() *moveList {
	bitboards := pos.getBitboards(pos.Turn)
	ml := newMoveList()

	for piece, bb := range bitboards {
		for bb > 0 {
			pieceBB := bb.NextBit()
			switch piece {
			case 0:
				genKingMoves(&pieceBB, pos, pos.Turn, ml)
			case 1:
				genQueenMoves(&pieceBB, pos, pos.Turn, ml)
			case 2:
				genRookMoves(&pieceBB, pos, pos.Turn, ml)
			case 3:
				genBishopMoves(&pieceBB, pos, pos.Turn, ml)
			case 4:
				genKnightMoves(&pieceBB, pos, pos.Turn, ml)
			case 5:
				genPawnMoves(&pieceBB, pos, pos.Turn, ml)
			}
		}
	}
	genEpPawnCaptures(pos, pos.Turn, ml)
	return ml
}

// MakeMove executes a chess move, updating the board state
func (pos *Position) MakeMove(move *Move) {
	pieceToMove := pos.PieceAt(squareReference[move.from()])
	pieceCaptured := pos.getCapturedPiece(move)

	positionBefore := encodePositionBefore(
		uint16(pieceToMove),
		uint16(pieceCaptured),
		uint16(Bsf(pos.enPassantTarget)),
		uint16(pos.halfmoveClock),
	)
	pos.positionHistory.add(positionBefore, pos.castlingRights)
	UpdateCastleRights(pos, move)

	pos.RemovePiece(bitboardFromIndex(move.from()))
	pos.updateMoveState()
	pos.handleSpecialMoveTypes(move, &pieceToMove)

	pos.Turn = pos.Turn.Opponent()
	pos.AddPiece(pieceToMove, squareReference[move.to()])
	pos.Hash = zobristHash(pos)
}

// updateMoveState resets and updates move-related state
func (pos *Position) updateMoveState() {
	pos.enPassantTarget = 0
	pos.halfmoveClock++

	if pos.Turn == Black {
		pos.fullmoveNumber++
	}
}

// handleSpecialMoveTypes processes different types of chess moves
func (pos *Position) handleSpecialMoveTypes(move *Move, pieceToMove *Piece) {
	switch flag := move.flag(); flag {
	case quiet:
		pos.handleQuietMove(*pieceToMove)
	case doublePawnPush:
		pos.handleDoublePawnPush(*move, *pieceToMove)
	case capture:
		pos.handleCapture(*move)
	case knightPromotion, bishopPromotion, rookPromotion, queenPromotion,
		knightCapturePromotion, bishopCapturePromotion, rookCapturePromotion, queenCapturePromotion:
		pos.handlePromotion(*move, flag, pieceToMove)
	case kingsideCastle, queensideCastle:
		moveRookOnCastleMove(pos, castleType[move.String()])
	case epCapture:
		pos.handleEnPassantCapture(*move, *pieceToMove)
	}
}

// handleQuietMove resets halfmove clock for pawn moves
func (pos *Position) handleQuietMove(pieceToMove Piece) {
	if pieceToMove == WhitePawn || pieceToMove == BlackPawn {
		pos.halfmoveClock = 0
	}
}

// handleDoublePawnPush sets en passant target square
func (pos *Position) handleDoublePawnPush(move Move, pieceToMove Piece) {
	if pieceToMove == WhitePawn {
		pos.enPassantTarget = bitboardFromIndex(move.to()) >> 8
	} else {
		pos.enPassantTarget = bitboardFromIndex(move.to()) << 8
	}
}

// handleCapture removes captured piece and resets halfmove clock
func (pos *Position) handleCapture(move Move) {
	pos.RemovePiece(bitboardFromIndex(move.to()))
	pos.halfmoveClock = 0
}

// handlePromotion processes pawn promotion
func (pos *Position) handlePromotion(move Move, flag int, pieceToMove *Piece) {
	pos.RemovePiece(bitboardFromIndex(move.to()))
	*pieceToMove = getPromotedToPiece(flag, pos.Turn)

	pos.halfmoveClock = 0
}

// handleEnPassantCapture removes the captured pawn in en passant
func (pos *Position) handleEnPassantCapture(move Move, pieceToMove Piece) {
	if pieceToMove == WhitePawn {
		pos.Bitboards[BlackPawn] ^= bitboardFromIndex(move.to()) >> 8
	} else {
		pos.Bitboards[WhitePawn] ^= bitboardFromIndex(move.to()) << 8
	}
	pos.halfmoveClock = 0
}

// getPromotedToPiece returns the piece promoted to based on the flag passed
func getPromotedToPiece(flag int, side Color) (piece Piece) {
	switch flag {
	case knightPromotion, knightCapturePromotion:
		return pieceOfColor[Knight][side]
	case bishopPromotion, bishopCapturePromotion:
		return pieceOfColor[Bishop][side]
	case rookPromotion, rookCapturePromotion:
		return pieceOfColor[Rook][side]
	case queenCapturePromotion, queenPromotion:
		return pieceOfColor[Queen][side]
	}
	return NoPiece
}

// getCapturedPiece determines the piece captured in the move
func (pos *Position) getCapturedPiece(move *Move) Piece {
	if move.flag() == epCapture {
		return pieceOfColor[Pawn][pos.Turn.Opponent()]
	}
	pieceCaptured := pos.PieceAt(squareReference[move.to()])
	return pieceCaptured
}

// UnmakeMove undoes a the passed move in the postion
func (pos *Position) UnmakeMove(move *Move) {
	prevState, castle := pos.positionHistory.pop()
	pos.RemovePiece(bitboardFromIndex(move.to()))

	if pos.Turn == White {
		pos.fullmoveNumber--
	}

	if prevState.epTarget() > 0 {
		pos.enPassantTarget = bitboardFromIndex(prevState.epTarget())
	}
	pos.halfmoveClock = prevState.rule50()
	pos.castlingRights = castle

	pieceToAdd := Piece(prevState.pieceMoved())

	switch flag := move.flag(); flag {
	case capture, knightCapturePromotion, bishopCapturePromotion, rookCapturePromotion, queenCapturePromotion:
		pos.AddPiece(Piece(prevState.pieceCaptured()), squareReference[move.to()])
	case queensideCastle, kingsideCastle:
		restoreRookOnCastle(pos, castleType[move.String()])
	case epCapture:
		restoreSq := move.to() + 8           // 1 rank up
		if pieceColor[pieceToAdd] == White { // White ep capture
			restoreSq = move.to() - 8 // 1 rank down
		}
		pos.AddPiece(pieceOfColor[Pawn][pos.Turn], squareReference[restoreSq])
	case doublePawnPush:
		pos.enPassantTarget = 0
	}

	pos.Turn = pos.Turn.Opponent()
	pos.AddPiece(pieceToAdd, squareReference[move.from()])
	pos.Hash = zobristHash(pos)
	return
}

// moveRook updates the rook position for the castle passed
func moveRookOnCastleMove(pos *Position, castle castling) {
	rookFrom := rookOrigin[castle]
	rookTo := rookDestination[castle]

	pos.RemovePiece(bitboardFromIndex(rookFrom))
	pos.AddPiece(castleRook[castle], squareReference[rookTo])
}

// restoreRookOnCastle restores the rook when undoing a castle
func restoreRookOnCastle(pos *Position, castle castling) {
	rookTo := rookOrigin[castle]
	rookFrom := rookDestination[castle]

	pos.RemovePiece(bitboardFromIndex(rookFrom))
	pos.AddPiece(castleRook[castle], squareReference[rookTo])
}

func UpdateCastleRights(pos *Position, move *Move) {
	piece := pos.PieceAt(squareReference[move.from()])
	switch piece {
	case WhiteKing:
		pos.castlingRights.remove(K | Q)
	case WhiteRook:
		if move.from() == Bsf(bitboardFromCoordinate("h1")) {
			pos.castlingRights.remove(K)
		} else {
			pos.castlingRights.remove(Q)
		}
	case BlackKing:
		pos.castlingRights.remove(k | q)
	case BlackRook:
		if move.from() == Bsf(bitboardFromCoordinate("h8")) {
			pos.castlingRights.remove(k)
		} else {
			pos.castlingRights.remove(q)
		}
	}

	capturedPiece := pos.PieceAt(squareReference[move.to()])
	if move.flag() == capture || move.flag() > 5 {
		switch {
		case move.to() == 0 && capturedPiece == WhiteRook:
			pos.castlingRights.remove(Q)
		case move.to() == 7 && capturedPiece == WhiteRook:
			pos.castlingRights.remove(K)
		case move.to() == 56 && capturedPiece == BlackRook:
			pos.castlingRights.remove(q)
		case move.to() == 63 && capturedPiece == BlackRook:
			pos.castlingRights.remove(k)
		}
	}

}

// makeNullMove performs a null move
func (pos *Position) makeNullMove() Bitboard {
	pos.Turn = pos.Turn.Opponent()
	ep := pos.enPassantTarget
	pos.enPassantTarget = 0 // NOTE: IMPORTANT!!!! - If not done search goes inestable and outputs random moves eg. promotions....
	pos.Hash = zobristHash(pos)
	return ep
}

// unmakeNullMove restores the position after a null move
func (pos *Position) unmakeNullMove(ep Bitboard) {
	pos.Turn = pos.Turn.Opponent()
	pos.enPassantTarget = ep
	pos.Hash = zobristHash(pos)
}

// ToFen serializes the position as a fen string
func (pos *Position) ToFen() (fen string) {
	squares := toRuneArray(pos)

	for rank := 7; rank >= 0; rank-- {
		blankSquares := 0
		currentFenSqNumber := 8*(rank+1) - 8 // Fen string is read from a8 -> h8, a7 -> h7, ..., so i have to reverse the order

		for file := 0; file < 8; file++ {
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

	fen += " " + pos.castlingRights.toFen()
	if pos.enPassantTarget > 0 {
		fen += " " + squareReference[Bsf(pos.enPassantTarget)]
	} else {
		fen += " " + "-"
	}
	fen += " " + strconv.Itoa(pos.halfmoveClock)
	fen += " " + strconv.Itoa(pos.fullmoveNumber)
	return
}

// Checkmate returns if the passed side is in checkmate on the current position
func (pos *Position) Checkmate(side Color) (checkmate bool) {
	if pos.LegalMoves().length == 0 && pos.Check(side) {
		checkmate = true
	} else {
		checkmate = false
	}
	return
}

// Stealmate returns if the passed side is in stealmate on the current position
func (pos *Position) Stealmate(side Color) (stealmate bool) {
	// Cannot be in check, and cannot have any legal moves
	if pos.LegalMoves().length == 0 && !pos.Check(side) {
		stealmate = true
	} else {
		stealmate = false
	}
	return
}

func (pos *Position) InsuficientMaterial() bool {
	// Insuficient material on each side (according to FIDE rules):
	// - lone king
	// - king and bishop
	// - king and knight
	insuficientMaterialWhite := pos.Pieces(White) == pos.KingPosition(White) ||
		onlyKingAndBishop(pos, White) ||
		onlyKingAndKnight(pos, White)
	insuficientMaterialBlack := pos.Pieces(Black) == pos.KingPosition(Black) ||
		onlyKingAndBishop(pos, Black) ||
		onlyKingAndKnight(pos, Black)

	return insuficientMaterialWhite && insuficientMaterialBlack
}

// onlyKingAndKnight returns if in the passed position there is only a king piece
// and a knight piece for the side passed
func onlyKingAndKnight(pos *Position, side Color) bool {
	if pos.knights(side).count() > 1 {
		return false
	}
	return pos.Pieces(side) == (pos.knights(side) | pos.KingPosition(side))
}

// knights returns the bitboards with the knights of the side passed
func (pos *Position) knights(side Color) Bitboard {
	if side == White {
		return pos.Bitboards[WhiteKnight]
	} else {
		return pos.Bitboards[BlackKnight]
	}
}

// onlyKingAndBishop returns if in the passed position there is only a king piece
// and a bishop piece for the side passed
func onlyKingAndBishop(pos *Position, side Color) bool {
	if pos.bishops(side).count() > 1 {
		return false
	}
	return pos.Pieces(side) == (pos.bishops(side) | pos.KingPosition(side))
}

// bishops returns the bitboards with the bishops of the side passed
func (pos *Position) bishops(side Color) Bitboard {
	if side == White {
		return pos.Bitboards[WhiteBishop]
	} else {
		return pos.Bitboards[BlackBishop]
	}
}

// Print prints the Position to the terminal from white's view perspective
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
		for i := 0; i < len(squares); i++ {
			if bitboard&(0b1<<i) > 0 {
				squares[i] = pieceSymbol[pieceType]
			}
		}
	}
	return squares
}

// Utility functions

// From creates a new Position struct from a fen string
func From(fen string) (pos *Position) {
	// FIX: this does not validate the fen string at all!!!!!!
	pos = EmptyPosition()
	elements := strings.Split(fen, " ")

	// NOTE: Order is reversed to match the square mapping in bitboards
	currentSquare := 56
	for _, rank := range strings.Split(elements[0], "/") {
		for _, piece := range strings.Split(rank, "") {
			switch piece {
			case "k", "q", "r", "b", "n", "p", "K", "Q", "R", "B", "N", "P":
				pos.Bitboards[pieceReference[piece]] |= (0b1 << currentSquare)
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

	pos.castlingRights.fromFen(elements[2]) // Fen string not implies its a legal move. Only says its available
	if elements[3] != "-" {
		pos.enPassantTarget = bitboardFromCoordinate(elements[3])
	}
	pos.halfmoveClock, _ = strconv.Atoi(elements[4]) // TODO: handle errors
	pos.fullmoveNumber, _ = strconv.Atoi(elements[5])

	// Set zobrist
	pos.Hash = zobristHash(pos)
	return
}

// InitialPosition is a factory that returns an initial postion board
func InitialPosition() (pos *Position) {
	pos = From("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	return
}

// EmptyPosition returns an empty Position struct
func EmptyPosition() (pos *Position) {
	return &Position{}
}
