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

// rayAttacks is a precalculated table that contains the rays on each direction for each square
var rayAttacks [8][64]Bitboard

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
		kingZone[White][sq] |= pawnAttacks(&from, White)
		kingZone[White][sq] |= from >> 8

		// Black
		kingZone[Black][sq] = kingAttacksTable[sq]
		kingZone[Black][sq] |= pawnAttacks(&from, Black)
		kingZone[Black][sq] |= from << 8
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
var MiddlegamePieceValue = [6]int{10000, 1071, 445, 351, 317, 74}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{10000, 1006, 565, 340, 330, 80}

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-55, 22, 18, -20, -48, -29, 10, 6,
		6, -1, -27, 8, 1, 3, -14, -18,
		-37, 22, -19, -18, -2, 31, 20, -21,
		-36, -44, -49, -71, -64, -54, -48, -76,
		-64, -39, -72, -91, -98, -74, -81, -103,
		-22, 3, -52, -69, -59, -59, -19, -48,
		52, 17, 3, -35, -36, -16, 32, 34,
		36, 54, 30, -53, 3, -33, 31, 46,
	},
	// Queen
	{
		-26, -23, 0, 10, 27, 35, 50, 10,
		0, -29, -26, -31, -32, 15, 3, 62,
		3, -4, -5, -3, 4, 54, 57, 57,
		-12, -7, -14, -14, -14, -1, 8, 8,
		-4, -14, -13, -8, -9, -10, 2, 12,
		-5, -2, -9, -11, -5, -3, 13, 15,
		4, 0, 4, 7, 4, 16, 25, 37,
		-4, 0, 3, 3, 6, -5, 4, -1,
	},
	// Rook
	{
		0, 0, -4, 2, 18, 23, 31, 40,
		-12, -15, 2, 15, 18, 26, 22, 47,
		-20, 6, 2, 1, 31, 38, 78, 42,
		-23, -11, -13, -4, -1, 6, 16, 5,
		-35, -36, -28, -19, -19, -29, -2, -17,
		-37, -33, -30, -28, -16, -14, 19, -1,
		-34, -31, -19, -16, -10, -2, 13, -20,
		-19, -15, -14, -8, 0, -3, 3, -13,
	},
	// Bishop
	{
		-27, -33, -59, -76, -61, -75, -18, -42,
		-10, 6, -4, -16, 7, 1, -14, -31,
		-2, 0, 3, 5, -2, 36, 18, 15,
		-14, 1, 0, 12, 12, 10, 7, -13,
		-1, -15, -7, 15, 13, -4, -5, 18,
		1, 10, 3, 4, 7, 5, 11, 24,
		23, 5, 19, -3, 3, 18, 27, 25,
		11, 26, 1, -5, 3, -4, 13, 31,
	},
	// Knight
	{
		-128, -91, -48, -31, 32, -75, -25, -76,
		-17, 10, 42, 51, 27, 84, 8, 25,
		0, 19, 30, 29, 72, 79, 43, 22,
		5, 7, 19, 45, 24, 48, 13, 43,
		-2, 2, 6, 16, 21, 23, 31, 13,
		-19, -8, -10, -3, 13, -7, 15, -1,
		-19, -15, -10, 5, 2, 6, 6, 4,
		-61, -13, -28, -13, -7, -3, -11, -12,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		74, 92, 55, 74, 40, 72, -17, -44,
		16, 5, 38, 37, 50, 93, 54, 26,
		-7, -5, 2, 6, 27, 30, 10, 9,
		-10, -14, 0, 13, 13, 13, 0, -2,
		-15, -9, -5, 1, 14, 2, 17, 0,
		-13, -9, -5, -6, 5, 21, 25, -9,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-75, -44, -28, 0, -1, 7, 2, -59,
		-18, 24, 29, 27, 40, 51, 52, 16,
		2, 30, 45, 52, 59, 58, 55, 23,
		-3, 31, 49, 59, 60, 60, 48, 20,
		-9, 15, 42, 57, 57, 45, 33, 13,
		-20, 1, 23, 37, 36, 27, 7, -2,
		-36, -13, 0, 12, 15, 4, -16, -32,
		-74, -59, -37, -13, -35, -20, -51, -83,
	},
	// Queen
	{
		18, 25, 38, 42, 29, 32, 1, 18,
		12, 40, 64, 77, 98, 51, 51, 38,
		17, 26, 53, 63, 75, 59, 30, 31,
		28, 33, 45, 61, 75, 66, 65, 53,
		11, 32, 32, 50, 50, 46, 36, 35,
		-3, 6, 23, 23, 27, 24, 4, 0,
		-10, -16, -13, -7, -2, -27, -52, -54,
		-13, -12, -14, -2, -16, -19, -19, -13,
	},
	// Rook
	{
		43, 45, 52, 43, 36, 40, 39, 34,
		46, 57, 57, 46, 39, 41, 39, 29,
		46, 43, 44, 41, 30, 25, 19, 19,
		48, 44, 49, 43, 32, 28, 30, 28,
		39, 41, 39, 35, 33, 35, 24, 24,
		29, 27, 25, 26, 19, 14, -3, 1,
		19, 23, 22, 19, 11, 6, -1, 7,
		19, 19, 27, 20, 12, 15, 9, 4,
	},
	// Bishop
	{
		4, 3, 6, 8, 5, 0, -3, -9,
		-8, -8, -4, -2, -13, -11, -1, -2,
		7, -3, 0, -8, -4, 0, -4, 4,
		3, 2, 0, 14, 4, 3, -2, 7,
		-5, 3, 11, 6, 5, 2, 1, -17,
		-3, 2, 4, 8, 10, 4, -7, -8,
		6, -12, -13, -3, 0, -11, -7, -12,
		-6, 4, -8, -5, -9, 3, -9, -17,
	},
	// Knight
	{
		-30, -18, -5, -11, -22, -23, -34, -63,
		3, -2, -14, -14, -20, -32, -7, -22,
		-5, -11, 0, 3, -15, -27, -24, -20,
		0, 0, 8, 8, 10, 3, 4, -10,
		-1, -5, 11, 8, 15, 2, -6, -3,
		-11, -8, -3, 11, 9, -7, -12, -7,
		-8, -4, -10, -10, -11, -14, -13, 3,
		7, -19, -11, -10, -7, -17, -9, -15,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		148, 130, 139, 86, 91, 87, 132, 157,
		82, 86, 43, 11, 2, 17, 61, 63,
		58, 46, 30, 10, 9, 16, 33, 33,
		39, 36, 24, 16, 15, 20, 24, 22,
		34, 30, 26, 25, 26, 27, 20, 19,
		36, 33, 29, 29, 37, 28, 19, 19,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}
