package aconcagua

import "math"

// Constants for orthogonal directions in the board
const (
	North = iota
	NorthEast
	East
	SouthEast
	South
	SouthWest
	West
	NorthWest
	Invalid
)

// directions is a table that contains the compass directions between 2 squares in the board
var directions [64][64]uint64

// rayAttacks is a precalculated table that contains the rays on each direction for each square
var rayAttacks [8][64]Bitboard

// knightAttacksTable is a precalculated table that contains the squares that a knight can attack
var knightAttacksTable [64]Bitboard

// pieces square tables with the value of each piece + square value for middlegame
var middlegamePiecesScore [12][64]int

// pieces square tables with the value of each piece + square value for endgame
var endgamePiecesScore [12][64]int

// middlegamePieceValue is the value of each piece for middlegame
var middlegamepieceValue = [12]int{10000, 950, 500, 320, 300, 90}

// endgamePieceValue is the value of each piece for middlegame
var endgamepieceValue = [12]int{10000, 900, 510, 300, 280, 100}

// initializes various tables for usage within the engine
func init() {
	for piece := range 6 {
		whitePiece := piece
		blackPiece := piece + 6

		for sq := range 64 {

			middlegamePiecesScore[whitePiece][sq] = middlegamepieceValue[piece] + middlegamePSQT[piece][sq^56]
			endgamePiecesScore[whitePiece][sq] = endgamepieceValue[piece] + endgamePSQT[piece][sq^56]

			middlegamePiecesScore[blackPiece][sq] = middlegamepieceValue[piece] + middlegamePSQT[piece][sq]
			endgamePiecesScore[blackPiece][sq] = endgamepieceValue[piece] + endgamePSQT[piece][sq]
		}

		directions = generateDirections()
		rayAttacks = generateRayAttacks()
		attacksFrontSpans = generateAttacksFrontSpans()
		knightAttacksTable = generateKnightAttacks()
	}
}

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
			// calcuate the direction
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

// generateKnightAttacks returns a precalculated array for all posible knight moves from each square in the board
func generateKnightAttacks() (knightAttacksTable [64]Bitboard) {
	for sq := range 64 {
		from := Bitboard(1 << sq)

		notInHFile := from & ^(from & files[7])
		notInAFile := from & ^(from & files[0])
		notInABFiles := from & ^(from & (files[0] | files[1]))
		notInGHFiles := from & ^(from & (files[7] | files[6]))

		knightAttacksTable[sq] = notInAFile<<15 | notInHFile<<17 | notInGHFiles<<10 |
			notInABFiles<<6 | notInHFile>>15 | notInAFile>>17 |
			notInABFiles>>10 | notInGHFiles>>6

	}
	return
}

// raysDirection returns the rays along the direction passed that intersects the
// piece in the square passed
func raysDirection(square Bitboard, direction uint64) Bitboard {
	oppositeDirections := [8]uint64{South, SouthWest, West, NorthWest, North, NorthEast, East, SouthEast}

	return rayAttacks[direction][Bsf(square)] | square |
		rayAttacks[oppositeDirections[direction]][Bsf(square)]
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

	return rayAttacks[fromDirection][fromSq] &
		rayAttacks[toDirection][toSq]
}

// TODO: create a struct for the PSQTables/ evaluation tables
// PSQTable contains the pieces square tables
type PSQTable struct {
	middlegame [6][64]int
	endgame    [6][64]int
}

// middlegamePSQT are the pieces square tables for middlegame
var middlegamePSQT = [6][64]int{
	// King
	{
		-40, -40, -40, -40, -40, -40, -40, -40,
		-40, -40, -40, -40, -40, -40, -40, -40,
		-40, -40, -40, -40, -40, -40, -40, -40,
		-40, -40, -40, -40, -40, -40, -40, -40,
		-40, -40, -40, -40, -40, -40, -40, -40,
		-40, -40, -40, -40, -40, -40, -40, -40,
		-20, -20, -20, -20, -20, -20, -20, -20,
		0, 20, 40, -20, 0, -20, 40, 20,
	},
	// Queen
	{
		-20, -10, -10, -5, -5, -10, -10, -20,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-5, 0, 5, 5, 5, 5, 0, -5,
		0, 0, 5, 5, 5, 5, 0, -5,
		-10, 5, 5, 5, 5, 5, 0, -10,
		-10, 0, 5, 0, 0, 0, 0, -10,
		-20, -10, -10, -5, -5, -10, -10, -20,
	},
	// Rook
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		5, 10, 10, 10, 10, 10, 10, 5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		0, 0, 0, 5, 5, 0, 0, 0,
	},
	// Bishop
	{
		-10, -10, -10, -10, -10, -10, -10, -10,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-10, 0, 5, 10, 10, 5, 0, -10,
		-10, 0, 5, 10, 10, 5, 0, -10,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, -10, -20, -10, -10, -20, -10, -10,
	},
	// Knight
	{
		-10, -10, -10, -10, -10, -10, -10, -10,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-10, 0, 5, 10, 10, 5, 0, -10,
		-10, 0, 5, 10, 10, 5, 0, -10,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, -30, -10, -10, -10, -10, -30, -10,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		5, 10, 15, 20, 20, 15, 10, 5,
		4, 8, 12, 16, 16, 12, 8, 4,
		3, 6, 9, 12, 12, 9, 6, 3,
		2, 4, 6, 8, 8, 6, 4, 2,
		1, 2, 3, -10, -10, 3, 2, 1,
		0, 0, 0, -40, -40, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}

// endgamePSQT are the pieces square tables for endgame
var endgamePSQT = [6][64]int{
	// King
	{
		0, 10, 20, 30, 30, 20, 10, 0,
		10, 20, 30, 40, 40, 30, 20, 10,
		20, 30, 40, 50, 50, 40, 30, 20,
		30, 40, 50, 60, 60, 50, 40, 30,
		30, 40, 50, 60, 60, 50, 40, 30,
		20, 30, 40, 50, 50, 40, 30, 20,
		10, 20, 30, 40, 40, 30, 20, 10,
		0, 10, 20, 30, 30, 20, 10, 0,
	},
	// Queen
	{
		-20, -10, -10, -5, -5, -10, -10, -20,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-5, 0, 5, 5, 5, 5, 0, -5,
		0, 0, 5, 5, 5, 5, 0, -5,
		-10, 5, 5, 5, 5, 5, 0, -10,
		-10, 0, 5, 0, 0, 0, 0, -10,
		-20, -10, -10, -5, -5, -10, -10, -20,
	},
	// Rook
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		5, 10, 10, 10, 10, 10, 10, 5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		-5, 0, 0, 0, 0, 0, 0, -5,
		0, 0, 0, 5, 5, 0, 0, 0,
	},
	// Bishop
	{
		-10, -10, -10, -10, -10, -10, -10, -10,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-10, 0, 5, 10, 10, 5, 0, -10,
		-10, 0, 5, 10, 10, 5, 0, -10,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, -10, -20, -10, -10, -20, -10, -10,
	},
	// Knight
	{
		-10, -10, -10, -10, -10, -10, -10, -10,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-10, 0, 5, 10, 10, 5, 0, -10,
		-10, 0, 5, 10, 10, 5, 0, -10,
		-10, 0, 5, 5, 5, 5, 0, -10,
		-10, 0, 0, 0, 0, 0, 0, -10,
		-10, -30, -10, -10, -10, -10, -30, -10,
	},
	// Pawn
	{
		0, 0, 0, 0, 0, 0, 0, 0,
		50, 55, 60, 60, 60, 60, 55, 50,
		20, 25, 30, 35, 35, 30, 25, 20,
		5, 5, 10, 15, 15, 10, 5, 5,
		2, 4, 6, 8, 8, 6, 4, 2,
		1, 2, 3, 4, 4, 3, 2, 1,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
	},
}
