package aconcagua

import "math"

// Constants for orthogonal directions in the board
const (
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

// directions is a table that contains the compass directions between 2 squares in the board
var directions [64][64]uint64

// rayAttacks is a precalculated table that contains the rays on each direction for each square
var rayAttacks [8][64]Bitboard

// knightAttacksTable is a precalculated table that contains the squares that a knight can attack
var knightAttacksTable [64]Bitboard

// pieces square tables with the value of each piece + square value for middlegame
var middlegamePiecesScore [12][64]int

// pieces square tables with the value of each piece + square value for endgame
var endgamePiecesScore [12][64]int

// MiddlegamePieceValue is the value of each piece for middlegame
var MiddlegamePieceValue = [6]int{12000, 935, 471, 365, 345, 82}

// middlegamePieceValue is the value of each piece for middlegame
var EndgamePieceValue = [6]int{12011, 935, 528, 307, 293, 86}

// initializes various tables for usage within the engine
func init() {
	GeneratePiecesScoreTables()

	directions = generateDirections()
	rayAttacks = generateRayAttacks()
	knightAttacksTable = generateKnightAttacks()
	attacksFrontSpans = generateAttacksFrontSpans()
}

// GeneratePiecesScoreTables generates the tables with the value of each piece + square
func GeneratePiecesScoreTables() {
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
			// calcuate the direction
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

		notInHFile := from & ^(from & files[7])
		notInAFile := from & ^(from & files[0])
		notInABFiles := from & ^(from & (files[0] | files[1]))
		notInGHFiles := from & ^(from & (files[7] | files[6]))

		knightAttacksTable[sq] = notInAFile<<15 | notInHFile<<17 | notInGHFiles<<10 |
			notInABFiles<<6 | notInHFile>>15 | notInAFile>>17 |
			notInABFiles>>10 | notInGHFiles>>6

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

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-25, -8, -13, -29, -55, -18, -10, -5,
		-4, -6, -16, -23, -23, -8, -13, -16,
		2, -9, -10, -39, -35, -10, -5, 1,
		-8, -14, -29, -49, -52, -36, -9, -15,
		-35, -7, -39, -63, -65, -42, -29, -36,
		-6, -7, -25, -41, -44, -36, -3, -13,
		20, 29, -8, -35, -35, -10, 33, 40,
		5, 56, 35, -35, 26, -18, 59, 48,
	},
	// Queen
	{
		-29, 16, 22, 22, 30, 25, 25, 15,
		-24, -34, -3, 24, 20, 35, 35, 25,
		-12, -11, 6, 14, 38, 40, 35, 25,
		-26, -26, -15, -16, 9, 24, 17, 9,
		-6, -29, -13, -15, -7, -1, 9, 5,
		-19, 4, -13, -5, -12, -1, 11, 9,
		-27, -10, 15, 2, 10, 14, -10, 7,
		0, -8, 0, 22, -8, -23, -32, -41,
	},
	// Rook
	{
		20, 22, 18, 30, 29, 13, 19, 16,
		21, 31, 42, 41, 44, 45, 29, 24,
		5, 17, 19, 28, 16, 30, 32, 15,
		-18, -11, 11, 17, 21, 31, -4, -7,
		-32, -23, -7, -4, 9, -1, 13, -11,
		-39, -20, -14, -16, 1, -2, -6, -29,
		-39, -10, -15, -5, -2, 9, -11, -40,
		-16, -10, 8, 16, 16, 4, -28, -13,
	},
	// Bishop
	{
		-38, -7, -44, -39, -21, -24, -9, -11,
		-37, 8, -20, -15, 26, 32, 23, -43,
		-22, 22, 35, 32, 32, 35, 31, 1,
		-8, 5, 22, 44, 42, 36, 10, 1,
		-2, 16, 16, 30, 38, 14, 10, 10,
		4, 21, 20, 19, 17, 34, 20, 10,
		11, 21, 23, 6, 14, 24, 39, 10,
		-26, 6, -3, -12, -10, -7, -30, -19,
	},
	// Knight
	{
		-85, -66, -46, -41, 0, -64, -44, -85,
		-75, -48, 35, 20, 7, 33, 3, -25,
		-53, 35, 27, 49, 50, 45, 35, 5,
		-24, 8, 9, 51, 33, 50, 13, 3,
		-23, -1, 10, 4, 22, 14, 17, -14,
		-32, -18, 5, 2, 16, 13, 18, -25,
		-35, -53, -20, -5, -4, 12, -26, -19,
		-84, -26, -58, -41, -18, -31, -23, -35,
	},
	// Pawn
	{
		5, 1, 0, 1, 0, -1, 2, -2,
		85, 85, 59, 68, 61, 81, 37, 53,
		33, 31, 19, 12, 43, 55, 33, 7,
		-12, 11, 4, 18, 22, 22, 10, -20,
		-24, -9, -4, 11, 17, 7, 3, -24,
		-20, -9, -4, -10, 4, 7, 29, -8,
		-27, -6, -20, -20, -14, 27, 34, -15,
		7, 1, 1, 1, -1, 0, 0, 2,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-77, -34, -22, -23, -11, 1, -5, -20,
		-11, 13, 12, 14, 15, 25, 14, 0,
		3, 16, 20, 11, 13, 41, 25, -2,
		-20, 17, 21, 26, 24, 28, 18, -9,
		-26, -10, 19, 25, 27, 20, 3, -20,
		-28, -10, 5, 15, 19, 14, -2, -19,
		-43, -25, 0, 2, 8, -1, -18, -38,
		-75, -55, -39, -25, -47, -23, -50, -73,
	},
	// Queen
	{
		13, 24, 25, 29, 30, 25, 25, 15,
		-7, 11, 29, 32, 34, 35, 33, 25,
		-6, 7, 15, 40, 40, 40, 35, 25,
		7, 19, 20, 39, 40, 38, 35, 30,
		-2, 27, 29, 40, 39, 35, 35, 30,
		15, -12, 23, 17, 25, 33, 26, 21,
		-5, -7, -15, 1, -1, -6, -12, -10,
		-10, -14, -10, -27, 5, -20, -1, -27,
	},
	// Rook
	{
		24, 22, 25, 23, 24, 20, 19, 18,
		24, 21, 21, 22, 13, 15, 17, 17,
		16, 20, 18, 18, 12, 7, 10, 5,
		17, 16, 21, 13, 11, 10, 7, 13,
		16, 19, 18, 14, 4, 3, -1, 0,
		7, 10, 3, 8, -1, -4, 3, -1,
		1, 3, 6, 8, -1, -4, 1, -5,
		-3, 9, 7, 5, 3, 0, 10, -22,
	},
	// Bishop
	{
		-8, -18, -12, -6, -2, -7, -7, -22,
		1, -3, 5, -10, -4, -8, -12, -11,
		7, -6, -6, -8, -9, -2, -2, 5,
		3, 8, 4, 1, 0, 2, -1, 3,
		-3, -2, 8, 11, -2, 2, -6, -4,
		-6, 0, 3, 9, 11, -6, -5, -4,
		-13, -17, -9, -1, 5, -10, -12, -28,
		-20, -4, -21, 0, -4, -10, -1, -10,
	},
	// Knight
	{
		-82, -45, -13, -31, -22, -46, -63, -84,
		-26, -8, -22, -5, -11, -27, -32, -52,
		-26, -20, 3, 4, -2, 3, -17, -36,
		-15, 1, 18, 14, 16, 7, 3, -16,
		-15, -9, 10, 22, 10, 12, 2, -18,
		-20, -5, -10, 9, 1, -10, -23, -21,
		-42, -20, -15, -12, -11, -24, -20, -45,
		-42, -54, -26, -14, -23, -25, -50, -66,
	},
	// Pawn
	{
		5, 0, 1, 1, -1, -2, 1, 1,
		85, 85, 85, 65, 71, 67, 85, 85,
		45, 45, 47, 18, 7, 22, 44, 45,
		21, 8, 1, -10, -9, -5, 6, 13,
		8, 2, -7, -14, -12, -7, -6, 0,
		-2, -3, -9, -3, -1, -4, -16, -8,
		10, -2, 12, 4, 12, 0, -11, -6,
		13, 2, 1, 0, 0, -1, -2, 2,
	},
}
