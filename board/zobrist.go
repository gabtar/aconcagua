package board

import "math/rand"

// zobritsKeys contains the random keys to create the zobrist hash of a position
var zobristKeys = initZobristKeys()

// zorbistKeyIndex returns the index of the zorbistKey array for a given piece
// on a given square
func zobristKeyIndex(piece int, sq int) (i int) {
	return (64 * piece) + sq
}

// zorbistHash calculates the hash of a given position
func zobristHash(pos Position, keys [781]uint64) (hash uint64) {
	for pieceType, bb := range pos.Bitboards {
		for bb > 0 {
			sqNumber := Bsf(bb)
			hash = hash ^ keys[zobristKeyIndex(pieceType, sqNumber)]

			bb &= ^Bitboard(0b1 << sqNumber)
		}
	}
	// Black to move
	if pos.ToMove() == Black {
		hash = hash ^ keys[768]
	}

	// Castle availability
	hash = updateZobristCastle(pos, hash, keys)

	// Ep target
	if pos.enPassantTarget != 0 {
		hash = hash ^ keys[Bsf(pos.enPassantTarget)%8]
	}
	return
}

// updateZobristCastle updates a zobrist hash depending on the castle
// availability in the passed position
func updateZobristCastle(pos Position, hash uint64, keys [781]uint64) uint64 {
	if pos.castlingRights.canCastle(SHORT_CASTLE_WHITE) {
		hash = hash ^ keys[769]
	}
	if pos.castlingRights.canCastle(LONG_CASTLE_WHITE) {
		hash = hash ^ keys[770]
	}
	if pos.castlingRights.canCastle(SHORT_CASTLE_BLACK) {
		hash = hash ^ keys[771]
	}
	if pos.castlingRights.canCastle(LONG_CASTLE_BLACK) {
		hash = hash ^ keys[772]
	}
	return hash
}

// initZorbistKey returns an array of zorbist keys of random numbers
func initZobristKeys() (zorbistKeys [781]uint64) {
	//     One number for each piece at each square (12 x 64) pieceType (int) x square Number
	//     One number to indicate the side to move is black (1)  zorbistKey[768]
	//     Four numbers to indicate the castling rights, though usually 16 (2^4) are used for speed (4) zorbistKey[769 - 772]
	//     Eight numbers to indicate the file of a valid En passant square, if any (8) zorbistKey[773 - 780] only account the file
	//
	// This leaves us with an array with 781 (12*64 + 1 + 4 + 8) random numbers. Since pawns don't happen on first and eighth rank, one might be fine with 12*64 though. There are even proposals and implementations to use overlapping keys from unaligned access up to an array of only 12 numbers for every piece and to rotate that number by square [13] [14] .

	for i := range zorbistKeys {
		zorbistKeys[i] = uint64(rand.Int63())
	}

	return
}
