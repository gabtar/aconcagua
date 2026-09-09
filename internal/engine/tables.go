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
var MiddlegamePieceValue = [6]int{10000, 954, 434, 347, 309, 75}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{10000, 1115, 605, 358, 342, 102}

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
