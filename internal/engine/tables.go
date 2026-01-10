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
	rayAttacks = generateRayAttacks()
	knightAttacksTable = generateKnightAttacks()
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

// rayAttacks is a precalculated table that contains the rays on each direction for each square
var rayAttacks [8][64]Bitboard

// knightAttacksTable is a precalculated table that contains the squares that a knight can attack
var knightAttacksTable [64]Bitboard

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

var KingZone [64]Bitboard

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

// generateKnightAttacks returns a precalculated array for all posible knight moves from each square in the board
func generateKnightAttacks() (knightAttacksTable [64]Bitboard) {
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

// generateAttacksFrontSpans returns a precalculated table containing the front attack spans for each square
func generateAttacksFrontSpans() (attacksFrontSpans [2][64]Bitboard) {

	for sq := range 64 {
		file, rank := sq%8, sq/8
		eastFront, westFront := rank*8+file+1, rank*8+file-1

		if file < 7 {
			attacksFrontSpans[White][sq] |= rayAttacks[North][eastFront]
			attacksFrontSpans[Black][sq] |= rayAttacks[South][eastFront]
		}
		if file > 0 {
			attacksFrontSpans[White][sq] |= rayAttacks[North][westFront]
			attacksFrontSpans[Black][sq] |= rayAttacks[South][westFront]
		}
	}

	return
}

// generateKingZone returns a precalculated table containing the king zone for each square
func generateKingZone() (kingZone [64]Bitboard) {
	for sq := range 64 {
		from := bitboardFromIndex(sq)
		kingZone[sq] = kingAttacks(&from) | from
	}
	return
}

// raysDirection returns the rays along the direction passed that intersects the
// piece in the square passed
func raysDirection(square Bitboard, direction uint64) Bitboard {
	oppositeDirections := [8]uint64{South, SouthWest, West, NorthWest, North, NorthEast, East, SouthEast}

	return rayAttacks[direction][Bsf(square)] | square |
		rayAttacks[oppositeDirections[direction]][Bsf(square)]
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

	return rayAttacks[fromDirection][fromSq] &
		rayAttacks[toDirection][toSq]
}

// pieces square tables with the value of each piece + square value for middlegame
var middlegamePiecesScore [12][64]int

// pieces square tables with the value of each piece + square value for endgame
var endgamePiecesScore [12][64]int

// MiddlegamePieceValue is the value of each piece for middlegame phase
var MiddlegamePieceValue = [6]int{12000, 1050, 470, 374, 362, 83}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{12000, 1000, 566, 313, 307, 89}

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-42, 3, 12, -12, -63, -33, 0, -16,
		7, 9, -6, 22, 0, 11, -9, -37,
		2, 33, 18, -2, 10, 42, 39, -3,
		-9, -10, -4, -51, -43, -36, -12, -66,
		-56, 7, -58, -98, -94, -52, -60, -95,
		-2, 21, -38, -52, -44, -47, -10, -48,
		46, 24, 8, -41, -30, -16, 35, 33,
		19, 75, 44, -57, 9, -30, 48, 44,
	},
	// Queen
	{
		-18, -20, 0, 9, 41, 36, 43, 36,
		1, -30, -13, -15, -16, 34, 12, 63,
		8, -4, 17, -3, 25, 59, 64, 52,
		-12, -2, -1, -5, 4, 16, 25, 23,
		11, -10, -1, 4, 5, 4, 18, 21,
		3, 14, 5, 7, 9, 19, 27, 21,
		9, 12, 23, 25, 22, 33, 29, 30,
		7, 11, 21, 28, 22, 7, 9, -12,
	},
	// Rook
	{
		3, 11, -19, 18, 17, -11, 1, 21,
		2, 0, 34, 41, 50, 53, 26, 35,
		-15, 15, 10, 14, 25, 43, 80, 38,
		-20, -1, -7, 0, 6, 19, 16, 4,
		-33, -31, -23, -17, -12, -24, 4, -15,
		-32, -29, -24, -20, -11, -6, 17, -11,
		-31, -25, -14, -14, -6, 1, 17, -34,
		-8, -7, -5, 3, 9, 6, 0, -5,
	},
	// Bishop
	{
		-7, -27, -66, -71, -59, -60, -28, -9,
		-9, 6, -7, -24, 16, 24, 0, -20,
		3, 22, 22, 23, 25, 46, 45, 27,
		-3, 19, 16, 42, 23, 30, 19, 5,
		8, 2, 11, 33, 31, 7, 4, 23,
		11, 23, 18, 21, 22, 26, 19, 35,
		30, 24, 35, 11, 19, 30, 45, 24,
		7, 35, 15, 6, 10, 11, 11, 23,
	},
	// Knight
	{
		-144, -100, -75, -52, 13, -86, -43, -108,
		-66, -37, 38, 10, 12, 51, -12, 0,
		-43, 11, 15, 29, 55, 66, 37, 21,
		-4, -9, 3, 35, 12, 31, -2, 27,
		-14, -17, -8, 3, 4, 5, 11, -1,
		-32, -27, -25, -20, -5, -20, -5, -14,
		-39, -36, -28, -15, -15, -12, -17, -14,
		-91, -27, -42, -28, -22, -18, -24, -33,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		53, 80, 35, 69, 23, 72, -32, -35,
		13, 3, 34, 36, 49, 92, 50, 21,
		-14, -13, -2, 3, 19, 31, 11, 4,
		-21, -19, -5, 5, 8, 16, 2, -4,
		-23, -16, -9, -2, 11, 10, 21, -1,
		-19, -16, -6, -10, 3, 34, 36, -4,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-89, -57, -41, -18, -16, -7, -6, -51,
		-42, -5, -1, 1, 11, 25, 23, 0,
		-25, 1, 12, 24, 27, 33, 26, 0,
		-27, 2, 19, 31, 34, 33, 19, 0,
		-32, -14, 18, 39, 39, 23, 10, -4,
		-44, -19, 3, 18, 17, 10, -8, -19,
		-47, -28, -16, 0, 0, -6, -25, -41,
		-73, -69, -50, -31, -49, -31, -59, -90,
	},
	// Queen
	{
		24, 38, 55, 61, 32, 39, 22, 33,
		23, 40, 57, 72, 87, 36, 38, 47,
		23, 30, 19, 69, 56, 41, 27, 45,
		44, 33, 31, 45, 59, 58, 64, 54,
		7, 42, 28, 38, 38, 45, 42, 49,
		7, -1, 16, 17, 26, 16, 13, 27,
		5, -12, -18, -10, 1, -30, -26, -23,
		-9, -9, -17, -6, -11, -7, -4, 8,
	},
	// Rook
	{
		21, 20, 37, 17, 20, 28, 29, 19,
		16, 25, 17, 9, 2, 4, 10, 7,
		22, 18, 17, 15, 14, 0, -1, 1,
		25, 19, 27, 22, 7, 0, 8, 8,
		19, 22, 20, 16, 13, 13, 7, 7,
		11, 12, 5, 7, 3, -5, -12, -3,
		1, 6, 4, 4, -7, -9, -18, 1,
		3, 3, 9, 3, -4, -1, 4, -11,
	},
	// Bishop
	{
		-1, 2, 11, 13, 4, 0, 5, -20,
		-10, -7, -1, 0, -15, -20, -3, -5,
		8, -7, -9, -14, -14, -10, -12, 4,
		2, -1, -6, -3, -7, -3, -3, -3,
		-7, 1, 0, -6, -5, -1, 1, -18,
		-5, -7, -2, 2, 5, -7, -9, -17,
		-8, -22, -18, -3, -3, -10, -16, -21,
		-21, -3, -15, -6, -10, -3, -6, -28,
	},
	// Knight
	{
		-77, -29, 0, -12, -21, -20, -41, -85,
		1, 6, -24, -1, -12, -28, -12, -31,
		1, -7, 5, 0, -8, -20, -23, -30,
		-4, 8, 20, 14, 15, 11, 6, -12,
		-4, 2, 19, 14, 23, 9, -2, -10,
		-20, -7, -1, 11, 12, -8, -10, -16,
		-29, -6, -8, -7, -8, -8, -20, -24,
		-26, -47, -19, -18, -15, -22, -36, -48,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		90, 67, 78, 28, 33, 23, 73, 94,
		54, 58, 15, -20, -25, -9, 36, 39,
		35, 24, 7, -13, -9, -6, 10, 11,
		17, 13, 1, -6, -7, -3, 0, 0,
		11, 6, 2, 2, 2, 1, -3, -2,
		14, 11, 6, 6, 12, 1, -6, -3,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}
