package aconcagua

import "testing"

func TestDirections(t *testing.T) {
	directionsTestCases := []struct {
		name      string
		fromSq    int
		toSq      int
		direction uint64
	}{
		{"A1 to A8", 0, 56, North},
		{"A1 to H8", 0, 63, NorthEast},
		{"E4 to D5", 28, 35, NorthWest},
		{"E4 to C5", 28, 34, Invalid},
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
		{"North Ray", North, 0, files[0] &^ Bitboard(1)},
		{"Northeast Ray", NorthEast, 46, bitboardFromCoordinates("h7")},
		{"East Ray", East, 51, bitboardFromCoordinates("e7", "f7", "g7", "h7")},
		{"SouthEast Ray", SouthEast, 19, bitboardFromCoordinates("e2", "f1")},
		{"South Ray", South, 39, bitboardFromCoordinates("h4", "h3", "h2", "h1")},
		{"SouthWest Ray", SouthWest, 18, bitboardFromCoordinates("b2", "a1")},
		{"West Ray", West, 34, bitboardFromCoordinates("b5", "a5")},
		{"NorthWest Ray", NorthWest, 11, bitboardFromCoordinates("c3", "b4", "a5")},
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
