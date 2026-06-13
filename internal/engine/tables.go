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
var MiddlegamePieceValue = [6]int{10000, 957, 426, 344, 307, 71}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{10000, 1074, 604, 362, 349, 86}

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-35, 36, 40, -21, -34, -27, 9, 13,
		-30, 5, -18, 41, 14, 8, -3, -27,
		-78, 32, -8, -15, 10, 58, 0, -37,
		-59, -40, -54, -79, -79, -70, -83, -137,
		-68, -32, -52, -78, -80, -54, -79, -148,
		-32, 10, -30, -38, -28, -35, -4, -54,
		41, 10, 5, -19, -27, -3, 18, 28,
		36, 55, 36, -30, 17, -18, 37, 39,
	},
	// Queen
	{
		-44, -47, -34, -12, -15, 2, 33, -19,
		5, -19, -27, -53, -70, -25, -13, 44,
		6, -2, -6, -13, -26, 13, 14, 16,
		-5, -3, -9, -11, -14, -8, 0, 0,
		1, -7, -7, -1, -1, -7, 8, 9,
		1, 5, 0, -4, 2, 5, 20, 15,
		11, 11, 13, 15, 13, 23, 35, 48,
		1, 6, 11, 11, 15, 1, 16, 23,
	},
	// Rook
	{
		-8, -9, -13, -19, -1, 28, 33, 34,
		1, -2, 14, 33, 9, 38, 24, 45,
		-11, 12, 7, 6, 31, 39, 58, 26,
		-17, -3, -3, 7, 4, 14, 12, 0,
		-27, -27, -17, -8, -7, -23, -2, -17,
		-30, -25, -22, -17, -9, -6, 17, -1,
		-27, -23, -12, -10, -5, 0, 14, -15,
		-14, -10, -9, -2, 4, 0, 2, -10,
	},
	// Bishop
	{
		-29, -59, -55, -97, -84, -83, -44, -57,
		-9, 0, -2, -18, 3, -12, -21, -34,
		0, 3, 3, 14, 1, 46, 11, 22,
		-13, 0, 8, 12, 17, 5, 4, -19,
		-1, -8, -8, 18, 10, -3, -6, 19,
		4, 9, 6, 2, 6, 7, 12, 21,
		19, 7, 16, -3, 4, 15, 28, 26,
		12, 20, 0, -5, 4, -3, 18, 40,
	},
	// Knight
	{
		-108, -97, -62, -28, 5, -61, -56, -72,
		1, 13, 38, 41, 24, 73, 12, 22,
		7, 23, 19, 34, 65, 69, 32, 18,
		11, 9, 23, 31, 23, 36, 20, 42,
		3, 8, 10, 19, 19, 22, 14, 14,
		-13, -2, -6, 2, 15, -3, 16, 3,
		-14, -10, -5, 7, 5, 8, 13, 10,
		-44, -10, -22, -10, -3, 1, -8, -4,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		73, 75, 50, 96, 70, 56, -21, -25,
		22, 11, 44, 50, 62, 95, 63, 27,
		-5, -4, 5, 8, 26, 26, 6, 4,
		-7, -10, 2, 14, 17, 14, -1, -3,
		-13, -6, -5, 3, 11, 1, 13, -2,
		-11, -7, -5, -8, 1, 16, 19, -10,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-82, -42, -26, 7, 0, 0, -10, -99,
		-12, 29, 36, 31, 44, 56, 50, 10,
		8, 32, 52, 61, 66, 58, 59, 21,
		-1, 35, 57, 70, 70, 65, 52, 25,
		-11, 18, 45, 62, 60, 43, 30, 13,
		-24, 0, 26, 39, 37, 24, 0, -14,
		-48, -12, 4, 17, 19, 3, -21, -49,
		-87, -63, -35, -8, -31, -20, -57, -96,
	},
	// Queen
	{
		20, 24, 39, 32, 30, 26, -12, 23,
		14, 33, 62, 82, 100, 61, 46, 37,
		29, 38, 62, 63, 77, 50, 24, 29,
		39, 52, 60, 69, 70, 56, 60, 43,
		29, 53, 58, 70, 67, 58, 40, 36,
		15, 31, 53, 56, 59, 47, 22, 13,
		7, 9, 16, 27, 31, 3, -27, -49,
		14, 13, 15, 28, 6, 0, -12, -19,
	},
	// Rook
	{
		44, 47, 53, 48, 40, 37, 37, 35,
		42, 53, 53, 39, 43, 36, 38, 29,
		46, 44, 44, 40, 30, 23, 24, 24,
		51, 46, 50, 41, 31, 24, 30, 30,
		43, 44, 41, 36, 33, 34, 26, 27,
		35, 32, 29, 28, 22, 15, 0, 4,
		26, 30, 27, 23, 15, 10, 1, 10,
		27, 26, 32, 23, 15, 15, 13, 9,
	},
	// Bishop
	{
		5, 12, 5, 16, 10, 3, 5, -3,
		-7, -4, -3, -1, -12, -5, 1, 0,
		8, -2, 2, -8, -4, 0, -2, 5,
		4, 5, 1, 17, 8, 6, 1, 8,
		-3, 4, 12, 10, 7, 6, 3, -16,
		-2, 4, 7, 9, 13, 5, -4, -8,
		7, -11, -11, -1, 0, -7, -6, -10,
		-4, 7, -9, -4, -8, 3, -12, -25,
	},
	// Knight
	{
		-41, -14, -4, -15, -14, -31, -23, -62,
		-3, -2, -13, -13, -19, -29, -6, -22,
		-6, -11, 2, 2, -10, -26, -20, -18,
		5, 2, 9, 12, 15, 7, 7, -4,
		1, -2, 15, 12, 20, 6, 2, 1,
		-7, -4, 0, 14, 12, -4, -9, -3,
		-2, 0, -6, -6, -8, -9, -11, 11,
		13, -18, -8, -7, -5, -14, -5, 7,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		151, 139, 142, 76, 79, 100, 145, 163,
		78, 81, 37, -3, -10, 13, 57, 61,
		59, 47, 30, 10, 10, 17, 35, 33,
		39, 35, 23, 15, 13, 18, 24, 22,
		33, 29, 25, 24, 24, 25, 18, 19,
		35, 32, 27, 24, 34, 23, 17, 17,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}
