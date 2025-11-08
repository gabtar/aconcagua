package engine

import "testing"

func TestCalculateSearchTime(t *testing.T) {
	tc := TimeControl{}

	tc.init(MoveTimeStrategy, 1, 1, Clock{moveTime: 10})

	if tc.calculateSearchTime(MoveTimeStrategy, 1, 1, Clock{moveTime: 10}) != 10 {
		t.Errorf("Expected: %v, got: %v", 10, tc.calculateSearchTime(MoveTimeStrategy, 1, 1, Clock{moveTime: 10}))
	}
}

func TestCalculateSearchTimeWithTimeLeftStrategy(t *testing.T) {
	tc := TimeControl{}

	tc.init(TimeLeftStrategy, 1, 1, Clock{wtime: 100, btime: 100})

	expected := 3
	got := tc.calculateSearchTime(TimeLeftStrategy, 1, 1, Clock{wtime: 100, btime: 100})

	if expected != got {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestCalculateSearchTimeWithMoveTimeStrategy(t *testing.T) {
	tc := TimeControl{}

	tc.init(MoveTimeStrategy, 1, 1, Clock{moveTime: 10})
	expected := 10
	got := tc.calculateSearchTime(MoveTimeStrategy, 1, 1, Clock{moveTime: 10})

	if expected != got {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
