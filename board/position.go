package board

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

// TODO: extract to piece.go
// Color is an int referencing the white or black player in chess
type Color int

const (
	White Color = iota
	Black
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

// zobritsKeys contains the random keys to create the zobrist hash of a position
var zobristKeys = initZobristKeys()

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
	// Set all as empty
	// TODO: extract to constant
	emptySquares = 0xFFFFFFFFFFFFFFFF

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

func Attacks(piece Piece, sq Bitboard, pos *Position) (attacks Bitboard) {
	// TODO: Tests...
	switch piece {
	case WhiteKing, BlackKing:
		attacks |= kingAttacks(&sq, pos)
	case WhiteQueen, BlackQueen:
		attacks |= queenAttacks(&sq, pos)
	case WhiteRook, BlackRook:
		attacks |= rookAttacks(&sq, pos)
	case WhiteBishop, BlackBishop:
		attacks |= bishopAttacks(&sq, pos)
	case WhiteKnight, BlackKnight:
		attacks |= knightAttacks(&sq, pos)
	case WhitePawn:
		attacks |= pawnAttacks(&sq, pos, White)
	case BlackPawn:
		attacks |= pawnAttacks(&sq, pos, Black)
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
	indexBB := 0
	if side == Black {
		indexBB = 6
	}

	for pieceBB := indexBB; pieceBB-indexBB < 6; pieceBB++ {
		// TODO: also duplicated...
		pieces := pos.Bitboards[pieceBB]
		piece := Piece(pieceBB + indexBB)

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

	if (kingPos & pos.AttackedSquares(opponentSide(side))) > 0 {
		inCheck = true
	}
	return
}

// KingPosition returns the bitboard of the passed side king
func (pos *Position) KingPosition(side Color) (king Bitboard) {
	if side == White {
		king = pos.Bitboards[WhiteKing]
	} else {
		king = pos.Bitboards[BlackKing]
	}
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

// NumberOfPieces return the number of pieces of the type passed
func (pos *Position) NumberOfPieces(pieceType int) int {
	return pos.Bitboards[pieceType].count()
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
	var refactoredMoves []Move
	bitboards := pos.getBitboards(side)

	for piece, bb := range bitboards {
		for bb > 0 {
			pieceBB := bb.nextOne()
			switch piece {
			case 0:
				refactoredMoves = append(refactoredMoves, getKingMoves(&pieceBB, pos, side)...)
			case 1:
				refactoredMoves = append(refactoredMoves, getQueenMoves(&pieceBB, pos, side)...)
			case 2:
				refactoredMoves = append(refactoredMoves, getRookMoves(&pieceBB, pos, side)...)
			case 3:
				refactoredMoves = append(refactoredMoves, getBishopMoves(&pieceBB, pos, side)...)
			case 4:
				refactoredMoves = append(refactoredMoves, getKnightMoves(&pieceBB, pos, side)...)
			case 5:
				refactoredMoves = append(refactoredMoves, getPawnMoves(&pieceBB, pos, side)...)
			}
		}
	}

	// TODO: Move ordering. Captures are more likely to generate cutoff in the search
	// Need to find how to evaluate how good a move is to generate cutoff...
	sort.Slice(legalMoves, func(i, j int) bool {
		return legalMoves[i].moveType() > legalMoves[j].moveType()
	})
	return
}

// legalCastles returns the castles moves the passed side can make
func (pos *Position) legalCastles(side Color) (castle []Move) {
	castlePathsBB := []Bitboard{0x60, 0xE, 0x6000000000000000, 0xE00000000000000}
	passingKingPathBB := []Bitboard{0x60, 0xC, 0x6000000000000000, 0xC00000000000000}

	if pos.Check(side) {
		return
	}

	if side == White {
		king := WhiteKing
		canCastleShort := castlePathsBB[0]&(^pos.EmptySquares()|pos.AttackedSquares(opponentSide(side))) == 0
		canCastleLong := (castlePathsBB[1] & ^pos.EmptySquares())|(passingKingPathBB[1]&pos.AttackedSquares(opponentSide(side))) == 0

		if pos.castlingRights.canCastle(SHORT_CASTLE_WHITE) && canCastleShort {
			castle = append(castle, MoveEncode(4, 6, int(king), 0, CASTLE))
			// Move{from: "e1", to: "g1", piece: WHITE_KING, moveType: CASTLE})
		}
		if pos.castlingRights.canCastle(LONG_CASTLE_WHITE) && canCastleLong {
			castle = append(castle, MoveEncode(4, 2, int(king), 0, CASTLE))
			// Move{from: "e1", to: "c1", piece: WHITE_KING, moveType: CASTLE})
		}
	} else {
		king := BlackKing
		canCastleShort := castlePathsBB[2]&(^pos.EmptySquares()|pos.AttackedSquares(opponentSide(side))) == 0
		canCastleLong := (castlePathsBB[3] & ^pos.EmptySquares())|(passingKingPathBB[3]&pos.AttackedSquares(opponentSide(side))) == 0
		if pos.castlingRights.canCastle(SHORT_CASTLE_BLACK) && canCastleShort {
			castle = append(castle, MoveEncode(60, 62, int(king), 0, CASTLE))
			// Move{from: "e8", to: "g8", piece: BLACK_KING, moveType: CASTLE})
		}
		if pos.castlingRights.canCastle(SHORT_CASTLE_BLACK) && canCastleLong {
			castle = append(castle, MoveEncode(60, 58, int(king), 0, CASTLE))
			// Move{from: "e8", to: "c8", piece: BLACK_KING, moveType: CASTLE})
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
	piece := WhitePawn
	if side == Black {
		piece = BlackPawn
	}

	posiblesCapturers := posiblesEpCapturers(epTarget, side, pos)

	for posiblesCapturers > 0 {
		capturer := posiblesCapturers.nextOne()
		p, _ := pos.PieceAt(squareReference[Bsf(capturer)])

		enPassantAvailable := epTarget &
			pinRestrictedDirection(capturer, pieceColor[p], pos) &
			checkRestrictedMoves(capturer, pieceColor[p], pos)

		if enPassantAvailable > 0 {
			enPassant = append(enPassant, MoveEncode(Bsf(capturer), Bsf(epTarget), int(piece), 0, EN_PASSANT))
		}

	}

	// for _, sq := range posiblesCapturers.ToStringSlice() {
	// 	p, _ := pos.PieceAt(sq)
	//
	// 	enPassantAvailable := epTarget &
	// 		pinRestrictedDirection(p.Square(), p.Color(), pos) &
	// 		checkRestrictedMoves(p.Square(), p.Color(), pos)
	//
	// 	if enPassantAvailable > 0 {
	// 		enPassant = append(enPassant, MoveEncode(Bsf(p.Square()), Bsf(epTarget), piece, 0, EN_PASSANT))
	// 		// Move{from: p.Square().ToStringSlice()[0], to: pos.enPassantTarget.ToStringSlice()[0], piece: p.role(), moveType: EN_PASSANT})
	// 	}
	// }
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

	// clear EnPassant target if was setted
	newPos.enPassantTarget &= 0

	// fullmoveNumber+1 if it's black turn
	if pos.turn == Black {
		newPos.fullmoveNumber++
	}
	newPos.halfmoveClock++

	// Check for special moves:
	switch move.moveType() {
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
	newPos.turn = opponentSide(pos.turn)
	// Add piece to destination sqaure
	newPos.AddPiece(pieceToAdd, squareReference[move.to()])

	// Update zobrist hash for the position
	// TODO: refactor to a proper update function to avoid calculating the hash from scratch
	newPos.Zobrist = zobristHash(newPos, zobristKeys)
	return
}

func moveRookOnCastleMove(newPos Position, move *Move) Position {
	// Move the rook
	// Si el rey mueve hacia la derecha -> torre en h1/h8
	// Kingside castling
	// Si el rey mueve hacia la izuqierda -> torren en a1/a8
	// Kingside castling black
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
		fen += "w"
	} else {
		fen += "b"
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
	// TODO: add coordinates/unicode chars for pieces
	board := toRuneArray(pos)

	currentSq := 63
	fmt.Println("\n  -------------------------------") // Break line
	for rank := 7; rank >= 0; rank-- {
		for file := 7; file >= 0; file-- {
			piece := board[currentSq-file]
			if piece == rune(0) { // Default rune char
				piece = ' '
			}
			fmt.Print(" | " + string(piece))
		}
		fmt.Println(" |\n  -------------------------------") // Break line
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

// makePiece is a factory function that returns a Piece based on the role and square passed
// func makePiece(role int, square Bitboard) (piece Piece) {
// 	switch role {
// 	case WHITE_KING:
// 		piece = &King{color: WHITE, square: square}
// 	case WHITE_QUEEN:
// 		piece = &Queen{color: WHITE, square: square}
// 	case WHITE_ROOK:
// 		piece = &Rook{color: WHITE, square: square}
// 	case WHITE_BISHOP:
// 		piece = &Bishop{color: WHITE, square: square}
// 	case WHITE_KNIGHT:
// 		piece = &Knight{color: WHITE, square: square}
// 	case WHITE_PAWN:
// 		piece = &Pawn{color: WHITE, square: square}
// 	case BLACK_KING:
// 		piece = &King{color: BLACK, square: square}
// 	case BLACK_QUEEN:
// 		piece = &Queen{color: BLACK, square: square}
// 	case BLACK_ROOK:
// 		piece = &Rook{color: BLACK, square: square}
// 	case BLACK_BISHOP:
// 		piece = &Bishop{color: BLACK, square: square}
// 	case BLACK_KNIGHT:
// 		piece = &Knight{color: BLACK, square: square}
// 	case BLACK_PAWN:
// 		piece = &Pawn{color: BLACK, square: square}
// 	}
// 	return
// }

// squareToBitboard returns a bitboard containing the position of the squares coordinates passed
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
	// FIX: can be null square '-' and goes segmentation fault/panic!
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

// zorbistKeyIndex returns the index of the zorbistKey array for a given piece
// on a given square
func zobristKeyIndex(piece int, sq int) (i int) {
	return (64 * piece) + sq
}

// zorbistHash calculates the hash of a given position
func zobristHash(pos Position, keys [781]uint64) (hash uint64) {
	for pieceType, bb := range pos.Bitboards {
		for bb > 0 {
			sqNumber := Bsf(bb)
			hash = hash ^ keys[zobristKeyIndex(pieceType, sqNumber)]

			bb &= ^Bitboard(0b1 << sqNumber)
		}
	}
	// Black to move
	if pos.ToMove() == Black {
		hash = hash ^ keys[768]
	}

	// Castle availability
	hash = updateZobristCastle(pos, hash, keys)

	// Ep target
	if pos.enPassantTarget != 0 {
		hash = hash ^ keys[Bsf(pos.enPassantTarget)%8]
	}
	return
}

// updateZobristCastle updates a zobrist hash depending on the castle
// availability in the passed position
func updateZobristCastle(pos Position, hash uint64, keys [781]uint64) uint64 {
	if pos.castlingRights.canCastle(SHORT_CASTLE_WHITE) {
		hash = hash ^ keys[769]
	}
	if pos.castlingRights.canCastle(LONG_CASTLE_WHITE) {
		hash = hash ^ keys[770]
	}
	if pos.castlingRights.canCastle(SHORT_CASTLE_BLACK) {
		hash = hash ^ keys[771]
	}
	if pos.castlingRights.canCastle(LONG_CASTLE_BLACK) {
		hash = hash ^ keys[772]
	}
	return hash
}

// initZorbistKey returns an array of zorbist keys of random numbers
func initZobristKeys() (zorbistKeys [781]uint64) {
	//     One number for each piece at each square (12 x 64) pieceType (int) x square Number
	//     One number to indicate the side to move is black (1)  zorbistKey[768]
	//     Four numbers to indicate the castling rights, though usually 16 (2^4) are used for speed (4) zorbistKey[769 - 772]
	//     Eight numbers to indicate the file of a valid En passant square, if any (8) zorbistKey[773 - 780] only account the file
	//
	// This leaves us with an array with 781 (12*64 + 1 + 4 + 8) random numbers. Since pawns don't happen on first and eighth rank, one might be fine with 12*64 though. There are even proposals and implementations to use overlapping keys from unaligned access up to an array of only 12 numbers for every piece and to rotate that number by square [13] [14] .

	for i := range zorbistKeys {
		zorbistKeys[i] = uint64(rand.Int63())
	}

	return
}
