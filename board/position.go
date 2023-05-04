package board

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	WHITE rune = 'w'
	BLACK rune = 'b'
)

// References for pieces role/color for bitboards in position struct
const (
	WHITE_KING   int = 0
	WHITE_QUEEN  int = 1
	WHITE_ROOK   int = 2
	WHITE_BISHOP int = 3
	WHITE_KNIGHT int = 4
	WHITE_PAWN   int = 5
	BLACK_KING   int = 6
	BLACK_QUEEN  int = 7
	BLACK_ROOK   int = 8
	BLACK_BISHOP int = 9
	BLACK_KNIGHT int = 10
	BLACK_PAWN   int = 11
)

// Maps squares to uint64 (the index of the array is the bit position on a bitboard that represents that square)
var squareMap = []string{"a1", "b1", "c1", "d1", "e1", "f1", "g1", "h1",
	"a2", "b2", "c2", "d2", "e2", "f2", "g2", "h2",
	"a3", "b3", "c3", "d3", "e3", "f3", "g3", "h3",
	"a4", "b4", "c4", "d4", "e4", "f4", "g4", "h4",
	"a5", "b5", "c5", "d5", "e5", "f5", "g5", "h5",
	"a6", "b6", "c6", "d6", "e6", "f6", "g6", "h6",
	"a7", "b7", "c7", "d7", "e7", "f7", "g7", "h7",
	"a8", "b8", "c8", "d8", "e8", "f8", "g8", "h8"}

// Reference for fen string conversion to internal struct
var pieceReference = map[string]int{
	"k": BLACK_KING,
	"q": BLACK_QUEEN,
	"r": BLACK_ROOK,
	"b": BLACK_BISHOP,
	"n": BLACK_KNIGHT,
	"p": BLACK_PAWN,
	"K": WHITE_KING,
	"Q": WHITE_QUEEN,
	"R": WHITE_ROOK,
	"B": WHITE_BISHOP,
	"N": WHITE_KNIGHT,
	"P": WHITE_PAWN,
}

// Position contains all information about a chess position
type Position struct {
	// bitboards piece order -> King, Queen, Rook, Bishop, Knight, Pawn (first white, second black)
	bitboards       [12]Bitboard
	turn            rune
	castlingRights  string
	enPassantTarget Bitboard
	halfmoveClock   int
	fullmoveNumber  int
}

// pieceAt returns a Piece at the given square coordinate in the Position
func (pos *Position) PieceAt(square string) (piece Piece, e error) {
	bitboardSquare := squareToBitboard([]string{square})

	for role, bitboard := range pos.bitboards {
		if bitboard&bitboardSquare > 0 {
			piece = makePiece(role, bitboardSquare)
		}
	}
	if piece == nil {
		e = errors.New("No piece")
	}
	return
}

// addPiece adds a new Piece to the Position
func (pos *Position) AddPiece(role int, square string) {
	bitboardSquare := squareToBitboard([]string{square})
	pos.bitboards[role] |= bitboardSquare
}

// emptySquares returns a Bitboard with the empty sqaures of the position
func (pos *Position) EmptySquares() (emptySquares Bitboard) {
	// Set all as empty
	emptySquares = 0xFFFFFFFFFFFFFFFF

	for _, bitboard := range pos.bitboards {
		emptySquares &= ^bitboard
	}
	return
}

// attackedSquares returns a bitboard with all squares attacked by the passed side
func (pos *Position) AttackedSquares(side rune) (attackedSquares Bitboard) {
	startingBitboard := 0
	if side != WHITE {
		startingBitboard = 6
	}

	for currentBitboard := startingBitboard; currentBitboard-startingBitboard < 6; currentBitboard++ {
		for _, square := range pos.bitboards[currentBitboard].ToStringSlice() {
			piece, _ := pos.PieceAt(square)
			attackedSquares |= piece.Attacks(pos)
		}
	}

	return
}

// checkingPieces returns an slice of Piece{} that are checking the passed side king
func (pos *Position) CheckingPieces(side rune) (pieces []Piece) {
	if !pos.Check(side) {
		return
	}

	kingSq := pos.KingPosition(side)
	// iterate over all opponent pieces an add the ones that are attacking the king
	for _, sq := range pos.Pieces(opponentSide(side)).ToStringSlice() {
		piece, _ := pos.PieceAt(sq)

		if (kingSq & piece.Attacks(pos)) > 0 {
			pieces = append(pieces, piece)
		}
	}
	return
}

// pieces returns a Bitboard with the pieces of the color pased
func (pos *Position) Pieces(side rune) (pieces Bitboard) {
	startingBitboard := 0
	if side != WHITE {
		startingBitboard = 6
	}
	for currentBitboard := startingBitboard; currentBitboard-startingBitboard < 6; currentBitboard++ {
		pieces |= pos.bitboards[currentBitboard]
	}
	return
}

// check returns if the side passed is in check
func (pos *Position) Check(side rune) (inCheck bool) {
	kingPos := pos.KingPosition(side)

	if (kingPos & pos.AttackedSquares(opponentSide(side))) > 0 {
		inCheck = true
	}
	return
}

// KingPosition returns the bitboard of the passed side king
func (pos *Position) KingPosition(side rune) (king Bitboard) {
	if side == WHITE {
		king = pos.bitboards[WHITE_KING]
	} else {
		king = pos.bitboards[BLACK_KING]
	}
	return
}

// Remove Piece returns a new position without the piece passed
func (pos Position) RemovePiece(piece Bitboard) Position {
	newPos := pos

	for role, bitboard := range newPos.bitboards {
		if bitboard&piece > 0 {
			newPos.bitboards[role] &= ^piece
		}
	}
	return newPos
}

// LegalMoves returns an slice of Move of all legal moves for the side passed
func (pos *Position) LegalMoves(side rune) (legalMoves []Move) {
	opponentPieces := pos.Pieces(opponentSide(side))
	piecesSq := pos.Pieces(side).ToStringSlice()

	for _, from := range piecesSq {
		piece, _ := pos.PieceAt(from)
		destinations := piece.Moves(pos).ToStringSlice()

		// MOVES / CAPTURES / PROMOTIONS
		for _, to := range destinations {
			pieceBB := piece.Square()
			isWhitePawn := pieceBB&pos.bitboards[WHITE_PAWN] > 0
			isBlackPawn := pieceBB&pos.bitboards[BLACK_PAWN] > 0

			if opponentPieces&pieceBB > 0 {
				legalMoves = append(legalMoves, Move{from: from, to: to, piece: piece.role(), moveType: CAPTURE})
			} else if isWhitePawn && (from[1] == '7') {
				for _, promotedRole := range []int{WHITE_KNIGHT, WHITE_BISHOP, WHITE_ROOK, WHITE_QUEEN} {
					legalMoves = append(legalMoves, Move{from: from, to: to, piece: piece.role(), moveType: PROMOTION, promotedTo: promotedRole})
				}
			} else if isBlackPawn && (from[1] == '2') {
				for _, promotedRole := range []int{BLACK_KNIGHT, BLACK_BISHOP, BLACK_ROOK, BLACK_QUEEN} {
					legalMoves = append(legalMoves, Move{from: from, to: to, piece: piece.role(), moveType: PROMOTION, promotedTo: promotedRole})
				}
			} else {
				legalMoves = append(legalMoves, Move{from: from, to: to, piece: piece.role(), moveType: NORMAL})
			}
		}
	}
	// CASTLE
	legalMoves = append(legalMoves, pos.legalCastles(side)...)
	// EN PASSANT
	legalMoves = append(legalMoves, pos.legalEnPassant(side)...)
	return
}

// legalCastles returns the castles moves the passed side can make
func (pos *Position) legalCastles(side rune) (castle []Move) {
	castlePathsBB := []Bitboard{0x60, 0xE, 0x6000000000000000, 0xE00000000000000}
	passingKingPathBB := []Bitboard{0x60, 0xC, 0x6000000000000000, 0xC00000000000000}

	if pos.Check(side) {
		return
	}

	if side == WHITE {
		canCastleShort := castlePathsBB[0]&(^pos.EmptySquares()|pos.AttackedSquares(opponentSide(side))) == 0
		canCastleLong := (castlePathsBB[1] & ^pos.EmptySquares())|(passingKingPathBB[1]&pos.AttackedSquares(opponentSide(side))) == 0
		if strings.Contains(pos.castlingRights, "K") && canCastleShort {
			castle = append(castle, Move{from: "e1", to: "g1", piece: WHITE_KING, moveType: CASTLE})
		}
		if strings.Contains(pos.castlingRights, "Q") && canCastleLong {
			castle = append(castle, Move{from: "e1", to: "c1", piece: WHITE_KING, moveType: CASTLE})
		}
	} else {
		canCastleShort := castlePathsBB[2]&(^pos.EmptySquares()|pos.AttackedSquares(opponentSide(side))) == 0
		canCastleLong := (castlePathsBB[3] & ^pos.EmptySquares())|(passingKingPathBB[3]&pos.AttackedSquares(opponentSide(side))) == 0
		if strings.Contains(pos.castlingRights, "k") && canCastleShort {
			castle = append(castle, Move{from: "e8", to: "g8", piece: BLACK_KING, moveType: CASTLE})
		}
		if strings.Contains(pos.castlingRights, "q") && canCastleLong {
			castle = append(castle, Move{from: "e8", to: "c8", piece: BLACK_KING, moveType: CASTLE})
		}
	}
	return
}

// legalEnPassant returns if there are any legal en passant move for the side passed
func (pos *Position) legalEnPassant(side rune) (enPassant []Move) {
	epTarget := pos.enPassantTarget

	if epTarget == 0 {
		return
	}

	posiblesCapturers := posiblesEpCapturers(epTarget, side, pos)
	for _, sq := range posiblesCapturers.ToStringSlice() {
		p, _ := pos.PieceAt(sq)

		enPassantAvailable := epTarget &
			pinRestrictedDirection(p.Square(), p.Color(), pos) &
			checkRestrictedMoves(p.Square(), p.Color(), pos)

		if enPassantAvailable > 0 {
			enPassant = append(enPassant, Move{from: p.Square().ToStringSlice()[0], to: pos.enPassantTarget.ToStringSlice()[0], piece: p.role(), moveType: EN_PASSANT})
		}
	}
	return
}

// posiblesEpCapturers returns a bitboard with the pawns that are attacking
// the en passant target square passed
func posiblesEpCapturers(target Bitboard, side rune, pos *Position) (squares Bitboard) {
	if side == WHITE {
		squares = (target & ^(target & files[7]) >> 7) |
			(target & ^(target & files[0]) >> 9)
		squares &= pos.bitboards[WHITE_PAWN]
	} else {
		squares |= (target & ^(target & files[7]) << 7) |
			(target & ^(target & files[0]) << 9)
		squares &= pos.bitboards[BLACK_PAWN]
	}
	return
}

// MakeMove updates the position by making the move passed as parameter
func (pos *Position) MakeMove(move *Move) (newPos Position) {
	// TODO/FIX validate the move is legal?!!!!!!
	// TODO REFACTOR!!!
	// Remove piece from origin
	pieceToAdd := move.piece
	newPos = pos.RemovePiece(squareToBitboard([]string{move.from}))
	// clear EnPassant target if was setted
	newPos.enPassantTarget &= 0

	// fullmoveNumber+1 if it's black turn
	if pos.turn == BLACK {
		newPos.fullmoveNumber++
	}
	newPos.halfmoveClock++

	// Check for special moves:
	switch move.moveType {
	case NORMAL:
		// Reset halfmoveClock also resets with single pawn push
		if move.piece == WHITE_PAWN || move.piece == BLACK_PAWN {
			newPos.halfmoveClock = 0
		}
		updateCastleRights(&newPos, move)
	case PAWN_DOUBLE_PUSH:
		// TODO/FIX need to generate the dobule push type when moving a pawn in move generation function
		// Add ep target sq
		if move.piece == WHITE_PAWN {
			newPos.enPassantTarget = squareToBitboard([]string{move.to}) >> 8
		} else {
			newPos.enPassantTarget = squareToBitboard([]string{move.to}) << 8
		}
		newPos.halfmoveClock = 0
	case CAPTURE:
		// Remove piece at destination
		newPos = newPos.RemovePiece(squareToBitboard([]string{move.from}))
		// Reset halfmoveClock
		newPos.halfmoveClock = 0
		updateCastleRights(&newPos, move)
	case PROMOTION:
		// change piece to add to the board
		pieceToAdd = move.promotedTo
		// Reset halfmoveClock
		newPos.halfmoveClock = 0
	case CASTLE:
		// TODO i need to figure out another way of implement this
		newPos = moveRookOnCastleMove(newPos, move)
		// Update castle rights
		updateCastleRights(&newPos, move)
	case EN_PASSANT:
		// Remove the captured pawn
		// Es hacia abajo(blancas) / arriba (negras) del epTarget
		if move.piece == WHITE_PAWN {
			newPos = newPos.RemovePiece(pos.enPassantTarget >> 8)
		} else {
			newPos = newPos.RemovePiece(pos.enPassantTarget << 8)
		}
		// Reset halfmoveClock
		newPos.halfmoveClock = 0
	}

	// Update Turn/color to move
	newPos.turn = opponentSide(newPos.turn)
	// Add piece to destination sqaure
	newPos.AddPiece(pieceToAdd, move.to)
	return
}

func moveRookOnCastleMove(newPos Position, move *Move) Position {
	// Move the rook
	// Si el rey mueve hacia la derecha -> torre en h1/h8
	// Kingside castling
	// Si el rey mueve hacia la izuqierda -> torren en a1/a8
	// Kingside castling black
	if move.piece == WHITE_KING {
		if move.to == "g1" {
			newPos = newPos.RemovePiece(squareToBitboard([]string{"h1"}))
			newPos.AddPiece(WHITE_ROOK, "f1")
		} else {
			newPos = newPos.RemovePiece(squareToBitboard([]string{"a1"}))
			newPos.AddPiece(WHITE_ROOK, "c1")
		}
	} else {

		if move.to == "g8" {
			newPos = newPos.RemovePiece(squareToBitboard([]string{"h8"}))
			newPos.AddPiece(BLACK_ROOK, "f8")
		} else {
			newPos = newPos.RemovePiece(squareToBitboard([]string{"a8"}))
			newPos.AddPiece(BLACK_ROOK, "c8")
		}
	}
	return newPos
}

// updateCastleRights updates the castle rigths based on the move passed if the
// rook or the king has been moved
func updateCastleRights(pos *Position, move *Move) {
	currentRights := pos.castlingRights
	switch move.piece {
	case WHITE_KING:
		currentRights = strings.Replace(currentRights, "K", "", 1)
		currentRights = strings.Replace(currentRights, "Q", "", 1)
	case WHITE_ROOK:
		if move.from == "h1" {
			currentRights = strings.Replace(currentRights, "K", "", 1)
		} else {
			currentRights = strings.Replace(currentRights, "Q", "", 1)
		}
	case BLACK_KING:
		currentRights = strings.Replace(currentRights, "k", "", 1)
		currentRights = strings.Replace(currentRights, "q", "", 1)
	case BLACK_ROOK:
		if move.from == "h8" {
			currentRights = strings.Replace(currentRights, "k", "", 1)
		} else {
			currentRights = strings.Replace(currentRights, "q", "", 1)
		}
	}
	if currentRights == "" {
		currentRights = "-"
	}
	pos.castlingRights = currentRights
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
	fen += " " + string(pos.turn)
	fen += " " + pos.castlingRights
	if pos.enPassantTarget > 0 {
		fen += " " + pos.enPassantTarget.ToStringSlice()[0]
	} else {
		fen += " " + "-"
	}
	fen += " " + strconv.Itoa(pos.halfmoveClock)
	fen += " " + strconv.Itoa(pos.fullmoveNumber)
	return
}

// Print prints the Position to the terminal from white's view perspective
func (pos *Position) Print() {
	// TODO add coordinates/unicode chars for pieces
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
	for pieceType, bitboard := range pos.bitboards {
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
func makePiece(role int, square Bitboard) (piece Piece) {
	switch role {
	case WHITE_KING:
		piece = &King{color: WHITE, square: square}
	case WHITE_QUEEN:
		piece = &Queen{color: WHITE, square: square}
	case WHITE_ROOK:
		piece = &Rook{color: WHITE, square: square}
	case WHITE_BISHOP:
		piece = &Bishop{color: WHITE, square: square}
	case WHITE_KNIGHT:
		piece = &Knight{color: WHITE, square: square}
	case WHITE_PAWN:
		piece = &Pawn{color: WHITE, square: square}
	case BLACK_KING:
		piece = &King{color: BLACK, square: square}
	case BLACK_QUEEN:
		piece = &Queen{color: BLACK, square: square}
	case BLACK_ROOK:
		piece = &Rook{color: BLACK, square: square}
	case BLACK_BISHOP:
		piece = &Bishop{color: BLACK, square: square}
	case BLACK_KNIGHT:
		piece = &Knight{color: BLACK, square: square}
	case BLACK_PAWN:
		piece = &Pawn{color: BLACK, square: square}
	}
	return
}

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
				pos.bitboards[pieceReference[piece]] |= (0b1 << currentSquare)
				currentSquare++
			default:
				currentSquare += int(piece[0]) - 48 // Updates square number
			}
		}
		currentSquare -= 16
	}
	pos.turn = rune(elements[1][0])
	pos.castlingRights = elements[2] // Fen string not implies its a legal move. Only says its available
	// FIX can be null square '-' and goes segmentation fault/panic!
	if elements[3] != "-" {
		pos.enPassantTarget = squareToBitboard([]string{elements[3]})
	}
	pos.halfmoveClock, _ = strconv.Atoi(elements[4]) // TODO handle errors
	pos.fullmoveNumber, _ = strconv.Atoi(elements[5])

	return
}

// InitialPosition is a factory that returns an initial postion board
func InitialPosition() (pos *Position) {
	pos = From("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	return
}

// EmptyPosition returns an empty Position struct
func EmptyPosition() (pos *Position) {
	pos = &Position{}
	return
}
