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
	// TODO, add all fields of a fen string to the position, so its easy to serialize to fen
	// bitboards piece order -> King, Queen, Rook, Bishop, Knight, Pawn (first white, second black)
	bitboards       [12]Bitboard
	turn            rune
	castlingRights  string
	enPassantTarget Bitboard // TODO, not sure to save as a bitboard or a coordinate string
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

	kingSq := pos.bitboards[WHITE_KING]
	if side != WHITE {
		kingSq = pos.bitboards[BLACK_KING]
	}
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
	king := WHITE_KING
	if side == BLACK {
		king = BLACK_KING
	}
	kingPos := pos.bitboards[king]

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
// This will be the input for the search function when generating the search tree!!!!
func (pos *Position) LegalMoves(side rune) (legalMoves []Move) {
	// TODO, not properly tested yet!!!
	// Need to check if this work as expected for any position!!!
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
	// Castling rules:
	// Neither the king nor the rook has previously moved. - Is checked in the pos.castlingRights
	// Need to check on the fly to see if its is legal
	// The king is not currently in check.
	// There are no pieces between the king and the rook.
	// The king does not pass through or finish on a square that is attacked by an enemy piece.

	// If its in the fen string, then its posible, so check if its legal
	// Castles KQkq -> KQ white, kq black
	// Need to check if not in check, or if the squares the king will pass over are not attacked
	// I can use a fixed bitboard for each square of the king and check vs attacks bitmap
	// Also check if the squares are not occupied by pieces
	// Array of castles squares [WHITE_SHORT, WHITE_LONG, black_short, black_long]
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
	if pos.enPassantTarget == 0 {
		return
	}

	if side == WHITE {
		southWeastPawn := pos.bitboards[WHITE_PAWN] & (pos.enPassantTarget & ^(pos.enPassantTarget & files[7]) >> 7)
		southEastPawn := pos.bitboards[WHITE_PAWN] & (pos.enPassantTarget & ^(pos.enPassantTarget & files[0]) >> 9)

		for _, sq := range (southEastPawn | southWeastPawn).ToStringSlice() {
			p, _ := pos.PieceAt(sq)
			legalEP := true

			// Validate pawn not pinned except direction is equal to capture direction
			if isPinned(p.Square(), WHITE, pos) {
				pinnedAlongEpCaptureDirection := (getDirection(pos.bitboards[WHITE_KING], p.Square()) != getDirection(pos.enPassantTarget, p.Square()))
				if !pinnedAlongEpCaptureDirection {
					legalEP = false
				}
			}
			// Validate if king is in check can block the attack
			// Get the checker if its 1 check if can block the check
			if pos.Check(WHITE) {
				checkingPieces := pos.CheckingPieces(WHITE)

				if len(checkingPieces) == 1 {
					checkerKingPath := getRayPath(checkingPieces[0].Square(), pos.bitboards[WHITE_KING])
					if checkerKingPath&pos.enPassantTarget == 0 {
						legalEP = false
					}
				} else {
					legalEP = false
				}
			}

			if legalEP {
				// Add the move
				enPassant = append(enPassant, Move{from: p.Square().ToStringSlice()[0], to: pos.enPassantTarget.ToStringSlice()[0], piece: p.role(), moveType: EN_PASSANT})
			}
		}
	} else {
		northWeastPawn := pos.bitboards[BLACK_PAWN] & (pos.enPassantTarget & ^(pos.enPassantTarget & files[7]) << 7)
		northEastPawn := pos.bitboards[BLACK_PAWN] & (pos.enPassantTarget & ^(pos.enPassantTarget & files[0]) << 9)

		for _, sq := range (northEastPawn | northWeastPawn).ToStringSlice() {
			p, _ := pos.PieceAt(sq)
			legalEP := true

			// Validate pawn not pinned except direction is equal to capture direction
			if isPinned(p.Square(), BLACK, pos) {
				pinnedAlongEpCaptureDirection := (getDirection(pos.bitboards[BLACK_KING], p.Square()) != getDirection(pos.enPassantTarget, p.Square()))
				if !pinnedAlongEpCaptureDirection {
					legalEP = false
				}
			}
			// Validate if king is in check can block the attack
			// Get the checker if its 1 check if can block the check
			if pos.Check(BLACK) {
				checkingPieces := pos.CheckingPieces(BLACK)

				if len(checkingPieces) == 1 {
					checkerKingPath := getRayPath(checkingPieces[0].Square(), pos.bitboards[BLACK_KING])
					if checkerKingPath&pos.enPassantTarget == 0 {
						legalEP = false
					}
				} else {
					legalEP = false
				}
			}

			if legalEP {
				// Add the move
				enPassant = append(enPassant, Move{from: p.Square().ToStringSlice()[0], to: pos.enPassantTarget.ToStringSlice()[0], piece: p.role(), moveType: EN_PASSANT})
			}
		}
	}
	return
}

// TODO update position
// MakeMove updates the position by making the move passed as parameter
func (pos *Position) MakeMove(move *Move) (newPosition *Position) {
	// TODO perform the move
	// TODO update position eg king has moved so no castle, ep squares leaved behind, etc
	return
}

// TODO return fen string from Position struct
// ToFen returns the fen string of the position (struct)
func (pos *Position) ToFen() (fen string) {
	return
}

// Print prints the Position to the terminal from white's view perspective
func (pos *Position) Print() {
	// TODO add coordinates/unicode chars for pieces
	board := [64]rune{}
	for i := 0; i < 64; i++ {
		board[i] = ' '
	}

	pieceSymbol := [12]rune{'K', 'Q', 'R', 'B', 'N', 'P', 'k', 'q', 'r', 'b', 'n', 'p'}
	for pieceType, bitboard := range pos.bitboards {
		for i := 0; i < len(board); i++ {
			if bitboard&(0b1<<i) > 0 {
				board[i] = pieceSymbol[pieceType]
			}
		}
	}

	currentSq := 63
	fmt.Println("\n  -------------------------------") // Break line
	for rank := 7; rank >= 0; rank-- {
		for file := 7; file >= 0; file-- {
			fmt.Print(" | " + string(board[currentSq-file]))
		}
		fmt.Println(" |\n  -------------------------------") // Break line
		currentSq -= 8
	}
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
	pos = &Position{turn: WHITE}
	return
}
