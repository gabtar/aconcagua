package aconcagua

// castlingRights represents the castling rights available in the position
// Represented in a binary of 4 bits where 0 is no castlingRights available and 1 is all castlingRights
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
type castlingRights int8

const (
	noCastling = castlingRights(0b0000)
	q          = castlingRights(0b0001)
	k          = castlingRights(0b0010)
	kq         = castlingRights(0b0011)
	Q          = castlingRights(0b0100)
	Qq         = castlingRights(0b0101)
	Qk         = castlingRights(0b0110)
	Qkq        = castlingRights(0b0111)
	K          = castlingRights(0b1000)
	Kq         = castlingRights(0b1001)
	Kk         = castlingRights(0b1010)
	Kkq        = castlingRights(0b1011)
	KQ         = castlingRights(0b1100)
	KQq        = castlingRights(0b1101)
	KQk        = castlingRights(0b1110)
	KQkq       = castlingRights(0b1111)
)

// TODO: does not work with chess960!!!!
// castleType matchs a move string to the castle type
var castleType = map[string]castlingRights{
	"e1g1": K,
	"e1c1": Q,
	"e8g8": k,
	"e8c8": q,
}

// TODO: add 960 support for castling
type castling struct {
	castlingRights castlingRights
	// TODO: add destination squares for rook/king?
	kingsStartSquare [2]int    // King starting square for castling(0-63)
	rooksStartSquare [2][2]int // Rook starting square for castling rookStartSquare[0] = white rooks, rookStartSquare[1] = black rooks
	kingsEndSquare   [2][2]int
	rooksEndSquare   [2][2]int
	chess960         bool
}

// castleRook matchs a castling to the rook that participates in the castle move
var castleRook = map[castlingRights]int{
	K: WhiteRook,
	Q: WhiteRook,
	k: BlackRook,
	q: BlackRook,
}

// NewCaslting returns a new castling struct
func NewCastling(whiteKingStart int, whiteKingsideRook int, whiteQueensideRook int) *castling {
	blackKingStart := whiteKingStart ^ 56 // flips the board to get the black king
	blackKingsideRook := whiteKingsideRook ^ 56
	blackQueensideRook := whiteQueensideRook ^ 56

	return &castling{
		castlingRights:   noCastling,
		kingsStartSquare: [2]int{whiteKingStart, blackKingStart},
		rooksStartSquare: [2][2]int{ // WhiteRook(shortcaslte, longcastle), BlackRook(shortcaslte, longcastle)
			{whiteKingsideRook, whiteQueensideRook},
			{blackKingsideRook, blackQueensideRook},
		},
		kingsEndSquare: [2][2]int{{6, 2}, {62, 58}}, // Fixed for both stardard and 960
		rooksEndSquare: [2][2]int{{5, 3}, {61, 59}},

		chess960: false, // set the flag manually if we are on 960
	}
}

// NewCastlingFromShredderFenCastlingCode returns a new castling struct from the fen castling encode
// NOTE: this only works to set up intial squares for castling in 960 when the castling codes
// are given with the with upper case and lower case file characters of the affected rooks for
// white and black castling rights (Shredder-FEN style). Must be position from move 0 for a given
// valid chess960 position, otherwise it will not work
func NewCastlingFromShredderFenCastlingCode(whiteKingStart int, castlEncode string) *castling {
	kingFile := whiteKingStart % 8
	whiteKingsideRook, whiteQueensideRook := -1, -1
	castlingRights := castlingRights(0)
	for _, char := range castlEncode {
		switch {
		case char >= 'A' && char <= 'H':
			file := int(char - 'A')
			if file > kingFile {
				castlingRights.add(K)
				whiteKingsideRook = file
			} else {
				castlingRights.add(Q)
				whiteQueensideRook = file
			}
		case char >= 'a' && char <= 'h':
			file := int(char - 'a')
			if file > kingFile {
				castlingRights.add(k)
				whiteKingsideRook = file
			} else {
				castlingRights.add(q)
				whiteQueensideRook = file
			}
		}
	}

	castling := NewCastling(whiteKingStart, whiteKingsideRook, whiteQueensideRook)
	castling.castlingRights = castlingRights
	castling.chess960 = true
	return castling
}

// NewCastlingFromBackrank returns a new castling struct for chess 960 by parsing the backrank of the position
// NOTE: only works for setting up the chess 960 initial position. It may be wrong for other positions,
// different from move number 0
func NewCastlingFromBackrank(backrank string) *castling {
	kingSq, queenSideRookSq, kingSideRookSq := -1, -1, -1
	for file, char := range backrank {
		if char == 'k' {
			kingSq = file
		}
		if char == 'r' {
			if queenSideRookSq == -1 {
				queenSideRookSq = file
			}
			kingSideRookSq = file
		}
	}

	castling := NewCastling(kingSq, kingSideRookSq, queenSideRookSq)
	castling.castlingRights = KQkq
	castling.chess960 = true
	return castling
}

// toFen returs the fen string of castlingRights
func (c *castlingRights) toFen() (castles string) {
	if *c == 0 {
		castles += "-"
		return
	}

	castleChar := []string{"K", "Q", "k", "q"}
	for idx, castl := range []castlingRights{K, Q, k, q} {
		if *c&castl > 0 {
			castles += castleChar[idx]
		}
	}
	return
}

// fromFen sets the castling to match the string passed
func (c *castlingRights) fromFen(castleFen string) {
	if castleFen == "-" {
		*c = castlingRights(0)
		return
	}

	castleReference := map[rune]castlingRights{
		'K': K, 'Q': Q, 'k': k, 'q': q,
	}

	for _, castleType := range castleFen {
		*c |= castleReference[castleType]
	}
}

// add adds a castling right to the castling
func (c *castlingRights) add(castle castlingRights) {
	*c |= castle
}

// remove removes a castling right to the castling
func (c *castlingRights) remove(castle castlingRights) {
	*c &^= castle
}

// canCastle returns if the passed castle is allowed
func (c *castlingRights) canCastle(to castlingRights) bool {
	return *c&to > 0
}

// updateCastleRights updates the castling rights when making a move
func (c *castling) updateCastleRights(from int, to int) {
	// got the idea from Tom Kerrigan's TSCP engine
	// based on the from/to squares of the move we can update the castle rights after making a move as following
	// with the from square, we are either moving a rook (disables the castling right asociated to that rook)
	// or the king, in that case disables both castling rights for that color(black/white)
	// with the to square, means we are attaking a rook/capturing move, so we wil have to disable the castling
	// right asociated with that rook after that move
	if c.castlingRights == 0 {
		return
	}

	casltesToDisable := castlingRights(0)
	sqModifierOrder := []int{
		c.rooksStartSquare[White][1],
		c.rooksStartSquare[White][0],
		c.kingsStartSquare[White],
		c.rooksStartSquare[Black][1],
		c.kingsStartSquare[Black],
		c.rooksStartSquare[Black][0],
	}
	fromModifier := []castlingRights{Q, K, KQ, q, kq, k}
	toModifier := []castlingRights{Q, K, 0, q, 0, k}

	for i, sq := range sqModifierOrder {
		if from == sq {
			casltesToDisable |= fromModifier[i]
		}
		if to == sq {
			casltesToDisable |= toModifier[i]
		}
	}

	c.castlingRights.remove(casltesToDisable)
}

// TODO: refactor to use a type/flag of caslte to remove duplication. Eg 0 for shortCaslte and 1 for longCastle...

// canCastleShort checks if the king can castle short on the pased position
func (pos *Position) canCastleShort(side Color) bool {
	if !pos.castling.castlingRights.canCastle(K) && side == White {
		return false
	}
	if !pos.castling.castlingRights.canCastle(k) && side == Black {
		return false
	}

	rookBB := bitboardFromIndex(pos.castling.rooksStartSquare[side][0])
	kingBB := bitboardFromIndex(pos.castling.kingsStartSquare[side])

	kingToSq := bitboardFromIndex(pos.castling.kingsEndSquare[side][0])
	kingFromToPath := getRayPath(&kingBB, &kingToSq) | kingToSq

	rookEndBB := bitboardFromIndex(pos.castling.rooksEndSquare[side][0])
	rookFromToPath := getRayPath(&rookBB, &rookEndBB) | rookEndBB

	emptySquares := pos.EmptySquares() | kingBB | rookBB // NOTE: Avoid king and rook, so are not taken into account when calculating whenever the path is clear for castle
	kingPathClear := (emptySquares & kingFromToPath) == kingFromToPath
	rookPathClear := (emptySquares & rookFromToPath) == rookFromToPath
	kingSquaresNotAttacked := pos.AttackedSquares(side.Opponent())&(kingFromToPath|kingToSq|kingBB) == 0

	if kingPathClear && rookPathClear && kingSquaresNotAttacked {
		return true
	}
	return false
}

// canCastleLong checks if the king can castle long
func (pos *Position) canCastleLong(side Color) bool {
	if !pos.castling.castlingRights.canCastle(Q) && side == White {
		return false
	}
	if !pos.castling.castlingRights.canCastle(q) && side == Black {
		return false
	}

	rookBB := bitboardFromIndex(pos.castling.rooksStartSquare[side][1])
	kingBB := bitboardFromIndex(pos.castling.kingsStartSquare[side])

	kingToSq := bitboardFromIndex(pos.castling.kingsEndSquare[side][1])
	kingFromToPath := getRayPath(&kingBB, &kingToSq) | kingToSq

	rookEndBB := bitboardFromIndex(pos.castling.rooksEndSquare[side][1])
	rookFromToPath := getRayPath(&rookBB, &rookEndBB) | rookEndBB

	emptySquares := pos.EmptySquares() | kingBB | rookBB // NOTE: Avoid king and rook, so are not taken into account when calculating whenever the path is clear for castle
	kingPathClear := (emptySquares & kingFromToPath) == kingFromToPath
	rookPathClear := (emptySquares & rookFromToPath) == rookFromToPath
	kingSquaresNotAttacked := pos.AttackedSquares(side.Opponent())&(kingFromToPath|kingToSq|kingBB) == 0

	if kingPathClear && rookPathClear && kingSquaresNotAttacked {
		return true
	}
	return false
}
