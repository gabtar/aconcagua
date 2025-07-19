package aconcagua

import "math/rand/v2"

// zobristHashKeys contains the random keys to create the zobrist hash of a position
var zobristHashKeys = HashKeys{}

// HashKeys contains the random keys to create the zobrist hash of a position
// piecesKey - 1 number for each piece at each square 12 pieces x 64 squares = 768
// sideKey - 1 number for the side to move
// castleKey - 1 number for each castle right (16 posibles combinations)
// epKey - 1 number for each en passant square (8 files)
type HashKeys struct {
	piecesSquaresKey [769]uint64
	casltesKey       [16]uint64
	epKey            [8]uint64
	sideKey          uint64
}

func init() {
	r := rand.NewPCG(43, 28)
	for i := range zobristHashKeys.piecesSquaresKey {
		zobristHashKeys.piecesSquaresKey[i] = uint64(r.Uint64())
	}
	for i := range zobristHashKeys.casltesKey {
		zobristHashKeys.casltesKey[i] = uint64(r.Uint64())
	}
	zobristHashKeys.sideKey = uint64(r.Uint64())
	for i := range zobristHashKeys.epKey {
		zobristHashKeys.epKey[i] = uint64(r.Uint64())
	}
}

// fullZorbistHash calculates the hash of a given position
func (hk *HashKeys) fullZobristHash(pos *Position) (hash uint64) {
	for pieceType, bb := range pos.Bitboards {
		for bb > 0 {
			sqNumber := Bsf(bb)
			hash = hash ^ zobristHashKeys.piecesSquaresKey[64*pieceType+sqNumber]
			bb &= ^Bitboard(0b1 << sqNumber)
		}
	}

	if pos.Turn == Black {
		hash = hash ^ zobristHashKeys.sideKey
	}

	hash = hash ^ zobristHashKeys.casltesKey[int(pos.castling.castlingRights)]

	if pos.enPassantTarget != 0 {
		hash = hash ^ zobristHashKeys.epKey[Bsf(pos.enPassantTarget)%8]
	}
	return
}

// getPieceSquareKey returns the zobrist key for a given piece and square
func (hk *HashKeys) getPieceSquareKey(pieceType, sqNumber int) uint64 {
	return hk.piecesSquaresKey[64*pieceType+sqNumber]
}

// getCastleKey returns the zobrist key for a given castle
func (hk *HashKeys) getCastleKey(castl castlingRights) uint64 {
	return hk.casltesKey[int(castl)]
}

// getSideKey returns the zobrist key for the side to move
func (hk *HashKeys) getSideKey() uint64 {
	return hk.sideKey
}

// getEpKey returns the zobrist key for a given en passant square
func (hk *HashKeys) getEpKey(sqNumber int) uint64 {
	if sqNumber == 0 {
		return 0
	}
	return hk.epKey[sqNumber%8]
}
