package aconcagua

import "testing"

func TestSetSerachTimeForMoves1To20(t *testing.T) {
	tc := TimeControl{
		timeLeftInMiliseconds: 60 * 1000, // 60 seconds
	}

	tc.setSearchTime(5)

	expected := 1800 // moves 1 - 20
	got := tc.searchTimeInMiliseconds

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestSetSerachTimeForMoves21To40(t *testing.T) {
	tc := TimeControl{
		timeLeftInMiliseconds: 60 * 1000, // 60 seconds
	}

	tc.setSearchTime(25)

	expected := 900 // moves 21 - 40
	got := tc.searchTimeInMiliseconds

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
