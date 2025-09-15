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

// middlegamePieceValue is the value of each piece for middlegame
var middlegamePieceValue = [6]int{12000, 1025, 477, 365, 337, 82}

// middlegamePieceValue is the value of each piece for middlegame
var endgamePieceValue = [6]int{12000, 936, 512, 297, 281, 94}

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

			middlegamePiecesScore[whitePiece][sq] = middlegamePieceValue[piece] + MiddlegamePSQT[piece][sq^56]
			endgamePiecesScore[whitePiece][sq] = endgamePieceValue[piece] + EndgamePSQT[piece][sq^56]

			middlegamePiecesScore[blackPiece][sq] = middlegamePieceValue[piece] + MiddlegamePSQT[piece][sq]
			endgamePiecesScore[blackPiece][sq] = endgamePieceValue[piece] + EndgamePSQT[piece][sq]
		}
	}
}

// SetPsqt updates the middlegamePiecesScore and endgamePiecesScore according to the new value
func SetPsqt(index int, value int) {
	square := index % 64
	piece := pieceRole(index / 64)

	if index >= 384 {
		EndgamePSQT[piece][square] = value

		endgamePiecesScore[piece][square^56] = endgamePieceValue[piece] + value
		endgamePiecesScore[piece+6][square] = endgamePieceValue[piece] + value
	} else {
		MiddlegamePSQT[piece][square] = value

		middlegamePiecesScore[piece][square^56] = middlegamePieceValue[piece] + value
		middlegamePiecesScore[piece+6][square] = middlegamePieceValue[piece] + value
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
		-76, 12, -1, -39, -50, -42, 1, -27,
		48, -4, -27, -11, 14, 11, -34, -50,
		4, 57, 34, -1, 6, 12, 32, -22,
		3, 10, 10, -10, 5, 11, 23, -13,
		-20, 39, -2, -3, -6, -4, 8, -26,
		9, 27, 19, -6, -3, 11, 26, -1,
		40, 27, 30, -61, -3, 25, 48, 5,
		-10, 28, -6, -14, -33, 13, -14, -27,
	},
	// Queen
	{
		13, -41, 15, -29, 37, 24, 10, 86,
		17, 2, 36, 7, 20, 98, 69, 92,
		18, 24, 48, -8, 26, 47, 6, 16,
		10, -12, 0, 25, 40, -24, -28, -40,
		-50, -30, 31, 31, 39, 2, -37, -44,
		-41, 43, 30, -18, 26, 26, 24, -33,
		-25, 25, 20, 15, -1, 28, 33, -11,
		24, -57, -10, -31, -56, 0, -28, -19,
	},
	// Rook
	{
		43, 17, -9, 10, 22, -32, -7, 4,
		68, 71, 96, 85, 100, 80, 33, 42,
		36, 60, 55, 57, 14, 22, 32, 47,
		17, 29, 48, 47, 28, 53, 21, 20,
		5, 12, 29, 34, 33, 19, 26, 18,
		-9, 12, 23, -21, -3, 13, 15, -7,
		-37, 19, 11, 22, 38, 51, 23, -77,
		-26, -15, 21, 23, 30, 14, -47, -67,
	},
	// Bishop
	{
		-65, -37, -121, -78, -66, -83, -34, 13,
		-8, 57, 23, -54, 37, 73, 59, -10,
		-57, 48, 14, 26, -3, 9, -4, -43,
		-7, -36, 40, 88, 72, -4, -34, -43,
		28, 23, 45, 64, 26, 4, 4, 0,
		41, 40, 56, -2, 37, 68, 34, 34,
		7, 56, 12, 39, 48, 61, 74, 28,
		8, 25, 3, 20, 27, 19, -3, -6,
	},
	// Knight
	{
		-126, -80, -75, -83, 21, -138, -56, -67,
		-32, 0, 113, 4, -18, 49, -34, 21,
		-8, 101, 78, 106, 109, 88, 33, 46,
		29, 58, 60, 94, 42, 110, 33, 57,
		25, 45, 57, 54, 68, 60, 60, -48,
		-36, 32, 25, 51, 49, 26, 62, -37,
		-57, -15, 29, 10, 38, 41, 27, 0,
		-93, -62, -17, 8, -9, 12, -60, -38,
	},
	// Pawn
	{
		2, 0, 2, -1, 1, -2, 0, -1,
		139, 175, 102, 136, 109, 167, 75, 30,
		-47, -34, -15, -10, 24, 15, -16, -61,
		-55, 54, -21, 49, 64, 10, 58, -50,
		-68, -43, 36, 11, -6, 26, -31, -58,
		-36, 2, 37, 4, 39, 44, 33, -15,
		-76, -24, -61, -64, -56, -17, -3, -63,
		0, 1, 1, 2, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-103, -79, -69, -66, -36, -39, -33, -71,
		-39, -36, -20, -33, 16, 88, 62, -42,
		42, 69, 73, -9, 23, 60, 74, -20,
		35, 75, 59, 47, 80, 87, 80, 32,
		19, 50, 49, 76, 77, 77, 63, 35,
		8, 51, 65, 65, 76, 70, 61, 9,
		22, 42, 48, 66, 53, 45, -4, -70,
		-66, -73, -38, -11, -77, -42, -78, -97,
	},
	// Queen
	{
		-27, -32, -25, -27, -26, -35, -44, -34,
		35, 74, 74, 38, 42, 79, 44, -46,
		33, 60, 63, 20, 5, 73, 27, -45,
		41, 69, 64, 86, 5, -12, 3, 19,
		-15, 55, -13, 69, -17, 15, -2, -31,
		-57, -17, 31, -46, -3, 21, -28, -44,
		-29, -55, -76, -52, -23, -1, -52, -71,
		-58, -55, -44, -97, -59, -58, -40, -83,
	},
	// Rook
	{
		3, -11, -36, -39, -42, -42, -46, -33,
		65, 61, 54, 18, -30, -7, 14, 32,
		61, 61, 61, 30, -11, 24, 29, 51,
		58, 57, 67, 43, 32, 48, 53, 56,
		57, 59, 62, 51, 24, 44, 21, 43,
		50, 53, 42, 18, 22, 34, 10, 38,
		48, 23, 47, 36, 39, 45, 30, 51,
		25, 53, 9, -30, -43, -64, 0, 27,
	},
	// Bishop
	{
		-68, -36, 40, -21, 22, -62, -70, -75,
		-10, 47, 59, 40, -17, -31, -5, -63,
		-8, 37, 31, -7, -41, -47, -9, -50,
		26, 54, 11, 40, 9, -22, -5, -15,
		18, -10, 18, 33, 24, 4, -14, 7,
		42, 51, 62, 40, 54, 52, 46, 24,
		40, 33, 37, 46, -1, 44, 37, 12,
		31, 35, 18, 49, 45, 38, 30, 21,
	},
	// Knight
	{
		-88, 2, -64, -42, -81, -81, -117, -150,
		5, 44, 26, -35, -63, -79, -72, -97,
		30, 14, 60, 23, -55, -63, -73, -28,
		34, 55, 75, -1, 9, 15, 6, -44,
		34, 48, 70, 60, 11, 69, 54, -1,
		30, 51, 47, 68, 37, 8, 26, 31,
		12, 34, 44, 47, 10, 33, 31, -2,
		15, -68, 30, 38, 23, 11, -73, -20,
	},
	// Pawn
	{
		5, -1, 1, 1, 1, 0, 1, 1,
		232, 227, 212, 188, 201, 186, 219, 241,
		40, 46, 31, 13, 2, -1, 28, 30,
		-22, -30, -41, -49, -52, -50, -37, -37,
		-41, -45, -57, -61, -59, -60, -51, -54,
		-50, -46, -60, -21, -54, -59, -55, -61,
		-41, -46, -45, -25, -41, -54, -52, -61,
		6, -2, 0, 1, 1, 0, 1, -1,
	},
}
