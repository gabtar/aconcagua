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
var MiddlegamePieceValue = [6]int{12000, 1032, 497, 397, 374, 88}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{12000, 978, 566, 322, 307, 90}

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-48, 4, 8, -13, -69, -35, -6, -11,
		21, 13, 1, 9, -8, 5, -20, -45,
		20, 22, 29, -13, -6, 28, 30, -7,
		2, -1, 12, -33, -36, -28, 2, -48,
		-61, 28, -42, -92, -91, -62, -53, -74,
		14, 10, -31, -59, -64, -51, -4, -26,
		25, 29, -11, -65, -47, -19, 27, 37,
		-5, 64, 36, -55, 20, -18, 61, 45,
	},
	// Queen
	{
		1, -7, 15, 6, 61, 52, 49, 56,
		-15, -47, -11, 1, -21, 54, 26, 60,
		2, -13, 15, -7, 33, 70, 58, 65,
		-30, -19, -20, -18, -4, 12, 10, 15,
		10, -30, 0, -9, -2, 1, 8, 9,
		-8, 18, -3, 11, 2, 10, 19, 17,
		-9, 6, 28, 22, 31, 31, 11, 27,
		19, 28, 32, 42, 22, -2, -4, -34,
	},
	// Rook
	{
		17, 27, -11, 41, 34, -21, -6, 10,
		24, 22, 62, 64, 78, 74, 16, 35,
		-11, 20, 21, 36, 6, 53, 68, 21,
		-27, -14, 5, 15, 20, 36, 0, -20,
		-50, -32, -13, -13, 1, -11, 13, -34,
		-49, -28, -18, -22, -1, -2, -6, -35,
		-43, -16, -18, -7, 6, 10, 2, -59,
		-13, -5, 11, 19, 23, 6, -20, -3,
	},
	// Bishop
	{
		-11, -11, -80, -65, -43, -45, -24, 8,
		-27, 12, -28, -26, 27, 48, 16, -44,
		-6, 35, 41, 22, 26, 50, 36, 11,
		8, 30, 16, 54, 34, 36, 29, 18,
		22, 23, 27, 31, 47, 14, 23, 37,
		31, 40, 29, 33, 31, 51, 31, 41,
		47, 46, 40, 25, 35, 41, 66, 36,
		-1, 41, 35, 23, 28, 29, -12, 16,
	},
	// Knight
	{
		-125, -88, -65, -61, 27, -89, -38, -97,
		-91, -61, 66, 2, -1, 45, -5, -20,
		-66, 34, -4, 24, 59, 81, 57, 40,
		-9, 1, -13, 39, 23, 43, 8, 25,
		-13, -4, -11, -4, 5, 4, 13, 0,
		-23, -32, -17, -23, -3, -12, 9, -18,
		-17, -53, -32, -2, -6, 5, -14, -1,
		-106, -3, -51, -30, 0, -24, 0, -13,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		45, 90, 11, 52, 36, 86, -15, -23,
		-1, -6, 15, 19, 60, 85, 28, 0,
		-21, -12, -2, 18, 16, 16, -4, -17,
		-27, -25, -2, 13, 19, 15, -2, -17,
		-19, -16, -4, 2, 16, 13, 31, 0,
		-16, -9, -12, -1, 3, 43, 40, 2,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-97, -51, -34, -31, -14, 6, -9, -29,
		-29, 2, 2, 4, 8, 26, 14, 3,
		-9, 5, 6, 6, 7, 35, 31, -3,
		-29, 8, 12, 22, 21, 28, 15, -5,
		-26, -21, 17, 30, 33, 23, 6, -15,
		-38, -17, 6, 19, 24, 16, -4, -19,
		-50, -29, -3, 8, 10, -1, -21, -43,
		-81, -66, -44, -19, -45, -25, -58, -83,
	},
	// Queen
	{
		8, 46, 44, 47, 44, 47, 37, 50,
		12, 31, 38, 50, 66, 39, 50, 52,
		11, 18, -6, 56, 47, 41, 48, 59,
		48, 32, 18, 26, 50, 45, 77, 75,
		3, 41, 12, 27, 20, 34, 64, 68,
		23, -26, 6, -4, 9, 20, 34, 50,
		8, -15, -25, -18, -18, -18, -25, -3,
		-15, -28, -26, -31, -3, -3, 7, -7,
	},
	// Rook
	{
		19, 15, 30, 15, 19, 30, 25, 17,
		17, 21, 10, 10, -3, 4, 21, 12,
		22, 18, 15, 13, 14, 2, 0, 5,
		22, 18, 25, 12, 13, 10, 10, 22,
		25, 23, 21, 17, 8, 9, 2, 12,
		16, 15, 5, 10, 1, 0, 5, 3,
		11, 5, 8, 9, -3, -2, -4, 14,
		-1, 7, 4, 0, -3, 0, 10, -23,
	},
	// Bishop
	{
		-8, -12, 1, 5, 7, 2, 0, -18,
		4, -2, 13, -3, -1, -7, -3, 0,
		8, -9, -7, -3, -6, -4, 1, 11,
		3, 1, 7, -1, 3, 3, -4, 7,
		-2, 0, 6, 11, -4, 5, -5, -5,
		-8, -2, 6, 7, 10, -8, -4, -8,
		-10, -20, -5, 0, 0, -9, -21, -26,
		-12, -2, -15, -1, -6, -12, 6, -7,
	},
	// Knight
	{
		-67, -36, 0, -24, -25, -26, -59, -92,
		-14, 2, -32, 7, -1, -29, -20, -47,
		-15, -18, 15, 12, -7, -6, -22, -42,
		-11, 5, 29, 20, 22, 10, 10, -15,
		-12, -3, 19, 30, 20, 20, 7, -15,
		-21, 4, 0, 19, 13, -1, -18, -17,
		-34, -12, -3, -8, -3, -17, -17, -45,
		-13, -50, -17, -7, -22, -15, -53, -68,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		105, 86, 74, 39, 52, 36, 84, 109,
		53, 49, 22, -15, -30, -10, 23, 34,
		32, 21, 7, -15, -7, -2, 14, 16,
		20, 14, 0, -9, -7, -6, 1, 4,
		7, 6, -1, -1, 0, -1, -10, -5,
		12, 3, 11, 0, 11, -2, -9, -10,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}
