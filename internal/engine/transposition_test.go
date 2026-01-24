package engine

import (
	"math"
	"testing"
	"unsafe"
)

func TestNewTranspositionTable(t *testing.T) {
	ttSizeInMb := 64
	tt := NewTranspositionTable(ttSizeInMb)

	expected := uint64(ttSizeInMb * 1024 * 1024 / (16 * BucketSize))
	got := tt.size

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestBucketsSizeInMemory(t *testing.T) {
	ttSizeInMb := 64
	tt := NewTranspositionTable(ttSizeInMb)

	bucketsSizeInMb := float64(unsafe.Sizeof(tt.buckets[0])) * float64(cap(tt.buckets)) / (1024.0 * 1024.0)

	got := int(math.Ceil(bucketsSizeInMb))

	if got != ttSizeInMb {
		t.Errorf("Expected: %v, got: %v", ttSizeInMb, got)
	}
}

func TestStoreEntryWithExactMatch(t *testing.T) {
	tt := NewTranspositionTable(1)

	key := uint64(1)
	depth := 2
	ply := 3
	flag := FlagExact
	score := 10
	eval := 0
	move := NoMove

	tt.store(key, depth, ply, flag, score, eval, move)

	index := key % tt.size
	bucket := &tt.buckets[index]

	expected := uint32(key >> 32)
	got := bucket.entries[0].key32

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestReplaceEntryWithExactMatch(t *testing.T) {
	tt := NewTranspositionTable(1)

	key := uint64(0)
	depth := 2
	ply := 3
	flag := FlagExact
	score := 10
	eval := 0
	move := NoMove

	tt.store(key, depth, ply, flag, score, eval, move)
	tt.store(key, depth+3, ply, flag, score, eval, move)

	index := key % tt.size
	bucket := &tt.buckets[index]

	expected := depth + 3
	got := int(bucket.entries[0].depth)

	if got != expected {
		t.Errorf("Expected: %v, got: %v", expected, got)
	}
}

func TestProbeEntryWithExactMatch(t *testing.T) {
	tt := NewTranspositionTable(1)

	key := uint64(0)
	depth := 2
	ply := 3
	flag := FlagExact
	score := 10
	eval := 0
	move := NoMove

	tt.store(key, depth, ply, flag, score, eval, move)
	ttScore, ttEval, ttMove, ttHit := tt.probe(key, depth, ply, MinInt, MaxInt)

	expectedScore := score
	expectedEval := eval
	expectedMove := move
	expectedHit := true

	if ttScore != expectedScore {
		t.Errorf("Expected: %v, got: %v", expectedScore, ttScore)
	}
	if ttEval != expectedEval {
		t.Errorf("Expected: %v, got: %v", expectedEval, ttEval)
	}
	if ttMove != expectedMove {
		t.Errorf("Expected: %v, got: %v", expectedMove, ttMove)
	}
	if ttHit != expectedHit {
		t.Errorf("Expected: %v, got: %v", expectedHit, ttHit)
	}
}

func TestProbeEntryInsufficientDepth(t *testing.T) {
	tt := NewTranspositionTable(1)

	key := uint64(10240)
	depth := 2
	ply := 3
	flag := FlagExact
	score := 10
	eval := 0
	move := Move(55)

	tt.store(key, depth, ply, flag, score, eval, move)
	ttScore, ttEval, ttMove, ttHit := tt.probe(key, depth+1, ply, MinInt, MaxInt)

	expectedScore := 0
	expectedEval := eval
	expectedMove := move
	expectedHit := false

	if ttScore != expectedScore {
		t.Errorf("Expected: %v, got: %v", expectedScore, ttScore)
	}
	if ttEval != expectedEval {
		t.Errorf("Expected: %v, got: %v", expectedEval, ttEval)
	}
	if ttMove != expectedMove {
		t.Errorf("Expected: %v, got: %v", expectedMove, ttMove)
	}
	if ttHit != expectedHit {
		t.Errorf("Expected: %v, got: %v", expectedHit, ttHit)
	}
}
