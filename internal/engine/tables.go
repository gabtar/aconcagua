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
var PiecesValues = [6]Score{S(0, 0), S(943, 1117), S(417, 615), S(327, 371), S(292, 358), S(43, 91)}

// Psqt stores the values for mg and eg phase for pieces square tables
var Psqt = [6][64]Score{
	// King
	{
		S(0, -115), S(45, -49), S(59, -31), S(-8, 0), S(-8, -2), S(-10, 10), S(16, 0), S(29, -105),
		S(-41, -30), S(16, 23), S(-1, 35), S(75, 24), S(25, 37), S(14, 56), S(15, 44), S(-25, -6),
		S(-71, -12), S(45, 26), S(10, 46), S(-9, 56), S(22, 57), S(73, 46), S(28, 48), S(-16, -6),
		S(-36, -30), S(-24, 24), S(-37, 48), S(-69, 64), S(-65, 59), S(-52, 51), S(-52, 32), S(-99, -6),
		S(-45, -42), S(-31, 9), S(-59, 40), S(-94, 60), S(-93, 57), S(-61, 35), S(-61, 16), S(-105, -19),
		S(-13, -51), S(-2, 0), S(-43, 28), S(-55, 43), S(-46, 40), S(-48, 26), S(-15, 4), S(-35, -32),
		S(47, -64), S(3, -5), S(-9, 12), S(-29, 23), S(-36, 28), S(-19, 15), S(4, -4), S(31, -53),
		S(29, -95), S(31, -40), S(21, -15), S(-42, 3), S(5, -14), S(-38, 2), S(16, -26), S(33, -92),
	},
	// Queen
	{
		S(-40, 25), S(-48, 39), S(-36, 55), S(-11, 45), S(-17, 45), S(-2, 34), S(31, -11), S(-10, 35),
		S(6, -6), S(-18, 15), S(-23, 49), S(-48, 77), S(-61, 93), S(-17, 48), S(-8, 29), S(51, 11),
		S(8, 16), S(0, 28), S(-1, 55), S(-7, 63), S(-11, 72), S(19, 44), S(24, 17), S(26, 24),
		S(-9, 42), S(-3, 54), S(-6, 55), S(-11, 71), S(-7, 68), S(4, 48), S(18, 56), S(12, 46),
		S(2, 30), S(-7, 56), S(-5, 60), S(1, 73), S(3, 68), S(6, 53), S(15, 45), S(18, 37),
		S(0, 16), S(8, 28), S(3, 51), S(3, 55), S(9, 58), S(12, 47), S(24, 26), S(18, 20),
		S(9, 5), S(9, 11), S(16, 15), S(20, 26), S(16, 29), S(24, 0), S(32, -24), S(40, -40),
		S(6, 7), S(7, 10), S(11, 17), S(18, 18), S(15, 5), S(1, 1), S(10, -8), S(18, -20),
	},
	// Rook
	{
		S(-10, 48), S(-14, 51), S(-26, 61), S(-22, 55), S(-8, 50), S(17, 44), S(28, 42), S(27, 41),
		S(-4, 31), S(-6, 40), S(7, 41), S(20, 31), S(5, 31), S(28, 27), S(16, 29), S(37, 19),
		S(-11, 41), S(14, 39), S(10, 40), S(12, 37), S(32, 26), S(46, 19), S(65, 20), S(33, 18),
		S(-13, 43), S(1, 37), S(2, 44), S(11, 34), S(10, 27), S(20, 23), S(19, 26), S(7, 25),
		S(-21, 35), S(-22, 38), S(-12, 34), S(-5, 31), S(-3, 28), S(-12, 29), S(7, 22), S(-8, 22),
		S(-23, 29), S(-18, 26), S(-13, 23), S(-11, 24), S(-3, 20), S(3, 12), S(23, 0), S(4, 3),
		S(-22, 22), S(-15, 24), S(-6, 22), S(-4, 21), S(0, 13), S(9, 7), S(23, -2), S(-17, 13),
		S(-3, 28), S(-3, 25), S(0, 29), S(6, 21), S(11, 15), S(8, 19), S(5, 15), S(2, 10),
	},
	// Bishop
	{
		S(-27, 4), S(-53, 11), S(-61, 9), S(-100, 21), S(-89, 17), S(-68, 3), S(-45, 8), S(-58, 0),
		S(-23, -6), S(-20, 1), S(-18, 1), S(-26, 2), S(-15, -1), S(-6, -2), S(-25, 4), S(-25, -4),
		S(-4, 10), S(-3, 2), S(-5, 5), S(0, -2), S(-6, 2), S(34, 3), S(8, 2), S(17, 8),
		S(-18, 6), S(-1, 7), S(-3, 6), S(2, 20), S(11, 8), S(-1, 11), S(3, 2), S(-23, 14),
		S(-5, 1), S(-13, 7), S(-7, 16), S(12, 14), S(6, 12), S(-5, 9), S(-7, 8), S(8, -7),
		S(0, 2), S(5, 11), S(5, 11), S(3, 16), S(5, 21), S(9, 8), S(10, 0), S(14, 0),
		S(13, 13), S(8, -7), S(14, -3), S(-3, 4), S(6, 6), S(13, 0), S(33, -5), S(20, -5),
		S(6, -3), S(18, 2), S(1, -3), S(-7, 2), S(1, -1), S(-1, 10), S(11, -3), S(28, -16),
	},
	// Knight
	{
		S(-118, -30), S(-97, -7), S(-66, 8), S(-28, -9), S(-2, 1), S(-58, -19), S(-71, -6), S(-80, -57),
		S(-12, -5), S(-3, 4), S(18, -1), S(22, 6), S(26, -4), S(58, -20), S(19, -6), S(20, -23),
		S(0, -8), S(13, 0), S(16, 14), S(27, 13), S(48, 3), S(70, -15), S(27, -10), S(27, -19),
		S(10, 4), S(9, 8), S(24, 16), S(30, 24), S(34, 19), S(40, 16), S(26, 9), S(37, -1),
		S(5, 4), S(9, 1), S(15, 20), S(21, 20), S(18, 29), S(23, 13), S(18, 8), S(12, 5),
		S(-8, -3), S(1, 1), S(4, 5), S(6, 22), S(19, 19), S(6, 2), S(20, -2), S(3, 4),
		S(-9, 0), S(-7, 1), S(-1, 0), S(14, 1), S(12, 0), S(14, -2), S(17, -5), S(17, 13),
		S(-42, 13), S(-2, -9), S(-19, -4), S(-6, -3), S(2, 0), S(6, -8), S(1, -2), S(-5, 4),
	},
	// Pawn
	{
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
		S(83, 177), S(77, 169), S(65, 161), S(103, 106), S(83, 106), S(75, 110), S(-11, 163), S(-21, 185),
		S(17, 71), S(4, 73), S(21, 35), S(25, -7), S(32, -12), S(67, 9), S(33, 51), S(5, 54),
		S(-2, 49), S(-4, 36), S(0, 19), S(1, 1), S(18, 1), S(14, 10), S(2, 27), S(2, 24),
		S(-3, 29), S(-7, 26), S(0, 11), S(10, 2), S(13, 2), S(7, 10), S(0, 16), S(-2, 12),
		S(-10, 22), S(-10, 17), S(-8, 12), S(-2, 11), S(3, 13), S(-4, 13), S(8, 4), S(0, 4),
		S(-3, 26), S(-1, 23), S(-4, 18), S(-10, 14), S(-4, 25), S(13, 14), S(22, 7), S(-1, 5),
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
