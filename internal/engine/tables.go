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

	// AllSquares contains the bitboard mask for all squares in the board
	AllSquares Bitboard = 0xFFFFFFFFFFFFFFFF

	// notAFile/notHFile contains a bitboard mask without the A and H file
	notAFile Bitboard = 0xfefefefefefefefe
	notHFile Bitboard = 0x7f7f7f7f7f7f7f7f
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

// IsolatedAdjacentFilesMask contains the adjacent files of a pawn to test if it is isolated
var IsolatedAdjacentFilesMask = [8]Bitboard{
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
var MiddlegamePieceValue = [6]int{10000, 930, 423, 342, 305, 71}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{10000, 1093, 607, 365, 351, 87}

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-24, 43, 49, -18, -26, -27, 6, 19,
		-40, 7, -11, 57, 22, 10, -1, -34,
		-89, 35, -1, -8, 19, 63, -8, -44,
		-64, -39, -53, -71, -76, -73, -92, -154,
		-70, -32, -51, -76, -81, -56, -83, -155,
		-34, 8, -31, -39, -30, -36, -5, -55,
		40, 10, 5, -20, -28, -4, 17, 27,
		36, 55, 36, -30, 16, -18, 37, 39,
	},
	// Queen
	{
		-47, -52, -44, -13, -30, -9, 23, -24,
		6, -18, -26, -51, -72, -30, -12, 46,
		7, -1, -6, -13, -29, 13, 11, 15,
		-5, -3, -10, -13, -14, -8, 1, 0,
		1, -7, -8, -2, -2, -7, 8, 9,
		1, 5, 0, -4, 1, 4, 20, 15,
		11, 10, 12, 15, 13, 22, 34, 49,
		0, 6, 10, 10, 15, 1, 17, 23,
	},
	// Rook
	{
		-10, -12, -20, -26, -7, 24, 30, 30,
		2, -1, 14, 33, 10, 39, 24, 45,
		-11, 13, 8, 7, 32, 39, 59, 27,
		-16, -2, -2, 8, 5, 15, 13, 1,
		-26, -26, -16, -7, -6, -22, -1, -16,
		-29, -24, -21, -16, -8, -5, 18, -1,
		-26, -22, -12, -9, -4, 1, 16, -14,
		-13, -9, -8, -1, 4, 0, 2, -9,
	},
	// Bishop
	{
		-29, -61, -56, -99, -91, -84, -51, -57,
		-8, 1, -2, -18, 3, -12, -21, -34,
		1, 4, 3, 14, 1, 46, 11, 22,
		-12, 1, 8, 12, 17, 5, 4, -19,
		-1, -8, -8, 18, 10, -3, -6, 19,
		4, 9, 6, 2, 6, 7, 12, 21,
		19, 7, 16, -3, 4, 15, 28, 26,
		12, 20, 0, -5, 4, -4, 18, 40,
	},
	// Knight
	{
		-107, -97, -66, -28, 1, -61, -65, -70,
		1, 13, 38, 41, 24, 72, 13, 22,
		7, 23, 18, 34, 65, 68, 32, 18,
		11, 9, 23, 31, 23, 36, 20, 41,
		3, 8, 10, 19, 19, 22, 14, 14,
		-13, -2, -6, 2, 15, -3, 16, 3,
		-14, -10, -6, 6, 5, 7, 13, 10,
		-44, -10, -22, -10, -3, 1, -8, -4,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		71, 74, 45, 92, 68, 54, -16, -17,
		22, 11, 43, 48, 61, 94, 63, 28,
		-6, -4, 4, 8, 25, 26, 6, 4,
		-7, -11, 2, 13, 16, 13, -1, -3,
		-14, -6, -5, 2, 11, 1, 13, -2,
		-12, -7, -5, -8, 1, 15, 18, -10,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-85, -44, -28, 6, -1, 0, -10, -101,
		-11, 28, 35, 28, 42, 55, 49, 11,
		10, 32, 50, 60, 64, 57, 61, 22,
		-1, 34, 56, 68, 70, 65, 53, 28,
		-12, 18, 45, 61, 60, 43, 30, 14,
		-24, 0, 26, 39, 37, 24, 0, -14,
		-49, -12, 4, 17, 19, 3, -21, -49,
		-87, -63, -35, -8, -31, -20, -57, -97,
	},
	// Queen
	{
		20, 25, 43, 29, 37, 30, -9, 25,
		10, 30, 60, 78, 100, 63, 43, 31,
		25, 35, 61, 61, 78, 48, 27, 27,
		36, 51, 60, 72, 70, 55, 56, 41,
		28, 53, 59, 72, 68, 58, 38, 33,
		14, 31, 54, 57, 59, 47, 21, 10,
		6, 9, 16, 28, 31, 3, -27, -51,
		13, 12, 15, 29, 6, -1, -13, -20,
	},
	// Rook
	{
		44, 47, 54, 50, 41, 37, 37, 36,
		41, 51, 52, 38, 42, 34, 37, 28,
		44, 43, 43, 38, 28, 22, 23, 23,
		50, 45, 49, 40, 29, 23, 29, 29,
		42, 43, 40, 35, 31, 33, 24, 26,
		34, 31, 28, 27, 20, 14, -2, 3,
		25, 28, 26, 22, 14, 8, 0, 8,
		26, 25, 31, 22, 13, 14, 11, 8,
	},
	// Bishop
	{
		5, 12, 5, 16, 12, 3, 6, -3,
		-7, -4, -3, -1, -12, -5, 1, -1,
		8, -2, 2, -8, -4, 0, -3, 5,
		4, 4, 0, 17, 8, 6, 1, 8,
		-3, 4, 12, 10, 7, 6, 3, -16,
		-2, 4, 7, 9, 13, 6, -4, -7,
		7, -11, -11, -1, 0, -6, -6, -9,
		-4, 7, -9, -3, -7, 4, -11, -25,
	},
	// Knight
	{
		-43, -14, -2, -14, -12, -31, -20, -64,
		-2, -2, -13, -13, -18, -28, -6, -22,
		-6, -10, 2, 2, -10, -25, -20, -17,
		5, 2, 10, 12, 16, 8, 8, -4,
		2, -1, 16, 12, 20, 7, 3, 2,
		-7, -4, 1, 15, 13, -3, -8, -3,
		-2, 0, -5, -5, -7, -8, -10, 12,
		13, -17, -7, -7, -4, -13, -5, 8,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		151, 138, 142, 76, 78, 99, 142, 159,
		78, 81, 37, -3, -10, 13, 57, 60,
		59, 47, 30, 10, 10, 17, 35, 33,
		39, 35, 23, 15, 13, 18, 24, 22,
		33, 29, 25, 23, 24, 24, 18, 19,
		35, 31, 26, 24, 33, 23, 16, 17,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}
