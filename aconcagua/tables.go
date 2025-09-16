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
var MiddlegamePieceValue = [6]int{12000, 900, 500, 330, 320, 100}

// middlegamePieceValue is the value of each piece for middlegame
var EndgamePieceValue = [6]int{12000, 900, 500, 330, 320, 100}

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

/* piece/sq tables */
/* values from Rofchade: http://www.talkchess.com/forum3/viewtopic.php?f=2&t=68311&start=19 */

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-21, -31, -38, -40, -51, -39, -41, -35,
		-33, -48, -41, -48, -43, -38, -42, -26,
		-26, -44, -42, -55, -52, -39, -36, -33,
		-28, -41, -40, -55, -48, -32, -36, -28,
		-20, -23, -25, -39, -35, -25, -28, -19,
		-10, -14, -18, -20, -20, -20, -18, -9,
		21, 24, -3, -4, -1, 2, 22, 19,
		25, 30, 14, 0, -3, 8, 31, 18,
	},
	// Queen
	{
		-21, -12, -7, 0, -4, -1, -4, -25,
		-13, 9, 4, 7, 8, 8, 9, -6,
		-6, 2, 8, 4, 7, 6, -4, -12,
		-3, 3, 2, 8, 8, 2, 1, -9,
		4, 0, 8, 8, 7, 8, -5, -2,
		-12, 8, 11, 3, 11, 9, -2, -13,
		-9, 0, 2, -1, 1, 0, 4, -10,
		-25, -6, -10, -6, -12, -5, -15, -11,
	},
	// Rook
	{
		6, 0, -2, 3, 0, -1, -5, -3,
		5, 9, 10, 8, 14, 15, 16, 8,
		-1, 6, 2, -1, -4, -3, 4, -2,
		-1, 2, 0, 1, 3, -1, 2, -5,
		-3, 1, 3, 3, -2, -3, 1, -6,
		-7, 2, 3, -1, 0, 3, -2, 1,
		-9, 2, 2, -1, 1, -1, 7, -11,
		0, -1, -1, 5, 5, 1, -1, -1,
	},
	// Bishop
	{
		-17, -6, -9, -18, -16, -15, -14, -13,
		-9, 9, 4, 6, 1, 8, 8, 1,
		-21, -7, 11, 4, 10, 8, -9, -18,
		-5, -2, -3, 11, 10, 2, -2, -15,
		-1, 0, 7, 11, 11, 10, 0, -14,
		-9, 12, 9, 9, 7, 12, 12, -13,
		-8, 12, 2, 1, 2, -1, 6, -2,
		-21, -8, -10, -4, -11, -11, -13, -23,
	},
	// Knight
	{
		-47, -41, -32, -35, -33, -31, -52, -48,
		-34, -14, 6, -4, -1, 3, -25, -32,
		-28, 7, 14, 10, 12, 11, 2, -22,
		-24, 8, 19, 21, 20, 16, 5, -26,
		-24, 7, 19, 25, 22, 21, 1, -32,
		-31, 11, 9, 14, 8, 10, 8, -29,
		-42, -15, 3, 4, 5, 4, -14, -44,
		-42, -40, -25, -37, -33, -31, -46, -37,
	},
	// Pawn
	{
		-2, 1, 0, 0, 1, 0, -1, 0,
		56, 59, 62, 56, 58, 56, 60, 51,
		8, 10, 17, 31, 32, 23, 8, 9,
		5, 6, 11, 26, 24, 10, 4, 4,
		0, -1, 3, 19, 16, -1, -2, 0,
		5, -4, -7, 2, 1, -6, -3, 5,
		4, 8, 3, -21, -23, 7, 7, 5,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-46, -34, -28, -15, -21, -30, -35, -63,
		-38, -22, -15, -1, 12, -5, -22, -30,
		-27, -8, 25, 28, 31, 23, -3, -35,
		-26, -11, 35, 39, 37, 32, -9, -32,
		-35, -8, 34, 42, 44, 33, -10, -27,
		-25, -4, 16, 29, 33, 20, -8, -30,
		-27, -25, 4, 2, 3, 0, -29, -30,
		-45, -28, -28, -30, -33, -33, -32, -53,
	},
	// Queen
	{
		-23, -12, -9, -5, -4, -5, -3, -21,
		-7, 0, 2, 3, 10, 11, 7, -6,
		-3, 6, 10, 1, 6, 7, 11, -4,
		-5, -1, 7, 5, 8, 2, 3, -4,
		2, 8, 7, 5, 9, 11, -4, -9,
		-14, 4, 7, 5, 5, 0, -2, -18,
		-10, -4, -1, 4, 2, 1, 3, -16,
		-19, -10, -10, -9, -4, -11, -8, -14,
	},
	// Rook
	{
		3, -1, -2, -2, -1, -6, -6, 0,
		5, 9, 12, 10, 14, 11, 14, 7,
		-4, 3, 2, 3, -4, -1, 0, 1,
		-3, 2, 1, 5, 4, 5, 4, -1,
		-6, 8, 5, 3, 4, 5, -2, 1,
		-5, 1, 1, -3, 1, 4, -1, -4,
		-4, 4, -3, -1, 0, -1, 4, -3,
		0, 0, 0, 5, 3, -1, -5, -2,
	},
	// Bishop
	{
		-16, -3, -10, -18, -8, -11, -4, -14,
		-8, 3, 5, -6, 1, 12, 2, -11,
		-12, -5, 13, 13, 5, 6, -8, -4,
		-5, 6, -3, 10, 9, 8, 8, -9,
		-8, 0, 4, 9, 12, 11, -2, -12,
		-8, 8, 14, 9, 12, 9, 10, -3,
		-6, 2, -4, 2, 0, -3, 8, -8,
		-22, -7, -10, -10, -11, -10, -11, -17,
	},
	// Knight
	{
		-49, -34, -27, -27, -28, -31, -53, -49,
		-34, -14, 4, -7, -3, -2, -17, -36,
		-32, -3, 10, 18, 12, 6, 0, -29,
		-29, 9, 22, 19, 18, 20, 2, -29,
		-31, -1, 18, 21, 14, 13, -1, -39,
		-31, 4, 8, 17, 12, 9, 4, -31,
		-27, -19, 1, 4, 5, 4, -13, -44,
		-52, -42, -29, -26, -24, -33, -42, -46,
	},
	// Pawn
	{
		2, -1, 0, -1, -1, 0, 0, 0,
		55, 58, 65, 57, 52, 60, 57, 57,
		9, 7, 21, 28, 33, 24, 10, 7,
		4, 6, 9, 24, 24, 11, 4, 4,
		-1, 0, 0, 19, 19, 0, 0, -1,
		5, -5, -10, 0, 0, -9, -5, 4,
		5, 10, 9, -17, -18, 9, 10, 4,
		1, 1, 1, 0, -1, 0, 0, 0,
	},
}
