package board

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// TODO: extract to piece.go?
// Color is an int referencing the white or black player in chess
type Color int

const (
	White Color = iota
	Black
)

// opponent returns the opponent color to the actual color
func (c Color) opponent() Color {
	if c == White {
		return Black
	} else {
		return White
	}
}

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
	turn            Color
	castlingRights  castling
	enPassantTarget Bitboard
	halfmoveClock   int
	fullmoveNumber  int
	Zobrist         uint64
}

// TODO: Idea...
// PieceAt should return a bitboard with a piece or null bitboard...
// PieceType should return the type of the piece at the specified square...

// pieceAt returns a Piece at the given square coordinate in the Position or error
func (pos *Position) PieceAt(square string) (piece Piece, e error) {
	bitboardSquare := squareToBitboard([]string{square})

	for index, bitboard := range pos.Bitboards {
		if bitboard&bitboardSquare > 0 {
			piece = Piece(index)
			return
		}
	}

	e = errors.New("no piece")
	return
}

// addPiece adds to the position the Piece passed in the square passed
func (pos *Position) AddPiece(piece Piece, square string) {
	bitboardSquare := squareToBitboard([]string{square})
	pos.Bitboards[piece] |= bitboardSquare
}

// emptySquares returns a Bitboard with the empty sqaures of the position
func (pos *Position) EmptySquares() (emptySquares Bitboard) {
	emptySquares = ALL_SQUARES

	for _, bitboard := range pos.Bitboards {
		emptySquares &= ^bitboard
	}
	return
}

// attackedSquares returns a bitboard with all squares attacked by the passed side
func (pos *Position) AttackedSquares(side Color) (attackedSquares Bitboard) {
	startingBitboard := 0
	if side != White {
		startingBitboard = 6
	}

	for currentBitboard := startingBitboard; currentBitboard-startingBitboard < 6; currentBitboard++ {
		piece := Piece(currentBitboard)
		pieces := pos.Bitboards[currentBitboard]
		sq := pieces.nextOne()

		for sq > 0 {
			attackedSquares |= Attacks(piece, sq, pos)
			sq = pieces.nextOne()
		}

	}
	return
}

func Attacks(piece Piece, from Bitboard, pos *Position) (attacks Bitboard) {
	switch piece {
	case WhiteKing, BlackKing:
		attacks |= kingAttacks(&from, pos)
	case WhiteQueen, BlackQueen:
		attacks |= queenAttacks(&from, pos)
	case WhiteRook, BlackRook:
		attacks |= rookAttacks(&from, pos)
	case WhiteBishop, BlackBishop:
		attacks |= bishopAttacks(&from, pos)
	case WhiteKnight, BlackKnight:
		attacks |= knightAttacks(&from, pos)
	case WhitePawn:
		attacks |= pawnAttacks(&from, pos, White)
	case BlackPawn:
		attacks |= pawnAttacks(&from, pos, Black)
	}
	return
}

// checkingPieces returns an Bitboard of squares that contain pieces attacking the king of the side passed
func (pos *Position) CheckingPieces(side Color) (kingAttackers Bitboard) {
	if !pos.Check(side) {
		return
	}

	kingSq := pos.KingPosition(side)
	// iterate over all opponent bitboards and check who is attacking the king
	// TODO: refactor... duplicated in pos.Pieces method
	// FIX: need to get the opponent pieces bitboards...
	indexBB := 0
	if side == White {
		indexBB = 6
	}

	for pieceBB := indexBB; pieceBB-indexBB < 6; pieceBB++ {
		// TODO: also duplicated...
		pieces := pos.Bitboards[pieceBB]
		piece := Piece(pieceBB)

		for pieces > 0 {
			sq := pieces.nextOne()

			if kingSq&Attacks(piece, sq, pos) > 0 {
				kingAttackers |= sq
			}
		}
	}
	return
}

// Pieces returns a Bitboard with the pieces of the color pased
func (pos *Position) Pieces(side Color) (pieces Bitboard) {
	// TODO: Also duplicated, triplicated, etc...
	startingBitboard := 0
	if side != White {
		startingBitboard = 6
	}
	for currentBitboard := startingBitboard; currentBitboard-startingBitboard < 6; currentBitboard++ {
		pieces |= pos.Bitboards[currentBitboard]
	}
	return
}

// Check returns if the side passed is in check
func (pos *Position) Check(side Color) (inCheck bool) {
	kingPos := pos.KingPosition(side)

	if (kingPos & pos.AttackedSquares(side.opponent())) > 0 {
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
func (pos Position) RemovePiece(piece Bitboard) Position {
	newPos := pos

	for role, bitboard := range newPos.Bitboards {
		if bitboard&piece > 0 {
			newPos.Bitboards[role] &= ^piece
		}
	}
	return newPos
}

// ToMove returns the side to move in the current position
func (pos *Position) ToMove() Color {
	return pos.turn
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
func (pos *Position) LegalMoves(side Color) (legalMoves []Move) {
	bitboards := pos.getBitboards(side)

	for piece, bb := range bitboards {
		for bb > 0 {
			pieceBB := bb.nextOne()
			switch piece {
			case 0:
				legalMoves = append(legalMoves, getKingMoves(&pieceBB, pos, side)...)
			case 1:
				legalMoves = append(legalMoves, getQueenMoves(&pieceBB, pos, side)...)
			case 2:
				legalMoves = append(legalMoves, getRookMoves(&pieceBB, pos, side)...)
			case 3:
				legalMoves = append(legalMoves, getBishopMoves(&pieceBB, pos, side)...)
			case 4:
				legalMoves = append(legalMoves, getKnightMoves(&pieceBB, pos, side)...)
			case 5:
				legalMoves = append(legalMoves, getPawnMoves(&pieceBB, pos, side)...)
			}
		}
	}
	legalMoves = append(legalMoves, pos.legalCastles(side)...)
	legalMoves = append(legalMoves, pos.legalEnPassant(side)...)

	return
}

// legalCastles returns the castles moves the passed side can make
func (pos *Position) legalCastles(side Color) (castle []Move) {
	if pos.Check(side) {
		return
	}
	king := pieceOfColor[King][side]

	castlePathsBB := []Bitboard{0x60, 0xE, 0x6000000000000000, 0xE00000000000000}
	passingKingPathBB := []Bitboard{0x60, 0xC, 0x6000000000000000, 0xC00000000000000}

	move := newMove().
		setRule50Before(pos.halfmoveClock).
		setCastleRightsBefore(pos.castlingRights).
		setEpTargetBefore(pos.enPassantTarget).
		setPiece(king)

	if side == White {
		canCastleShort := castlePathsBB[0]&(^pos.EmptySquares()|pos.AttackedSquares(side.opponent())) == 0
		canCastleLong := (castlePathsBB[1] & ^pos.EmptySquares())|(passingKingPathBB[1]&pos.AttackedSquares(side.opponent())) == 0

		if pos.castlingRights.canCastle(SHORT_CASTLE_WHITE) && canCastleShort {
			castle = append(castle, *move.setFromSq(4).setToSq(6).setMoveType(CASTLE))
		}
		if pos.castlingRights.canCastle(LONG_CASTLE_WHITE) && canCastleLong {
			castle = append(castle, *move.setFromSq(4).setToSq(2).setMoveType(CASTLE))
		}
	} else {
		canCastleShort := castlePathsBB[2]&(^pos.EmptySquares()|pos.AttackedSquares(side.opponent())) == 0
		canCastleLong := (castlePathsBB[3] & ^pos.EmptySquares())|(passingKingPathBB[3]&pos.AttackedSquares(side.opponent())) == 0

		if pos.castlingRights.canCastle(SHORT_CASTLE_BLACK) && canCastleShort {
			castle = append(castle, *move.setFromSq(60).setToSq(62).setMoveType(CASTLE))
		}
		if pos.castlingRights.canCastle(SHORT_CASTLE_BLACK) && canCastleLong {
			castle = append(castle, *move.setFromSq(60).setToSq(58).setMoveType(CASTLE))
		}
	}
	return
}

// legalEnPassant returns the en passant moves in the position
func (pos *Position) legalEnPassant(side Color) (enPassant []Move) {
	epTarget := pos.enPassantTarget

	if epTarget == 0 {
		return
	}
	piece := pieceOfColor[Pawn][side]
	posiblesCapturers := posiblesEpCapturers(epTarget, side, pos)
	move := newMove().
		setRule50Before(pos.halfmoveClock).
		setCastleRightsBefore(pos.castlingRights).
		setEpTargetBefore(pos.enPassantTarget).
		setPiece(piece)

	for posiblesCapturers > 0 {
		capturer := posiblesCapturers.nextOne()
		p, _ := pos.PieceAt(squareReference[Bsf(capturer)])

		enPassantAvailable := epTarget &
			pinRestrictedDirection(capturer, pieceColor[p], pos) &
			checkRestrictedMoves(capturer, pieceColor[p], pos)

		if enPassantAvailable > 0 {
			move.setFromSq(Bsf(capturer)).
				setToSq(Bsf(epTarget)).
				setMoveType(EN_PASSANT)

			enPassant = append(enPassant, *move)
		}
	}
	return
}

// Endgame criteria
// Both sides have no queens or
// TODO: Every side which has a queen has additionally no other pieces or one minorpiece maximum.
func (pos *Position) IsEndgame() bool {
	// Both sides no queen
	if (pos.Bitboards[WhiteQueen] | pos.Bitboards[BlackQueen]) > 0 {
		return false
	}
	return true
}

// posiblesEpCapturers returns a bitboard with the pawns that are attacking
// the en passant target square passed
func posiblesEpCapturers(target Bitboard, side Color, pos *Position) (squares Bitboard) {
	if side == White {
		squares = (target & ^(target & files[7]) >> 7) |
			(target & ^(target & files[0]) >> 9)
		squares &= pos.Bitboards[WhitePawn]
	} else {
		squares |= (target & ^(target & files[7]) << 7) |
			(target & ^(target & files[0]) << 9)
		squares &= pos.Bitboards[BlackPawn]
	}
	return
}

// MakeMove updates the position by making the move passed as parameter
func (pos *Position) MakeMove(move *Move) (newPos Position) {
	pieceToAdd := Piece(move.piece())
	newPos = pos.RemovePiece(BitboardFromIndex(move.from()))

	// clear ep target
	newPos.enPassantTarget &= 0

	// fullmoveNumber+1 if it's black turn
	if pos.turn == Black {
		newPos.fullmoveNumber++
	}
	newPos.halfmoveClock++

	// Check for special moves:
	switch move.MoveType() {
	case NORMAL:
		// Reset halfmoveClock also resets with single pawn push
		if move.piece() == int(WhitePawn) || move.piece() == int(BlackPawn) {
			newPos.halfmoveClock = 0
		}
		updateCastleRights(&newPos, move)
	case PAWN_DOUBLE_PUSH:
		// Add ep target sq
		if move.piece() == int(WhitePawn) {
			newPos.enPassantTarget = BitboardFromIndex(move.to()) >> 8
		} else {
			newPos.enPassantTarget = BitboardFromIndex(move.to()) << 8
		}
		newPos.halfmoveClock = 0
	case CAPTURE:
		// Remove piece at destination
		newPos = newPos.RemovePiece(BitboardFromIndex(move.to()))
		// Reset halfmoveClock
		newPos.halfmoveClock = 0
		updateCastleRights(&newPos, move)
	case PROMOTION:
		// change piece to add to the board
		pieceToAdd = Piece(move.promotedTo())
		// Reset halfmoveClock
		newPos.halfmoveClock = 0
	case CASTLE:
		// TODO: i need to figure out another way of implement this
		newPos = moveRookOnCastleMove(newPos, move)
		// Update castle rights
		updateCastleRights(&newPos, move)
	case EN_PASSANT:
		// Remove the captured pawn
		// Es hacia abajo(blancas) / arriba (negras) del epTarget
		if move.piece() == int(WhitePawn) {
			newPos = newPos.RemovePiece(pos.enPassantTarget >> 8)
		} else {
			newPos = newPos.RemovePiece(pos.enPassantTarget << 8)
		}
		// Reset halfmoveClock
		newPos.halfmoveClock = 0
	}

	// Update Turn/color to move
	newPos.turn = pos.turn.opponent()
	// Add piece to destination sqaure
	newPos.AddPiece(pieceToAdd, squareReference[move.to()])

	// Update zobrist hash for the position
	// TODO: refactor to a proper update function to avoid calculating the hash from scratch
	newPos.Zobrist = zobristHash(newPos, zobristKeys)
	return
}

// UnmakeMove undoes a the passed move in the postion
func (pos *Position) UnmakeMove(move Move) (oldPos Position) {
	// pos is the position with the move make
	// oldPos will be the old position before the move
	oldPos = pos.RemovePiece(BitboardFromIndex(move.to()))
	// fullmoveNumber+1 if it's black turn

	if pos.turn == White {
		oldPos.fullmoveNumber--
	}
	oldPos.halfmoveClock--

	// restore to previous ep target
	// oldPos.enPassantTarget = Bitboard(0)
	// if move.oldEpTarget() > 0 {
	// 	oldPos.enPassantTarget = BitboardFromIndex(move.oldEpTarget())
	// }

	// pos.fullmoveNumber
	// NOTE: not needed to update because it does not modifies zorbist hash fullmoveNumber

	switch move.MoveType() {
	// case NORMAL: noting special needed
	// case PAWN_DOUBLE_PUSH: ep target managed by move.oldEpTarget()
	case CAPTURE:
		// oldPos.AddPiece(Piece(move.capturedPiece()), squareReference[move.to()])
		// restore updateCastleRights
	case PROMOTION:
	// remove promoted piece
	// add pawn to from
	// check castle if capture promotion(?)
	case CASTLE:
	// move back the rook
	case EN_PASSANT:
		// add captured pawn to original square
	}

	oldPos.AddPiece(Piece(move.piece()), squareReference[move.from()])

	// Update Turn/color to move to opponent
	oldPos.turn = pos.turn.opponent()

	// Update zobrist hash for the position
	oldPos.Zobrist = zobristHash(oldPos, zobristKeys)
	return
}

func moveRookOnCastleMove(newPos Position, move *Move) Position {
	if move.piece() == int(WhiteKing) {
		// TODO: map integer number to square string....
		if move.to() == Bsf(squareToBitboard([]string{"g1"})) {
			newPos = newPos.RemovePiece(squareToBitboard([]string{"h1"}))
			newPos.AddPiece(WhiteRook, "f1")
		} else {
			newPos = newPos.RemovePiece(squareToBitboard([]string{"a1"}))
			newPos.AddPiece(WhiteRook, "c1")
		}
	} else {

		if move.to() == Bsf(squareToBitboard([]string{"g8"})) {
			newPos = newPos.RemovePiece(squareToBitboard([]string{"h8"}))
			newPos.AddPiece(BlackRook, "f8")
		} else {
			newPos = newPos.RemovePiece(squareToBitboard([]string{"a8"}))
			newPos.AddPiece(BlackRook, "c8")
		}
	}
	return newPos
}

// updateCastleRights updates the castle rigths based on the move passed if the
// rook or the king has been moved
func updateCastleRights(pos *Position, move *Move) {
	piece := Piece(move.piece())
	switch piece {
	case WhiteKing:
		pos.castlingRights.remove(SHORT_CASTLE_WHITE | LONG_CASTLE_WHITE)
	case WhiteRook:
		if move.from() == Bsf(squareToBitboard([]string{"h1"})) {
			pos.castlingRights.remove(SHORT_CASTLE_WHITE)
		} else {
			pos.castlingRights.remove(LONG_CASTLE_WHITE)
		}
	case BlackKing:
		pos.castlingRights.remove(SHORT_CASTLE_BLACK | LONG_CASTLE_BLACK)
	case BlackRook:
		if move.from() == Bsf(squareToBitboard([]string{"h8"})) {
			pos.castlingRights.remove(SHORT_CASTLE_BLACK)
		} else {
			pos.castlingRights.remove(LONG_CASTLE_BLACK)
		}
	}
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
		// En todos menos el ultimo
		if rank != 0 {
			fen += "/"
		}
	}

	if pos.turn == White {
		fen += " w"
	} else {
		fen += " b"
	}

	fen += " " + pos.castlingRights.toFen()
	if pos.enPassantTarget > 0 {
		fen += " " + pos.enPassantTarget.ToStringSlice()[0]
	} else {
		fen += " " + "-"
	}
	fen += " " + strconv.Itoa(pos.halfmoveClock)
	fen += " " + strconv.Itoa(pos.fullmoveNumber)
	return
}

// Checkmate returns if the passed side is in checkmate on the current position
func (pos *Position) Checkmate(side Color) (checkmate bool) {
	if len(pos.LegalMoves(side)) == 0 && pos.Check(side) {
		checkmate = true
	} else {
		checkmate = false
	}
	return
}

// Stealmate returns if the passed side is in stealmate on the current position
func (pos *Position) Stealmate(side Color) (stealmate bool) {
	// Cannot be in check, and cannot have any legal moves
	if len(pos.LegalMoves(side)) == 0 && !pos.Check(side) {
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

// drawAvailableBy50MoveRule returns whenever if possible to claim draw by the
// 50 move rule
func (pos *Position) drawAvailableBy50MoveRule() bool {
	return pos.halfmoveClock >= 50
}

// Print prints the Position to the terminal from white's view perspective
func (pos *Position) Print() {
	board := toRuneArray(pos)

	currentSq := 63
	fmt.Println("\n  -------------------------------")
	for rank := 7; rank >= 0; rank-- {
		for file := 7; file >= 0; file-- {
			piece := board[currentSq-file]
			if piece == rune(0) { // Default rune char
				piece = ' '
			}
			fmt.Print(" | " + string(piece))
		}
		fmt.Println(" |\n  -------------------------------")
		currentSq -= 8
	}
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

// squareToBitboard returns a bitboard containing the position of the squares coordinates passed
// TODO: Refactor/improve
func squareToBitboard(coordinates []string) (bitboard Bitboard) {
	for _, coordinate := range coordinates {
		fileNumber := int(coordinate[0]) - 96
		rankNumber := int(coordinate[1]) - 48
		squareNumber := (fileNumber - 1) + 8*(rankNumber-1)

		// displaces 1 bit to the coordinate passed
		bitboard |= 0b1 << squareNumber
	}
	return
}

// From creates a new Position struct from a fen string
func From(fen string) (pos *Position) {
	// FIX this does not validate the fen string at all!!!!!!
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
		pos.turn = White
	} else {
		pos.turn = Black
	}

	pos.castlingRights.fromFen(elements[2]) // Fen string not implies its a legal move. Only says its available
	if elements[3] != "-" {
		pos.enPassantTarget = squareToBitboard([]string{elements[3]})
	}
	pos.halfmoveClock, _ = strconv.Atoi(elements[4]) // TODO handle errors
	pos.fullmoveNumber, _ = strconv.Atoi(elements[5])

	// Set zobrist
	pos.Zobrist = zobristHash(*pos, zobristKeys)
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
