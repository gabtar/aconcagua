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
var MiddlegamePieceValue = [6]int{10000, 1069, 459, 356, 321, 84}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{10000, 986, 550, 330, 319, 78}

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-58, 20, 15, -19, -50, -32, 7, 6,
		11, -2, -26, 4, 0, 1, -19, -21,
		-31, 21, -15, -12, 0, 25, 20, -22,
		-33, -39, -40, -59, -52, -46, -41, -67,
		-64, -31, -63, -78, -86, -73, -73, -91,
		-27, -4, -55, -72, -63, -66, -27, -60,
		39, 12, 0, -39, -43, -22, 23, 25,
		25, 68, 33, -63, -5, -35, 38, 39,
	},
	// Queen
	{
		-26, -19, 6, 11, 33, 37, 46, 15,
		0, -31, -25, -25, -28, 20, 0, 61,
		6, -4, -2, -3, 6, 55, 59, 63,
		-12, -7, -15, -14, -13, -1, 8, 9,
		-4, -15, -13, -9, -8, -11, 1, 11,
		-6, -2, -9, -11, -5, -4, 12, 14,
		0, 0, 4, 7, 5, 15, 21, 33,
		-3, 0, 4, 4, 6, -5, 0, -11,
	},
	// Rook
	{
		5, 10, 3, 15, 30, 22, 32, 43,
		-9, -9, 14, 23, 34, 31, 25, 51,
		-15, 11, 4, 3, 33, 39, 80, 44,
		-17, -7, -13, -2, 0, 5, 17, 9,
		-33, -33, -27, -20, -18, -32, 0, -13,
		-36, -32, -30, -27, -15, -15, 18, -1,
		-33, -31, -19, -17, -10, -5, 8, -30,
		-19, -15, -14, -7, 1, -4, 1, -18,
	},
	// Bishop
	{
		-28, -24, -63, -67, -53, -68, -12, -36,
		-12, 6, -5, -15, 8, 15, -11, -26,
		-1, 4, 8, 8, 1, 38, 22, 17,
		-14, 2, 1, 14, 14, 12, 7, -13,
		-2, -14, -6, 16, 15, -4, -4, 18,
		2, 10, 3, 5, 7, 6, 11, 24,
		23, 5, 19, -3, 4, 19, 26, 24,
		4, 26, 2, -4, 2, -4, 2, 19,
	},
	// Knight
	{
		-137, -89, -45, -32, 39, -80, -20, -81,
		-32, -1, 44, 50, 30, 85, 10, 16,
		-10, 24, 32, 33, 75, 91, 48, 28,
		6, 8, 20, 47, 26, 49, 16, 46,
		-1, 3, 8, 17, 23, 24, 31, 14,
		-18, -6, -8, -1, 14, -5, 16, 0,
		-18, -17, -9, 7, 4, 7, 5, 4,
		-70, -10, -26, -11, -6, -1, -7, -12,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		67, 95, 46, 61, 30, 84, -5, -40,
		9, -1, 28, 22, 36, 84, 55, 23,
		-13, -12, -2, 2, 20, 28, 9, 11,
		-16, -21, -6, 7, 8, 13, 0, 0,
		-20, -16, -10, -2, 11, 4, 22, 1,
		-18, -16, -10, -7, 3, 29, 33, -2,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-77, -47, -30, -4, -4, 4, 0, -49,
		-22, 19, 24, 23, 35, 46, 47, 12,
		-2, 26, 39, 45, 53, 54, 50, 18,
		-8, 26, 43, 52, 53, 53, 42, 14,
		-13, 10, 36, 50, 51, 40, 26, 5,
		-23, 0, 20, 34, 33, 25, 6, -3,
		-32, -12, -1, 10, 14, 5, -13, -30,
		-65, -56, -35, -14, -33, -17, -46, -77,
	},
	// Queen
	{
		13, 17, 28, 36, 19, 25, -2, 9,
		8, 36, 56, 65, 88, 37, 42, 28,
		9, 21, 43, 57, 65, 52, 24, 24,
		23, 29, 39, 54, 67, 58, 59, 48,
		6, 28, 26, 44, 43, 40, 30, 31,
		-7, 0, 18, 18, 22, 19, 0, -3,
		-7, -21, -19, -12, -6, -30, -54, -48,
		-18, -16, -18, -8, -19, -22, -13, -12,
	},
	// Rook
	{
		35, 35, 43, 32, 25, 34, 31, 26,
		39, 48, 45, 37, 26, 33, 31, 21,
		38, 36, 38, 35, 24, 20, 13, 14,
		38, 35, 43, 35, 26, 23, 24, 22,
		33, 34, 33, 30, 27, 29, 17, 17,
		23, 21, 19, 19, 13, 8, -8, -3,
		13, 18, 16, 14, 6, 2, -5, 10,
		14, 14, 22, 14, 7, 10, 8, 1,
	},
	// Bishop
	{
		4, 0, 6, 5, 1, -3, -6, -12,
		-8, -9, -6, -4, -15, -20, -5, -5,
		5, -6, -4, -12, -9, -2, -7, 3,
		2, 0, -1, 9, 2, 0, -3, 6,
		-6, 1, 8, 4, 3, 0, 0, -18,
		-4, 0, 2, 6, 9, 2, -9, -10,
		4, -14, -15, -4, -2, -12, -8, -13,
		0, 3, -8, -6, -9, 2, -1, -8,
	},
	// Knight
	{
		-29, -20, -6, -10, -23, -20, -40, -69,
		5, 3, -14, -13, -20, -32, -9, -24,
		-1, -12, 0, 2, -15, -31, -25, -22,
		0, 0, 8, 8, 10, 3, 4, -11,
		-3, -4, 11, 8, 15, 2, -6, -6,
		-12, -9, -3, 11, 9, -7, -12, -10,
		-14, -2, -9, -10, -11, -13, -14, -7,
		2, -22, -11, -11, -7, -17, -17, -27,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		141, 132, 130, 93, 103, 90, 126, 150,
		86, 91, 52, 27, 17, 24, 64, 68,
		56, 44, 29, 8, 9, 14, 32, 31,
		38, 35, 23, 14, 14, 19, 24, 21,
		32, 29, 25, 23, 24, 24, 18, 18,
		34, 32, 28, 27, 35, 24, 16, 16,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}
