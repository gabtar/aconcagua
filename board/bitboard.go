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

// coordinateToSquareNumber maps a coordinate string to a square number
var coordinateToSquareNumber = map[string]int{
	"a1": 0, "b1": 1, "c1": 2, "d1": 3, "e1": 4, "f1": 5, "g1": 6, "h1": 7,
	"a2": 8, "b2": 9, "c2": 10, "d2": 11, "e2": 12, "f2": 13, "g2": 14, "h2": 15,
	"a3": 16, "b3": 17, "c3": 18, "d3": 19, "e3": 20, "f3": 21, "g3": 22, "h3": 23,
	"a4": 24, "b4": 25, "c4": 26, "d4": 27, "e4": 28, "f4": 29, "g4": 30, "h4": 31,
	"a5": 32, "b5": 33, "c5": 34, "d5": 35, "e5": 36, "f5": 37, "g5": 38, "h5": 39,
	"a6": 40, "b6": 41, "c6": 42, "d6": 43, "e6": 44, "f6": 45, "g6": 46, "h6": 47,
	"a7": 48, "b7": 49, "c7": 50, "d7": 51, "e7": 52, "f7": 53, "g7": 54, "h7": 55,
	"a8": 56, "b8": 57, "c8": 58, "d8": 59, "e8": 60, "f8": 61, "g8": 62, "h8": 63,
}

const AllSquares Bitboard = 0xFFFFFFFFFFFFFFFF

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

// NextBit removes the next bit of the bitboard and returns it
func (b *Bitboard) NextBit() (bb Bitboard) {
	bb = bitboardFromIndex(Bsf(*b))
	*b ^= bb
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
