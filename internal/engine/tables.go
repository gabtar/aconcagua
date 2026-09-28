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
var PiecesValues = [6]Score{S(0, 0), S(932, 1087), S(422, 583), S(335, 351), S(299, 339), S(45, 86)}

// Psqt stores the values for mg and eg phase for pieces square tables
var Psqt = [6][64]Score{
	// King
	{
		S(-4, -109), S(46, -48), S(59, -29), S(-12, 1), S(-12, -1), S(-14, 8), S(11, 0), S(26, -95),
		S(-46, -28), S(14, 23), S(-1, 32), S(75, 22), S(24, 36), S(11, 54), S(9, 42), S(-31, -5),
		S(-77, -14), S(48, 22), S(13, 40), S(-8, 50), S(22, 51), S(79, 42), S(28, 43), S(-28, -4),
		S(-43, -29), S(-25, 21), S(-39, 43), S(-72, 58), S(-71, 54), S(-56, 47), S(-59, 30), S(-117, -3),
		S(-53, -39), S(-27, 7), S(-60, 37), S(-100, 56), S(-95, 53), S(-65, 33), S(-67, 17), S(-113, -17),
		S(-12, -48), S(0, 0), S(-41, 26), S(-55, 40), S(-45, 38), S(-50, 25), S(-13, 4), S(-32, -30),
		S(55, -64), S(7, -5), S(-7, 11), S(-27, 21), S(-34, 25), S(-17, 14), S(7, -4), S(36, -52),
		S(35, -93), S(36, -39), S(25, -16), S(-39, 2), S(8, -15), S(-33, 0), S(21, -27), S(39, -91),
	},
	// Queen
	{
		S(-25, 15), S(-36, 32), S(-23, 49), S(0, 40), S(-4, 39), S(0, 31), S(42, -13), S(4, 22),
		S(-1, 3), S(-28, 30), S(-31, 62), S(-53, 84), S(-65, 98), S(-18, 51), S(-14, 37), S(47, 16),
		S(0, 16), S(-9, 30), S(-6, 53), S(-13, 62), S(-5, 65), S(19, 38), S(24, 10), S(24, 12),
		S(-17, 40), S(-8, 52), S(-14, 55), S(-18, 72), S(-13, 69), S(-2, 49), S(12, 53), S(3, 46),
		S(0, 24), S(-16, 56), S(-12, 61), S(-8, 76), S(-5, 67), S(-1, 51), S(6, 46), S(12, 35),
		S(-5, 14), S(2, 27), S(-4, 50), S(-3, 52), S(2, 54), S(4, 44), S(17, 23), S(12, 14),
		S(4, 4), S(4, 7), S(11, 10), S(15, 19), S(11, 21), S(18, -8), S(26, -32), S(36, -47),
		S(0, 6), S(2, 6), S(6, 12), S(13, 13), S(10, 0), S(-2, -7), S(5, -16), S(13, -25),
	},
	// Rook
	{
		S(5, 46), S(1, 48), S(-6, 57), S(-3, 52), S(11, 46), S(30, 41), S(38, 40), S(41, 38),
		S(-8, 37), S(-10, 46), S(4, 46), S(16, 37), S(0, 38), S(24, 32), S(14, 32), S(38, 23),
		S(-9, 42), S(17, 40), S(11, 41), S(12, 38), S(32, 28), S(46, 22), S(69, 20), S(33, 20),
		S(-10, 42), S(4, 37), S(5, 43), S(13, 34), S(11, 27), S(20, 24), S(19, 27), S(7, 26),
		S(-17, 34), S(-20, 36), S(-10, 34), S(-4, 31), S(-1, 27), S(-12, 29), S(7, 23), S(-5, 21),
		S(-20, 27), S(-14, 25), S(-11, 22), S(-8, 23), S(0, 19), S(4, 12), S(22, 3), S(7, 3),
		S(-17, 21), S(-10, 21), S(-1, 20), S(0, 19), S(5, 11), S(10, 6), S(24, -2), S(-13, 13),
		S(1, 25), S(2, 22), S(5, 27), S(12, 20), S(17, 13), S(14, 17), S(10, 15), S(8, 9),
	},
	// Bishop
	{
		S(-24, 0), S(-52, 6), S(-60, 3), S(-98, 14), S(-85, 11), S(-69, 0), S(-42, 1), S(-58, -7),
		S(-18, -11), S(1, -7), S(-11, -2), S(-17, -2), S(4, -8), S(-3, -6), S(-7, -2), S(-25, -7),
		S(-4, 6), S(0, -3), S(2, 2), S(5, -4), S(-4, 0), S(33, 3), S(4, 0), S(13, 7),
		S(-14, 2), S(0, 5), S(-1, 4), S(1, 19), S(10, 8), S(-3, 10), S(2, 1), S(-23, 9),
		S(-3, -1), S(-12, 6), S(-9, 15), S(10, 13), S(3, 12), S(-7, 8), S(-7, 5), S(7, -8),
		S(1, 1), S(5, 8), S(4, 10), S(2, 13), S(4, 17), S(9, 5), S(10, -2), S(16, -3),
		S(14, 10), S(8, -9), S(15, -7), S(-2, 0), S(7, 2), S(14, -5), S(33, -9), S(21, -8),
		S(7, -5), S(17, 6), S(1, -6), S(-6, -1), S(2, -5), S(-1, 4), S(10, -8), S(28, -19),
	},
	// Knight
	{
		S(-118, -37), S(-97, -13), S(-62, 4), S(-22, -12), S(4, -5), S(-53, -25), S(-70, -15), S(-80, -62),
		S(-6, -9), S(7, 0), S(42, -11), S(44, -3), S(30, -9), S(84, -29), S(22, -8), S(27, -27),
		S(3, -11), S(23, -6), S(26, 11), S(32, 10), S(66, -4), S(74, -17), S(34, -14), S(22, -19),
		S(9, 3), S(7, 7), S(23, 15), S(30, 21), S(32, 18), S(39, 14), S(23, 6), S(34, -3),
		S(2, 1), S(6, 0), S(13, 19), S(18, 19), S(16, 26), S(21, 11), S(15, 5), S(9, 3),
		S(-10, -4), S(0, 0), S(1, 5), S(4, 19), S(17, 17), S(4, 1), S(18, -4), S(1, 1),
		S(-11, -3), S(-9, -1), S(-2, -2), S(13, -1), S(11, -3), S(11, -7), S(14, -10), S(14, 9),
		S(-46, 11), S(-4, -12), S(-21, -7), S(-8, -6), S(0, -3), S(4, -13), S(-1, -5), S(-10, 2),
	},
	// Pawn
	{
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
		S(81, 171), S(95, 156), S(69, 153), S(108, 102), S(86, 105), S(68, 119), S(-12, 165), S(-34, 186),
		S(18, 65), S(6, 68), S(26, 29), S(29, -8), S(37, -13), S(70, 10), S(32, 49), S(3, 54),
		S(-1, 45), S(-2, 34), S(0, 18), S(4, 0), S(22, 0), S(15, 10), S(4, 26), S(2, 23),
		S(-3, 27), S(-7, 25), S(0, 10), S(11, 2), S(14, 2), S(7, 10), S(0, 16), S(-3, 12),
		S(-10, 20), S(-10, 17), S(-8, 12), S(-2, 11), S(3, 13), S(-4, 13), S(8, 5), S(-1, 4),
		S(-3, 25), S(-2, 22), S(-5, 18), S(-10, 14), S(-5, 25), S(12, 15), S(20, 8), S(-3, 5),
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
