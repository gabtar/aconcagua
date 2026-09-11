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
var MiddlegamePieceValue = [6]int{0, 954, 434, 347, 309, 75}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{0, 1115, 605, 358, 342, 102}

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-16, 44, 52, -18, -21, -24, 8, 26,
		-45, 3, -12, 60, 20, 7, -1, -35,
		-93, 30, -6, -16, 10, 61, -3, -44,
		-67, -46, -61, -81, -87, -78, -89, -151,
		-74, -43, -67, -94, -98, -72, -92, -143,
		-35, -8, -49, -59, -52, -56, -24, -56,
		47, 10, -4, -41, -42, -21, 28, 37,
		41, 74, 47, -55, 10, -35, 53, 52,
	},
	// Queen
	{
		-30, -33, -21, 7, -5, 6, 38, -1,
		-2, -20, -10, -26, -47, -1, 7, 54,
		0, 0, 3, 9, 0, 38, 38, 42,
		-17, -10, -3, -5, 2, 13, 12, 15,
		-11, -14, -12, -4, -3, -4, 7, 8,
		-13, -3, -9, -10, -8, 0, 13, 5,
		-11, -7, 3, 2, 0, 9, 14, 27,
		-14, -17, -13, -2, -7, -20, -3, 0,
	},
	// Rook
	{
		11, 7, -1, 1, 15, 33, 36, 43,
		22, 21, 38, 56, 37, 60, 41, 64,
		2, 25, 26, 28, 52, 56, 79, 47,
		-12, 2, 6, 17, 21, 24, 24, 21,
		-30, -30, -16, -2, -2, -17, 2, -5,
		-40, -28, -19, -19, -13, -14, 18, -5,
		-43, -28, -13, -16, -10, -8, 7, -32,
		-21, -19, -8, -1, 3, -7, 0, -19,
	},
	// Bishop
	{
		-32, -55, -58, -93, -85, -73, -47, -52,
		-15, 14, 5, -11, 17, 10, 4, -16,
		0, 25, 25, 39, 24, 66, 35, 36,
		-6, 8, 28, 38, 39, 29, 11, -4,
		-11, 1, 9, 30, 28, 11, 2, -2,
		0, 7, 6, 10, 11, 7, 8, 12,
		2, 3, 15, -8, 0, 14, 22, 5,
		-13, 0, -17, -23, -16, -22, 2, 10,
	},
	// Knight
	{
		-133, -101, -63, -26, 7, -61, -71, -85,
		-19, 3, 42, 49, 34, 90, 14, 23,
		0, 43, 44, 61, 91, 96, 59, 33,
		3, 18, 44, 61, 48, 65, 30, 41,
		-10, 5, 21, 23, 33, 28, 26, 1,
		-30, -6, 8, 12, 23, 13, 17, -16,
		-36, -28, -13, -1, 0, 4, -7, -12,
		-68, -33, -42, -29, -24, -14, -28, -27,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		91, 96, 67, 114, 90, 74, 3, -3,
		23, 32, 63, 71, 83, 98, 76, 23,
		-19, 5, 7, 11, 33, 23, 25, 0,
		-29, -3, -4, 13, 13, 4, 11, -12,
		-32, -7, -8, -7, 7, -3, 27, -4,
		-32, -6, -13, -25, -4, 15, 37, -12,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-84, -49, -33, 0, -5, 6, 1, -86,
		-7, 20, 28, 14, 31, 48, 43, 15,
		13, 25, 40, 46, 46, 47, 52, 22,
		0, 31, 48, 58, 58, 57, 50, 31,
		-11, 11, 37, 51, 51, 39, 29, 17,
		-20, -1, 17, 29, 28, 21, 5, -2,
		-39, -13, 0, 8, 13, 3, -14, -34,
		-73, -57, -38, -23, -46, -20, -47, -76,
	},
	// Queen
	{
		24, 45, 66, 51, 63, 52, 9, 40,
		0, 38, 72, 95, 121, 83, 55, 30,
		11, 27, 62, 80, 99, 72, 48, 46,
		23, 43, 57, 81, 88, 75, 66, 54,
		15, 46, 53, 77, 70, 61, 42, 32,
		5, 15, 43, 41, 47, 37, 14, 5,
		-9, -4, -3, 8, 12, -12, -43, -64,
		-5, -12, -5, -2, -12, -16, -26, -38,
	},
	// Rook
	{
		49, 53, 67, 65, 58, 43, 39, 38,
		39, 51, 56, 48, 50, 38, 35, 23,
		38, 40, 42, 41, 29, 22, 18, 17,
		38, 36, 46, 40, 27, 21, 18, 17,
		30, 37, 38, 36, 30, 27, 17, 11,
		27, 27, 26, 31, 25, 16, -1, 0,
		19, 23, 25, 27, 17, 13, 1, 12,
		15, 25, 32, 30, 22, 15, 15, 5,
	},
	// Bishop
	{
		-6, 3, 4, 17, 12, 3, -1, -9,
		-17, 0, 6, 6, 1, 1, 7, -13,
		7, 3, 14, 8, 11, 9, 5, -2,
		2, 20, 17, 30, 21, 20, 14, 2,
		-3, 16, 24, 21, 18, 18, 11, -16,
		-3, 6, 16, 15, 20, 14, -1, -12,
		-10, -8, -10, 4, 8, -5, -3, -27,
		-27, -9, -28, -9, -14, -10, -20, -42,
	},
	// Knight
	{
		-63, -23, -1, -15, -11, -32, -25, -82,
		-23, -5, -3, 1, -7, -19, -15, -39,
		-12, 1, 24, 23, 10, 0, -4, -26,
		-4, 17, 30, 35, 34, 30, 16, -12,
		-3, 6, 31, 33, 34, 24, 10, -12,
		-20, 0, 10, 24, 22, 6, -6, -14,
		-24, -16, -4, 0, 0, -7, -24, -13,
		-7, -40, -23, -19, -19, -27, -32, -17,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		179, 168, 165, 105, 106, 123, 169, 186,
		102, 106, 64, 23, 16, 36, 80, 84,
		55, 45, 26, 16, 7, 11, 31, 30,
		28, 27, 9, 6, 3, 5, 19, 10,
		22, 25, 7, 20, 13, 10, 16, 5,
		28, 30, 17, 23, 30, 16, 17, 7,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// S is an alias/shorthand for NewScore()
var S = NewScore

// PiecesScore stores the score for each piece for mg and eg
var PiecesScore = [6]Score{S(0, 0), S(954, 1115), S(434, 605), S(347, 358), S(309, 342), S(75, 102)}

// Psqt stores the values for mg and eg phase for pieces square tables
var Psqt = [6][64]Score{
	// King
	{
		S(-16, -84), S(44, -49), S(52, -33), S(-18, 0), S(-21, -5), S(-24, 6), S(8, 1), S(26, -86),
		S(-45, -7), S(3, 20), S(-12, 28), S(60, 14), S(20, 31), S(7, 48), S(-1, 43), S(-35, 15),
		S(-93, 13), S(30, 25), S(-6, 40), S(-16, 46), S(10, 46), S(61, 47), S(-3, 52), S(-44, 22),
		S(-67, 0), S(-46, 31), S(-61, 48), S(-81, 58), S(-87, 58), S(-78, 57), S(-89, 50), S(-151, 31),
		S(-74, -11), S(-43, 11), S(-67, 37), S(-94, 51), S(-98, 51), S(-72, 39), S(-92, 29), S(-143, 17),
		S(-35, -20), S(-8, -1), S(-49, 17), S(-59, 29), S(-52, 28), S(-56, 21), S(-24, 5), S(-56, -2),
		S(47, -39), S(10, -13), S(-4, 0), S(-41, 8), S(-42, 13), S(-21, 3), S(28, -14), S(37, -34),
		S(41, -73), S(74, -57), S(47, -38), S(-55, -23), S(10, -46), S(-35, -20), S(53, -47), S(52, -76),
	},
	// Queen
	{
		S(-30, 24), S(-33, 45), S(-21, 66), S(7, 51), S(-5, 63), S(6, 52), S(38, 9), S(-1, 40),
		S(-2, 0), S(-20, 38), S(-10, 72), S(-26, 95), S(-47, 121), S(-1, 83), S(7, 55), S(54, 30),
		S(0, 11), S(0, 27), S(3, 62), S(9, 80), S(0, 99), S(38, 72), S(38, 48), S(42, 46),
		S(-17, 23), S(-10, 43), S(-3, 57), S(-5, 81), S(2, 88), S(13, 75), S(12, 66), S(15, 54),
		S(-11, 15), S(-14, 46), S(-12, 53), S(-4, 77), S(-3, 70), S(-4, 61), S(7, 42), S(8, 32),
		S(-13, 5), S(-3, 15), S(-9, 43), S(-10, 41), S(-8, 47), S(0, 37), S(13, 14), S(5, 5),
		S(-11, -9), S(-7, -4), S(3, -3), S(2, 8), S(0, 12), S(9, -12), S(14, -43), S(27, -64),
		S(-14, -5), S(-17, -12), S(-13, -5), S(-2, -2), S(-7, -12), S(-20, -16), S(-3, -26), S(0, -38),
	},
	// Rook
	{
		S(11, 49), S(7, 53), S(-1, 67), S(1, 65), S(15, 58), S(33, 43), S(36, 39), S(43, 38),
		S(22, 39), S(21, 51), S(38, 56), S(56, 48), S(37, 50), S(60, 38), S(41, 35), S(64, 23),
		S(2, 38), S(25, 40), S(26, 42), S(28, 41), S(52, 29), S(56, 22), S(79, 18), S(47, 17),
		S(-12, 38), S(2, 36), S(6, 46), S(17, 40), S(21, 27), S(24, 21), S(24, 18), S(21, 17),
		S(-30, 30), S(-30, 37), S(-16, 38), S(-2, 36), S(-2, 30), S(-17, 27), S(2, 17), S(-5, 11),
		S(-40, 27), S(-28, 27), S(-19, 26), S(-19, 31), S(-13, 25), S(-14, 16), S(18, -1), S(-5, 0),
		S(-43, 19), S(-28, 23), S(-13, 25), S(-16, 27), S(-10, 17), S(-8, 13), S(7, 1), S(-32, 12),
		S(-21, 15), S(-19, 25), S(-8, 32), S(-1, 30), S(3, 22), S(-7, 15), S(0, 15), S(-19, 5),
	},
	// Bishop
	{
		S(-32, -6), S(-55, 3), S(-58, 4), S(-93, 17), S(-85, 12), S(-73, 3), S(-47, -1), S(-52, -9),
		S(-15, -17), S(14, 0), S(5, 6), S(-11, 6), S(17, 1), S(10, 1), S(4, 7), S(-16, -13),
		S(0, 7), S(25, 3), S(25, 14), S(39, 8), S(24, 11), S(66, 9), S(35, 5), S(36, -2),
		S(-6, 2), S(8, 20), S(28, 17), S(38, 30), S(39, 21), S(29, 20), S(11, 14), S(-4, 2),
		S(-11, -3), S(1, 16), S(9, 24), S(30, 21), S(28, 18), S(11, 18), S(2, 11), S(-2, -16),
		S(0, -3), S(7, 6), S(6, 16), S(10, 15), S(11, 20), S(7, 14), S(8, -1), S(12, -12),
		S(2, -10), S(3, -8), S(15, -10), S(-8, 4), S(0, 8), S(14, -5), S(22, -3), S(5, -27),
		S(-13, -27), S(0, -9), S(-17, -28), S(-23, -9), S(-16, -14), S(-22, -10), S(2, -20), S(10, -42),
	},
	// Knight
	{
		S(-133, -63), S(-101, -23), S(-63, -1), S(-26, -15), S(7, -11), S(-61, -32), S(-71, -25), S(-85, -82),
		S(-19, -23), S(3, -5), S(42, -3), S(49, 1), S(34, -7), S(90, -19), S(14, -15), S(23, -39),
		S(0, -12), S(43, 1), S(44, 24), S(61, 23), S(91, 10), S(96, 0), S(59, -4), S(33, -26),
		S(3, -4), S(18, 17), S(44, 30), S(61, 35), S(48, 34), S(65, 30), S(30, 16), S(41, -12),
		S(-10, -3), S(5, 6), S(21, 31), S(23, 33), S(33, 34), S(28, 24), S(26, 10), S(1, -12),
		S(-30, -20), S(-6, 0), S(8, 10), S(12, 24), S(23, 22), S(13, 6), S(17, -6), S(-16, -14),
		S(-36, -24), S(-28, -16), S(-13, -4), S(-1, 0), S(0, 0), S(4, -7), S(-7, -24), S(-12, -13),
		S(-68, -7), S(-33, -40), S(-42, -23), S(-29, -19), S(-24, -19), S(-14, -27), S(-28, -32), S(-27, -17),
	},
	// Pawn
	{
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
		S(91, 179), S(96, 168), S(67, 165), S(114, 105), S(90, 106), S(74, 123), S(3, 169), S(-3, 186),
		S(23, 102), S(32, 106), S(63, 64), S(71, 23), S(83, 16), S(98, 36), S(76, 80), S(23, 84),
		S(-19, 55), S(5, 45), S(7, 26), S(11, 16), S(33, 7), S(23, 11), S(25, 31), S(0, 30),
		S(-29, 28), S(-3, 27), S(-4, 9), S(13, 6), S(13, 3), S(4, 5), S(11, 19), S(-12, 10),
		S(-32, 22), S(-7, 25), S(-8, 7), S(-7, 20), S(7, 13), S(-3, 10), S(27, 16), S(-4, 5),
		S(-32, 28), S(-6, 30), S(-13, 17), S(-25, 23), S(-4, 30), S(15, 16), S(37, 17), S(-12, 7),
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
	},
}
