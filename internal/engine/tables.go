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
var PieceValues = [6]Score{S(0, 0), S(943, 1111), S(438, 593), S(353, 356), S(302, 337), S(72, 110)}

// Psqt stores the values for mg and eg phase for pieces square tables
var Psqt = [6][64]Score{
	// King
	{
		S(-16, -85), S(43, -50), S(50, -34), S(-18, -1), S(-20, -6), S(-23, 5), S(8, 1), S(26, -83),
		S(-45, -7), S(1, 21), S(-13, 29), S(59, 16), S(18, 33), S(5, 50), S(0, 43), S(-33, 15),
		S(-92, 12), S(29, 28), S(-8, 44), S(-17, 51), S(7, 50), S(61, 51), S(0, 55), S(-41, 23),
		S(-66, 0), S(-48, 34), S(-63, 52), S(-85, 62), S(-91, 62), S(-77, 61), S(-87, 52), S(-146, 31),
		S(-74, -12), S(-45, 13), S(-73, 40), S(-101, 56), S(-105, 56), S(-78, 42), S(-94, 31), S(-140, 16),
		S(-33, -21), S(-10, 0), S(-56, 20), S(-68, 33), S(-61, 32), S(-66, 24), S(-31, 7), S(-56, -3),
		S(50, -42), S(10, -14), S(-4, -1), S(-44, 9), S(-44, 13), S(-25, 5), S(28, -16), S(35, -35),
		S(41, -78), S(77, -62), S(51, -39), S(-51, -21), S(16, -42), S(-26, -22), S(56, -52), S(53, -83),
	},
	// Queen
	{
		S(-26, 21), S(-32, 44), S(-16, 68), S(11, 55), S(1, 66), S(10, 55), S(41, 10), S(4, 40),
		S(0, 0), S(-27, 35), S(-17, 68), S(-28, 93), S(-39, 125), S(6, 84), S(5, 53), S(59, 31),
		S(2, 9), S(-4, 23), S(-5, 59), S(2, 76), S(6, 101), S(46, 78), S(47, 52), S(51, 52),
		S(-16, 20), S(-14, 39), S(-11, 51), S(-14, 77), S(-8, 89), S(7, 74), S(12, 67), S(18, 55),
		S(-7, 12), S(-19, 42), S(-18, 49), S(-13, 72), S(-11, 65), S(-10, 57), S(2, 45), S(12, 35),
		S(-11, 4), S(-5, 12), S(-15, 37), S(-15, 35), S(-12, 42), S(-2, 31), S(10, 12), S(10, 7),
		S(-5, -6), S(-8, -4), S(1, -6), S(0, 4), S(0, 6), S(7, -15), S(15, -45), S(30, -62),
		S(-12, -6), S(-13, -7), S(-7, -1), S(5, 2), S(0, -7), S(-14, -12), S(0, -25), S(1, -38),
	},
	// Rook
	{
		S(14, 40), S(6, 45), S(-2, 58), S(-1, 55), S(14, 48), S(32, 37), S(36, 34), S(46, 32),
		S(14, 33), S(14, 45), S(31, 49), S(50, 40), S(35, 41), S(61, 30), S(46, 28), S(70, 16),
		S(-3, 36), S(22, 38), S(22, 39), S(26, 37), S(54, 23), S(58, 18), S(85, 15), S(54, 13),
		S(-16, 38), S(1, 35), S(3, 44), S(10, 37), S(17, 25), S(20, 19), S(28, 18), S(24, 15),
		S(-33, 32), S(-31, 35), S(-19, 35), S(-8, 32), S(-7, 28), S(-20, 27), S(5, 18), S(-4, 12),
		S(-40, 28), S(-29, 26), S(-21, 22), S(-22, 27), S(-13, 22), S(-12, 13), S(19, -2), S(-5, 0),
		S(-40, 20), S(-27, 20), S(-11, 20), S(-12, 21), S(-6, 13), S(-6, 11), S(9, 1), S(-30, 13),
		S(-21, 18), S(-16, 21), S(-4, 25), S(1, 23), S(7, 14), S(-3, 16), S(4, 11), S(-17, 5),
	},
	// Bishop
	{
		S(-29, -6), S(-53, 0), S(-56, 1), S(-93, 12), S(-83, 8), S(-69, -1), S(-45, -4), S(-50, -10),
		S(-14, -16), S(6, -6), S(-4, -1), S(-15, 0), S(11, -6), S(6, -6), S(3, 0), S(-11, -15),
		S(-4, 5), S(17, -4), S(16, 5), S(29, -1), S(20, 1), S(59, 1), S(34, -2), S(29, -2),
		S(-9, -1), S(1, 11), S(16, 7), S(30, 20), S(28, 12), S(22, 10), S(4, 5), S(-5, 0),
		S(-5, -4), S(-8, 8), S(0, 14), S(20, 12), S(18, 11), S(0, 8), S(-4, 3), S(7, -10),
		S(3, -3), S(8, 1), S(2, 7), S(4, 9), S(5, 14), S(6, 6), S(6, -5), S(17, -7),
		S(11, -1), S(5, -10), S(14, -13), S(-6, 0), S(0, 2), S(11, -7), S(25, -4), S(14, -19),
		S(-3, -18), S(10, 0), S(-5, -17), S(-14, -3), S(-8, -7), S(-11, 0), S(8, -16), S(17, -33),
	},
	// Knight
	{
		S(-127, -57), S(-101, -21), S(-62, 0), S(-24, -14), S(8, -9), S(-58, -30), S(-69, -23), S(-79, -77),
		S(-12, -16), S(4, -1), S(39, -4), S(48, -1), S(33, -8), S(93, -23), S(17, -11), S(28, -33),
		S(1, -9), S(39, -3), S(39, 18), S(56, 17), S(92, 3), S(97, -1), S(62, -9), S(37, -21),
		S(6, 0), S(16, 13), S(35, 24), S(60, 29), S(45, 27), S(63, 23), S(32, 12), S(46, -6),
		S(-3, 3), S(4, 3), S(13, 26), S(22, 28), S(26, 29), S(29, 19), S(28, 8), S(9, -3),
		S(-22, -13), S(-7, -3), S(0, 2), S(3, 19), S(16, 18), S(3, -1), S(18, -9), S(-9, -8),
		S(-28, -19), S(-21, -9), S(-9, -7), S(6, -3), S(4, -3), S(6, -8), S(-1, -17), S(-5, -6),
		S(-66, -5), S(-24, -31), S(-35, -17), S(-20, -11), S(-15, -9), S(-8, -19), S(-19, -23), S(-27, -17),
	},
	// Pawn
	{
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
		S(94, 188), S(103, 178), S(73, 175), S(119, 115), S(94, 115), S(79, 131), S(7, 178), S(-4, 195),
		S(12, 113), S(23, 117), S(53, 76), S(68, 35), S(75, 28), S(86, 47), S(70, 91), S(15, 95),
		S(-24, 60), S(-4, 49), S(-2, 28), S(1, 16), S(23, 5), S(14, 13), S(16, 33), S(-3, 34),
		S(-29, 32), S(-9, 30), S(-7, 10), S(6, 6), S(6, 5), S(2, 7), S(6, 21), S(-11, 11),
		S(-30, 25), S(-4, 26), S(-7, 9), S(-2, 20), S(12, 14), S(-1, 12), S(30, 16), S(-4, 7),
		S(-27, 30), S(-2, 30), S(-6, 16), S(-14, 30), S(3, 32), S(22, 17), S(42, 17), S(-7, 7),
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
	},
}

// PieceSquaresScores stores the score of each piece on each square from Black persctive/ point of view
var PieceSquaresScores = [6][64]Score{}

// generatePsqtScores generates an array of scores with the sum of the Psqt and Piece Value for each piece on each square
func generatePsqtScores() {
	for p := range 6 {
		mgVal, egVal := PieceValues[p].Get()
		for sq := range 64 {
			mgPsqt, egPsqt := Psqt[p][sq].Get()

			PieceSquaresScores[p][sq] = S(mgVal+mgPsqt, egVal+egPsqt)
		}
	}
}
