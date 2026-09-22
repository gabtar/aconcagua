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
var PiecesValues = [6]Score{S(0, 0), S(951, 1130), S(429, 591), S(337, 353), S(300, 340), S(48, 85)}

// Psqt stores the values for mg and eg phase for pieces square tables
var Psqt = [6][64]Score{
	// King
	{
		S(-11, -92), S(40, -47), S(51, -27), S(-17, 2), S(-17, 2), S(-14, 12), S(13, 6), S(27, -79),
		S(-46, -17), S(1, 23), S(-15, 32), S(57, 22), S(13, 36), S(3, 56), S(2, 46), S(-25, 9),
		S(-86, -1), S(28, 26), S(-10, 44), S(-27, 53), S(0, 55), S(60, 47), S(13, 49), S(-33, 9),
		S(-62, -15), S(-50, 26), S(-68, 48), S(-100, 63), S(-103, 59), S(-78, 52), S(-76, 35), S(-132, 11),
		S(-70, -24), S(-48, 11), S(-85, 41), S(-126, 61), S(-124, 58), S(-88, 37), S(-88, 22), S(-130, -3),
		S(-20, -35), S(1, 0), S(-56, 29), S(-73, 44), S(-65, 42), S(-67, 29), S(-20, 5), S(-37, -20),
		S(69, -60), S(17, -11), S(0, 7), S(-24, 19), S(-30, 22), S(-16, 10), S(21, -12), S(51, -50),
		S(53, -94), S(62, -54), S(48, -30), S(-27, -6), S(28, -27), S(-18, -11), S(44, -40), S(62, -96),
	},
	// Queen
	{
		S(-15, 18), S(-29, 49), S(-11, 70), S(15, 62), S(15, 65), S(21, 57), S(56, 9), S(18, 37),
		S(-4, 7), S(-37, 48), S(-31, 79), S(-39, 101), S(-38, 122), S(7, 76), S(-2, 59), S(58, 36),
		S(0, 13), S(-10, 28), S(-10, 61), S(-6, 79), S(7, 91), S(51, 67), S(59, 32), S(62, 30),
		S(-17, 27), S(-12, 41), S(-15, 53), S(-19, 80), S(-17, 97), S(0, 80), S(9, 78), S(10, 66),
		S(-6, 15), S(-20, 46), S(-17, 49), S(-15, 73), S(-14, 69), S(-11, 63), S(-1, 55), S(6, 50),
		S(-10, 5), S(-3, 15), S(-11, 37), S(-11, 35), S(-9, 41), S(-4, 40), S(8, 22), S(8, 13),
		S(-1, -5), S(-2, -4), S(4, -5), S(6, 0), S(5, 4), S(13, -16), S(20, -43), S(32, -58),
		S(-6, -7), S(-2, -10), S(0, -7), S(9, -4), S(6, -11), S(-6, -10), S(2, -25), S(6, -34),
	},
	// Rook
	{
		S(6, 43), S(0, 47), S(-5, 54), S(0, 49), S(15, 42), S(33, 38), S(35, 39), S(49, 34),
		S(-12, 45), S(-11, 53), S(5, 54), S(20, 44), S(11, 42), S(47, 31), S(39, 30), S(60, 21),
		S(-22, 41), S(4, 39), S(0, 40), S(3, 36), S(30, 23), S(49, 15), S(91, 9), S(53, 9),
		S(-22, 41), S(-7, 37), S(-6, 43), S(1, 34), S(3, 25), S(14, 21), S(26, 20), S(12, 19),
		S(-29, 34), S(-32, 36), S(-22, 34), S(-14, 31), S(-11, 27), S(-16, 27), S(11, 16), S(-8, 16),
		S(-31, 27), S(-26, 25), S(-22, 22), S(-19, 24), S(-6, 17), S(-1, 10), S(26, -3), S(0, 0),
		S(-29, 20), S(-22, 21), S(-12, 20), S(-9, 18), S(-2, 10), S(7, 4), S(23, -4), S(-22, 10),
		S(-9, 24), S(-8, 21), S(-4, 26), S(3, 18), S(10, 11), S(7, 16), S(7, 11), S(0, 7),
	},
	// Bishop
	{
		S(-24, -1), S(-51, 5), S(-60, 2), S(-96, 13), S(-76, 9), S(-67, -2), S(-38, 0), S(-51, -9),
		S(-21, -10), S(-5, -5), S(-18, 0), S(-24, 0), S(0, -8), S(-4, -7), S(-14, -3), S(-16, -12),
		S(-8, 8), S(-4, -1), S(0, 3), S(0, -4), S(-3, -2), S(30, 1), S(17, -5), S(13, 2),
		S(-18, 3), S(-1, 6), S(-4, 4), S(8, 16), S(8, 6), S(4, 6), S(0, 0), S(-15, 4),
		S(-4, -1), S(-16, 6), S(-6, 14), S(10, 12), S(8, 9), S(-8, 7), S(-8, 5), S(6, -10),
		S(-1, 1), S(7, 8), S(2, 10), S(4, 14), S(5, 17), S(9, 7), S(9, -1), S(16, -3),
		S(17, 9), S(7, -9), S(15, -7), S(-3, 1), S(5, 4), S(15, -4), S(31, -5), S(20, -8),
		S(5, -5), S(18, 5), S(1, -6), S(-7, 0), S(1, -3), S(-2, 7), S(10, -8), S(26, -15),
	},
	// Knight
	{
		S(-117, -40), S(-99, -13), S(-57, 2), S(-20, -12), S(10, -5), S(-49, -25), S(-65, -16), S(-69, -66),
		S(-7, -8), S(4, 1), S(39, -8), S(46, -3), S(31, -11), S(86, -32), S(13, -6), S(36, -29),
		S(0, -9), S(18, -4), S(29, 10), S(29, 10), S(68, -7), S(79, -18), S(45, -18), S(37, -24),
		S(6, 5), S(9, 5), S(22, 15), S(44, 16), S(32, 15), S(50, 10), S(22, 6), S(42, -9),
		S(0, 3), S(3, 0), S(11, 19), S(18, 18), S(17, 25), S(21, 10), S(25, 0), S(8, 2),
		S(-13, -2), S(-3, 1), S(0, 5), S(1, 20), S(18, 17), S(3, 2), S(20, -2), S(0, 2),
		S(-15, -1), S(-12, 0), S(-5, -1), S(12, 0), S(10, -1), S(13, -5), S(11, -3), S(12, 12),
		S(-55, 9), S(-5, -9), S(-24, -5), S(-10, -3), S(0, 0), S(2, -8), S(-1, -3), S(-18, 0),
	},
	// Pawn
	{
		S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0), S(0, 0),
		S(78, 172), S(89, 157), S(62, 155), S(102, 101), S(74, 103), S(61, 117), S(-19, 161), S(-29, 182),
		S(14, 73), S(0, 77), S(22, 38), S(25, 0), S(31, -6), S(66, 14), S(30, 54), S(9, 57),
		S(-5, 47), S(-7, 35), S(-3, 20), S(-1, 0), S(19, 0), S(13, 9), S(7, 24), S(5, 22),
		S(-7, 29), S(-11, 27), S(-3, 12), S(8, 4), S(11, 3), S(9, 10), S(2, 15), S(-1, 12),
		S(-14, 22), S(-14, 19), S(-11, 13), S(-5, 12), S(5, 13), S(-2, 13), S(15, 4), S(1, 4),
		S(-6, 26), S(-5, 24), S(-7, 20), S(-11, 15), S(-3, 24), S(16, 15), S(28, 6), S(-1, 5),
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
