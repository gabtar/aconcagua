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
var PiecesValues = [6]Score{S(0, 0), S(932, 1105), S(429, 597), S(337, 359), S(300, 347), S(46, 88)}

// Psqt stores the values for mg and eg phase for pieces square tables
var Psqt = [6][64]Score{
	// King
	{
		S(-7, -102), S(44, -46), S(56, -27), S(-13, 2), S(-14, 0), S(-14, 8), S(11, 0), S(25, -88),
		S(-47, -26), S(9, 25), S(-6, 34), S(67, 25), S(19, 37), S(7, 55), S(5, 42), S(-30, -4),
		S(-82, -11), S(40, 24), S(4, 43), S(-15, 53), S(12, 55), S(71, 46), S(22, 46), S(-32, -1),
		S(-52, -25), S(-35, 24), S(-50, 47), S(-83, 63), S(-84, 59), S(-63, 51), S(-64, 33), S(-126, 0),
		S(-60, -35), S(-33, 10), S(-67, 41), S(-109, 61), S(-104, 57), S(-70, 37), S(-73, 19), S(-123, -13),
		S(-13, -44), S(6, 0), S(-40, 29), S(-54, 44), S(-46, 42), S(-52, 30), S(-11, 6), S(-35, -27),
		S(58, -64), S(12, -5), S(-1, 12), S(-21, 23), S(-29, 27), S(-13, 15), S(11, -4), S(39, -53),
		S(41, -100), S(44, -47), S(33, -21), S(-32, -1), S(15, -19), S(-27, -4), S(28, -34), S(46, -99),
	},
	// Queen
	{
		S(-25, 14), S(-43, 33), S(-28, 51), S(-4, 43), S(-3, 42), S(3, 36), S(42, -8), S(3, 22),
		S(-9, 14), S(-35, 33), S(-41, 63), S(-58, 83), S(-57, 101), S(-10, 53), S(-14, 40), S(43, 19),
		S(1, 23), S(-9, 36), S(-6, 53), S(-17, 61), S(-9, 69), S(29, 43), S(38, 10), S(38, 10),
		S(-13, 44), S(-6, 55), S(-12, 53), S(-17, 65), S(-24, 76), S(-11, 58), S(6, 62), S(0, 51),
		S(1, 28), S(-13, 58), S(-8, 59), S(-4, 70), S(-3, 62), S(-2, 52), S(5, 49), S(9, 42),
		S(-2, 16), S(5, 27), S(0, 49), S(0, 49), S(5, 52), S(6, 46), S(18, 25), S(12, 17),
		S(8, 5), S(7, 8), S(14, 10), S(18, 16), S(14, 21), S(19, -2), S(26, -27), S(38, -47),
		S(6, 5), S(6, 6), S(11, 9), S(16, 13), S(12, 5), S(-1, 1), S(10, -13), S(17, -25),
	},
	// Rook
	{
		S(5, 45), S(2, 48), S(-4, 55), S(-1, 49), S(14, 43), S(31, 39), S(35, 40), S(43, 38),
		S(-4, 44), S(-5, 53), S(13, 52), S(26, 43), S(11, 42), S(39, 32), S(28, 33), S(47, 26),
		S(-13, 41), S(13, 39), S(9, 39), S(11, 35), S(32, 25), S(45, 18), S(74, 15), S(37, 15),
		S(-15, 42), S(0, 37), S(0, 43), S(9, 33), S(7, 26), S(19, 20), S(21, 24), S(7, 23),
		S(-23, 35), S(-24, 37), S(-15, 34), S(-9, 31), S(-6, 27), S(-15, 27), S(7, 19), S(-8, 20),
		S(-26, 29), S(-20, 26), S(-17, 23), S(-14, 24), S(-6, 19), S(0, 11), S(21, 0), S(2, 3),
		S(-24, 23), S(-16, 23), S(-8, 22), S(-6, 20), S(-1, 12), S(4, 6), S(20, -3), S(-20, 14),
		S(-5, 28), S(-3, 24), S(0, 28), S(6, 21), S(11, 14), S(8, 17), S(5, 15), S(1, 10),
	},
	// Bishop
	{
		S(-24, -1), S(-51, 6), S(-59, 3), S(-95, 14), S(-81, 8), S(-68, 0), S(-40, 1), S(-57, -8),
		S(-18, -10), S(0, -6), S(-11, -1), S(-16, -1), S(6, -7), S(-1, -7), S(-6, 0), S(-23, -8),
		S(-4, 7), S(0, -1), S(2, 3), S(6, -3), S(-2, -1), S(35, 4), S(7, -1), S(16, 6),
		S(-14, 3), S(1, 6), S(0, 5), S(2, 21), S(11, 10), S(-1, 10), S(3, 2), S(-21, 8),
		S(-2, -1), S(-11, 6), S(-9, 16), S(10, 15), S(3, 13), S(-7, 10), S(-7, 5), S(7, -8),
		S(1, 2), S(6, 9), S(5, 11), S(2, 15), S(4, 19), S(8, 7), S(11, -1), S(16, -3),
		S(15, 10), S(9, -8), S(15, -6), S(-2, 1), S(7, 3), S(14, -4), S(32, -8), S(22, -8),
		S(7, -4), S(18, 7), S(2, -5), S(-6, -1), S(2, -4), S(0, 4), S(10, -7), S(27, -17),
	},
	// Knight
	{
		S(-118, -37), S(-97, -13), S(-59, 3), S(-21, -13), S(5, -6), S(-52, -27), S(-69, -17), S(-78, -67),
		S(-7, -9), S(7, 0), S(42, -11), S(46, -3), S(33, -9), S(86, -28), S(21, -6), S(28, -27),
		S(3, -11), S(23, -6), S(25, 12), S(35, 10), S(70, -5), S(75, -18), S(36, -17), S(24, -21),
		S(9, 4), S(7, 7), S(23, 16), S(31, 21), S(34, 18), S(41, 13), S(25, 5), S(36, -4),
		S(2, 3), S(7, 0), S(12, 20), S(18, 20), S(17, 27), S(23, 11), S(17, 3), S(10, 5),
		S(-11, -3), S(0, 1), S(1, 6), S(4, 21), S(18, 18), S(4, 2), S(18, -4), S(0, 2),
		S(-12, -1), S(-9, 0), S(-2, -1), S(13, 0), S(11, -2), S(11, -6), S(13, -8), S(12, 12),
		S(-46, 15), S(-5, -10), S(-21, -6), S(-8, -5), S(0, -2), S(3, -12), S(-1, -4), S(-10, 4),
	},
	// Pawn
	{
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
		S(84, 173), S(96, 158), S(70, 154), S(109, 103), S(85, 109), S(68, 122), S(-13, 167), S(-34, 188),
		S(18, 70), S(6, 71), S(26, 33), S(29, -6), S(38, -11), S(72, 12), S(34, 52), S(4, 57),
		S(-1, 47), S(-2, 35), S(0, 20), S(3, 1), S(23, 1), S(15, 11), S(4, 27), S(1, 24),
		S(-3, 29), S(-7, 26), S(0, 12), S(10, 3), S(13, 3), S(7, 11), S(-1, 18), S(-4, 13),
		S(-10, 21), S(-10, 18), S(-8, 13), S(-2, 11), S(3, 13), S(-5, 14), S(7, 6), S(-1, 5),
		S(-3, 26), S(-2, 23), S(-5, 18), S(-10, 14), S(-4, 24), S(12, 15), S(20, 9), S(-4, 7),
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
