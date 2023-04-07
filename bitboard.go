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

const aFile Bitboard = 0x0101010101010101

// Get another file using bitwise operations
// aFile << (number of file to displace the 'a' file (1 - 7))
// const bFile uint64 = aFile << 1
// const hFile uint64 = aFile << 7
const hFile Bitboard = 0x8080808080808080

const rank1 Bitboard = 0x00000000000000FF

// Get other rank using bitwise operations
// const rank2 uint64 = rank1 << 8
// rank1 << 8*(number of files to displace the '1' rank (1-7))
const rank8 Bitboard = 0xFF00000000000000

const a1h8diagonal Bitboard = 0x8040201008040201
const h1a8antidiagonal Bitboard = 0x0102040810204080
const lightSquares Bitboard = 0x55AA55AA55AA55AA
const darkSquares Bitboard = 0xAA55AA55AA55AA55

// Bitboard represents a bitboard as a 64bit integer
type Bitboard uint64

type Printer interface {
	Print()
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
}

// reverseArray reverses an array of strings
func reverseArray(arr []string) []string {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}
