package aconcagua

import "math"

// directions is a table that contains the compass directions between 2 squares in the board
var directions [64][64]uint64

// TODO: generate raysMask table

// init initializes the precalculated tables
func init() {
	directions = generateDirections()
}

// generateDirections generates all posible directions between all squares in the board
func generateDirections() (directions [64][64]uint64) {
	for from := 0; from < 64; from++ {
		for to := 0; to < 64; to++ {
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
				directions[from][to] = NORTH
			case fileDiff == 0 && rankDiff < 0:
				directions[from][to] = SOUTH
			case fileDiff > 0 && rankDiff == 0:
				directions[from][to] = EAST
			case fileDiff < 0 && rankDiff == 0:
				directions[from][to] = WEST
			case absFileDiff == absRankDiff && fileDiff < 0 && rankDiff < 0:
				directions[from][to] = SOUTHWEST
			case absFileDiff == absRankDiff && fileDiff > 0 && rankDiff > 0:
				directions[from][to] = NORTHEAST
			case absFileDiff == absRankDiff && fileDiff > 0 && rankDiff < 0:
				directions[from][to] = SOUTHEAST
			case absFileDiff == absRankDiff && fileDiff < 0 && rankDiff > 0:
				directions[from][to] = NORTHWEST
			default:
				directions[from][to] = INVALID
			}
		}
	}
	return
}
