package aconcagua

import "math/rand"

// zobritsKeys contains the random keys to create the zobrist hash of a position
var zobristKeys = initZobristKeys()

// zorbistKeyIndex returns the index of the zorbistKey array for a given piece
// on a given square
func zobristKeyIndex(piece int, sq int) int {
	return (64 * piece) + sq
}

// zorbistHash calculates the hash of a given position
func zobristHash(pos *Position) (hash uint64) {
	for pieceType, bb := range pos.Bitboards {
		for bb > 0 {
			sqNumber := Bsf(bb)
			hash = hash ^ zobristKeys[zobristKeyIndex(pieceType, sqNumber)]
			bb &= ^Bitboard(0b1 << sqNumber)
		}
	}

	if pos.Turn == Black {
		hash = hash ^ zobristKeys[768]
	}

	castleKey := 769
	for idx, castl := range []castling{K, Q, k, q} {
		if pos.castlingRights.canCastle(castl) {
			hash = hash ^ zobristKeys[castleKey+idx]
		}
	}

	if pos.enPassantTarget != 0 {
		hash = hash ^ zobristKeys[772+Bsf(pos.enPassantTarget)%8]
	}
	return
}

// initZorbistKey returns an array of random numbers
//
//	One number for each piece at each square (12 x 64) pieceType (int) x square Number
//	One number to indicate the side to move is black (1)  zorbistKey[768]
//	Four numbers to indicate the castling rights, though usually 16 (2^4) are used for speed (4) zorbistKey[769 - 772]
//	Eight numbers to indicate the file of a valid En passant square, if any (8) zorbistKey[773 - 780] only account the file
func initZobristKeys() (zorbistKeys [781]uint64) {
	for i := range zorbistKeys {
		zorbistKeys[i] = uint64(rand.Int63())
	}

	return
}
