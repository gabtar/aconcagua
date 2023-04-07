package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Useful constans for mapping squares to uint64
// https://www.chessprogramming.org/Square_Mapping_Considerations#Some_hexadecimal_Constants
// Little-Endian Rank-File Mapping
//
//  Chess board sqaures mapping
//
//  8 | 56 57 58 59 60 61 62 63
//  7 | 48 49 50 51 52 53 54 55
//  6 | 40 41 42 43 44 45 46 47
//  5 | 32 33 34 35 36 37 38 39
//  4 | 24 25 26 27 28 29 30 31
//  3 | 16 17 18 19 20 21 22 23
//  2 |  8  9 10 11 12 13 14 15
//  1 |  0  1  2  3  4  5  6  7
//    --------------------------
//    a  b  c  d  e  f  g  h

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

// TODO, I need to map diagonals in the same way as files and ranks

const a1h8diagonal Bitboard = 0x8040201008040201
const h1a8antidiagonal Bitboard = 0x0102040810204080
const lightSquares Bitboard = 0x55AA55AA55AA55AA
const darkSquares Bitboard = 0xAA55AA55AA55AA55

// Bitboard represents a bitboard as a 64bit integer
type Bitboard uint64

type Printer interface {
	Print()
}

type BitboardConverter interface {
  ToStringSlice() (squares []string)
}

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


// ToStringSlice returns an slice of (string)coordinates with the squares occupied
// in the bitboard
func (b Bitboard) ToStringSlice() (squares []string){
	binaryString := strconv.FormatUint(uint64(b), 2)

  for i := len(binaryString)-1; i >= 0; i-- {
    if binaryString[i] == '1' {
      squares = append(squares, squareMap[len(binaryString) - i - 1])
    }
  }
  return
}
