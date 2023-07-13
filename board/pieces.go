// Pieces movements in a chess board
package board

import (
	"math"
)

// Constants for orthogonal directions in the board
const (
	NORTH     uint64 = 0
	NORTHEAST uint64 = 1
	EAST      uint64 = 2
	SOUTHEAST uint64 = 3
	SOUTH     uint64 = 4
	SOUTHWEST uint64 = 5
	WEST      uint64 = 6
	NORTHWEST uint64 = 7
	INVALID   uint64 = 8
)

// rays contains all precalculated rays for a given square in all possible 8 directions
// useful with calculating attacks/moves on sliding pieces(Rook, Bishop, Queens)
// https://gekomad.github.io/Cinnamon/BitboardCalculator/?type=2
var raysAttacks [8][64]Bitboard = [8][64]Bitboard{
	NORTH: {0x101010101010100, 0x202020202020200, 0x404040404040400, 0x808080808080800,
		0x1010101010101000, 0x2020202020202000, 0x4040404040404000, 0x8080808080808000,
		0x101010101010000, 0x202020202020000, 0x404040404040000, 0x808080808080000,
		0x1010101010100000, 0x2020202020200000, 0x4040404040400000, 0x8080808080800000,
		0x101010101000000, 0x202020202000000, 0x404040404000000, 0x808080808000000,
		0x1010101010000000, 0x2020202020000000, 0x4040404040000000, 0x8080808080000000,
		0x101010100000000, 0x202020200000000, 0x404040400000000, 0x808080800000000,
		0x1010101000000000, 0x2020202000000000, 0x4040404000000000, 0x8080808000000000,
		0x101010000000000, 0x202020000000000, 0x404040000000000, 0x808080000000000,
		0x1010100000000000, 0x2020200000000000, 0x4040400000000000, 0x8080800000000000,
		0x101000000000000, 0x202000000000000, 0x404000000000000, 0x808000000000000,
		0x1010000000000000, 0x2020000000000000, 0x4040000000000000, 0x8080000000000000,
		0x100000000000000, 0x200000000000000, 0x400000000000000, 0x800000000000000,
		0x1000000000000000, 0x2000000000000000, 0x4000000000000000, 0x8000000000000000,
		0x000000000000000, 0x000000000000000, 0x000000000000000, 0x000000000000000,
		0x0000000000000000, 0x0000000000000000, 0x0000000000000000, 0x0000000000000000,
	},
	NORTHEAST: {0x8040201008040200, 0x80402010080400, 0x804020100800, 0x8040201000, 0x80402000, 0x804000, 0x8000, 0,
		0x4020100804020000, 0x8040201008040000, 0x80402010080000, 0x804020100000, 0x8040200000, 0x80400000, 0x800000, 0,
		0x2010080402000000, 0x4020100804000000, 0x8040201008000000, 0x80402010000000, 0x804020000000, 0x8040000000, 0x80000000, 0,
		0x1008040200000000, 0x2010080400000000, 0x4020100800000000, 0x8040201000000000, 0x80402000000000, 0x804000000000, 0x8000000000, 0,
		0x804020000000000, 0x1008040000000000, 0x2010080000000000, 0x4020100000000000, 0x8040200000000000, 0x80400000000000, 0x800000000000, 0,
		0x402000000000000, 0x804000000000000, 0x1008000000000000, 0x2010000000000000, 0x4020000000000000, 0x8040000000000000, 0x80000000000000, 0,
		0x200000000000000, 0x400000000000000, 0x800000000000000, 0x1000000000000000, 0x2000000000000000, 0x4000000000000000, 0x8000000000000000, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
	EAST: {0xfe, 0xfc, 0xf8, 0xf0, 0xe0, 0xc0, 0x80, 0,
		0xfe << 8, 0xfc << 8, 0xf8 << 8, 0xf0 << 8, 0xe0 << 8, 0xc0 << 8, 0x80 << 8, 0,
		0xfe << 16, 0xfc << 16, 0xf8 << 16, 0xf0 << 16, 0xe0 << 16, 0xc0 << 16, 0x80 << 16, 0,
		0xfe << 24, 0xfc << 24, 0xf8 << 24, 0xf0 << 24, 0xe0 << 24, 0xc0 << 24, 0x80 << 24, 0,
		0xfe << 32, 0xfc << 32, 0xf8 << 32, 0xf0 << 32, 0xe0 << 32, 0xc0 << 32, 0x80 << 32, 0,
		0xfe << 40, 0xfc << 40, 0xf8 << 40, 0xf0 << 40, 0xe0 << 40, 0xc0 << 40, 0x80 << 40, 0,
		0xfe << 48, 0xfc << 48, 0xf8 << 48, 0xf0 << 48, 0xe0 << 48, 0xc0 << 48, 0x80 << 48, 0,
		0xfe << 56, 0xfc << 56, 0xf8 << 56, 0xf0 << 56, 0xe0 << 56, 0xc0 << 56, 0x80 << 56, 0,
	},
	SOUTHEAST: {0, 0, 0, 0, 0, 0, 0, 0,
		0x2, 0x4, 0x8, 0x10, 0x20, 0x40, 0x80, 0,
		0x204, 0x408, 0x810, 0x1020, 0x2040, 0x4080, 0x8000, 0,
		0x20408, 0x40810, 0x81020, 0x102040, 0x204080, 0x408000, 0x800000, 0,
		0x2040810, 0x4081020, 0x8102040, 0x10204080, 0x20408000, 0x40800000, 0x80000000, 0,
		0x204081020, 0x408102040, 0x810204080, 0x1020408000, 0x2040800000, 0x4080000000, 0x8000000000, 0,
		0x20408102040, 0x40810204080, 0x81020408000, 0x102040800000, 0x204080000000, 0x408000000000, 0x800000000000, 0,
		0x2040810204080, 0x4081020408000, 0x8102040800000, 0x10204080000000, 0x20408000000000, 0x40800000000000, 0x80000000000000, 0,
	},
	SOUTH: {0, 0, 0, 0, 0, 0, 0, 0,
		0x1, 0x2, 0x4, 0x8, 0x10, 0x20, 0x40, 0x80,
		0x101, 0x202, 0x404, 0x808, 0x1010, 0x2020, 0x4040, 0x8080,
		0x10101, 0x20202, 0x40404, 0x80808, 0x101010, 0x202020, 0x404040, 0x808080,
		0x1010101, 0x2020202, 0x4040404, 0x8080808, 0x10101010, 0x20202020, 0x40404040, 0x80808080,
		0x101010101, 0x202020202, 0x404040404, 0x808080808, 0x1010101010, 0x2020202020, 0x4040404040, 0x8080808080,
		0x10101010101, 0x20202020202, 0x40404040404, 0x80808080808, 0x101010101010, 0x202020202020, 0x404040404040, 0x808080808080,
		0x1010101010101, 0x2020202020202, 0x4040404040404, 0x8080808080808, 0x10101010101010, 0x20202020202020, 0x40404040404040, 0x80808080808080,
	},
	SOUTHWEST: {0, 0, 0, 0, 0, 0, 0, 0,
		0, 0x1, 0x2, 0x4, 0x8, 0x10, 0x20, 0x40,
		0, 0x100, 0x201, 0x402, 0x804, 0x1008, 0x2010, 0x4020,
		0, 0x10000, 0x20100, 0x40201, 0x80402, 0x100804, 0x201008, 0x402010,
		0, 0x1000000, 0x2010000, 0x4020100, 0x8040201, 0x10080402, 0x20100804, 0x40201008,
		0, 0x100000000, 0x201000000, 0x402010000, 0x804020100, 0x1008040201, 0x2010080402, 0x4020100804,
		0, 0x10000000000, 0x20100000000, 0x40201000000, 0x80402010000, 0x100804020100, 0x201008040201, 0x402010080402,
		0, 0x1000000000000, 0x2010000000000, 0x4020100000000, 0x8040201000000, 0x10080402010000, 0x20100804020100, 0x40201008040201,
	},
	WEST: {0, 0x1, 0x3, 0x7, 0xf, 0x1f, 0x3f, 0x7f,
		0, 0x100, 0x300, 0x700, 0xf00, 0x1f00, 0x3f00, 0x7f00,
		0, 0x10000, 0x30000, 0x70000, 0xf0000, 0x1f0000, 0x3f0000, 0x7f0000,
		0, 0x1000000, 0x3000000, 0x7000000, 0xf000000, 0x1f000000, 0x3f000000, 0x7f000000,
		0, 0x100000000, 0x300000000, 0x700000000, 0xf00000000, 0x1f00000000, 0x3f00000000, 0x7f00000000,
		0, 0x10000000000, 0x30000000000, 0x70000000000, 0xf0000000000, 0x1f0000000000, 0x3f0000000000, 0x7f0000000000,
		0, 0x1000000000000, 0x3000000000000, 0x7000000000000, 0xf000000000000, 0x1f000000000000, 0x3f000000000000, 0x7f000000000000,
		0, 0x100000000000000, 0x300000000000000, 0x700000000000000, 0xf00000000000000, 0x1f00000000000000, 0x3f00000000000000, 0x7f00000000000000,
	},
	NORTHWEST: {0, 0x100, 0x10200, 0x1020400, 0x102040800, 0x10204081000, 0x1020408102000, 0x102040810204000,
		0, 0x10000, 0x1020000, 0x102040000, 0x10204080000, 0x1020408100000, 0x102040810200000, 0x204081020400000,
		0, 0x1000000, 0x102000000, 0x10204000000, 0x1020408000000, 0x102040810000000, 0x204081020000000, 0x408102040000000,
		0, 0x100000000, 0x10200000000, 0x1020400000000, 0x102040800000000, 0x204081000000000, 0x408102000000000, 0x810204000000000,
		0, 0x10000000000, 0x1020000000000, 0x102040000000000, 0x204080000000000, 0x408100000000000, 0x810200000000000, 0x1020400000000000,
		0, 0x1000000000000, 0x102000000000000, 0x204000000000000, 0x408000000000000, 0x810000000000000, 0x1020000000000000, 0x2040000000000000,
		0, 0x100000000000000, 0x200000000000000, 0x400000000000000, 0x800000000000000, 0x1000000000000000, 0x2000000000000000, 0x4000000000000000,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// isPinned returns if the passed piece is pinned in the passed position
func isPinned(piece Bitboard, side Color, pos *Position) bool {
	// TODO: quedo medio fea..., pero funciona bien
	kingBB := pos.KingPosition(side)
	// FIX: Es necesario porque en algunos test no hay rey en la posicion
	if kingBB == 0 {
		return false
	}

	pinDirection := getDirection(piece, kingBB) // Direction from king to posible pinned piece
	if pinDirection == INVALID {
		return false
	}
	// Ray from king to piece
	possibleAttackers := raysAttacks[pinDirection][Bsf(kingBB)]

	// Get the pieces along that line
	piecesInLine := possibleAttackers & ^pos.EmptySquares()
	if piecesInLine.count() <= 1 { // Not enought pieces for a pin
		return false
	}

	// Get the first 2 pieces along the direction
	switch pinDirection {
	case NORTH, EAST, NORTHEAST, NORTHWEST:
		firstBB := BitboardFromIndex(Bsf(piecesInLine))
		secondBB := BitboardFromIndex(Bsf(piecesInLine & ^firstBB))

		pieceOne, _ := pos.PieceAt(squareReference[Bsf(firstBB)])
		pieceTwo, _ := pos.PieceAt(squareReference[Bsf(secondBB)])

		withoutPinnedPiece := pos.RemovePiece(firstBB)

		if pieceColor[pieceOne] != pieceColor[pieceTwo] &&
			(Attacks(pieceTwo, secondBB, &withoutPinnedPiece)&kingBB) > 0 {
			return true
		}
		// if pieceOne.Color() != pieceTwo.Color() && (pieceTwo.Attacks(&withoutPinnedPiece)&kingBB) > 0 {
		// 	return true
		// }

	case SOUTH, WEST, SOUTHEAST, SOUTHWEST:
		firstBB := BitboardFromIndex(63 - Bsr(piecesInLine))
		secondBB := BitboardFromIndex(63 - Bsr(piecesInLine & ^firstBB))

		pieceOne, _ := pos.PieceAt(squareReference[Bsf(firstBB)])
		pieceTwo, _ := pos.PieceAt(squareReference[Bsf(secondBB)])

		withoutPinnedPiece := pos.RemovePiece(firstBB)

		if pieceColor[pieceOne] != pieceColor[pieceTwo] &&
			(Attacks(pieceTwo, secondBB, &withoutPinnedPiece)&kingBB) > 0 {
			return true
		}
	}
	return false
}

// opponentSide returns the opposite color of the passed
func opponentSide(color Color) Color {
	if color == White {
		return Black
	}
	return White
}

// getDirection returns the direction from piece2 towards piece1 (piece2 -> piece1)
func getDirection(piece1 Bitboard, piece2 Bitboard) (dir uint64) {
	//  Direction of piece2 towards piece1 (piece2 -> piece1)
	//  Based on ±File±Column difference
	//   ---------------------
	//   | +1-1 | -1+0 | +1+1 |
	//   ----------------------
	//   | -1+0 |  P2  | +0+1 |
	//   ----------------------
	//   | -1-1 | +1+0 | -1+1 |
	//   ----------------------
	fileDiff := (Bsf(piece1) % 8) - (Bsf(piece2) % 8)
	rankDiff := (Bsf(piece1) / 8) - (Bsf(piece2) / 8)
	absFileDiff := math.Abs(float64(fileDiff))
	absRankDiff := math.Abs(float64(rankDiff))

	switch {
	case fileDiff == 0 && rankDiff > 0:
		dir = NORTH
	case fileDiff == 0 && rankDiff < 0:
		dir = SOUTH
	case fileDiff > 0 && rankDiff == 0:
		dir = EAST
	case fileDiff < 0 && rankDiff == 0:
		dir = WEST
	case absFileDiff == absRankDiff && fileDiff < 0 && rankDiff < 0:
		dir = SOUTHWEST
	case absFileDiff == absRankDiff && fileDiff > 0 && rankDiff > 0:
		dir = NORTHEAST
	case absFileDiff == absRankDiff && fileDiff > 0 && rankDiff < 0:
		dir = SOUTHEAST
	case absFileDiff == absRankDiff && fileDiff < 0 && rankDiff > 0:
		dir = NORTHWEST
	default:
		dir = INVALID
	}
	return
}

// raysDirection returns the rays along the direction passed that intersects the
// piece in the square passed
func raysDirection(square Bitboard, direction uint64) Bitboard {
	rays := raysAttacks[direction][Bsf(square)] | square

	// Need to complement the opposite of the direction passed
	switch direction {
	case NORTH:
		rays |= raysAttacks[SOUTH][Bsf(square)]
	case NORTHEAST:
		rays |= raysAttacks[SOUTHWEST][Bsf(square)]
	case EAST:
		rays |= raysAttacks[WEST][Bsf(square)]
	case SOUTHEAST:
		rays |= raysAttacks[NORTHWEST][Bsf(square)]
	case SOUTH:
		rays |= raysAttacks[NORTH][Bsf(square)]
	case SOUTHWEST:
		rays |= raysAttacks[NORTHEAST][Bsf(square)]
	case WEST:
		rays |= raysAttacks[EAST][Bsf(square)]
	case NORTHWEST:
		rays |= raysAttacks[SOUTHEAST][Bsf(square)]
	}

	return rays
}

// getRayPath returns a Bitboard with the path between 2 bitboards pieces
// (not including the 2 pieces)
func getRayPath(from Bitboard, to Bitboard) (rayPath Bitboard) {
	fromDirection := getDirection(to, from)
	toDirection := getDirection(from, to)

	if fromDirection == INVALID || toDirection == INVALID {
		return
	}

	rayPath = (raysAttacks[fromDirection][Bsf(from)] & raysAttacks[toDirection][Bsf(to)])
	return
}

// pinRestrictedDirection returns a bitboard with the restricted direction of moves
func pinRestrictedDirection(piece Bitboard, side Color, pos *Position) (restrictedDirection Bitboard) {
	restrictedDirection = ALL_SQUARES // No initial restrictions
	kingBB := pos.KingPosition(side)

	if isPinned(piece, side, pos) {
		direction := getDirection(kingBB, piece)
		allowedMovesDirection := raysDirection(kingBB, direction)
		restrictedDirection = allowedMovesDirection
	}
	return
}

// checkRestrictedMoves returns a bitboard with the allowed squares to block
// the path or capture the checking piece if in check
func checkRestrictedMoves(piece Bitboard, side Color, pos *Position) (allowedSquares Bitboard) {
	// FIX: need to refactor witout using the piece struct...
	// Need to return a bitboard of the checking piece...
	// Need to find a way to detect the type of piece checking -> sliding or not...
	checkingPieces := pos.CheckingPieces(side)

	switch {
	case checkingPieces.count() == 0:
		// No restriction
		allowedSquares = ALL_SQUARES
	case checkingPieces.count() == 1:
		// Capture or block the path
		checker := checkingPieces.nextOne()
		piece, _ := pos.PieceAt(squareReference[Bsf(checker)])

		if isSliding(piece) {
			allowedSquares |= getRayPath(checker, pos.KingPosition(side))
		}
		allowedSquares |= checker
	}
	// If there are more than 2 CheckingPieces it cant move at all (default allowedSquares value = 0)
	return
}

// isSliding returns a the passed Piece is an sliding piece(Queen, Rook or Bishop)
func isSliding(piece Piece) bool {
	// TODO: refactor, use a map instead?
	if piece == WhiteQueen || piece == WhiteRook ||
		piece == WhiteBishop || piece == BlackQueen ||
		piece == BlackRook || piece == BlackBishop {
		return true
	}
	return false
}
