package aconcagua

import (
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

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
	for i := range 8 {
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
	return bits.OnesCount64(uint64(b))
}

// NextBit removes the next bit of the bitboard and returns it
func (b *Bitboard) NextBit() (bb Bitboard) {
	bb = bitboardFromIndex(Bsf(*b))
	*b ^= bb
	return
}

// Bsf (bit scan forward) returns the bit-index of the least significant 1
// bit (LS1B) in an integer Bitboard(uint64)
func Bsf(bitboard Bitboard) int {
	return bits.TrailingZeros64(uint64(bitboard))
}

// squareNumberFromCoordinate returns the square number from a string coordinate
func squareNumberFromCoordinate(coordinate string) int {
	fileNumber := int(coordinate[0]) - 96
	rankNumber := int(coordinate[1]) - 48
	return (fileNumber - 1) + 8*(rankNumber-1)
}

// bitboardFromCoordinates is a factory that returns a bitboard from an array of string coordinates
func bitboardFromCoordinates(coordinates ...string) (bitboard Bitboard) {
	for _, c := range coordinates {
		bitboard |= 1 << squareNumberFromCoordinate(c)
	}
	return
}

// bitboardFromIndex is a factory that returns a bitboard from an index square
func bitboardFromIndex(index int) (bitboard Bitboard) {
	// NOTE: Since bitscan cannot be used with empty sets i use this guard clause to
	// ensure returning a valid bitboard for the engine
	if index > 63 || index < 0 {
		bitboard = Bitboard(0)
	} else {
		bitboard = Bitboard(0b1 << index)
	}
	return
}
