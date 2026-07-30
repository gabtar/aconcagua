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
var MiddlegamePieceValue = [6]int{10000, 938, 424, 343, 306, 71}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{10000, 1086, 606, 364, 350, 86}

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-29, 40, 44, -19, -30, -27, 7, 16,
		-35, 6, -14, 49, 18, 9, -2, -31,
		-83, 34, -4, -11, 15, 61, -4, -41,
		-62, -39, -53, -73, -76, -71, -88, -146,
		-69, -31, -51, -75, -80, -55, -81, -153,
		-33, 9, -31, -39, -29, -35, -4, -55,
		40, 10, 5, -20, -28, -3, 17, 27,
		36, 55, 36, -30, 16, -18, 37, 39,
	},
	// Queen
	{
		-47, -50, -40, -13, -24, -4, 26, -23,
		6, -18, -26, -52, -72, -29, -12, 45,
		7, -1, -6, -13, -27, 13, 11, 14,
		-5, -3, -10, -13, -14, -8, 1, 0,
		1, -7, -8, -2, -2, -7, 8, 9,
		1, 5, 0, -4, 2, 4, 20, 15,
		11, 10, 12, 15, 13, 22, 34, 48,
		0, 6, 10, 10, 15, 1, 16, 23,
	},
	// Rook
	{
		-9, -11, -18, -24, -5, 25, 30, 30,
		2, -1, 14, 33, 10, 38, 24, 45,
		-11, 12, 8, 6, 31, 39, 59, 27,
		-16, -2, -2, 8, 5, 14, 13, 0,
		-26, -26, -16, -7, -7, -22, -1, -16,
		-30, -24, -21, -16, -8, -6, 18, -1,
		-27, -23, -12, -9, -5, 1, 15, -14,
		-13, -9, -8, -2, 4, 0, 2, -9,
	},
	// Bishop
	{
		-29, -60, -56, -98, -88, -84, -48, -57,
		-8, 0, -2, -18, 3, -12, -21, -34,
		0, 3, 3, 14, 1, 46, 11, 22,
		-12, 1, 8, 12, 17, 5, 4, -19,
		-1, -8, -8, 18, 10, -3, -6, 19,
		4, 9, 6, 2, 6, 7, 12, 21,
		19, 7, 16, -3, 4, 15, 28, 26,
		12, 20, 0, -5, 3, -4, 18, 40,
	},
	// Knight
	{
		-107, -97, -64, -28, 2, -61, -60, -71,
		1, 13, 38, 40, 24, 72, 13, 22,
		7, 23, 18, 34, 65, 68, 32, 18,
		11, 9, 23, 31, 23, 36, 20, 41,
		3, 8, 10, 19, 19, 22, 14, 14,
		-13, -2, -6, 1, 15, -3, 16, 3,
		-14, -10, -6, 6, 5, 7, 13, 10,
		-44, -10, -22, -10, -3, 1, -8, -4,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		71, 74, 46, 93, 68, 54, -17, -18,
		22, 11, 43, 49, 62, 94, 63, 28,
		-5, -4, 4, 8, 25, 26, 6, 4,
		-7, -11, 2, 14, 16, 13, -1, -3,
		-14, -6, -5, 2, 11, 1, 13, -2,
		-12, -7, -5, -8, 1, 15, 18, -10,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-83, -43, -27, 6, 0, 0, -10, -100,
		-11, 29, 36, 30, 43, 55, 49, 11,
		9, 32, 51, 60, 65, 58, 60, 21,
		-1, 34, 56, 69, 70, 65, 53, 26,
		-12, 18, 45, 61, 60, 43, 30, 13,
		-24, 1, 26, 39, 37, 24, 0, -14,
		-48, -12, 4, 17, 19, 3, -21, -49,
		-87, -63, -34, -8, -31, -20, -57, -96,
	},
	// Queen
	{
		21, 24, 41, 29, 33, 27, -11, 25,
		11, 31, 60, 79, 100, 64, 44, 33,
		26, 36, 61, 61, 77, 48, 27, 28,
		37, 51, 60, 71, 69, 55, 57, 41,
		28, 53, 58, 72, 68, 58, 39, 34,
		14, 32, 54, 57, 59, 47, 21, 11,
		6, 9, 16, 28, 31, 4, -26, -50,
		14, 13, 15, 29, 6, 0, -12, -19,
	},
	// Rook
	{
		44, 47, 54, 49, 40, 37, 37, 36,
		41, 52, 52, 38, 42, 35, 37, 28,
		45, 43, 43, 39, 29, 23, 24, 23,
		51, 45, 50, 40, 30, 23, 29, 29,
		43, 43, 40, 35, 32, 33, 25, 26,
		34, 31, 28, 27, 21, 14, -1, 3,
		25, 29, 26, 22, 14, 9, 0, 9,
		26, 26, 32, 22, 14, 14, 12, 8,
	},
	// Bishop
	{
		5, 12, 5, 16, 11, 3, 6, -3,
		-7, -4, -3, -1, -12, -5, 1, 0,
		8, -2, 2, -8, -4, 0, -2, 5,
		4, 4, 0, 17, 8, 6, 1, 8,
		-3, 4, 12, 10, 7, 6, 3, -16,
		-2, 4, 7, 9, 13, 5, -4, -7,
		7, -11, -11, -1, 0, -7, -6, -9,
		-4, 7, -9, -3, -7, 3, -12, -25,
	},
	// Knight
	{
		-42, -14, -3, -14, -13, -31, -21, -63,
		-2, -2, -13, -13, -18, -28, -6, -22,
		-6, -11, 2, 2, -10, -26, -20, -17,
		5, 2, 9, 12, 16, 7, 8, -4,
		2, -1, 15, 12, 20, 6, 2, 2,
		-7, -4, 1, 15, 13, -4, -8, -3,
		-2, 0, -6, -5, -7, -8, -10, 11,
		13, -17, -8, -7, -4, -13, -5, 7,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		151, 138, 142, 76, 79, 99, 143, 160,
		78, 81, 37, -4, -10, 13, 57, 60,
		59, 46, 30, 10, 10, 17, 35, 33,
		39, 35, 22, 15, 13, 18, 24, 22,
		33, 29, 25, 23, 23, 24, 18, 19,
		35, 31, 26, 23, 33, 23, 16, 17,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}
