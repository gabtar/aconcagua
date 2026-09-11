package engine

import "testing"

func TestScore(t *testing.T) {
	sc := NewScore(10, -10)
	sc2 := NewScore(20, -20)

	result := sc + sc2

	mg, eg := result.Get()

	if mg != 30 {
		t.Errorf("Expected mg: %v, got: %v", 30, mg)
	}
	if eg != -30 {
		t.Errorf("Expected eg: %v, got: %v", -30, eg)
	}
}
