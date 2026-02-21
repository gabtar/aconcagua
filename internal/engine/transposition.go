package engine

import (
	"runtime"
	"runtime/debug"
	"strconv"
)

const (
	FlagExact = uint8(0)
	FlagAlpha = uint8(1)
	FlagBeta  = uint8(2)
)

const (
	DefaultTableSizeInMb         = 64
	DefaultPawnHashTableSizeInMb = 4
	BucketSize                   = 3
)

// TTEntry represents a transposition table entry
// For key it uses only the upper 32 bits of the zobrist hash.
// Collisions are reduced by using multiple entries per bucket, and when
// practically minimal if the size of the table is large enough
type TTEntry struct {
	key32 uint32 // Upper 32 bits of hash for verification
	move  Move
	score int16
	eval  int16 // Static evaluation
	depth uint8
	flag  uint8
	age   uint8   // Age of the entry
	_     [3]byte // Alignment
}

// TTBucket holds multiple entries to reduce collisions
type TTBucket struct {
	entries [BucketSize]TTEntry
}

// TranspositionTable is a database of previously evaluated positions
type TranspositionTable struct {
	buckets []TTBucket
	size    uint64
	age     uint8 // Last age/generation of the table

	// Stats
	stored int
	tried  int
	hits   int
	pruned int
}

// NewTranspositionTable returns a pointer to a new TranspositionTable with the passed size
func NewTranspositionTable(sizeInMb int) *TranspositionTable {
	bucketSizeInBytes := 16 * BucketSize // 16 bytes(TTEntry) per entry * 3 entries
	numBuckets := uint64(sizeInMb * 1024 * 1024 / bucketSizeInBytes)

	return &TranspositionTable{
		buckets: make([]TTBucket, numBuckets),
		size:    numBuckets,
		age:     0,
		stored:  0,
		tried:   0,
		hits:    0,
		pruned:  0,
	}
}

func (tt *TranspositionTable) Resize(sizeInMb int) {
	bucketSizeInBytes := 16 * BucketSize
	numBuckets := uint64(sizeInMb * 1024 * 1024 / bucketSizeInBytes)

	// Force Clear
	tt.buckets = nil
	runtime.GC()
	debug.FreeOSMemory()

	tt.buckets = make([]TTBucket, numBuckets)
	tt.size = numBuckets
}

// Clear clears the transposition table entries
func (tt *TranspositionTable) Clear() {
	for i := range tt.buckets {
		for j := range tt.buckets[i].entries {
			tt.buckets[i].entries[j].key32 = 0
			tt.buckets[i].entries[j].move = NoMove
			tt.buckets[i].entries[j].score = 0
			tt.buckets[i].entries[j].eval = 0
			tt.buckets[i].entries[j].depth = 0
			tt.buckets[i].entries[j].flag = 0
			tt.buckets[i].entries[j].age = 0
		}
	}
	tt.stored = 0
	tt.tried = 0
	tt.hits = 0
	tt.pruned = 0
	tt.age = 0
}

// newSearch resets the stats for a new search
func (tt *TranspositionTable) newSearch() {
	tt.age++
	tt.tried = 0
	tt.hits = 0
	tt.pruned = 0
	tt.stored = 0
}

// store stores a new entry in the transposition table
func (tt *TranspositionTable) store(key uint64, depth int, ply int, flag uint8, score int, eval int, move Move) {
	index := key % tt.size
	key32 := uint32(key >> 32) // Use only upper 32 bits for hash verification
	bucket := &tt.buckets[index]

	// Look for existing entry with same key or find best slot to replace
	replaceIdx, replaceDepth := 0, 0

	for i := range BucketSize {
		entry := &bucket.entries[i]

		// Always replace entry for exact key match
		if entry.key32 == key32 {
			replaceIdx = i
			break
		}

		// Empty slot
		if entry.depth == 0 {
			replaceIdx = i
			tt.stored++
			break
		}

		// Replacement strategy
		// New entry replaces the entry with lowest depth
		if replaceDepth == 0 || int(entry.depth) < replaceDepth {
			replaceIdx = i
			replaceDepth = int(entry.depth)
		}
	}

	bucket.entries[replaceIdx] = TTEntry{
		key32: key32,
		move:  move,
		score: int16(adjustMateScoreForTT(score, ply)),
		eval:  int16(eval),
		depth: uint8(depth),
		flag:  flag,
		age:   tt.age,
	}
}

// probe tries to find an entry in the transposition table
func (tt *TranspositionTable) probe(key uint64, depth int, ply int, alpha int, beta int) (int, int, Move, bool) {
	tt.tried++
	index := key % tt.size
	key32 := uint32(key >> 32)
	bucket := &tt.buckets[index]

	for i := range BucketSize {
		entry := &bucket.entries[i]

		if entry.key32 == key32 {
			tt.hits++
			move := entry.move
			eval := int(entry.eval)

			if entry.depth >= uint8(depth) {
				score := adjustMateScoreFromTT(int(entry.score), ply)

				// Update age to mark as recently used
				entry.age = tt.age

				if entry.flag == FlagExact {
					tt.pruned++
					return score, eval, move, true
				}
				if entry.flag == FlagAlpha && score <= alpha {
					tt.pruned++
					return alpha, eval, move, true
				}
				if entry.flag == FlagBeta && score >= beta {
					tt.pruned++
					return beta, eval, move, true
				}
			}
			// Found entry but can't use score, return move and eval
			return 0, eval, move, false
		}
	}
	return 0, 0, NoMove, false
}

// hashfull returns the approximate percentage of the transposition table that is used
func (tt *TranspositionTable) hashfull() int {
	// Take a sample of only 1000 buckets to improve performance
	sampleSize := min(int(tt.size), 1000)

	used := 0
	for i := range sampleSize {
		bucket := &tt.buckets[i]
		for j := range BucketSize {
			// Check if entry is occupied (depth > 0)
			if tt.age == bucket.entries[j].age && bucket.entries[j].depth > 0 {
				used++
				break // Only count one entry per bucket
			}
		}
	}

	return 1000 * used / sampleSize
}

// adjustMateScoreForTT converts mate scores to be ply-independent for storage
func adjustMateScoreForTT(score int, ply int) int {
	if score >= MateScore-MaxSearchDepth {
		return score + ply
	}
	if score <= -MateScore+MaxSearchDepth {
		return score - ply
	}
	return score
}

// adjustMateScoreFromTT converts mate scores from TT back to ply-dependent
func adjustMateScoreFromTT(score int, ply int) int {
	if score >= MateScore-MaxSearchDepth {
		return score - ply
	}
	if score <= -MateScore+MaxSearchDepth {
		return score + ply
	}
	return score
}

// Stats returns an string with useful Stats about the transposition table
func (tt *TranspositionTable) Stats() string {
	return "Hashfull: " + strconv.Itoa(tt.hashfull()) + " Age: " + strconv.Itoa(int(tt.age)) + "\n" +
		"Stored: " + strconv.Itoa(tt.stored) +
		" Tried: " + strconv.Itoa(tt.tried) + " Hits: " + strconv.Itoa(tt.hits) +
		" Pruned: " + strconv.Itoa(tt.pruned)
}

// PawnHashEntry stores the score of the previously evaluated pawn strucure
type PawnHashEntry struct {
	key     uint64
	mgScore int16
	egScore int16
	age     uint8
	_       [3]byte // Alignment
}

// PawnHashBucketSize is the number of entries per bucket for the pawn hash table
const PawnHashBucketSize = 2

// PawnHashBucket holds multiple entries to reduce collisions
type PawnHashBucket struct {
	entries [PawnHashBucketSize]PawnHashEntry
}

// PawnHashTable contains the score of the previously evaluated pawn strucure
type PawnHashTable struct {
	buckets []PawnHashBucket
	size    uint64
	age     uint8
	stores  int
	probes  int
	found   int
}

// NewPawnHashTable returns a pointer to a new PawnHashTable with the passed size
func NewPawnHashTable(sizeInMb int) *PawnHashTable {
	bucketSize := 32 // 16 bytes per entry * 2
	numBuckets := uint64(sizeInMb * 1024 * 1024 / bucketSize)
	return &PawnHashTable{
		buckets: make([]PawnHashBucket, numBuckets),
		size:    numBuckets,
		age:     0,
	}
}

// clear resets the PawnHashTable
func (pht *PawnHashTable) clear() {
	for i := range pht.buckets {
		for j := range pht.buckets[i].entries {
			pht.buckets[i].entries[j].key = 0
			pht.buckets[i].entries[j].mgScore = 0
			pht.buckets[i].entries[j].egScore = 0
			pht.buckets[i].entries[j].age = 0
		}
	}
	pht.stores = 0
	pht.probes = 0
	pht.found = 0
}

// store stores a new entry in the PawnHashTable
func (pht *PawnHashTable) store(key uint64, mgScore int, egScore int) {
	pht.stores++
	index := key % pht.size
	bucket := &pht.buckets[index]

	// First check for existing entry or empty slot
	replaceIdx := 0

	for i := range PawnHashBucketSize {
		entry := &bucket.entries[i]

		// Always replace exact match
		if entry.key == key {
			replaceIdx = i
			break
		}

		// Found empty slot
		if entry.key == 0 {
			replaceIdx = i
			break
		}

		// Replace older entries
		if entry.age < bucket.entries[replaceIdx].age {
			replaceIdx = i
		}
	}

	bucket.entries[replaceIdx] = PawnHashEntry{
		key:     key,
		mgScore: int16(mgScore),
		egScore: int16(egScore),
		age:     uint8(pht.age),
	}
}

// probe tries to find an entry in the PawnHashTable
func (pht *PawnHashTable) probe(key uint64, side Color) (int, int, bool) {
	pht.probes++
	index := key % pht.size
	bucket := &pht.buckets[index]

	for i := range PawnHashBucketSize {
		entry := &bucket.entries[i]
		if entry.key == key {
			pht.found++
			entry.age = uint8(pht.age)

			if side == Black {
				return -int(entry.mgScore), -int(entry.egScore), true
			}
			return int(entry.mgScore), int(entry.egScore), true
		}
	}
	return 0, 0, false
}

func (pht *PawnHashTable) newSearch() {
	pht.age++
	pht.probes = 0
	pht.found = 0
	pht.stores = 0
}

// Stats returns an string with useful Stats about the PawnHashTable
func (pht *PawnHashTable) Stats() string {
	return "Hashfull: " + strconv.Itoa(pht.hashfull()) + " Age: " + strconv.Itoa(int(pht.age)) + "\n" +
		"Stored: " + strconv.Itoa(pht.stores) +
		" Tried: " + strconv.Itoa(pht.probes) + " Hits: " + strconv.Itoa(pht.found) + "\n" +
		"Hitrate: " + strconv.FormatFloat(pht.hitRate(), 'f', 2, 64)
}

// hashfull returns the approximate percentage of the PawnHashTable that is used
func (pht *PawnHashTable) hashfull() int {
	sampleSize := min(int(pht.size), 1000)

	used := 0
	for i := range sampleSize {
		bucket := &pht.buckets[i]
		for j := range PawnHashBucketSize {
			if pht.age == bucket.entries[j].age {
				used++
				break // Only count one entry per bucket
			}
		}
	}

	return 1000 * used / sampleSize
}

// hitRate returns the hit rate of the PawnHashTable
func (pht *PawnHashTable) hitRate() float64 {
	if pht.probes == 0 {
		return 0
	}
	return float64(pht.found) / float64(pht.probes)
}
