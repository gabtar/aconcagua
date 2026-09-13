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
	initBitboards()
	generatePsqtScores()
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

// PieceValues stores the score for each piece for mg and eg
var PiecesValues = [6]Score{S(0, 0), S(940, 1106), S(429, 580), S(341, 346), S(300, 333), S(70, 119)}

// Psqt stores the values for mg and eg phase for pieces square tables
var Psqt = [6][64]Score{
	// King
	{
		S(-14, -85), S(41, -49), S(50, -32), S(-18, 0), S(-19, -3), S(-21, 7), S(8, 2), S(26, -81),
		S(-45, -9), S(0, 20), S(-15, 28), S(57, 15), S(16, 32), S(3, 50), S(0, 43), S(-31, 13),
		S(-90, 10), S(27, 27), S(-10, 43), S(-21, 49), S(3, 50), S(59, 51), S(2, 54), S(-39, 22),
		S(-66, -1), S(-51, 32), S(-67, 51), S(-91, 61), S(-96, 62), S(-79, 60), S(-85, 51), S(-142, 29),
		S(-72, -13), S(-49, 13), S(-79, 41), S(-109, 56), S(-114, 57), S(-85, 43), S(-97, 32), S(-136, 16),
		S(-30, -22), S(-8, 0), S(-61, 22), S(-77, 36), S(-72, 35), S(-75, 28), S(-29, 8), S(-52, -3),
		S(58, -43), S(17, -15), S(2, -1), S(-37, 10), S(-40, 14), S(-19, 4), S(33, -16), S(41, -35),
		S(48, -82), S(84, -64), S(56, -42), S(-48, -20), S(18, -41), S(-20, -22), S(59, -53), S(58, -84),
	},
	// Queen
	{
		S(-20, 19), S(-30, 44), S(-13, 68), S(13, 56), S(6, 66), S(13, 56), S(45, 9), S(11, 38),
		S(-1, -2), S(-33, 38), S(-25, 69), S(-33, 91), S(-39, 122), S(4, 78), S(1, 51), S(61, 30),
		S(1, 7), S(-8, 22), S(-9, 60), S(-4, 76), S(6, 95), S(51, 76), S(55, 47), S(63, 48),
		S(-16, 19), S(-16, 39), S(-18, 51), S(-21, 80), S(-17, 93), S(1, 76), S(12, 69), S(18, 56),
		S(-8, 12), S(-21, 43), S(-21, 50), S(-19, 76), S(-16, 69), S(-13, 59), S(1, 48), S(13, 38),
		S(-12, 3), S(-7, 13), S(-17, 37), S(-15, 34), S(-12, 38), S(-5, 34), S(9, 15), S(12, 6),
		S(-3, -6), S(-7, -5), S(0, -6), S(3, -1), S(0, 3), S(8, -18), S(17, -46), S(33, -61),
		S(-9, -5), S(-6, -6), S(-1, -5), S(6, -2), S(1, -9), S(-11, -10), S(4, -24), S(3, -37),
	},
	// Rook
	{
		S(11, 37), S(4, 41), S(-4, 52), S(-2, 47), S(13, 41), S(32, 35), S(37, 33), S(48, 29),
		S(0, 34), S(0, 44), S(19, 46), S(38, 34), S(24, 35), S(53, 27), S(43, 26), S(70, 13),
		S(-14, 34), S(12, 35), S(9, 35), S(13, 31), S(41, 18), S(53, 15), S(89, 11), S(60, 7),
		S(-23, 37), S(-5, 32), S(-7, 41), S(-1, 34), S(5, 22), S(14, 19), S(30, 15), S(23, 13),
		S(-33, 29), S(-33, 32), S(-25, 31), S(-17, 28), S(-15, 25), S(-20, 25), S(10, 15), S(-3, 11),
		S(-35, 23), S(-31, 23), S(-27, 18), S(-25, 21), S(-14, 16), S(-5, 8), S(26, -5), S(2, -2),
		S(-32, 16), S(-27, 17), S(-18, 18), S(-15, 18), S(-7, 9), S(0, 5), S(13, -1), S(-23, 10),
		S(-12, 20), S(-11, 18), S(-8, 25), S(0, 17), S(7, 10), S(5, 15), S(9, 8), S(-5, 5),
	},
	// Bishop
	{
		S(-27, -5), S(-52, 0), S(-57, 0), S(-94, 10), S(-81, 7), S(-67, -3), S(-43, -5), S(-49, -11),
		S(-17, -17), S(-2, -9), S(-13, -5), S(-20, -4), S(4, -11), S(0, -11), S(-2, -5), S(-6, -17),
		S(-11, 5), S(4, -9), S(3, -2), S(14, -12), S(7, -8), S(47, -6), S(26, -8), S(24, 0),
		S(-16, -1), S(-5, 4), S(1, -1), S(15, 10), S(14, 2), S(11, 2), S(3, -1), S(-10, 2),
		S(-8, -7), S(-16, 2), S(-7, 9), S(10, 6), S(11, 5), S(-8, 3), S(-10, 1), S(11, -14),
		S(1, -2), S(6, 1), S(0, 4), S(2, 7), S(4, 10), S(2, 4), S(5, -6), S(18, -6),
		S(21, 2), S(4, -11), S(13, -12), S(-6, -2), S(0, 0), S(13, -9), S(24, -4), S(18, -14),
		S(5, -9), S(21, 0), S(3, -10), S(-11, -5), S(-3, -7), S(-6, 2), S(12, -14), S(22, -24),
	},
	// Knight
	{
		S(-120, -49), S(-100, -18), S(-60, 0), S(-22, -13), S(10, -7), S(-54, -29), S(-67, -21), S(-74, -73),
		S(-7, -14), S(6, 0), S(41, -5), S(48, -2), S(36, -8), S(95, -27), S(18, -9), S(34, -32),
		S(0, -11), S(26, -8), S(29, 10), S(42, 7), S(82, -6), S(89, -11), S(55, -16), S(41, -22),
		S(2, 1), S(8, 6), S(20, 14), S(47, 17), S(37, 16), S(56, 11), S(29, 6), S(50, -8),
		S(-4, 1), S(2, -1), S(7, 19), S(17, 18), S(24, 23), S(26, 12), S(31, 2), S(12, -1),
		S(-17, -7), S(-7, -1), S(-4, 4), S(0, 19), S(14, 18), S(0, 0), S(18, -5), S(-3, -2),
		S(-21, -13), S(-15, -4), S(-9, -5), S(10, -3), S(7, -2), S(8, -7), S(3, -10), S(2, 1),
		S(-64, -2), S(-8, -17), S(-27, -11), S(-12, -7), S(-6, -3), S(-2, -13), S(-4, -12), S(-27, -16),
	},
	// Pawn
	{
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
		S(91, 198), S(106, 189), S(75, 185), S(121, 127), S(95, 126), S(80, 141), S(4, 188), S(-10, 205),
		S(3, 126), S(10, 132), S(42, 91), S(61, 51), S(65, 44), S(72, 55), S(57, 102), S(9, 103),
		S(-20, 54), S(-4, 42), S(-4, 23), S(0, 13), S(18, 3), S(19, 4), S(16, 25), S(1, 25),
		S(-24, 25), S(-10, 24), S(-7, 3), S(6, 0), S(8, -2), S(9, -3), S(8, 12), S(-6, 3),
		S(-25, 18), S(-6, 19), S(-11, 2), S(-3, 13), S(13, 5), S(1, 2), S(31, 7), S(0, 0),
		S(-20, 22), S(-1, 23), S(-7, 9), S(-3, 19), S(8, 21), S(29, 6), S(45, 7), S(-2, -1),
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
	},
}

// PieceSquaresScores stores the score of each piece on each square from Black persctive/ point of view
var PieceSquaresScores = [6][64]Score{}

// generatePsqtScores generates an array of scores with the sum of the Psqt and Piece Value for each piece on each square
func generatePsqtScores() {
	for p := range 6 {
		mgVal, egVal := PiecesValues[p].Get()
		for sq := range 64 {
			mgPsqt, egPsqt := Psqt[p][sq].Get()

			PieceSquaresScores[p][sq] = S(mgVal+mgPsqt, egVal+egPsqt)
		}
	}
}
