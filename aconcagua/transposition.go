package aconcagua

const (
	FlagExact = iota
	FlagAlpha
	FlagBeta
)

const DefaultTableSizeInMb = 64

// TTEntry represents a transposition table entry
// key is the zobrist hash of the position - 64bits
// depth is the depth of the position - 32bits
// flag is the evaluation flag - 32bits
// score is the evaluation score - 32bits
type TTEntry struct {
	key   uint64
	depth int
	flag  int
	score int
}

// TranspositionTable is a database of previously evaluated positions
type TranspositionTable struct {
	entries []TTEntry
	size    uint64
}

// NewTranspositionTable returns a pointer to a new TranspositionTable with the passed size
func NewTranspositionTable(sizeInMb int) *TranspositionTable {
	entrySizeInBytes := 20 // 64bits + 32bits + 32bits + 32bits = 160bits / 8 = 20 bytes
	size := uint64(sizeInMb * 1024 * 1024 / entrySizeInBytes)
	return &TranspositionTable{
		entries: make([]TTEntry, size),
		size:    size,
	}
}

// store stores a new entry in the transposition table
func (tt *TranspositionTable) store(key uint64, depth int, flag int, score int) {
	index := key % tt.size

	tt.entries[index] = TTEntry{
		key:   key,
		depth: depth,
		flag:  flag,
		score: score,
	}
}

// probe tries to find an entry in the transposition table
func (tt *TranspositionTable) probe(key uint64, depth int, alpha int, beta int) (int, bool) {
	index := key % tt.size
	entry := tt.entries[index]

	if entry.key == key && entry.depth >= depth {
		if entry.flag == FlagExact {
			return entry.score, true
		}
		if entry.flag == FlagAlpha && entry.score <= alpha {
			return alpha, true
		}
		if entry.flag == FlagBeta && entry.score >= beta {
			return beta, true
		}
	}
	return 0, false
}
