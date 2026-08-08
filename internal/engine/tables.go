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
var MiddlegamePieceValue = [6]int{10000, 947, 436, 351, 313, 56}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{10000, 1096, 605, 363, 348, 81}

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-19, 47, 55, -17, -21, -28, 5, 24,
		-45, 10, -6, 67, 27, 11, -2, -39,
		-95, 36, 3, -5, 23, 67, -8, -49,
		-68, -37, -51, -67, -74, -71, -91, -162,
		-73, -30, -50, -76, -79, -55, -79, -155,
		-35, 7, -30, -38, -30, -34, -4, -55,
		41, 15, 5, -22, -29, -4, 21, 29,
		30, 53, 34, -33, 14, -23, 35, 35,
	},
	// Queen
	{
		-42, -52, -45, -15, -30, -13, 22, -20,
		4, -21, -26, -49, -74, -27, -10, 46,
		7, 0, -2, -13, -25, 13, 14, 17,
		-7, -5, -10, -15, -14, -8, 2, -1,
		0, -9, -7, -4, -2, -6, 8, 7,
		0, 6, 0, -2, 2, 5, 20, 14,
		10, 11, 14, 15, 15, 23, 35, 48,
		3, 8, 11, 13, 15, 1, 15, 21,
	},
	// Rook
	{
		-10, -10, -23, -22, -7, 21, 26, 26,
		2, 0, 17, 35, 15, 40, 23, 44,
		-11, 14, 10, 10, 34, 40, 63, 26,
		-15, 0, 0, 13, 8, 19, 15, 2,
		-25, -26, -13, -6, -3, -19, 2, -15,
		-30, -22, -19, -15, -6, -3, 18, -2,
		-27, -20, -11, -8, -2, 4, 18, -23,
		-12, -8, -6, 0, 7, 2, 0, -6,
	},
	// Bishop
	{
		-29, -59, -62, -101, -93, -85, -53, -58,
		-10, 3, -6, -18, 3, -9, -17, -37,
		1, 7, 7, 15, 3, 48, 14, 24,
		-11, 2, 8, 14, 18, 7, 6, -17,
		0, -6, -5, 19, 12, -1, -4, 19,
		3, 10, 7, 3, 7, 10, 13, 19,
		18, 10, 17, -2, 6, 15, 31, 23,
		11, 19, 0, -3, 5, -4, 14, 37,
	},
	// Knight
	{
		-111, -96, -68, -30, 2, -64, -70, -72,
		-4, 10, 43, 40, 25, 72, 14, 19,
		5, 26, 20, 37, 67, 72, 38, 22,
		11, 11, 24, 36, 27, 40, 23, 41,
		3, 9, 13, 21, 21, 24, 16, 15,
		-13, -1, -4, 2, 16, -1, 18, 0,
		-13, -10, -5, 8, 7, 10, 14, 11,
		-46, -10, -22, -9, -2, 2, -6, -4,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		72, 74, 44, 91, 68, 54, -19, -17,
		26, 11, 39, 46, 59, 90, 62, 30,
		-5, -5, -1, 2, 18, 18, 1, 3,
		-6, -13, -3, 7, 10, 5, -6, -5,
		-16, -16, -16, -7, -2, -9, 2, -5,
		-9, -7, -10, -17, -8, 10, 18, -8,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-89, -48, -30, 2, -2, 1, -8, -95,
		-11, 27, 34, 24, 38, 55, 46, 11,
		11, 31, 48, 55, 59, 56, 61, 21,
		-2, 34, 55, 66, 67, 64, 53, 28,
		-13, 17, 44, 61, 59, 43, 29, 12,
		-25, 0, 26, 40, 38, 26, 2, -13,
		-50, -14, 5, 19, 22, 5, -21, -49,
		-88, -64, -34, -6, -30, -17, -56, -95,
	},
	// Queen
	{
		13, 26, 44, 29, 40, 34, -5, 28,
		7, 30, 57, 75, 98, 61, 43, 27,
		23, 34, 55, 62, 77, 49, 27, 26,
		37, 52, 57, 74, 70, 55, 58, 44,
		29, 55, 59, 76, 68, 57, 40, 33,
		16, 28, 52, 54, 58, 46, 22, 14,
		5, 7, 13, 27, 29, 2, -27, -50,
		10, 8, 15, 22, 5, -2, -14, -24,
	},
	// Rook
	{
		44, 45, 53, 48, 41, 37, 36, 35,
		40, 49, 49, 37, 38, 32, 36, 27,
		44, 41, 40, 36, 26, 22, 22, 22,
		48, 42, 48, 36, 29, 24, 28, 29,
		41, 43, 39, 34, 30, 31, 23, 26,
		34, 31, 27, 26, 20, 13, 0, 4,
		26, 27, 25, 22, 13, 8, -1, 14,
		24, 26, 31, 21, 13, 12, 14, 6,
	},
	// Bishop
	{
		2, 10, 4, 15, 12, 2, 6, -4,
		-6, -5, -2, -3, -10, -4, 0, 0,
		7, -3, 1, -8, -4, 0, -1, 4,
		3, 4, 2, 17, 9, 7, 0, 8,
		-4, 4, 12, 12, 8, 7, 2, -14,
		-2, 4, 8, 10, 15, 5, -3, -8,
		5, -13, -11, -1, 0, -5, -9, -12,
		-6, 6, -9, -3, -7, 4, -9, -22,
	},
	// Knight
	{
		-42, -15, -1, -16, -14, -32, -22, -66,
		-5, -1, -16, -11, -17, -31, -7, -24,
		-7, -11, 4, 2, -10, -23, -20, -20,
		3, 2, 11, 13, 17, 8, 8, -4,
		0, 0, 16, 15, 20, 8, 4, 1,
		-7, -3, 3, 17, 14, -1, -8, 0,
		-3, -1, -5, -5, -6, -9, -9, 9,
		15, -16, -7, -6, -3, -13, -8, 5,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		154, 143, 140, 79, 81, 100, 146, 164,
		75, 78, 36, -5, -12, 10, 54, 58,
		54, 42, 25, 5, 6, 13, 31, 29,
		34, 32, 17, 8, 8, 12, 20, 18,
		27, 24, 19, 18, 18, 18, 11, 13,
		32, 28, 23, 21, 31, 18, 13, 12,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}
