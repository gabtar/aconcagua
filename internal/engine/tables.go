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
var MiddlegamePieceValue = [6]int{10000, 1057, 440, 348, 314, 74}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{10000, 1025, 576, 348, 337, 81}

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-54, 19, 16, -25, -50, -30, 9, 6,
		-1, -8, -35, 2, -5, -2, -18, -20,
		-45, 15, -28, -28, -10, 29, 13, -25,
		-41, -50, -58, -83, -76, -62, -56, -85,
		-62, -41, -75, -101, -104, -68, -79, -112,
		-19, 8, -45, -59, -48, -51, -13, -46,
		52, 8, 0, -29, -34, -15, 19, 30,
		43, 55, 35, -40, 14, -23, 38, 50,
	},
	// Queen
	{
		-23, -23, -2, 10, 26, 35, 53, 10,
		4, -23, -21, -31, -38, 10, 2, 60,
		2, -5, -8, -8, -3, 45, 46, 46,
		-11, -5, -9, -9, -15, -6, 0, 1,
		0, -10, -8, -5, -6, -13, 1, 5,
		-2, 2, -5, -8, -2, -1, 14, 12,
		8, 4, 8, 11, 9, 19, 29, 40,
		-1, 2, 5, 6, 9, -1, 8, 7,
	},
	// Rook
	{
		1, -2, -3, -4, 13, 25, 32, 42,
		-7, -10, 5, 21, 6, 31, 24, 50,
		-23, 2, -3, -6, 19, 28, 67, 31,
		-24, -11, -11, -4, -3, 1, 10, -2,
		-33, -35, -25, -16, -15, -29, -6, -21,
		-35, -30, -27, -23, -11, -12, 15, -5,
		-31, -28, -15, -13, -7, 0, 14, -18,
		-16, -13, -11, -5, 3, -2, 2, -10,
	},
	// Bishop
	{
		-25, -40, -57, -84, -67, -79, -23, -48,
		-10, 9, -4, -16, 3, -11, -22, -37,
		-1, 0, 5, 9, -2, 30, 9, 7,
		-14, 3, 3, 18, 11, 7, 0, -16,
		0, -12, -4, 15, 13, -6, -4, 16,
		3, 13, 4, 4, 7, 8, 12, 25,
		25, 6, 18, -3, 5, 19, 30, 26,
		12, 23, 1, -4, 4, -1, 17, 39,
	},
	// Knight
	{
		-124, -93, -51, -30, 25, -72, -30, -76,
		-6, 16, 41, 50, 20, 76, 5, 24,
		4, 21, 32, 31, 65, 69, 34, 16,
		7, 7, 20, 41, 19, 43, 13, 38,
		0, 2, 6, 16, 18, 18, 18, 9,
		-17, -5, -8, -1, 16, -4, 16, 0,
		-18, -13, -7, 7, 4, 9, 10, 6,
		-54, -12, -25, -12, -5, 0, -10, -10,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		77, 89, 60, 85, 50, 62, -28, -53,
		18, 10, 44, 48, 57, 93, 52, 24,
		-7, -5, 3, 7, 28, 28, 11, 8,
		-11, -13, 0, 14, 16, 16, 2, -2,
		-15, -8, -5, 2, 16, 4, 19, 0,
		-14, -10, -6, -7, 2, 18, 22, -11,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-79, -51, -35, -6, -7, 0, -4, -68,
		-24, 15, 20, 17, 31, 43, 43, 9,
		0, 25, 40, 47, 54, 53, 50, 19,
		-3, 32, 50, 60, 61, 60, 48, 21,
		-7, 20, 47, 62, 62, 48, 36, 17,
		-18, 4, 28, 41, 39, 30, 8, -1,
		-38, -10, 3, 15, 18, 6, -15, -32,
		-77, -57, -34, -14, -34, -19, -49, -83,
	},
	// Queen
	{
		23, 32, 46, 47, 38, 38, 4, 26,
		16, 43, 69, 86, 107, 64, 58, 47,
		23, 30, 59, 69, 82, 60, 30, 30,
		31, 36, 48, 64, 80, 68, 64, 53,
		13, 35, 35, 53, 53, 50, 39, 35,
		0, 10, 26, 26, 29, 28, 7, 3,
		-7, -13, -11, -7, 0, -23, -49, -57,
		-9, -9, -12, 0, -12, -16, -20, -15,
	},
	// Rook
	{
		45, 49, 56, 49, 41, 43, 41, 36,
		47, 58, 60, 48, 49, 43, 41, 31,
		49, 47, 48, 46, 35, 30, 24, 24,
		51, 46, 51, 45, 34, 31, 33, 32,
		41, 43, 40, 36, 34, 36, 28, 28,
		30, 28, 26, 26, 19, 15, 0, 3,
		20, 24, 22, 20, 12, 7, -1, 7,
		19, 19, 27, 21, 13, 15, 10, 4,
	},
	// Bishop
	{
		4, 5, 6, 10, 7, 1, -1, -8,
		-8, -8, -3, -1, -12, -5, 0, 0,
		7, -2, 0, -9, -3, 2, -1, 8,
		3, 2, 0, 14, 6, 4, 0, 8,
		-5, 3, 11, 7, 7, 3, 1, -17,
		-2, 2, 5, 8, 10, 4, -7, -7,
		6, -12, -13, -3, -1, -10, -7, -11,
		-7, 5, -10, -5, -9, 2, -13, -22,
	},
	// Knight
	{
		-34, -16, -5, -12, -20, -25, -31, -61,
		-1, -4, -13, -15, -19, -31, -6, -23,
		-7, -11, 0, 3, -12, -23, -20, -18,
		1, 0, 7, 10, 13, 6, 5, -7,
		-1, -5, 11, 8, 16, 3, -3, -1,
		-11, -9, -4, 11, 8, -8, -12, -7,
		-6, -4, -10, -11, -12, -14, -13, 6,
		9, -20, -11, -10, -7, -17, -8, -6,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		149, 133, 139, 81, 84, 92, 137, 162,
		79, 82, 37, 0, -6, 14, 59, 62,
		59, 46, 31, 10, 9, 17, 34, 34,
		40, 36, 24, 16, 14, 19, 25, 23,
		34, 30, 26, 25, 25, 27, 20, 20,
		36, 33, 30, 28, 37, 28, 20, 19,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}
