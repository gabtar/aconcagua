package search

var tt *TranspositionTable

// https://www.chessprogramming.org/Zobrist_Hashing#Initialization

// Initialization
//
// At program initialization, we generate an array of pseudorandom numbers [11] [12]:
//
//     One number for each piece at each square
//     One number to indicate the side to move is black
//     Four numbers to indicate the castling rights, though usually 16 (2^4) are used for speed
//     Eight numbers to indicate the file of a valid En passant square, if any
//
// This leaves us with an array with 781 (12*64 + 1 + 4 + 8) random numbers. Since pawns don't happen on first and eighth rank, one might be fine with 12*64 though. There are even proposals and implementations to use overlapping keys from unaligned access up to an array of only 12 numbers for every piece and to rotate that number by square [13] [14] .
//
// Programs usually implement their own Pseudorandom number generator (PRNG), both for better quality random numbers than standard library functions, and also for reproducibility. This means that whatever platform the program is run on, it will use the exact same set of Zobrist keys. This is also useful for things like opening books, where the positions in the book can be stored by hash key and be used portably across machines, considering endianness.

// Another resource -> https://mediocrechess.sourceforge.net/guides/zobristkeys.html
// The trick is if we have a random number and XOR (explained below) another random number, we get a third seemingly random number. By XORing all random numbers from pieces, en passant and castling rights we get a unique key for the position.

// TranspositionTable stores the information related to an specific position
type TranspositionTable struct {
	table map[uint64]struct {
		depth int
		score int
	}
}

// newTranspositionTable returns a new TranspositionTable
func newTranspositionTable() *TranspositionTable {
	return &TranspositionTable{
		table: make(map[uint64]struct {
			depth int
			score int
		}),
	}
}

// save saves a searched position in the TranspositionTable with it's score and depth
func (tt *TranspositionTable) save(key uint64, depth int, score int) {
	tt.table[key] = struct {
		depth int
		score int
	}{
		depth: depth,
		score: score,
	}
}

// find returns the value of a transposition table for a given key
func (tt *TranspositionTable) find(key uint64) (int, bool) {
	if entry, ok := tt.table[key]; ok {
		return entry.score, true
	}
	return 0, false
}
