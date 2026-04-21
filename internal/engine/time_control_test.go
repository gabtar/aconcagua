package engine

import "testing"

func TestCalculateSearchTime(t *testing.T) {
	tc := TimeControl{}

	tc.Initialize(MoveTimeStrategy, 1, 1, Clock{moveTime: 10})
	tc.setupLimits(MoveTimeStrategy, 1, 1, Clock{moveTime: 10})

	if tc.limits.hardLimit != 10 {
		t.Errorf("Expected: %v, got: %v", 10, tc.limits.hardLimit)
	}
}

func TestCalculateSearchTimeWithTimeLeftStrategy(t *testing.T) {
	tc := TimeControl{}

	tc.Initialize(TimeLeftStrategy, 1, 1, Clock{wtime: 100, btime: 100, movesToGo: -1})
	tc.setupLimits(TimeLeftStrategy, 1, 1, Clock{wtime: 100, btime: 100, movesToGo: -1})

	expected := int(100/float64(tc.estimatedMovesToGo(0))) + int(0)
	got := tc.limits.hardLimit

	if expected != got {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestCalculateSearchTimeWithMoveTimeStrategy(t *testing.T) {
	tc := TimeControl{}

	tc.Initialize(MoveTimeStrategy, 1, 1, Clock{moveTime: 10})
	tc.setupLimits(MoveTimeStrategy, 1, 1, Clock{moveTime: 10})

	expected := 10
	got := tc.limits.hardLimit

	if expected != got {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}
