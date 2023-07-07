package board

import (
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

// files contains bitboards representing a file in chess board. Eg -> files[0] -> a File, files[1] -> b file, etc
var files [8]Bitboard = [8]Bitboard{
	0x0101010101010101,
	0x0101010101010101 << 1,
	0x0101010101010101 << 2,
	0x0101010101010101 << 3,
	0x0101010101010101 << 4,
	0x0101010101010101 << 5,
	0x0101010101010101 << 6,
	0x0101010101010101 << 7,
}

// ranks contains bitboards representing a rank in chess board. Eg -> rank[0] -> rank 1, rank[1] -> rank 2, etc
var ranks [8]Bitboard = [8]Bitboard{
	0x00000000000000FF,
	0x00000000000000FF << 8,
	0x00000000000000FF << 16,
	0x00000000000000FF << 24,
	0x00000000000000FF << 32,
	0x00000000000000FF << 40,
	0x00000000000000FF << 48,
	0x00000000000000FF << 56,
}

const ALL_SQUARES Bitboard = 0xFFFFFFFFFFFFFFFF

// Bitboard represents a bitboard as a 64bit integer
type Bitboard uint64

// Print draws a Bitboard in the terminal in a 'prettier' way
func (b Bitboard) Print() {
	binary := strconv.FormatUint(uint64(b), 2)
	fill := ""
	if len(binary) < 64 {
		fill = strings.Repeat("0", 64-len(binary))
	}
	binary = fill + binary
	for i := 0; i < 8; i++ {
		fmt.Println(reverseArray(strings.Split(binary[i*8:i*8+8], "")))
	}
	fmt.Println()
}

// reverseArray reverses an array of strings
func reverseArray(arr []string) []string {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

// count returns the number of non zero bits in a bitboard
func (b Bitboard) count() int {
	count := 0
	for b > 0 {
		b &= ^(0b1 << Bsf(b))
		count++
	}

	return count
}

// nextOne returns a bitboard with the first bit
func (b *Bitboard) nextOne() (bb Bitboard) {
	bb = BitboardFromIndex(Bsf(*b))
	*b ^= bb
	return
}

// ToStringSlice returns an slice of (string)coordinates with the squares occupied
// in the bitboard
func (b Bitboard) ToStringSlice() (squares []string) {
	for b > 0 {
		squares = append(squares, squareReference[Bsf(b)])
		b -= (1 << Bsf(b))
	}
	return
}

// bsf (bit scan forward) returns the bit-index of the least significant 1
// bit (LS1B) in an integer Bitboard(uint64)
func Bsf(bitboard Bitboard) int {
	return bits.TrailingZeros64(uint64(bitboard))
}

// bsr (bit scan reverse) returns the bit-index of the most significant 1
// bit (MS1B) in an integer Bitboard(uint64)
func Bsr(bitboard Bitboard) int {
	return bits.LeadingZeros64(uint64(bitboard))
}

// bitboardFromIndex is a factory that returns a bitboard from an index square
func BitboardFromIndex(index int) (bitboard Bitboard) {
	// NOTE: Since bitscan cannot be used with empty sets i use this guard clause to
	// ensure returning a valid bitboard for the engine
	if index > 63 || index < 0 {
		bitboard = Bitboard(0)
	} else {
		bitboard = Bitboard(0b1 << index)
	}
	return
}
