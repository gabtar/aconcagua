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
	AllSquares   Bitboard = 0xFFFFFFFFFFFFFFFF
	LightSquares Bitboard = 0x55AA55AA55AA55AA

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
	KingFrontMask = generateKingFrontMasks()
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

// KingZone is a precalculated table for the squares surrounding the king for each side(used in king safety)
var KingZone [2][64]Bitboard

// KingFrontMask is a precalculated table for the squares ahead of a king and it adjacent files(used in pawn shield and storm)
var KingFrontMask [2][64]Bitboard

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

// generateKingFrontMasks generates a precalculated table for squares ahead the king and its adjacent files for each side
func generateKingFrontMasks() (kingFrontMask [2][64]Bitboard) {
	for sq := range 64 {
		file := sq % 8
		kingSquare := bitboardFromIndex(sq)

		whiteFontMask := fillUp(kingSquare)
		if file > 0 {
			whiteFontMask |= fillUp(bitboardFromIndex(sq - 1))
		}
		if file < 7 {
			whiteFontMask |= fillUp(bitboardFromIndex(sq + 1))
		}

		blackFrontMask := fillDown(kingSquare)
		if file > 0 {
			blackFrontMask |= fillDown(bitboardFromIndex(sq - 1))
		}
		if file < 7 {
			blackFrontMask |= fillDown(bitboardFromIndex(sq + 1))
		}

		kingFrontMask[White][sq] = whiteFontMask
		kingFrontMask[Black][sq] = blackFrontMask
	}
	return
}

// generateSquaresBetween generates a precalculated array of mask between 2 squares in orthogonal directions
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
var PiecesValues = [6]Score{S(0, 0), S(941, 1131), S(413, 623), S(325, 375), S(290, 364), S(43, 91)}

// Psqt stores the values for mg and eg phase for pieces square tables
var Psqt = [6][64]Score{
	// King
	{
		S(2, -109), S(46, -43), S(59, -26), S(-6, 4), S(-7, 1), S(-9, 14), S(18, 5), S(30, -102),
		S(-39, -24), S(17, 26), S(-3, 35), S(74, 23), S(25, 35), S(14, 56), S(17, 48), S(-22, 0),
		S(-69, -11), S(42, 23), S(6, 40), S(-13, 50), S(18, 51), S(70, 41), S(28, 45), S(-13, -6),
		S(-35, -34), S(-27, 17), S(-42, 40), S(-74, 55), S(-70, 51), S(-55, 44), S(-52, 26), S(-94, -9),
		S(-45, -46), S(-34, 2), S(-64, 32), S(-100, 52), S(-99, 49), S(-65, 28), S(-62, 10), S(-102, -21),
		S(-14, -49), S(-1, -3), S(-44, 23), S(-57, 37), S(-48, 34), S(-48, 21), S(-14, 0), S(-33, -30),
		S(49, -56), S(3, -3), S(-7, 12), S(-27, 21), S(-35, 26), S(-18, 14), S(5, -3), S(33, -46),
		S(33, -84), S(32, -32), S(22, -9), S(-41, 8), S(7, -10), S(-36, 6), S(17, -20), S(35, -82),
	},
	// Queen
	{
		S(-41, 29), S(-49, 42), S(-37, 58), S(-11, 48), S(-18, 47), S(-3, 36), S(30, -9), S(-12, 39),
		S(8, -7), S(-17, 17), S(-22, 50), S(-46, 78), S(-59, 94), S(-15, 48), S(-7, 30), S(52, 11),
		S(8, 17), S(0, 29), S(-2, 58), S(-7, 64), S(-11, 73), S(19, 45), S(26, 15), S(27, 24),
		S(-9, 44), S(-4, 57), S(-6, 58), S(-12, 73), S(-7, 70), S(5, 48), S(18, 57), S(12, 47),
		S(2, 32), S(-8, 59), S(-5, 62), S(0, 76), S(2, 71), S(6, 55), S(15, 46), S(18, 39),
		S(0, 18), S(8, 30), S(2, 54), S(3, 58), S(8, 61), S(11, 50), S(23, 28), S(17, 22),
		S(8, 7), S(8, 14), S(16, 18), S(19, 29), S(15, 32), S(23, 3), S(31, -22), S(39, -37),
		S(5, 10), S(6, 13), S(10, 21), S(17, 21), S(14, 8), S(0, 4), S(9, -5), S(18, -17),
	},
	// Rook
	{
		S(-11, 50), S(-14, 53), S(-27, 63), S(-22, 57), S(-9, 52), S(16, 47), S(27, 45), S(26, 44),
		S(-4, 33), S(-6, 43), S(8, 44), S(22, 33), S(6, 34), S(29, 29), S(17, 31), S(37, 21),
		S(-12, 44), S(14, 41), S(9, 42), S(11, 39), S(32, 28), S(46, 21), S(65, 21), S(32, 20),
		S(-14, 44), S(0, 39), S(1, 45), S(10, 36), S(9, 28), S(20, 24), S(19, 27), S(7, 26),
		S(-21, 36), S(-23, 39), S(-12, 36), S(-6, 33), S(-4, 30), S(-12, 30), S(7, 23), S(-8, 23),
		S(-24, 30), S(-18, 28), S(-14, 25), S(-11, 26), S(-3, 21), S(3, 13), S(22, 2), S(4, 4),
		S(-23, 24), S(-15, 25), S(-6, 23), S(-4, 22), S(0, 14), S(9, 8), S(23, -1), S(-18, 14),
		S(-3, 29), S(-3, 26), S(0, 30), S(6, 23), S(11, 16), S(8, 20), S(5, 17), S(3, 12),
	},
	// Bishop
	{
		S(-28, 7), S(-54, 12), S(-62, 11), S(-100, 22), S(-91, 20), S(-68, 4), S(-47, 11), S(-59, 0),
		S(-23, -5), S(-21, 3), S(-19, 2), S(-27, 4), S(-15, 0), S(-7, 0), S(-26, 5), S(-25, -4),
		S(-4, 10), S(-4, 3), S(-5, 7), S(0, -2), S(-6, 4), S(35, 4), S(8, 3), S(17, 9),
		S(-18, 7), S(-1, 8), S(-3, 6), S(2, 22), S(11, 8), S(-1, 12), S(3, 2), S(-23, 16),
		S(-5, 1), S(-13, 7), S(-7, 17), S(12, 15), S(6, 13), S(-5, 10), S(-8, 9), S(8, -6),
		S(0, 3), S(5, 12), S(5, 12), S(3, 17), S(5, 22), S(9, 10), S(10, 0), S(14, 0),
		S(13, 14), S(8, -6), S(14, -2), S(-3, 5), S(6, 8), S(13, 0), S(32, -4), S(20, -5),
		S(6, -3), S(18, 2), S(1, -2), S(-7, 3), S(1, 0), S(-2, 11), S(10, -2), S(28, -15),
	},
	// Knight
	{
		S(-119, -29), S(-99, -5), S(-68, 10), S(-29, -7), S(-4, 2), S(-60, -18), S(-73, -3), S(-81, -57),
		S(-12, -5), S(-3, 5), S(17, 0), S(21, 6), S(25, -4), S(58, -19), S(19, -4), S(20, -23),
		S(0, -8), S(13, 0), S(16, 15), S(27, 13), S(48, 3), S(70, -15), S(27, -10), S(28, -19),
		S(10, 5), S(9, 8), S(24, 16), S(30, 24), S(34, 20), S(40, 16), S(26, 9), S(37, 0),
		S(4, 4), S(9, 0), S(15, 20), S(21, 21), S(17, 30), S(23, 13), S(18, 8), S(12, 5),
		S(-8, -3), S(1, 1), S(3, 5), S(5, 22), S(19, 20), S(6, 2), S(20, -2), S(3, 4),
		S(-9, 0), S(-7, 2), S(-1, -1), S(14, 1), S(12, 0), S(13, -1), S(16, -5), S(17, 13),
		S(-42, 12), S(-2, -9), S(-19, -4), S(-7, -2), S(2, 0), S(5, -7), S(0, -2), S(-5, 3),
	},
	// Pawn
	{
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
		S(87, 185), S(79, 176), S(67, 169), S(107, 113), S(87, 112), S(78, 116), S(-11, 168), S(-21, 192),
		S(15, 77), S(2, 78), S(18, 40), S(22, -1), S(30, -8), S(67, 10), S(32, 51), S(5, 57),
		S(-3, 48), S(-4, 34), S(0, 17), S(1, 0), S(19, 0), S(15, 7), S(4, 23), S(3, 22),
		S(-3, 27), S(-6, 23), S(1, 9), S(11, 0), S(14, 0), S(8, 8), S(0, 13), S(-1, 10),
		S(-10, 20), S(-10, 15), S(-8, 10), S(-2, 9), S(3, 11), S(-3, 11), S(9, 1), S(0, 2),
		S(-2, 24), S(0, 20), S(-4, 15), S(-10, 13), S(-4, 23), S(13, 12), S(23, 4), S(-1, 2),
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
