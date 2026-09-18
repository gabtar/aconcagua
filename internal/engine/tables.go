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
var PiecesValues = [6]Score{S(0, 0), S(961, 1130), S(422, 596), S(336, 358), S(298, 345), S(52, 98)}

// Psqt stores the values for mg and eg phase for pieces square tables
var Psqt = [6][64]Score{
	// King
	{
		S(-14, -91), S(38, -56), S(49, -37), S(-19, -4), S(-19, -4), S(-19, 7), S(10, 1), S(27, -81),
		S(-47, -17), S(-2, 13), S(-19, 21), S(54, 11), S(13, 28), S(1, 47), S(0, 39), S(-28, 12),
		S(-90, 3), S(25, 23), S(-14, 38), S(-27, 46), S(-1, 49), S(58, 49), S(7, 52), S(-36, 18),
		S(-66, -7), S(-54, 28), S(-71, 48), S(-99, 58), S(-104, 61), S(-81, 59), S(-82, 48), S(-138, 25),
		S(-72, -15), S(-52, 13), S(-87, 41), S(-122, 59), S(-126, 59), S(-92, 46), S(-99, 32), S(-135, 14),
		S(-26, -25), S(-4, 0), S(-66, 26), S(-86, 40), S(-80, 41), S(-77, 32), S(-29, 10), S(-48, -3),
		S(69, -46), S(24, -15), S(4, 0), S(-35, 13), S(-39, 17), S(-15, 6), S(37, -14), S(49, -35),
		S(55, -86), S(86, -63), S(59, -41), S(-46, -17), S(20, -39), S(-17, -20), S(64, -53), S(62, -84),
	},
	// Queen
	{
		S(-15, 20), S(-29, 48), S(-12, 70), S(15, 61), S(12, 67), S(19, 58), S(52, 11), S(18, 41),
		S(-3, 3), S(-38, 45), S(-31, 76), S(-37, 97), S(-38, 123), S(4, 77), S(-2, 57), S(62, 35),
		S(1, 11), S(-8, 26), S(-11, 63), S(-6, 80), S(8, 94), S(55, 74), S(63, 41), S(71, 39),
		S(-15, 25), S(-14, 43), S(-18, 55), S(-21, 84), S(-19, 101), S(0, 82), S(13, 77), S(18, 63),
		S(-7, 16), S(-20, 47), S(-18, 51), S(-17, 77), S(-16, 73), S(-11, 63), S(2, 53), S(13, 46),
		S(-10, 5), S(-5, 16), S(-14, 40), S(-12, 36), S(-10, 41), S(-3, 36), S(11, 18), S(14, 9),
		S(-1, -5), S(-5, -4), S(2, -5), S(4, 0), S(3, 5), S(10, -18), S(19, -47), S(35, -60),
		S(-7, -5), S(-2, -7), S(1, -7), S(8, -4), S(4, -9), S(-8, -10), S(7, -23), S(7, -34),
	},
	// Rook
	{
		S(10, 44), S(3, 47), S(-4, 56), S(-1, 51), S(16, 45), S(34, 41), S(38, 40), S(51, 36),
		S(-9, 45), S(-10, 55), S(7, 55), S(27, 43), S(17, 42), S(47, 33), S(43, 32), S(70, 21),
		S(-21, 43), S(6, 41), S(2, 42), S(6, 39), S(36, 25), S(53, 19), S(96, 13), S(65, 10),
		S(-24, 44), S(-7, 39), S(-11, 47), S(-4, 39), S(2, 28), S(16, 23), S(34, 22), S(25, 19),
		S(-33, 37), S(-33, 38), S(-26, 37), S(-20, 33), S(-16, 30), S(-18, 29), S(16, 18), S(-2, 18),
		S(-35, 28), S(-29, 26), S(-27, 23), S(-25, 25), S(-12, 18), S(-3, 11), S(31, -4), S(6, 0),
		S(-33, 21), S(-27, 22), S(-18, 22), S(-16, 21), S(-7, 11), S(2, 6), S(18, -1), S(-21, 12),
		S(-12, 25), S(-10, 23), S(-8, 28), S(0, 20), S(8, 12), S(5, 19), S(12, 12), S(-3, 7),
	},
	// Bishop
	{
		S(-26, 0), S(-51, 6), S(-58, 4), S(-95, 15), S(-78, 12), S(-66, 0), S(-40, 2), S(-48, -6),
		S(-20, -10), S(-9, -6), S(-19, -1), S(-23, -1), S(2, -9), S(-3, -5), S(-7, -4), S(-7, -11),
		S(-9, 9), S(-3, -2), S(-3, 2), S(3, -6), S(0, -1), S(38, 1), S(25, -2), S(25, 6),
		S(-18, 5), S(-3, 5), S(-4, 3), S(11, 15), S(10, 6), S(10, 6), S(5, 0), S(-10, 9),
		S(-6, 0), S(-16, 6), S(-4, 12), S(10, 11), S(12, 8), S(-6, 6), S(-9, 5), S(12, -10),
		S(0, 2), S(8, 6), S(0, 10), S(5, 13), S(5, 17), S(5, 7), S(6, -2), S(17, -3),
		S(20, 9), S(5, -8), S(16, -7), S(-5, 2), S(2, 3), S(12, -5), S(25, -4), S(16, -7),
		S(4, -3), S(23, 7), S(3, -4), S(-8, 0), S(-2, -2), S(-5, 8), S(14, -8), S(21, -14),
	},
	// Knight
	{
		S(-117, -42), S(-99, -14), S(-57, 2), S(-20, -11), S(12, -4), S(-50, -24), S(-65, -18), S(-69, -67),
		S(-6, -12), S(1, 0), S(37, -8), S(47, -1), S(36, -9), S(92, -30), S(16, -8), S(40, -29),
		S(0, -10), S(17, -3), S(25, 12), S(35, 12), S(76, -4), S(86, -15), S(54, -16), S(46, -22),
		S(3, 4), S(8, 6), S(18, 18), S(48, 18), S(39, 16), S(56, 12), S(28, 8), S(52, -7),
		S(-3, 3), S(0, 2), S(8, 20), S(17, 20), S(23, 26), S(25, 13), S(32, 2), S(14, 2),
		S(-17, -3), S(-8, 2), S(-4, 6), S(-2, 23), S(13, 20), S(0, 2), S(16, -3), S(-3, 1),
		S(-19, -4), S(-16, 0), S(-10, 0), S(8, 0), S(6, -1), S(8, -5), S(4, -6), S(3, 8),
		S(-60, 5), S(-6, -11), S(-27, -5), S(-12, -4), S(-5, 0), S(-1, -9), S(-1, -8), S(-24, -6),
	},
	// Pawn
	{
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
		S(73, 175), S(86, 166), S(55, 162), S(100, 103), S(74, 103), S(61, 118), S(-13, 165), S(-26, 183),
		S(4, 100), S(-6, 106), S(23, 65), S(38, 25), S(43, 18), S(68, 30), S(38, 76), S(18, 77),
		S(-11, 45), S(-13, 32), S(-9, 16), S(-5, -3), S(13, -3), S(18, 1), S(5, 19), S(11, 19),
		S(-13, 26), S(-19, 23), S(-9, 9), S(3, -1), S(4, 0), S(7, 5), S(-1, 11), S(0, 7),
		S(-18, 19), S(-21, 15), S(-16, 11), S(-9, 9), S(2, 11), S(-2, 12), S(13, 3), S(1, 2),
		S(-9, 24), S(-11, 21), S(-10, 18), S(-9, 15), S(0, 23), S(26, 13), S(36, 4), S(4, 2),
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
