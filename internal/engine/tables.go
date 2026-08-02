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
var MiddlegamePieceValue = [6]int{10000, 933, 426, 344, 307, 59}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{10000, 1097, 608, 365, 351, 80}

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-21, 45, 52, -18, -23, -27, 5, 21,
		-43, 8, -9, 62, 25, 11, -1, -36,
		-93, 35, 0, -6, 22, 64, -10, -46,
		-66, -39, -54, -70, -76, -74, -93, -159,
		-71, -33, -51, -76, -80, -56, -83, -157,
		-35, 7, -32, -39, -30, -36, -5, -56,
		42, 12, 6, -18, -26, -2, 21, 28,
		32, 54, 35, -31, 16, -20, 36, 36,
	},
	// Queen
	{
		-48, -53, -46, -13, -33, -12, 22, -25,
		5, -17, -26, -51, -72, -30, -12, 45,
		7, 0, -5, -12, -28, 13, 13, 15,
		-5, -3, -9, -13, -13, -8, 2, 0,
		1, -7, -7, -2, -2, -7, 8, 8,
		1, 5, 0, -3, 2, 4, 20, 15,
		11, 11, 13, 14, 13, 22, 34, 48,
		1, 7, 10, 10, 14, 1, 16, 22,
	},
	// Rook
	{
		-10, -13, -21, -26, -7, 24, 29, 30,
		1, -1, 14, 34, 11, 39, 24, 45,
		-11, 13, 8, 8, 33, 40, 60, 27,
		-16, -1, -1, 9, 7, 17, 15, 2,
		-25, -26, -15, -7, -5, -21, 0, -15,
		-30, -23, -21, -16, -7, -5, 19, 0,
		-26, -22, -11, -9, -4, 2, 17, -14,
		-13, -9, -8, -1, 5, 0, 3, -9,
	},
	// Bishop
	{
		-29, -61, -57, -99, -92, -85, -52, -59,
		-9, 2, -2, -18, 4, -11, -21, -34,
		1, 5, 4, 14, 2, 48, 13, 23,
		-13, 1, 7, 12, 17, 5, 5, -18,
		0, -8, -7, 18, 10, -2, -6, 17,
		2, 9, 6, 2, 6, 8, 12, 19,
		16, 7, 16, -3, 5, 14, 29, 22,
		11, 18, 0, -5, 4, -4, 16, 38,
	},
	// Knight
	{
		-109, -97, -67, -29, 0, -61, -68, -70,
		0, 13, 37, 42, 24, 72, 12, 21,
		7, 24, 20, 35, 66, 69, 34, 18,
		10, 10, 24, 33, 26, 38, 21, 41,
		2, 8, 12, 20, 20, 23, 15, 14,
		-14, -2, -6, 2, 15, -2, 16, 3,
		-15, -11, -5, 6, 5, 8, 13, 9,
		-45, -11, -22, -9, -3, 1, -8, -4,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		72, 75, 45, 93, 69, 52, -19, -16,
		25, 12, 41, 48, 61, 90, 63, 30,
		-5, -5, -1, 2, 20, 19, 2, 5,
		-6, -12, -3, 7, 10, 6, -4, -3,
		-17, -14, -15, -6, 0, -8, 2, -5,
		-10, -7, -9, -16, -5, 9, 18, -9,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-86, -45, -29, 6, -2, 0, -10, -102,
		-11, 28, 34, 27, 42, 55, 49, 11,
		10, 31, 50, 59, 64, 57, 62, 22,
		-1, 34, 56, 68, 70, 66, 54, 29,
		-12, 18, 45, 62, 60, 44, 30, 14,
		-24, 0, 26, 40, 37, 25, 0, -13,
		-50, -13, 4, 17, 19, 3, -21, -49,
		-87, -63, -35, -8, -31, -19, -57, -97,
	},
	// Queen
	{
		19, 24, 44, 29, 38, 32, -8, 25,
		10, 30, 59, 77, 98, 63, 43, 30,
		25, 35, 61, 61, 78, 48, 27, 26,
		36, 50, 60, 72, 70, 56, 56, 41,
		27, 53, 59, 73, 69, 58, 38, 32,
		14, 31, 54, 56, 59, 47, 21, 10,
		5, 8, 15, 28, 31, 3, -27, -52,
		11, 10, 15, 29, 5, -2, -14, -21,
	},
	// Rook
	{
		44, 47, 54, 49, 41, 37, 37, 36,
		40, 51, 51, 38, 42, 34, 36, 27,
		45, 43, 43, 38, 28, 23, 24, 23,
		51, 45, 49, 40, 29, 23, 29, 29,
		42, 43, 40, 35, 31, 33, 24, 26,
		34, 31, 29, 27, 20, 14, -2, 3,
		25, 28, 25, 22, 13, 8, -1, 8,
		25, 25, 31, 22, 13, 13, 11, 7,
	},
	// Bishop
	{
		5, 12, 5, 16, 12, 3, 7, -3,
		-7, -4, -4, -1, -12, -5, 2, 0,
		8, -2, 2, -8, -4, 0, -2, 5,
		3, 4, 1, 17, 8, 7, 1, 8,
		-3, 4, 12, 11, 8, 6, 3, -15,
		-2, 4, 7, 10, 14, 6, -3, -8,
		6, -12, -12, -1, 0, -6, -7, -9,
		-5, 7, -9, -4, -7, 4, -11, -25,
	},
	// Knight
	{
		-43, -14, -2, -14, -12, -31, -19, -65,
		-3, -2, -12, -12, -18, -28, -5, -22,
		-6, -10, 3, 3, -9, -25, -19, -17,
		5, 3, 10, 13, 16, 8, 9, -3,
		1, 0, 16, 13, 21, 7, 3, 2,
		-8, -4, 3, 16, 14, -2, -8, -3,
		-2, 0, -5, -5, -7, -8, -10, 11,
		13, -17, -8, -8, -4, -14, -6, 8,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		150, 138, 142, 76, 78, 99, 142, 159,
		76, 80, 35, -5, -12, 11, 56, 59,
		54, 42, 25, 5, 5, 13, 30, 29,
		35, 31, 17, 9, 8, 13, 20, 18,
		28, 23, 19, 18, 18, 19, 11, 13,
		32, 29, 23, 20, 29, 19, 13, 13,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}
