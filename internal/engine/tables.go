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

	initBitboards()
	directions = generateDirections()
	RayAttacks = generateRayAttacks()
	squaresBetween = generateSquaresBetween()
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

// squaresBetween is a precalculated table that contains a bitboard with the squares between 2 squares in any of the 8 direction
var squaresBetween [64][64]Bitboard

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

func generateSquaresBetween() (squaresBetween [64][64]Bitboard) {
	for from := range 64 {
		for to := range 64 {
			fromBB := bitboardFromIndex(from)
			toBB := bitboardFromIndex(to)

			squaresBetween[from][to] = getRayPath(&fromBB, &toBB)
		}
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
var MiddlegamePieceValue = [6]int{10000, 982, 430, 348, 309, 71}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{10000, 1066, 601, 359, 347, 85}

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-40, 32, 35, -22, -38, -27, 10, 10,
		-24, 3, -22, 32, 9, 7, -4, -23,
		-71, 29, -14, -20, 4, 53, 6, -33,
		-57, -43, -57, -88, -84, -69, -77, -125,
		-69, -35, -60, -89, -90, -60, -78, -141,
		-30, 11, -34, -43, -35, -41, -6, -52,
		43, 10, 6, -19, -27, -6, 18, 29,
		40, 59, 40, -24, 21, -13, 42, 46,
	},
	// Queen
	{
		-32, -36, -21, 0, 2, 17, 46, -6,
		0, -27, -30, -48, -58, -9, -7, 45,
		0, -8, -10, -18, -16, 23, 29, 25,
		-12, -5, -13, -13, -20, -15, -4, -4,
		0, -12, -9, -6, -5, -12, 3, 2,
		-2, 0, -4, -8, -2, 0, 15, 9,
		7, 6, 8, 10, 9, 17, 29, 42,
		-3, 1, 5, 5, 8, -5, 6, 17,
	},
	// Rook
	{
		2, 0, -3, -9, 8, 32, 38, 43,
		0, -3, 11, 27, 5, 41, 30, 49,
		-14, 9, 4, 1, 25, 34, 61, 25,
		-19, -4, -4, 4, 2, 10, 13, -1,
		-28, -28, -18, -10, -9, -23, -2, -18,
		-31, -25, -22, -18, -10, -7, 17, -2,
		-28, -24, -13, -11, -6, -3, 13, -16,
		-14, -11, -10, -4, 2, -2, 1, -9,
	},
	// Bishop
	{
		-28, -56, -52, -94, -77, -79, -38, -56,
		-9, 0, -3, -17, 1, -10, -23, -33,
		0, 3, 5, 14, 4, 45, 14, 21,
		-14, 5, 7, 15, 16, 7, 3, -18,
		0, -9, -7, 18, 10, -3, -6, 18,
		3, 9, 5, 2, 6, 7, 11, 21,
		20, 6, 15, -4, 3, 14, 27, 25,
		11, 19, -1, -7, 2, -5, 14, 40,
	},
	// Knight
	{
		-111, -97, -57, -27, 9, -60, -50, -73,
		1, 13, 38, 42, 21, 72, 11, 23,
		8, 23, 21, 34, 64, 67, 32, 18,
		10, 10, 23, 35, 24, 39, 20, 42,
		3, 8, 11, 19, 19, 22, 15, 14,
		-13, -2, -6, 2, 15, -3, 16, 3,
		-14, -10, -6, 6, 5, 6, 12, 9,
		-45, -11, -23, -10, -3, 0, -10, -5,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		77, 78, 55, 100, 70, 57, -27, -35,
		22, 12, 44, 51, 63, 94, 62, 26,
		-5, -3, 5, 9, 26, 26, 6, 4,
		-7, -10, 2, 14, 17, 14, 0, -3,
		-13, -6, -4, 3, 12, 1, 14, -2,
		-12, -7, -5, -8, 1, 14, 18, -12,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-81, -42, -26, 7, 0, 0, -8, -96,
		-13, 29, 37, 33, 44, 56, 50, 11,
		8, 33, 52, 62, 66, 59, 58, 21,
		-1, 36, 57, 70, 71, 65, 52, 24,
		-11, 19, 46, 64, 62, 44, 30, 13,
		-23, 0, 26, 40, 38, 25, 0, -13,
		-47, -12, 3, 16, 18, 3, -21, -46,
		-86, -64, -36, -11, -36, -22, -59, -96,
	},
	// Queen
	{
		24, 32, 49, 45, 44, 38, -3, 30,
		20, 42, 70, 90, 111, 73, 52, 43,
		32, 40, 65, 70, 88, 60, 27, 32,
		41, 48, 58, 70, 80, 67, 68, 49,
		24, 52, 51, 67, 63, 57, 42, 41,
		11, 29, 48, 48, 50, 41, 17, 14,
		6, 3, 8, 16, 19, -8, -37, -53,
		9, 7, 4, 14, -2, -8, -20, -23,
	},
	// Rook
	{
		45, 48, 54, 50, 41, 39, 38, 36,
		45, 55, 57, 44, 47, 37, 37, 29,
		48, 46, 46, 43, 34, 27, 25, 25,
		52, 47, 51, 43, 32, 27, 31, 31,
		43, 44, 41, 36, 33, 34, 26, 27,
		35, 31, 29, 27, 21, 14, -1, 4,
		26, 29, 26, 22, 14, 10, 0, 9,
		24, 25, 32, 22, 14, 14, 12, 5,
	},
	// Bishop
	{
		5, 12, 5, 15, 9, 2, 4, -3,
		-6, -4, -3, -1, -11, -5, 2, 0,
		9, -2, 2, -8, -5, 1, -3, 6,
		4, 4, 1, 16, 8, 5, 2, 7,
		-4, 5, 12, 10, 8, 5, 3, -16,
		-2, 4, 7, 8, 12, 5, -5, -8,
		6, -11, -12, -3, 0, -8, -8, -10,
		-4, 7, -10, -4, -8, 2, -14, -26,
	},
	// Knight
	{
		-39, -14, -4, -13, -13, -28, -22, -60,
		-2, -2, -13, -13, -16, -28, -4, -21,
		-6, -11, 0, 2, -10, -24, -19, -16,
		5, 1, 8, 9, 14, 6, 7, -4,
		2, -2, 14, 11, 18, 5, 1, 2,
		-7, -5, -1, 12, 10, -6, -10, -3,
		-3, -1, -8, -7, -9, -12, -13, 10,
		13, -17, -8, -8, -6, -16, -4, 6,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		151, 138, 141, 76, 79, 99, 146, 165,
		78, 81, 37, -4, -10, 13, 57, 61,
		59, 47, 30, 10, 10, 17, 35, 33,
		39, 35, 23, 15, 13, 18, 24, 22,
		34, 30, 25, 24, 24, 25, 18, 19,
		35, 32, 27, 25, 34, 24, 17, 17,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}
