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
var PiecesValues = [6]Score{S(0, 0), S(941, 1140), S(413, 629), S(327, 380), S(292, 368), S(45, 94)}

// Psqt stores the values for mg and eg phase for pieces square tables
var Psqt = [6][64]Score{
	// King
	{
		S(3, -105), S(46, -40), S(58, -25), S(-5, 4), S(-7, 0), S(-9, 15), S(18, 5), S(30, -101),
		S(-37, -20), S(17, 28), S(-4, 35), S(73, 21), S(25, 33), S(14, 56), S(17, 49), S(-22, 2),
		S(-67, -8), S(40, 23), S(5, 38), S(-14, 47), S(18, 50), S(69, 42), S(28, 46), S(-12, -5),
		S(-34, -32), S(-26, 15), S(-43, 35), S(-74, 50), S(-70, 47), S(-53, 41), S(-50, 24), S(-91, -8),
		S(-43, -44), S(-33, 0), S(-63, 28), S(-99, 47), S(-96, 44), S(-63, 24), S(-59, 7), S(-100, -21),
		S(-13, -46), S(1, -5), S(-40, 19), S(-52, 33), S(-43, 30), S(-44, 18), S(-11, -1), S(-30, -30),
		S(49, -51), S(3, -2), S(-5, 11), S(-25, 20), S(-33, 24), S(-16, 12), S(5, -4), S(32, -44),
		S(33, -77), S(31, -28), S(18, -8), S(-40, 9), S(3, -10), S(-35, 7), S(15, -18), S(35, -75),
	},
	// Queen
	{
		S(-41, 29), S(-50, 42), S(-38, 57), S(-13, 46), S(-18, 47), S(-3, 37), S(28, -9), S(-14, 39),
		S(11, -3), S(-13, 19), S(-19, 52), S(-43, 78), S(-57, 95), S(-12, 51), S(-2, 33), S(54, 13),
		S(11, 21), S(4, 33), S(2, 61), S(-3, 67), S(-10, 76), S(23, 49), S(29, 18), S(29, 27),
		S(-5, 47), S(-2, 59), S(-3, 61), S(-8, 76), S(-4, 73), S(6, 50), S(20, 59), S(13, 50),
		S(1, 33), S(-6, 60), S(-4, 62), S(1, 78), S(4, 72), S(6, 56), S(14, 46), S(18, 39),
		S(0, 18), S(6, 29), S(2, 54), S(0, 57), S(8, 61), S(9, 51), S(22, 28), S(17, 23),
		S(8, 7), S(8, 12), S(14, 15), S(17, 27), S(12, 30), S(22, 2), S(30, -22), S(40, -37),
		S(4, 10), S(8, 11), S(10, 20), S(12, 18), S(13, 7), S(0, 3), S(9, -4), S(18, -15),
	},
	// Rook
	{
		S(-9, 51), S(-13, 54), S(-26, 63), S(-21, 56), S(-9, 51), S(15, 48), S(25, 46), S(24, 44),
		S(0, 36), S(-3, 45), S(9, 46), S(23, 34), S(6, 35), S(31, 32), S(17, 34), S(38, 25),
		S(-8, 47), S(17, 43), S(12, 44), S(14, 40), S(36, 31), S(49, 24), S(67, 24), S(34, 23),
		S(-10, 47), S(3, 42), S(5, 47), S(14, 39), S(13, 31), S(23, 27), S(20, 30), S(10, 30),
		S(-18, 39), S(-19, 41), S(-8, 37), S(-2, 34), S(0, 31), S(-9, 31), S(9, 25), S(-7, 26),
		S(-22, 31), S(-15, 28), S(-11, 25), S(-9, 26), S(0, 20), S(4, 12), S(24, 1), S(4, 6),
		S(-21, 24), S(-13, 25), S(-4, 23), S(-2, 21), S(2, 13), S(10, 7), S(22, -1), S(-17, 14),
		S(-6, 26), S(-2, 25), S(0, 29), S(6, 21), S(11, 14), S(6, 17), S(4, 15), S(0, 9),
	},
	// Bishop
	{
		S(-28, 9), S(-55, 13), S(-62, 12), S(-100, 21), S(-92, 21), S(-70, 3), S(-48, 12), S(-57, 1),
		S(-17, -1), S(-15, 5), S(-14, 3), S(-26, 5), S(-12, 1), S(-6, 1), S(-22, 7), S(-27, -1),
		S(-3, 12), S(-2, 4), S(-3, 8), S(2, -1), S(-5, 5), S(37, 6), S(9, 5), S(18, 9),
		S(-15, 8), S(-1, 8), S(0, 6), S(5, 23), S(14, 10), S(1, 13), S(3, 4), S(-20, 17),
		S(-3, 2), S(-11, 8), S(-9, 18), S(14, 15), S(8, 14), S(-6, 10), S(-6, 9), S(9, -6),
		S(0, 2), S(5, 12), S(5, 12), S(1, 17), S(3, 21), S(9, 10), S(11, 0), S(15, 0),
		S(12, 14), S(7, -8), S(14, -2), S(-4, 5), S(4, 6), S(13, 0), S(30, -5), S(20, -6),
		S(7, -1), S(18, 6), S(-1, -3), S(-6, 3), S(2, 0), S(-5, 11), S(11, -1), S(31, -14),
	},
	// Knight
	{
		S(-117, -29), S(-99, -4), S(-68, 10), S(-30, -7), S(-5, 0), S(-61, -20), S(-74, -3), S(-80, -57),
		S(-8, -1), S(1, 8), S(21, 0), S(22, 5), S(25, -6), S(59, -18), S(19, -2), S(22, -21),
		S(3, -5), S(15, 0), S(18, 15), S(29, 14), S(50, 4), S(69, -14), S(29, -9), S(27, -18),
		S(13, 7), S(11, 8), S(25, 17), S(30, 25), S(32, 21), S(40, 18), S(27, 11), S(41, 1),
		S(6, 5), S(11, 2), S(15, 21), S(21, 21), S(18, 30), S(23, 14), S(19, 9), S(13, 6),
		S(-9, -3), S(2, 0), S(1, 5), S(6, 22), S(19, 20), S(3, 1), S(20, -2), S(2, 5),
		S(-8, 0), S(-6, 2), S(-1, -1), S(12, 0), S(10, -1), S(14, -1), S(17, -5), S(17, 12),
		S(-40, 14), S(-5, -10), S(-17, -3), S(-5, -1), S(3, 0), S(5, -7), S(-2, -2), S(-1, 5),
	},
	// Pawn
	{
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
		S(84, 190), S(78, 181), S(65, 174), S(106, 118), S(88, 116), S(78, 120), S(-11, 172), S(-24, 196),
		S(16, 75), S(2, 76), S(18, 39), S(21, -1), S(31, -10), S(69, 9), S(34, 50), S(8, 54),
		S(0, 48), S(-1, 34), S(1, 17), S(2, 0), S(21, 0), S(17, 9), S(6, 25), S(6, 23),
		S(-1, 27), S(-5, 24), S(2, 9), S(12, 1), S(14, 1), S(8, 10), S(2, 15), S(0, 11),
		S(-10, 20), S(-8, 15), S(-7, 11), S(-1, 9), S(4, 12), S(-3, 12), S(11, 3), S(1, 3),
		S(-3, 24), S(0, 19), S(-4, 15), S(-10, 12), S(-4, 23), S(14, 12), S(23, 5), S(-1, 3),
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
