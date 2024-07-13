package aconcagua

// castling represents the castling rights available in the position
// Represented in a binary of 4 bits where 0 is no castling available and 1 is available
// NNNq = 0001  or directly q
// NNkN = 0010
// NNkq = 0011
// NQNN = 0100
// NQNq = 0101
// NQkN = 0110
// NQkq = 0111
// KNNN = 1000
// KNNq = 1001
// KNkN = 1010
// KNkq = 1011
// KQNN = 1100
// KQNq = 1101
// KQkN = 1110
// KQkq = 1111
type castling int

const (
	q    = castling(0b0001)
	k    = castling(0b0010)
	kq   = castling(0b0011)
	Q    = castling(0b0100)
	Qq   = castling(0b0101)
	Qk   = castling(0b0110)
	Qkq  = castling(0b0111)
	K    = castling(0b1000)
	Kq   = castling(0b1001)
	Kk   = castling(0b1010)
	Kkq  = castling(0b1011)
	KQ   = castling(0b1100)
	KQq  = castling(0b1101)
	KQk  = castling(0b1110)
	KQkq = castling(0b1111)
)

// castleMap matchs a castle string to a castle type representation
var castleReference = map[rune]castling{
	'K': K,
	'Q': Q,
	'k': k,
	'q': q,
}

// rookOrigin matchs a castling to the corresponding rook origin square to move in the castle
var rookOrigin = map[castling]int{
	K: 7,
	Q: 0,
	k: 63,
	q: 56,
}

// rookDestination matchs a castling to the corresponding rook destination square to move in the castle
var rookDestination = map[castling]int{
	K: 5,
	Q: 3,
	k: 61,
	q: 59,
}

// castleType matchs a move string to the castle type
var castleType = map[string]castling{
	"e1g1": K,
	"e1c1": Q,
	"e8g8": k,
	"e8c8": q,
}

// castleRook matchs a castling to the rook that participates in the castle move
var castleRook = map[castling]Piece{
	K: WhiteRook,
	Q: WhiteRook,
	k: BlackRook,
	q: BlackRook,
}

// toFen returs the fen string of castlingRights
func (c *castling) toFen() (castlingRights string) {
	if *c == 0 {
		castlingRights += "-"
		return
	}

	castleChar := []string{"K", "Q", "k", "q"}
	for idx, castl := range []castling{K, Q, k, q} {
		if *c&castl > 0 {
			castlingRights += castleChar[idx]
		}
	}
	return
}

// fromFen sets the castling to match the string passed
func (c *castling) fromFen(castleFen string) {
	if castleFen == "-" {
		*c = castling(0)
		return
	}

	for _, castleType := range castleFen {
		*c |= castleReference[castleType]
	}
}

// add adds a castling right to the castling
func (c *castling) add(castle castling) {
	*c |= castle
}

// remove removes a castling right to the castling
func (c *castling) remove(castle castling) {
	*c &^= castle
}

// canCastle returns if the passed castle is allowed
func (c *castling) canCastle(to castling) bool {
	return *c&to > 0
}
