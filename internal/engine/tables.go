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
var PiecesValues = [6]Score{S(0, 0), S(936, 1091), S(428, 588), S(337, 353), S(300, 339), S(44, 87)}

// Psqt stores the values for mg and eg phase for pieces square tables
var Psqt = [6][64]Score{
	// King
	{
		S(-5, -107), S(45, -45), S(58, -27), S(-13, 3), S(-12, 0), S(-10, 10), S(16, 0), S(28, -92),
		S(-44, -28), S(11, 25), S(-7, 35), S(70, 24), S(20, 38), S(10, 55), S(11, 43), S(-24, -6),
		S(-76, -13), S(45, 24), S(7, 42), S(-14, 53), S(15, 53), S(74, 44), S(26, 45), S(-22, -4),
		S(-45, -28), S(-31, 23), S(-46, 46), S(-80, 61), S(-79, 57), S(-61, 49), S(-60, 31), S(-113, -3),
		S(-53, -38), S(-32, 9), S(-64, 39), S(-103, 58), S(-98, 54), S(-66, 34), S(-69, 18), S(-112, -17),
		S(-12, -48), S(1, 0), S(-40, 27), S(-54, 41), S(-43, 38), S(-48, 26), S(-12, 4), S(-33, -30),
		S(55, -64), S(8, -5), S(-5, 12), S(-24, 22), S(-32, 26), S(-15, 14), S(8, -3), S(36, -52),
		S(35, -93), S(36, -38), S(26, -15), S(-38, 3), S(10, -13), S(-33, 1), S(21, -25), S(39, -90),
	},
	// Queen
	{
		S(-33, 16), S(-44, 35), S(-30, 51), S(-5, 42), S(-6, 39), S(2, 34), S(40, -10), S(-3, 26),
		S(-3, 1), S(-30, 28), S(-34, 59), S(-57, 83), S(-65, 95), S(-18, 47), S(-16, 37), S(44, 18),
		S(7, 17), S(-1, 29), S(-1, 53), S(-8, 63), S(-10, 69), S(24, 39), S(32, 7), S(31, 16),
		S(-9, 39), S(-4, 52), S(-6, 52), S(-11, 69), S(-10, 72), S(1, 55), S(17, 59), S(10, 51),
		S(2, 28), S(-8, 54), S(-5, 57), S(1, 70), S(3, 67), S(6, 54), S(13, 48), S(17, 41),
		S(1, 14), S(9, 25), S(3, 47), S(4, 52), S(10, 55), S(12, 46), S(24, 26), S(17, 21),
		S(9, 4), S(10, 8), S(18, 12), S(21, 21), S(17, 24), S(25, -1), S(32, -25), S(42, -44),
		S(6, 5), S(8, 7), S(12, 13), S(19, 14), S(16, 0), S(1, 0), S(11, -11), S(19, -25),
	},
	// Rook
	{
		S(-8, 44), S(-8, 46), S(-17, 54), S(-13, 49), S(1, 43), S(26, 38), S(31, 38), S(34, 36),
		S(-11, 32), S(-13, 40), S(1, 41), S(13, 32), S(-3, 33), S(21, 27), S(14, 27), S(34, 18),
		S(-10, 38), S(17, 36), S(12, 37), S(13, 35), S(33, 25), S(48, 18), S(68, 18), S(33, 17),
		S(-12, 39), S(3, 34), S(4, 41), S(12, 33), S(11, 25), S(21, 21), S(18, 25), S(7, 24),
		S(-20, 32), S(-21, 35), S(-10, 32), S(-4, 30), S(-2, 27), S(-11, 28), S(6, 21), S(-7, 20),
		S(-22, 26), S(-16, 24), S(-13, 21), S(-10, 23), S(-2, 18), S(4, 11), S(22, 2), S(4, 2),
		S(-21, 19), S(-13, 20), S(-4, 19), S(-2, 18), S(1, 11), S(11, 6), S(24, -3), S(-17, 12),
		S(-2, 24), S(-1, 21), S(1, 26), S(8, 19), S(13, 13), S(10, 16), S(6, 14), S(4, 8),
	},
	// Bishop
	{
		S(-24, -1), S(-53, 5), S(-62, 2), S(-101, 14), S(-84, 9), S(-71, -1), S(-41, 0), S(-58, -7),
		S(-21, -10), S(-16, -5), S(-17, -3), S(-24, -2), S(-12, -7), S(-3, -8), S(-21, -1), S(-24, -8),
		S(-2, 4), S(0, -4), S(-3, 1), S(1, -6), S(-3, -3), S(37, 0), S(11, -2), S(19, 5),
		S(-16, 2), S(-1, 4), S(-2, 2), S(3, 16), S(12, 6), S(0, 7), S(5, 0), S(-21, 8),
		S(-4, -2), S(-12, 4), S(-6, 12), S(14, 11), S(7, 9), S(-4, 7), S(-6, 4), S(9, -9),
		S(0, 0), S(7, 7), S(6, 8), S(5, 12), S(6, 17), S(11, 5), S(12, -2), S(16, -4),
		S(14, 10), S(10, -9), S(16, -7), S(-1, 1), S(8, 3), S(15, -4), S(34, -9), S(21, -8),
		S(7, -6), S(18, 6), S(2, -5), S(-5, 0), S(2, -4), S(0, 6), S(12, -7), S(30, -19),
	},
	// Knight
	{
		S(-116, -36), S(-97, -13), S(-64, 4), S(-25, -12), S(3, -5), S(-54, -25), S(-68, -15), S(-78, -61),
		S(-10, -8), S(-1, 1), S(24, -8), S(27, 0), S(31, -9), S(65, -26), S(22, -9), S(23, -25),
		S(1, -11), S(17, -6), S(19, 10), S(30, 8), S(53, -1), S(75, -19), S(31, -14), S(29, -21),
		S(12, 2), S(11, 4), S(26, 13), S(32, 20), S(36, 16), S(42, 12), S(28, 5), S(39, -4),
		S(6, 1), S(10, -2), S(16, 17), S(22, 17), S(19, 26), S(25, 10), S(19, 4), S(14, 3),
		S(-7, -4), S(2, 0), S(4, 4), S(7, 19), S(20, 17), S(7, 1), S(21, -4), S(4, 1),
		S(-8, -2), S(-6, 0), S(0, -2), S(15, 0), S(13, -2), S(15, -5), S(18, -7), S(18, 10),
		S(-44, 13), S(-1, -11), S(-18, -7), S(-6, -5), S(3, -1), S(6, -9), S(2, -4), S(-5, 3),
	},
	// Pawn
	{
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
		S(83, 171), S(88, 158), S(68, 153), S(108, 101), S(83, 105), S(70, 117), S(-16, 166), S(-30, 186),
		S(19, 66), S(5, 68), S(24, 30), S(27, -8), S(35, -13), S(67, 11), S(32, 49), S(4, 54),
		S(-1, 45), S(-4, 35), S(0, 19), S(1, 0), S(19, 1), S(14, 10), S(1, 26), S(2, 23),
		S(-2, 27), S(-7, 25), S(0, 10), S(11, 2), S(14, 2), S(7, 10), S(-1, 16), S(-3, 12),
		S(-10, 20), S(-10, 17), S(-8, 12), S(-2, 11), S(3, 13), S(-5, 13), S(7, 5), S(-1, 5),
		S(-2, 25), S(-1, 22), S(-4, 17), S(-10, 14), S(-4, 24), S(13, 14), S(21, 8), S(-2, 6),
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
