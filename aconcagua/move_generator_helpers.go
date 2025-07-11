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

// init initializes the precalculated tables
func init() {
	directions = generateDirections()
	rayAttacks = generateRayAttacks()
	knightAttacksTable = generateKnightAttacks()
}

// PositionData contains relevant data for legal move validations of a position
type PositionData struct {
	kingPosition           Bitboard
	checkRestrictedSquares Bitboard
	pinnedPieces           Bitboard
	allies                 Bitboard
	enemies                Bitboard
}

// generatePositionData returns the position data for the current position
func (pos *Position) generatePositionData() PositionData {
	checkingPieces, checkingSliders := pos.CheckingPieces(pos.Turn)
	checkRestrictedSquares := checkRestrictedSquares(pos.KingPosition(pos.Turn), checkingSliders, checkingPieces&^checkingSliders)

	return PositionData{
		kingPosition:           pos.KingPosition(pos.Turn),
		checkRestrictedSquares: checkRestrictedSquares,
		pinnedPieces:           pos.pinnedPieces(pos.Turn),
		allies:                 pos.Pieces(pos.Turn),
		enemies:                pos.Pieces(pos.Turn.Opponent()),
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

// isSliding returns a the passed Piece is an sliding piece(Queen, Rook or Bishop)
func isSliding(piece int) bool {
	return (piece >= WhiteQueen && piece <= WhiteBishop) ||
		(piece >= BlackQueen && piece <= BlackBishop)
}

// canCastleShort checks if the king can castle short on the pased position
func canCastleShort(from *Bitboard, pos *Position, side Color) bool {
	if !pos.castlingRights.canCastle(K) && side == White {
		return false
	}
	if !pos.castlingRights.canCastle(k) && side == Black {
		return false
	}

	shortCastlePath := (files[5] | files[6]) & (*from<<2 | *from<<1)
	kingSquaresAttacked := pos.AttackedSquares(side.Opponent())&(shortCastlePath|*from) > 0
	kingSquaresClear := pos.EmptySquares()&shortCastlePath == shortCastlePath

	if !kingSquaresAttacked && kingSquaresClear {
		return true
	}

	return false
}

// canCastleLong checks if the king can castle long
func canCastleLong(from *Bitboard, pos *Position, side Color) bool {
	if !pos.castlingRights.canCastle(Q) && side == White {
		return false
	}
	if !pos.castlingRights.canCastle(q) && side == Black {
		return false
	}

	longCastlePath := (files[1] | files[2] | files[3]) & (*from>>3 | *from>>2 | *from>>1)
	kingPassSquares := *from>>2 | *from>>1 | *from
	kingSquaresAttacked := pos.AttackedSquares(side.Opponent())&(kingPassSquares) > 0
	kingSquaresClear := pos.EmptySquares()&longCastlePath == longCastlePath

	if !kingSquaresAttacked && kingSquaresClear {
		return true
	}

	return false
}

// pawnMoveFlag returns the move flag for the pawn move
func pawnMoveFlag(from *Bitboard, to *Bitboard, pd *PositionData, side Color) []uint16 {
	fromSq := Bsf(*from)
	toSq := Bsf(*to)
	promotion := lastRank(side) & *to

	switch {
	case promotion > 0 && pd.enemies&*to > 0:
		return []uint16{knightCapturePromotion, bishopCapturePromotion, rookCapturePromotion, queenCapturePromotion}
	case promotion > 0:
		return []uint16{knightPromotion, bishopPromotion, rookPromotion, queenPromotion}
	case toSq-fromSq == 16 || fromSq-toSq == 16:
		return []uint16{doublePawnPush}
	case pd.enemies&*to > 0:
		return []uint16{capture}
	default:
		return []uint16{quiet}
	}
}

// potentialEpCapturers returns a bitboard with the potential pawn that can caputure enPassant
func potentialEpCapturers(pos *Position, side Color) (epCaptures Bitboard) {
	epShift := pos.enPassantTarget >> 8
	if side == Black {
		epShift = epShift << 16
	}
	notInHFile := epShift & ^(epShift & files[7])
	notInAFile := epShift & ^(epShift & files[0])

	epCaptures |= pos.getBitboards(side)[Pawn] & (notInAFile>>1 | notInHFile<<1)
	return
}

// lastRank returns the rank of the last rank for the side passed
func lastRank(side Color) (rank Bitboard) {
	if side == White {
		rank = ranks[7]
	} else {
		rank = ranks[0]
	}
	return
}
