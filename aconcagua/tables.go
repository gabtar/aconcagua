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
var MiddlegamePieceValue = [6]int{12000, 1000, 469, 378, 358, 82}

// middlegamePieceValue is the value of each piece for middlegame
var EndgamePieceValue = [6]int{12023, 1000, 550, 309, 295, 87}

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
		-18, 29, 28, -5, -52, -28, 13, 4,
		31, 17, -4, 8, -10, 5, -15, -29,
		23, 10, 19, -18, -18, 13, 32, 6,
		9, -8, -5, -44, -48, -33, -3, -17,
		-48, 9, -37, -72, -70, -42, -28, -39,
		3, -2, -23, -46, -43, -36, -1, -16,
		23, 31, -4, -57, -38, -9, 34, 40,
		6, 65, 38, -47, 27, -18, 59, 50,
	},
	// Queen
	{
		-22, 14, 31, 29, 68, 55, 42, 55,
		-18, -37, -2, 19, 4, 58, 42, 55,
		-4, -6, 9, 10, 45, 73, 60, 63,
		-28, -24, -13, -18, 1, 18, 8, 7,
		0, -27, -9, -16, -6, -2, 6, 5,
		-15, 10, -12, 0, -8, 3, 16, 14,
		-23, -5, 20, 9, 16, 20, -1, 15,
		6, -3, 6, 27, -3, -21, -26, -44,
	},
	// Rook
	{
		20, 28, 20, 45, 44, 9, 23, 23,
		28, 35, 54, 57, 66, 69, 33, 31,
		-1, 20, 25, 37, 17, 45, 48, 19,
		-22, -12, 12, 24, 23, 37, -3, -8,
		-34, -27, -9, -4, 7, -5, 17, -18,
		-46, -23, -14, -16, 3, -1, -2, -31,
		-45, -12, -17, -7, -1, 8, -9, -67,
		-16, -10, 8, 16, 16, 4, -28, -13,
	},
	// Bishop
	{
		-25, -13, -86, -56, -27, -27, -14, 6,
		-29, 15, -20, -16, 36, 55, 29, -46,
		-18, 28, 43, 42, 38, 57, 40, 2,
		-6, 8, 21, 52, 46, 43, 12, 3,
		-2, 19, 17, 32, 41, 14, 15, 12,
		6, 24, 20, 20, 17, 37, 19, 15,
		18, 23, 26, 6, 16, 26, 40, 14,
		-33, 6, -2, -12, -9, -6, -28, -17,
	},
	// Knight
	{
		-149, -88, -48, -48, 21, -93, -42, -106,
		-96, -54, 65, 22, 13, 58, 0, -24,
		-63, 50, 27, 54, 80, 100, 69, 33,
		-20, 10, 10, 53, 36, 66, 11, 16,
		-22, -2, 8, 7, 22, 15, 17, -14,
		-32, -18, 5, 4, 14, 13, 18, -24,
		-36, -61, -20, -6, -4, 11, -25, -18,
		-113, -26, -67, -46, -19, -33, -23, -32,
	},
	// Pawn
	{
		10, 6, -2, -3, -2, -3, 4, -4,
		83, 101, 36, 65, 62, 86, 1, -4,
		0, 0, 14, 12, 52, 72, 28, -8,
		-15, 11, 4, 22, 23, 23, 12, -20,
		-25, -9, -4, 12, 17, 7, 4, -24,
		-20, -9, -4, -9, 4, 7, 29, -8,
		-28, -6, -20, -20, -13, 27, 34, -15,
		15, 2, -1, 1, -3, 1, 1, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-92, -47, -30, -27, -14, 8, -5, -24,
		-21, 9, 7, 10, 14, 28, 16, 4,
		-2, 13, 15, 11, 11, 38, 35, -2,
		-24, 15, 19, 27, 26, 30, 19, -9,
		-23, -12, 20, 28, 30, 21, 3, -20,
		-30, -10, 8, 17, 21, 15, -2, -19,
		-44, -24, -1, 8, 10, -1, -18, -38,
		-75, -58, -41, -20, -48, -23, -50, -74,
	},
	// Queen
	{
		20, 43, 46, 47, 52, 46, 41, 53,
		2, 33, 48, 58, 68, 50, 49, 45,
		-1, 19, 27, 67, 63, 58, 48, 48,
		29, 43, 38, 63, 74, 63, 78, 68,
		3, 48, 41, 69, 53, 51, 58, 55,
		20, -10, 37, 23, 36, 39, 33, 33,
		5, -2, -14, 1, 0, -3, -14, -8,
		-10, -10, -8, -31, 11, -8, 5, -15,
	},
	// Rook
	{
		24, 20, 26, 22, 22, 23, 18, 17,
		21, 20, 19, 20, 6, 10, 17, 15,
		19, 19, 17, 16, 13, 5, 6, 5,
		19, 18, 21, 11, 12, 9, 9, 15,
		19, 20, 21, 15, 6, 7, -1, 4,
		14, 13, 4, 8, -1, -2, 3, 0,
		9, 6, 8, 10, -1, -2, 0, 10,
		-3, 9, 7, 5, 3, 0, 12, -22,
	},
	// Bishop
	{
		-6, -13, 2, 2, 3, -5, -4, -22,
		2, -2, 10, -6, -4, -11, -8, -5,
		7, -5, -4, -7, -7, -3, 0, 8,
		4, 8, 7, 2, 3, 2, -1, 6,
		1, 0, 10, 14, -1, 5, -5, -3,
		-6, -1, 6, 9, 12, -6, -1, -6,
		-12, -17, -7, -1, 4, -8, -14, -26,
		-11, -1, -22, 1, -2, -11, 4, -7,
	},
	// Knight
	{
		-66, -37, -7, -27, -23, -28, -64, -96,
		-18, -3, -32, -1, -8, -32, -29, -50,
		-19, -23, 5, 4, -11, -13, -25, -44,
		-14, 2, 22, 16, 18, 5, 6, -19,
		-14, -7, 14, 25, 13, 15, 4, -18,
		-22, -4, -10, 10, 6, -10, -24, -22,
		-38, -14, -12, -10, -9, -21, -19, -50,
		-23, -56, -20, -9, -23, -22, -52, -68,
	},
	// Pawn
	{
		11, 3, 2, 7, 1, 1, 0, 4,
		128, 116, 100, 73, 80, 72, 111, 139,
		74, 71, 51, 21, 7, 19, 48, 59,
		23, 8, 1, -19, -9, -6, 6, 13,
		9, 2, -7, -15, -12, -7, -6, 0,
		-2, -3, -9, -5, -1, -4, -16, -8,
		11, -2, 13, 5, 12, 0, -11, -6,
		21, 1, 0, 1, 1, -1, -1, 3,
	},
}
