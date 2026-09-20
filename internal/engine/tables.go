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
var PiecesValues = [6]Score{S(0, 0), S(948, 1126), S(424, 591), S(337, 354), S(299, 341), S(51, 85)}

// Psqt stores the values for mg and eg phase for pieces square tables
var Psqt = [6][64]Score{
	// King
	{
		S(-12, -91), S(39, -51), S(50, -31), S(-18, 0), S(-18, 0), S(-16, 10), S(12, 4), S(27, -79),
		S(-47, -16), S(0, 20), S(-17, 29), S(55, 18), S(12, 33), S(2, 53), S(1, 44), S(-26, 11),
		S(-87, 2), S(26, 27), S(-12, 44), S(-28, 52), S(-1, 54), S(59, 49), S(11, 51), S(-33, 15),
		S(-63, -10), S(-52, 27), S(-70, 49), S(-101, 62), S(-105, 61), S(-79, 55), S(-78, 40), S(-133, 18),
		S(-71, -19), S(-50, 12), S(-86, 43), S(-126, 61), S(-126, 59), S(-88, 40), S(-91, 25), S(-132, 4),
		S(-22, -31), S(0, 0), S(-59, 29), S(-77, 44), S(-70, 42), S(-67, 29), S(-21, 6), S(-42, -14),
		S(67, -56), S(18, -13), S(3, 5), S(-25, 18), S(-32, 21), S(-17, 9), S(24, -13), S(48, -48),
		S(52, -92), S(68, -58), S(55, -34), S(-31, -7), S(29, -30), S(-15, -14), S(48, -45), S(61, -95),
	},
	// Queen
	{
		S(-15, 18), S(-30, 48), S(-12, 69), S(14, 60), S(14, 64), S(20, 57), S(55, 9), S(19, 38),
		S(-5, 7), S(-38, 48), S(-32, 79), S(-38, 100), S(-37, 122), S(8, 77), S(-2, 59), S(60, 36),
		S(0, 13), S(-9, 28), S(-11, 61), S(-5, 79), S(10, 92), S(55, 70), S(65, 36), S(69, 34),
		S(-18, 27), S(-14, 42), S(-17, 53), S(-19, 81), S(-17, 98), S(0, 81), S(11, 78), S(16, 66),
		S(-8, 17), S(-21, 45), S(-17, 50), S(-16, 73), S(-14, 69), S(-11, 62), S(2, 53), S(11, 49),
		S(-11, 5), S(-5, 16), S(-14, 37), S(-11, 34), S(-10, 40), S(-3, 36), S(11, 19), S(12, 11),
		S(-2, -5), S(-5, -4), S(2, -5), S(4, 0), S(2, 4), S(10, -18), S(18, -46), S(32, -60),
		S(-7, -7), S(-3, -10), S(0, -8), S(6, -4), S(2, -11), S(-9, -11), S(3, -25), S(5, -34),
	},
	// Rook
	{
		S(7, 44), S(0, 47), S(-5, 55), S(0, 50), S(16, 43), S(33, 39), S(36, 40), S(50, 34),
		S(-11, 45), S(-9, 54), S(7, 54), S(24, 44), S(16, 42), S(50, 32), S(43, 31), S(65, 21),
		S(-22, 42), S(5, 40), S(2, 40), S(7, 37), S(36, 24), S(55, 17), S(97, 11), S(60, 9),
		S(-25, 42), S(-9, 38), S(-10, 45), S(-1, 36), S(2, 26), S(18, 21), S(31, 21), S(18, 18),
		S(-33, 35), S(-34, 37), S(-25, 36), S(-18, 32), S(-16, 29), S(-15, 28), S(16, 17), S(-4, 17),
		S(-35, 28), S(-30, 26), S(-27, 23), S(-24, 25), S(-11, 18), S(-1, 11), S(31, -4), S(4, 0),
		S(-32, 20), S(-27, 22), S(-17, 21), S(-14, 20), S(-6, 11), S(5, 5), S(21, -3), S(-22, 10),
		S(-12, 24), S(-11, 21), S(-8, 26), S(0, 18), S(8, 11), S(6, 16), S(9, 10), S(0, 6),
	},
	// Bishop
	{
		S(-25, -1), S(-51, 6), S(-59, 3), S(-95, 14), S(-76, 11), S(-66, 0), S(-38, 1), S(-49, -7),
		S(-21, -10), S(-9, -5), S(-19, -1), S(-24, 0), S(2, -8), S(0, -5), S(-9, -2), S(-9, -10),
		S(-9, 8), S(-4, -1), S(-3, 2), S(2, -5), S(1, -1), S(36, 1), S(25, -3), S(23, 5),
		S(-18, 3), S(-3, 5), S(-4, 3), S(12, 15), S(11, 6), S(10, 6), S(5, 0), S(-8, 7),
		S(-6, -1), S(-17, 6), S(-4, 13), S(11, 11), S(12, 8), S(-7, 7), S(-8, 5), S(12, -10),
		S(-1, 1), S(9, 6), S(1, 10), S(5, 13), S(6, 17), S(5, 7), S(7, -2), S(17, -3),
		S(20, 8), S(5, -8), S(16, -7), S(-4, 1), S(2, 4), S(11, -4), S(28, -6), S(18, -8),
		S(5, -4), S(23, 5), S(2, -6), S(-8, 0), S(-1, -4), S(-5, 7), S(10, -9), S(22, -14),
	},
	// Knight
	{
		S(-116, -40), S(-99, -13), S(-57, 2), S(-20, -11), S(12, -4), S(-48, -24), S(-65, -16), S(-68, -66),
		S(-7, -8), S(1, 1), S(37, -8), S(47, -2), S(35, -9), S(92, -30), S(15, -6), S(40, -28),
		S(0, -9), S(16, -3), S(26, 10), S(35, 10), S(75, -6), S(84, -17), S(53, -16), S(44, -22),
		S(3, 5), S(7, 6), S(19, 16), S(49, 16), S(40, 15), S(56, 10), S(27, 7), S(50, -7),
		S(-3, 4), S(0, 0), S(8, 19), S(17, 19), S(23, 24), S(25, 11), S(33, 2), S(15, 3),
		S(-17, -2), S(-8, 1), S(-4, 5), S(-2, 20), S(14, 18), S(0, 2), S(16, -3), S(-3, 2),
		S(-19, -2), S(-16, 0), S(-9, -1), S(8, 0), S(6, -2), S(8, -6), S(6, -5), S(7, 11),
		S(-58, 7), S(-8, -10), S(-27, -6), S(-13, -4), S(-4, 0), S(-2, -9), S(-4, -6), S(-21, -2),
	},
	// Pawn
	{
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
		S(77, 173), S(87, 158), S(60, 157), S(101, 101), S(73, 101), S(60, 115), S(-17, 160), S(-27, 181),
		S(11, 80), S(-4, 86), S(19, 45), S(26, 6), S(32, 0), S(70, 16), S(33, 59), S(14, 60),
		S(-9, 47), S(-12, 36), S(-7, 19), S(-5, 0), S(16, 0), S(12, 7), S(4, 23), S(5, 21),
		S(-11, 29), S(-17, 27), S(-7, 13), S(5, 3), S(6, 4), S(6, 9), S(0, 15), S(-2, 11),
		S(-16, 22), S(-19, 19), S(-13, 13), S(-7, 12), S(3, 13), S(-3, 14), S(13, 5), S(1, 4),
		S(-8, 26), S(-9, 25), S(-8, 19), S(-11, 16), S(0, 25), S(19, 14), S(33, 6), S(0, 4),
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
