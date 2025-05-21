package aconcagua

import "testing"

func TestDirections(t *testing.T) {
	directionsTestCases := []struct {
		name      string
		fromSq    int
		toSq      int
		direction uint64
	}{
		{"A1 to A8", 0, 56, NORTH},
		{"A1 to H8", 0, 63, NORTHEAST},
		{"E4 to D5", 28, 35, NORTHWEST},
		{"E4 to C5", 28, 34, INVALID},
	}

	for _, tc := range directionsTestCases {
		t.Run(tc.name, func(t *testing.T) {
			got := directions[tc.fromSq][tc.toSq]

			if got != tc.direction {
				t.Errorf("Input: from: %d, to: %d, Expected: %d, Got: %d", tc.fromSq, tc.toSq, tc.direction, got)
			}
		})
	}
}

func TestRayAttack(t *testing.T) {
	rayTestCases := []struct {
		name      string
		direction uint64
		sq        int
		ray       Bitboard
	}{
		{"North Ray", NORTH, 0, files[0] &^ Bitboard(1)},
		{"Northeast Ray", NORTHEAST, 46, bitboardFromCoordinate("h7")},
		{"East Ray", EAST, 51, bitboardFromCoordinates([]string{"e7", "f7", "g7", "h7"})},
		{"SouthEast Ray", SOUTHEAST, 19, bitboardFromCoordinates([]string{"e2", "f1"})},
		{"South Ray", SOUTH, 39, bitboardFromCoordinates([]string{"h4", "h3", "h2", "h1"})},
		{"SouthWest Ray", SOUTHWEST, 18, bitboardFromCoordinates([]string{"b2", "a1"})},
		{"West Ray", WEST, 34, bitboardFromCoordinates([]string{"b5", "a5"})},
		{"NorthWest Ray", NORTHWEST, 11, bitboardFromCoordinates([]string{"c3", "b4", "a5"})},
	}

	for _, tc := range rayTestCases {
		t.Run(tc.name, func(t *testing.T) {
			got := rayAttacks[tc.direction][tc.sq]

			if got != tc.ray {
				got.Print()
				t.Errorf("Input: dir: %d, sq: %d, Expected: %d, Got: %d", tc.direction, tc.sq, tc.ray, got)
			}
		})
	}
}
