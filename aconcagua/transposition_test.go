package aconcagua

import "testing"

func TestNewTranspositionTable(t *testing.T) {
	tt := NewTranspositionTable(64)

	expected := uint64(64 * 1024 * 1024 / 18)
	got := tt.size

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestStore(t *testing.T) {
	entryTestCases := []struct {
		name  string
		key   uint64
		depth int
		flag  uint8
		score int
		move  Move
	}{
		{name: "store entry", key: 1, depth: 2, flag: 3, score: 4, move: 5},
		{name: "store entry", key: 2, depth: 2, flag: 3, score: 4, move: 5},
		{name: "not store if depth < stored depth", key: 2, depth: 1, flag: 3, score: 4, move: 5},
	}
	tt := NewTranspositionTable(64)

	for _, testCase := range entryTestCases {
		t.Run(testCase.name, func(t *testing.T) {

			tt.store(testCase.key, testCase.depth, testCase.flag, testCase.score, testCase.move)
			if tt.entries[testCase.key%tt.size].depth != 2 {
				t.Errorf("Expected: %v, got: %v", 1, tt.stored)
			}
		})
	}
}

func TestProbe(t *testing.T) {
	entryTestCases := []struct {
		name  string
		key   uint64
		depth int
		flag  uint8
		score int
		move  Move
	}{
		{name: "alpha", key: 1, depth: 2, flag: FlagAlpha, score: 10, move: 0},
		{name: "beta", key: 2, depth: 2, flag: FlagBeta, score: 10, move: 0},
		{name: "exact", key: 3, depth: 2, flag: FlagExact, score: 10, move: 0},
	}

	for _, testCase := range entryTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			tt := NewTranspositionTable(64)

			tt.store(testCase.key, testCase.depth, testCase.flag, testCase.score, testCase.move)
			_, _, found := tt.probe(testCase.key, testCase.depth, testCase.score, testCase.score)
			if !found {
				t.Errorf("Expected: %v, got: %v", true, found)
			}
		})
	}
}
