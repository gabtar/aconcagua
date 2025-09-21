package aconcagua

const (
	FlagExact = uint8(0)
	FlagAlpha = uint8(1)
	FlagBeta  = uint8(2)
)

const DefaultTableSizeInMb = 64

// TTEntry represents a transposition table entry
// key is the zobrist hash of the position - 64bits
// depth is the depth of the position - 32bits
// flag is the evaluation flag - 32bits
// score is the evaluation score - 32bits
type TTEntry struct {
	key   uint64
	score int32
	move  Move
	depth uint8
	flag  uint8
}

// TranspositionTable is a database of previously evaluated positions
type TranspositionTable struct {
	entries []TTEntry
	size    uint64
	stored  int // number of entries stored
	tried   int // number of entries tried
	found   int // number of entries found
	pruned  int // number of entries pruned
}

// NewTranspositionTable returns a pointer to a new TranspositionTable with the passed size
func NewTranspositionTable(sizeInMb int) *TranspositionTable {
	entrySizeInBytes := 16 // 64bits + 32bits + 16bits + 8bits + 8bits = 128 bits bits / 8 = 16 bits per entry
	size := uint64(sizeInMb * 1024 * 1024 / entrySizeInBytes)
	return &TranspositionTable{
		entries: make([]TTEntry, size),
		size:    size,
		stored:  0,
		tried:   0,
		found:   0,
		pruned:  0,
	}
}

// store stores a new entry in the transposition table
func (tt *TranspositionTable) store(key uint64, depth int, flag uint8, score int, move Move) {
	index := key % tt.size

	// replacement scheme -> Only replace entry if stored depth >= current depth
	if tt.entries[index].key == key && tt.entries[index].depth > uint8(depth) {
		return
	}

	tt.stored++
	tt.entries[index] = TTEntry{
		key:   key,
		depth: uint8(depth),
		flag:  flag,
		score: int32(score),
		move:  move,
	}
}

// probe tries to find an entry in the transposition table
func (tt *TranspositionTable) probe(key uint64, depth int, alpha int, beta int) (int, Move, bool) {
	tt.tried++
	index := key % tt.size
	entry := tt.entries[index]
	move := NoMove

	if entry.key == key && entry.depth >= uint8(depth) {
		tt.found++
		move = entry.move
		if entry.flag == FlagExact {
			tt.pruned++
			return int(entry.score), move, true
		}
		if entry.flag == FlagAlpha && entry.score <= int32(alpha) {
			tt.pruned++
			return alpha, move, true
		}
		if entry.flag == FlagBeta && entry.score >= int32(beta) {
			tt.pruned++
			return beta, move, true
		}
	}
	return 0, move, false
}

// pawnHashEntry stores the score of the previously evaluated pawn strucure
type PawnHashEntry struct {
	key     uint64
	mgScore int16
	egScore int16
	turn    int8
	_       [3]byte // Alignment
}

// PawnHashTable contains the score of the previously evaluated pawn strucure
type PawnHashTable struct {
	entries []PawnHashEntry
	size    uint64
	stores  int
	found   int
}

// NewPawnHashTable returns a pointer to a new PawnHashTable with the passed size
func NewPawnHashTable(sizeInMb int) *PawnHashTable {
	entrySizeInBytes := 16 // 64bits + 16bits + 16bits + 8bits + 3bits + 3*8bits(padding) = 128 bits / 8 = 16 bits per entry
	size := uint64(sizeInMb * 1024 * 1024 / entrySizeInBytes)
	return &PawnHashTable{
		entries: make([]PawnHashEntry, size),
		size:    size,
	}
}

// reset resets the PawnHashTable
func (pht *PawnHashTable) reset() {
	pht.entries = make([]PawnHashEntry, pht.size)
}

// store stores a new entry in the PawnHashTable
func (pht *PawnHashTable) store(key uint64, mgScore int, egScore int, turn Color) {
	pht.stores++
	index := key % pht.size
	pht.entries[index] = PawnHashEntry{
		key:     key,
		mgScore: int16(mgScore),
		egScore: int16(egScore),
		turn:    int8(turn),
	}
}

// probe tries to find an entry in the PawnHashTable
func (pht *PawnHashTable) probe(key uint64, side Color) (int, int, bool) {
	index := key % pht.size
	entry := pht.entries[index]
	if entry.key == key {
		// returns the score relative to the side passed
		opponentModifier := 1
		if side != Color(entry.turn) {
			opponentModifier = -1
		}

		return int(entry.mgScore) * opponentModifier, int(entry.egScore) * opponentModifier, true
	}
	return 0, 0, false
}
