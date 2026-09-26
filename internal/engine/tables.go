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
var PiecesValues = [6]Score{S(0, 0), S(931, 1090), S(420, 586), S(333, 353), S(298, 341), S(45, 87)}

// Psqt stores the values for mg and eg phase for pieces square tables
var Psqt = [6][64]Score{
	// King
	{
		S(-6, -107), S(45, -48), S(58, -29), S(-13, 1), S(-13, -1), S(-15, 7), S(10, -1), S(25, -93),
		S(-47, -28), S(12, 24), S(-3, 33), S(72, 23), S(22, 37), S(9, 54), S(7, 42), S(-32, -6),
		S(-80, -13), S(45, 23), S(10, 41), S(-10, 51), S(18, 53), S(76, 43), S(25, 44), S(-31, -4),
		S(-47, -28), S(-29, 22), S(-42, 45), S(-75, 59), S(-75, 56), S(-58, 48), S(-61, 31), S(-122, -2),
		S(-56, -38), S(-28, 7), S(-60, 38), S(-100, 57), S(-95, 53), S(-64, 33), S(-69, 17), S(-118, -16),
		S(-13, -48), S(2, 0), S(-38, 26), S(-52, 40), S(-43, 38), S(-48, 26), S(-12, 4), S(-34, -30),
		S(55, -65), S(7, -5), S(-6, 12), S(-25, 21), S(-33, 26), S(-16, 14), S(8, -4), S(37, -53),
		S(37, -96), S(38, -41), S(27, -17), S(-37, 1), S(10, -16), S(-32, 0), S(22, -28), S(41, -93),
	},
	// Queen
	{
		S(-24, 15), S(-38, 32), S(-26, 47), S(-2, 39), S(-6, 36), S(-1, 30), S(40, -13), S(5, 21),
		S(-8, 9), S(-34, 32), S(-36, 62), S(-56, 82), S(-65, 93), S(-17, 46), S(-15, 35), S(46, 15),
		S(0, 19), S(-9, 32), S(-5, 52), S(-11, 59), S(-2, 62), S(22, 34), S(30, 4), S(31, 5),
		S(-17, 44), S(-9, 55), S(-12, 55), S(-14, 68), S(-13, 68), S(-1, 48), S(13, 54), S(5, 47),
		S(0, 27), S(-16, 59), S(-12, 63), S(-7, 76), S(-3, 66), S(0, 50), S(7, 46), S(12, 36),
		S(-5, 17), S(2, 30), S(-4, 53), S(-4, 56), S(1, 58), S(4, 47), S(17, 25), S(12, 16),
		S(4, 7), S(4, 11), S(11, 14), S(14, 24), S(11, 25), S(18, -5), S(25, -29), S(36, -46),
		S(0, 8), S(1, 10), S(6, 16), S(13, 16), S(10, 3), S(-3, -3), S(6, -14), S(15, -24),
	},
	// Rook
	{
		S(6, 46), S(2, 48), S(-4, 57), S(-1, 51), S(13, 45), S(31, 41), S(37, 41), S(43, 38),
		S(-12, 40), S(-14, 48), S(3, 48), S(17, 38), S(1, 38), S(28, 31), S(18, 32), S(39, 24),
		S(-9, 43), S(17, 40), S(12, 41), S(13, 38), S(34, 28), S(48, 21), S(73, 19), S(36, 19),
		S(-11, 43), S(4, 38), S(4, 44), S(13, 35), S(11, 28), S(21, 24), S(21, 27), S(8, 26),
		S(-18, 35), S(-20, 37), S(-11, 35), S(-4, 32), S(-1, 28), S(-11, 30), S(9, 22), S(-4, 22),
		S(-20, 29), S(-15, 26), S(-11, 23), S(-9, 24), S(-1, 19), S(4, 13), S(23, 3), S(6, 5),
		S(-18, 22), S(-11, 23), S(-2, 21), S(0, 20), S(4, 12), S(10, 7), S(24, -1), S(-14, 14),
		S(1, 26), S(2, 23), S(5, 28), S(12, 21), S(17, 14), S(14, 17), S(10, 16), S(8, 10),
	},
	// Bishop
	{
		S(-24, 0), S(-52, 6), S(-60, 3), S(-96, 14), S(-83, 10), S(-68, 0), S(-41, 1), S(-57, -7),
		S(-18, -10), S(1, -7), S(-11, -2), S(-17, -1), S(5, -8), S(-3, -6), S(-6, -2), S(-25, -7),
		S(-4, 7), S(0, -2), S(2, 3), S(5, -4), S(-4, 0), S(33, 4), S(5, 0), S(13, 8),
		S(-14, 2), S(0, 5), S(-1, 4), S(1, 20), S(10, 9), S(-3, 11), S(2, 1), S(-23, 10),
		S(-3, -1), S(-12, 6), S(-9, 15), S(10, 14), S(3, 12), S(-7, 9), S(-7, 5), S(7, -7),
		S(1, 1), S(5, 9), S(4, 10), S(2, 13), S(4, 17), S(9, 6), S(10, -2), S(16, -3),
		S(14, 10), S(8, -8), S(15, -6), S(-2, 1), S(7, 2), S(14, -6), S(33, -9), S(21, -8),
		S(7, -5), S(17, 7), S(1, -6), S(-6, -1), S(2, -5), S(-1, 4), S(10, -8), S(28, -19),
	},
	// Knight
	{
		S(-118, -37), S(-97, -14), S(-61, 4), S(-21, -13), S(5, -5), S(-52, -26), S(-70, -16), S(-79, -64),
		S(-7, -9), S(7, 0), S(42, -10), S(45, -3), S(31, -9), S(84, -28), S(21, -8), S(28, -27),
		S(3, -11), S(23, -6), S(25, 11), S(32, 11), S(67, -3), S(74, -17), S(35, -14), S(23, -19),
		S(9, 3), S(7, 7), S(23, 15), S(30, 21), S(32, 19), S(39, 14), S(23, 7), S(35, -3),
		S(2, 2), S(6, 0), S(12, 19), S(18, 19), S(16, 27), S(21, 12), S(16, 5), S(9, 4),
		S(-10, -4), S(0, 0), S(1, 5), S(4, 19), S(17, 17), S(4, 1), S(18, -4), S(1, 1),
		S(-11, -3), S(-9, -1), S(-2, -2), S(13, -1), S(11, -3), S(12, -7), S(14, -10), S(14, 9),
		S(-47, 12), S(-4, -12), S(-21, -7), S(-8, -6), S(0, -3), S(4, -13), S(0, -6), S(-11, 3),
	},
	// Pawn
	{
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
		S(81, 172), S(94, 157), S(69, 154), S(109, 102), S(86, 106), S(68, 119), S(-13, 166), S(-34, 187),
		S(17, 66), S(6, 69), S(26, 30), S(28, -7), S(36, -13), S(70, 10), S(32, 50), S(3, 55),
		S(-2, 45), S(-2, 34), S(0, 19), S(3, 0), S(22, 0), S(14, 11), S(4, 26), S(2, 23),
		S(-4, 27), S(-7, 25), S(0, 11), S(10, 3), S(13, 2), S(6, 11), S(0, 17), S(-3, 12),
		S(-11, 20), S(-11, 17), S(-9, 12), S(-3, 11), S(3, 13), S(-4, 13), S(8, 6), S(-1, 5),
		S(-4, 25), S(-2, 23), S(-5, 18), S(-10, 14), S(-5, 25), S(12, 15), S(21, 8), S(-3, 6),
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
