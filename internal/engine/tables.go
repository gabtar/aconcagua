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

var KingZone [64]Bitboard

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
func generateKingZone() (kingZone [64]Bitboard) {
	for sq := range 64 {
		from := bitboardFromIndex(sq)
		kingZone[sq] = kingAttacksTable[sq] | from
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
var MiddlegamePieceValue = [6]int{12000, 1049, 432, 343, 338, 74}

// EndgamePieceValue is the value of each piece endgame phase
var EndgamePieceValue = [6]int{12000, 1036, 612, 338, 335, 98}

// MiddlegamePSQT are the pieces square tables for middlegame
var MiddlegamePSQT = [6][64]int{
	// King
	{
		-36, 12, 21, -14, -54, -28, 5, -16,
		-6, 16, 1, 39, 14, 21, 3, -32,
		-17, 44, 17, 1, 21, 59, 38, -9,
		-25, -12, -14, -55, -56, -37, -31, -92,
		-63, -7, -44, -75, -68, -46, -69, -122,
		-18, 22, -22, -29, -20, -39, -9, -57,
		52, 23, 17, -19, -21, -7, 34, 29,
		35, 58, 39, -35, 17, -20, 38, 42,
	},
	// Queen
	{
		-26, -29, -15, 3, 21, 18, 40, 14,
		14, -13, -7, -26, -20, 19, 14, 67,
		13, 6, 11, 9, 22, 61, 64, 53,
		2, 7, 6, 7, 10, 16, 24, 24,
		12, 2, 4, 10, 9, 7, 20, 28,
		12, 14, 7, 5, 10, 14, 30, 31,
		21, 16, 21, 23, 20, 32, 39, 45,
		11, 17, 21, 20, 24, 11, 21, 10,
	},
	// Rook
	{
		-2, -4, -18, -11, -2, 1, 9, 28,
		-5, -7, 8, 24, 16, 37, 28, 44,
		-16, 10, 9, 9, 36, 46, 80, 40,
		-20, -3, -7, 0, 4, 16, 20, 7,
		-32, -31, -21, -14, -13, -22, 4, -13,
		-33, -28, -24, -23, -11, -7, 27, 3,
		-30, -26, -14, -12, -6, 3, 19, -15,
		-15, -11, -10, -2, 5, 1, 7, -8,
	},
	// Bishop
	{
		-12, -42, -56, -77, -72, -80, -31, -30,
		0, 14, 3, -14, 19, 2, -3, -15,
		3, 7, 10, 14, 10, 42, 28, 24,
		-8, 7, 7, 23, 19, 18, 13, -3,
		6, -7, 0, 22, 22, 0, 0, 23,
		9, 17, 10, 10, 13, 11, 18, 31,
		32, 12, 26, 2, 9, 23, 34, 30,
		17, 36, 8, 3, 10, 4, 25, 44,
	},
	// Knight
	{
		-133, -111, -83, -50, -1, -89, -57, -98,
		-32, -7, 20, 28, 11, 59, -9, 12,
		-21, -3, 11, 10, 47, 45, 25, 4,
		-14, -14, -3, 22, 6, 28, -5, 24,
		-22, -20, -14, -4, 0, 1, 9, -5,
		-37, -29, -32, -24, -7, -28, -6, -21,
		-39, -34, -31, -16, -18, -14, -14, -14,
		-72, -31, -48, -35, -24, -24, -30, -33,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		66, 87, 49, 86, 44, 50, -43, -56,
		22, 11, 43, 47, 59, 93, 61, 26,
		-7, -4, 2, 5, 27, 29, 8, 6,
		-10, -12, 0, 13, 14, 16, 0, -2,
		-14, -8, -4, 2, 16, 1, 19, -2,
		-12, -8, -3, -5, 7, 20, 27, -11,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// EndgamePSQT are the pieces square tables for endgame
var EndgamePSQT = [6][64]int{
	// King
	{
		-87, -50, -35, -9, -13, -9, -9, -75,
		-33, 7, 12, 12, 25, 33, 33, 2,
		-16, 12, 27, 37, 43, 40, 36, 3,
		-21, 14, 32, 46, 48, 43, 30, 7,
		-27, -1, 27, 44, 43, 27, 15, 0,
		-39, -14, 8, 22, 19, 10, -10, -20,
		-56, -26, -13, 1, 2, -9, -31, -50,
		-94, -71, -49, -24, -46, -34, -67, -102,
	},
	// Queen
	{
		32, 44, 64, 60, 43, 49, 19, 31,
		23, 42, 68, 93, 105, 67, 54, 53,
		30, 35, 50, 70, 76, 58, 33, 45,
		36, 37, 42, 58, 68, 68, 68, 58,
		15, 37, 36, 48, 53, 50, 43, 41,
		2, 12, 27, 30, 33, 29, 10, 13,
		-6, -9, -8, -3, 2, -20, -43, -40,
		-3, -6, -13, 1, -13, -12, -17, 0,
	},
	// Rook
	{
		37, 38, 50, 39, 33, 36, 35, 29,
		35, 45, 46, 34, 31, 28, 27, 19,
		36, 33, 33, 29, 20, 11, 9, 10,
		39, 33, 40, 34, 21, 14, 18, 17,
		31, 33, 30, 26, 24, 22, 14, 14,
		22, 20, 17, 18, 12, 4, -13, -7,
		13, 15, 14, 11, 3, -2, -12, -3,
		13, 12, 20, 12, 3, 5, 0, -2,
	},
	// Bishop
	{
		10, 15, 16, 21, 17, 10, 10, -5,
		-1, 3, 6, 9, -5, 2, 9, 2,
		18, 6, 11, 0, 3, 10, 3, 12,
		12, 13, 9, 22, 16, 12, 8, 14,
		2, 15, 20, 17, 16, 15, 13, -8,
		7, 11, 15, 20, 22, 16, 4, 2,
		11, -1, -2, 8, 11, 2, 2, -4,
		0, 12, 0, 4, 0, 13, -6, -23,
	},
	// Knight
	{
		-50, -16, 5, -8, -12, -23, -28, -64,
		4, 3, -7, -6, -15, -22, -4, -24,
		0, 0, 9, 9, -4, -14, -17, -17,
		6, 8, 18, 18, 17, 10, 9, -5,
		4, 6, 22, 17, 25, 12, 2, 0,
		-8, 0, 7, 22, 19, 3, -2, -1,
		-6, 2, 0, -1, -1, -3, -7, 3,
		-1, -20, -5, -4, -4, -10, -9, -15,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		91, 69, 81, 23, 25, 34, 81, 102,
		54, 59, 15, -23, -28, -9, 33, 37,
		37, 24, 9, -10, -11, -4, 12, 11,
		18, 14, 2, -5, -6, -2, 3, 1,
		12, 8, 4, 4, 3, 5, -2, -1,
		14, 11, 6, 6, 13, 5, -3, -2,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}
