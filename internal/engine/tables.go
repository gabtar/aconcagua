package engine

import "math"

const (
	// Constants for orthogonal directions in the board
	North = iota
	NorthEast
	East
	SouthEast
	South
	SouthWest
	West
	NorthWest
	Invalid
)

// init initializes various tables for usage within the engine
func init() {
	generatePiecesScoreTables()

	directions = generateDirections()
	RayAttacks = generateRayAttacks()
	knightAttacksTable = generateKnightAttacksTable()
	kingAttacksTable = generateKingAttacksTable()
	pawnPushesTable = generatePawnPushesTable()
	pawnDoublePushesTable = generatePawnDoublePushesTable()
	attacksFrontSpans = generateAttacksFrontSpans()
	KingZone = generateKingZone()
}

// Files array contains the bitboard mask for each file in the board
var Files [8]Bitboard = [8]Bitboard{
	0x0101010101010101,
	0x0101010101010101 << 1,
	0x0101010101010101 << 2,
	0x0101010101010101 << 3,
	0x0101010101010101 << 4,
	0x0101010101010101 << 5,
	0x0101010101010101 << 6,
	0x0101010101010101 << 7,
}

// Ranks contains the bitboard mask for each rank in the board
var Ranks [8]Bitboard = [8]Bitboard{
	0x00000000000000FF,
	0x00000000000000FF << 8,
	0x00000000000000FF << 16,
	0x00000000000000FF << 24,
	0x00000000000000FF << 32,
	0x00000000000000FF << 40,
	0x00000000000000FF << 48,
	0x00000000000000FF << 56,
}

// AllSquares contains the bitboard mask for all squares in the board
const AllSquares Bitboard = 0xFFFFFFFFFFFFFFFF

// directions is a table that contains the compass directions between 2 squares in the board
var directions [64][64]uint64

// RayAttacks is a precalculated table that contains the rays on each direction for each square
var RayAttacks [8][64]Bitboard

// knightAttacksTable is a precalculated table that contains the squares that a knight can attack
var knightAttacksTable [64]Bitboard

// kingAttacksTable is a precalculated table that contains the squares that a king can attack
var kingAttacksTable [64]Bitboard

// pawnPushesTable is a precalculated table that contains the squares that a pawn can push
var pawnPushesTable [2][64]Bitboard

// pawnDoublePushesTable is a precalculated table that contains the squares that a pawn can double push
var pawnDoublePushesTable [2][64]Bitboard

// isolatedAdjacentFilesMask contains the adjacent files of a pawn to test if it is isolated
var isolatedAdjacentFilesMask = [8]Bitboard{
	Files[1],
	Files[0] | Files[2],
	Files[1] | Files[3],
	Files[2] | Files[4],
	Files[3] | Files[5],
	Files[4] | Files[6],
	Files[5] | Files[7],
	Files[6],
}

// attacksFrontSpans is a precalculated table containing the bitmask of front attack spans for each square
// The mask includes the attacked squares itself, thus it is like a fill of attacked squares in the appropriate
// direction front attack span for pawn on d4
// see: https://www.chessprogramming.org/Attack_Spans
// . . 1 . 1 . . .
// . . 1 . 1 . . .
// . . 1 . 1 . . .
// . . 1 . 1 . . .
// . . . w . . . .
// . . . . . . . .
// . . . . . . . .
// . . . . . . . .
var attacksFrontSpans [2][64]Bitboard

var KingZone [2][64]Bitboard

// generatePiecesScoreTables generates the tables with the value of each piece + square
func generatePiecesScoreTables() {
	for piece := range 6 {
		whitePiece := piece
		blackPiece := piece + 6

		for sq := range 64 {
			middlegamePiecesScore[whitePiece][sq] = MiddlegamePieceValue[piece] + MiddlegamePSQT[piece][sq^56]
			endgamePiecesScore[whitePiece][sq] = EndgamePieceValue[piece] + EndgamePSQT[piece][sq^56]

			middlegamePiecesScore[blackPiece][sq] = MiddlegamePieceValue[piece] + MiddlegamePSQT[piece][sq]
			endgamePiecesScore[blackPiece][sq] = EndgamePieceValue[piece] + EndgamePSQT[piece][sq]
		}
	}
}

// generateDirections generates all posible directions between all squares in the board
func generateDirections() (directions [64][64]uint64) {
	for from := range 64 {
		for to := range 64 {
			//  Direction of 2 squares
			//  Based on ±File±Column difference
			//   ---------------------
			//   | +1-1 | -1+0 | +1+1 |
			//   ----------------------
			//   | -1+0 |  P2  | +0+1 |
			//   ----------------------
			//   | -1-1 | +1+0 | -1+1 |
			//   ----------------------
			fileDiff := (to % 8) - (from % 8)
			rankDiff := (to / 8) - (from / 8)
			absFileDiff := math.Abs(float64(fileDiff))
			absRankDiff := math.Abs(float64(rankDiff))

			switch {
			case fileDiff == 0 && rankDiff > 0:
				directions[from][to] = North
			case fileDiff == 0 && rankDiff < 0:
				directions[from][to] = South
			case fileDiff > 0 && rankDiff == 0:
				directions[from][to] = East
			case fileDiff < 0 && rankDiff == 0:
				directions[from][to] = West
			case absFileDiff == absRankDiff && fileDiff < 0 && rankDiff < 0:
				directions[from][to] = SouthWest
			case absFileDiff == absRankDiff && fileDiff > 0 && rankDiff > 0:
				directions[from][to] = NorthEast
			case absFileDiff == absRankDiff && fileDiff > 0 && rankDiff < 0:
				directions[from][to] = SouthEast
			case absFileDiff == absRankDiff && fileDiff < 0 && rankDiff > 0:
				directions[from][to] = NorthWest
			default:
				directions[from][to] = Invalid
			}
		}
	}
	return
}

// generateRayAttacks returns a precalculated array for all posible rays on each direction from each square in the board
func generateRayAttacks() (rayAttacks [8][64]Bitboard) {
	directions := [8]uint64{North, NorthEast, East, SouthEast, South, SouthWest, West, NorthWest}

	for sq := range 64 {
		rank, file := sq/8, sq%8
		for _, dir := range directions {
			switch dir {
			case North:
				for r := rank + 1; r <= 7; r++ {
					rayAttacks[dir][sq] |= Bitboard(1 << (r*8 + file))
				}
			case NorthEast:
				for r, f := rank+1, file+1; r <= 7 && f <= 7; r, f = r+1, f+1 {
					rayAttacks[dir][sq] |= Bitboard(1 << (r*8 + f))
				}
			case East:
				for f := file + 1; f <= 7; f++ {
					rayAttacks[dir][sq] |= Bitboard(1 << (f + rank*8))
				}
			case SouthEast:
				for r, f := rank-1, file+1; r >= 0 && f <= 7; r, f = r-1, f+1 {
					rayAttacks[dir][sq] |= Bitboard(1 << (r*8 + f))
				}
			case South:
				for r := rank - 1; r >= 0; r-- {
					rayAttacks[dir][sq] |= Bitboard(1 << (r*8 + file))
				}
			case SouthWest:
				for r, f := rank-1, file-1; r >= 0 && f >= 0; r, f = r-1, f-1 {
					rayAttacks[dir][sq] |= Bitboard(1 << (r*8 + f))
				}
			case West:
				for f := file - 1; f >= 0; f-- {
					rayAttacks[dir][sq] |= Bitboard(1 << (f + rank*8))
				}
			case NorthWest:
				for r, f := rank+1, file-1; r <= 7 && f >= 0; r, f = r+1, f-1 {
					rayAttacks[dir][sq] |= Bitboard(1 << (r*8 + f))
				}
			}
		}
	}
	return
}

// generateKnightAttacksTable returns a precalculated array for all posible knight moves from each square in the board
func generateKnightAttacksTable() (knightAttacksTable [64]Bitboard) {
	for sq := range 64 {
		from := Bitboard(1 << sq)

		notInHFile := from & ^(from & Files[7])
		notInAFile := from & ^(from & Files[0])
		notInABFiles := from & ^(from & (Files[0] | Files[1]))
		notInGHFiles := from & ^(from & (Files[7] | Files[6]))

		knightAttacksTable[sq] = notInAFile<<15 | notInHFile<<17 | notInGHFiles<<10 |
			notInABFiles<<6 | notInHFile>>15 | notInAFile>>17 |
			notInABFiles>>10 | notInGHFiles>>6

	}
	return
}

// generateKingAttacksTable returns a precalculated array for all posible king moves from each square in the board
func generateKingAttacksTable() (kingAttacksTable [64]Bitboard) {
	for sq := range 64 {
		k := Bitboard(1 << sq)
		notInHFile := k & ^(k & Files[7])
		notInAFile := k & ^(k & Files[0])

		kingAttacksTable[sq] = notInAFile<<7 | k<<8 | notInHFile<<9 |
			notInHFile<<1 | notInAFile>>1 | notInHFile>>7 |
			k>>8 | notInAFile>>9
	}
	return
}

// generatePawnPushesTable returns a precalculated table containing the squares that a pawn can push
func generatePawnPushesTable() (pawnPushesTable [2][64]Bitboard) {
	for sq := a2; sq <= h7; sq++ { // Only from 2nd to 7th rank
		bb := bitboardFromIndex(sq)
		pawnPushesTable[White][sq] = bb << 8
		pawnPushesTable[Black][sq] = bb >> 8
	}
	return
}

// generateDoublePushesTable returns a precalculated table containing the squares that a pawn can double push
func generatePawnDoublePushesTable() (pawnDoublePushesTable [2][64]Bitboard) {
	for file := range 8 {
		whiteSq := a2 + file
		blackSq := a7 + file

		pawnDoublePushesTable[White][whiteSq] = bitboardFromIndex(whiteSq + 16)
		pawnDoublePushesTable[Black][blackSq] = bitboardFromIndex(blackSq - 16)
	}
	return
}

// generateAttacksFrontSpans returns a precalculated table containing the front attack spans for each square
func generateAttacksFrontSpans() (attacksFrontSpans [2][64]Bitboard) {

	for sq := range 64 {
		file, rank := sq%8, sq/8
		eastFront, westFront := rank*8+file+1, rank*8+file-1

		if file < 7 {
			attacksFrontSpans[White][sq] |= RayAttacks[North][eastFront]
			attacksFrontSpans[Black][sq] |= RayAttacks[South][eastFront]
		}
		if file > 0 {
			attacksFrontSpans[White][sq] |= RayAttacks[North][westFront]
			attacksFrontSpans[Black][sq] |= RayAttacks[South][westFront]
		}
	}

	return
}

// generateKingZone returns a precalculated table containing the king zone for each square
// King zone is defined as the squares a king can move plus the squares 2 ranks ahead, depending on the side
// Here is an example. White zone from g2, and black zone from b8
// x k x . . . . .
// x x x . . . . .
// x x x . . . . .
// . . . . . . . .
// . . . . . x x x
// . . . . . x x x
// . . . . . x K x
// . . . . . x x x
func generateKingZone() (kingZone [2][64]Bitboard) {
	for sq := range 64 {
		from := Bitboard(1 << sq)

		// White
		kingZone[White][sq] = kingAttacksTable[sq]
		fromUp := from << 8
		kingZone[White][sq] |= pawnAttacks(&fromUp, White)
		kingZone[White][sq] |= from << 16

		// Black
		kingZone[Black][sq] = kingAttacksTable[sq]
		fromDown := from >> 8
		kingZone[Black][sq] |= pawnAttacks(&fromDown, Black)
		kingZone[Black][sq] |= from >> 16
	}
	return
}

// raysDirection returns the rays along the direction passed that intersects the
// piece in the square passed
func raysDirection(square Bitboard, direction uint64) Bitboard {
	oppositeDirections := [8]uint64{South, SouthWest, West, NorthWest, North, NorthEast, East, SouthEast}

	return RayAttacks[direction][Bsf(square)] | square |
		RayAttacks[oppositeDirections[direction]][Bsf(square)]
}

// getRayPath returns a Bitboard with the path between 2 bitboards pieces
// (not including the 2 pieces)
func getRayPath(from *Bitboard, to *Bitboard) (rayPath Bitboard) {
	fromSq := Bsf(*from)
	toSq := Bsf(*to)

	fromDirection := directions[fromSq][toSq]
	toDirection := directions[toSq][fromSq]

	if fromDirection == Invalid || toDirection == Invalid {
		return
	}

	return RayAttacks[fromDirection][fromSq] &
		RayAttacks[toDirection][toSq]
}

// pieces square tables with the value of each piece + square value for middlegame
var middlegamePiecesScore [12][64]int

// pieces square tables with the value of each piece + square value for endgame
var endgamePiecesScore [12][64]int

// MiddlegamePieceValue is the value of each piece for middlegame phase
var MiddlegamePieceValue = [6]int{10000, 1021, 435, 344, 311, 73}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{10000, 1061, 592, 356, 344, 83}

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-48, 25, 24, -24, -45, -26, 13, 8,
		-9, -2, -30, 14, 1, 2, -9, -17,
		-55, 21, -24, -27, -6, 39, 14, -25,
		-48, -50, -63, -95, -89, -68, -64, -101,
		-66, -43, -73, -104, -102, -68, -79, -128,
		-25, 4, -46, -57, -47, -55, -16, -51,
		46, 4, 0, -27, -34, -13, 13, 30,
		43, 58, 40, -28, 21, -17, 41, 51,
	},
	// Queen
	{
		-25, -29, -10, 8, 17, 31, 56, 5,
		3, -24, -24, -37, -44, 3, 2, 57,
		0, -8, -10, -12, -9, 37, 37, 36,
		-13, -6, -11, -11, -17, -10, -7, -3,
		-3, -12, -10, -8, -9, -17, 0, 0,
		-3, 0, -7, -10, -4, -3, 12, 10,
		7, 3, 5, 8, 6, 17, 27, 40,
		-3, 1, 3, 4, 7, -3, 7, 13,
	},
	// Rook
	{
		1, -4, -4, -10, 8, 29, 35, 44,
		-6, -9, 6, 23, 3, 40, 30, 51,
		-22, 2, -3, -5, 17, 27, 57, 24,
		-25, -11, -10, -3, -3, 3, 9, -5,
		-33, -34, -25, -16, -15, -28, -5, -21,
		-34, -30, -27, -23, -11, -11, 16, -5,
		-31, -27, -15, -12, -7, 0, 15, -18,
		-16, -12, -11, -4, 3, -1, 2, -9,
	},
	// Bishop
	{
		-24, -50, -54, -94, -74, -84, -31, -55,
		-10, 9, -4, -16, 3, -14, -26, -39,
		-1, 0, 5, 10, 0, 30, 9, 8,
		-14, 3, 3, 18, 12, 7, 0, -15,
		0, -12, -4, 15, 13, -6, -4, 16,
		3, 13, 4, 5, 8, 8, 13, 25,
		26, 7, 19, -2, 5, 19, 30, 27,
		13, 23, 1, -3, 5, -1, 18, 42,
	},
	// Knight
	{
		-118, -96, -54, -29, 14, -69, -39, -76,
		1, 17, 41, 49, 16, 74, 6, 24,
		6, 21, 32, 32, 66, 66, 33, 11,
		7, 7, 20, 41, 19, 43, 13, 39,
		0, 3, 6, 16, 18, 18, 16, 9,
		-16, -5, -8, -1, 15, -5, 16, 0,
		-17, -12, -7, 7, 4, 8, 11, 7,
		-49, -12, -25, -12, -4, 0, -10, -10,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		75, 85, 59, 97, 63, 51, -36, -55,
		20, 12, 46, 52, 61, 94, 56, 25,
		-7, -4, 4, 7, 29, 24, 9, 4,
		-9, -12, 0, 14, 15, 14, 0, -4,
		-14, -7, -4, 2, 15, 3, 16, 0,
		-13, -8, -5, -6, 1, 18, 19, -12,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-79, -42, -25, 3, -1, 4, -1, -81,
		-17, 26, 34, 30, 42, 53, 51, 13,
		3, 32, 49, 58, 62, 58, 55, 21,
		-3, 34, 54, 66, 67, 62, 50, 21,
		-12, 17, 44, 61, 60, 43, 30, 13,
		-25, 0, 24, 38, 35, 25, 2, -9,
		-45, -13, 1, 13, 15, 0, -19, -42,
		-85, -63, -38, -14, -37, -24, -57, -93,
	},
	// Queen
	{
		27, 41, 58, 54, 50, 46, 6, 35,
		19, 47, 76, 98, 120, 80, 64, 53,
		27, 36, 66, 78, 95, 75, 47, 51,
		35, 41, 53, 71, 87, 79, 78, 66,
		17, 40, 41, 60, 59, 59, 45, 48,
		3, 15, 32, 32, 35, 34, 13, 11,
		-6, -9, -6, -1, 4, -17, -44, -57,
		-5, -6, -7, 5, -7, -11, -22, -23,
	},
	// Rook
	{
		47, 51, 57, 53, 44, 42, 41, 37,
		48, 59, 60, 48, 52, 41, 40, 31,
		49, 47, 49, 46, 37, 31, 28, 27,
		51, 46, 51, 45, 35, 31, 34, 33,
		41, 43, 41, 36, 34, 36, 28, 28,
		30, 28, 26, 27, 20, 16, 0, 3,
		20, 24, 23, 20, 13, 7, 0, 6,
		19, 19, 27, 21, 13, 15, 10, 3,
	},
	// Bishop
	{
		3, 8, 5, 14, 9, 2, 0, -6,
		-8, -8, -3, -1, -11, -2, 1, 0,
		8, -2, 1, -8, -3, 3, 0, 9,
		3, 3, 0, 15, 7, 5, 1, 8,
		-5, 4, 12, 8, 7, 4, 1, -16,
		-3, 2, 5, 9, 11, 4, -6, -7,
		6, -12, -13, -3, -1, -10, -7, -11,
		-7, 5, -10, -5, -9, 2, -14, -26,
	},
	// Knight
	{
		-39, -16, -3, -12, -16, -28, -28, -61,
		-4, -4, -13, -14, -16, -29, -5, -22,
		-7, -10, 1, 4, -11, -21, -18, -15,
		2, 0, 8, 11, 14, 7, 7, -6,
		0, -4, 12, 9, 18, 4, 0, 0,
		-11, -8, -3, 11, 9, -8, -12, -6,
		-6, -4, -10, -10, -11, -13, -13, 7,
		7, -19, -11, -10, -7, -16, -8, 1,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		149, 134, 138, 75, 78, 96, 142, 165,
		77, 80, 35, -5, -11, 11, 56, 60,
		59, 46, 30, 10, 8, 17, 34, 34,
		40, 36, 24, 15, 13, 19, 24, 23,
		34, 30, 26, 25, 25, 26, 20, 19,
		36, 33, 29, 26, 37, 27, 20, 19,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}
