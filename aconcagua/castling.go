package aconcagua

import "strings"

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

type castling struct {
	castlingRights   castlingRights
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

// NewCastlingFromFen returns a new castle struct from the fen string
func NewCastlingFromFen(fen string, is960 bool) *castling {
	fenElements := strings.Split(fen, " ")
	if !is960 {
		castling := NewCastling(4, 7, 0)
		castling.castlingRights.fromFen(fenElements[2])
		return castling
	} else {
		ranks := strings.Split(fenElements[0], "/")
		kingSq := strings.Index(ranks[7], "K")
		if kingSq == -1 {
			kingSq = strings.Index(ranks[0], "k")
		}
		castling := NewCastlingFromShredderFenCastlingCode(kingSq, fenElements[2])
		castling.chess960 = true
		return castling
	}
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
// Got the idea from Tom Kerrigan's TSCP engine
// based on the from/to squares of the move we can update the castle rights after making a move as following
// with the 'from' square, we are either moving a rook (disables the castling right asociated to that rook)
// or the king, in that case disables both castling rights for that color(black/white)
// with the 'to' square, means we are attaking a rook/capturing move, so we wil have to disable the castling
// right asociated with that rook after that move
func (c *castling) updateCastleRights(from int, to int) (newCastleRights castlingRights) {
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

	newCastleRights = *&c.castlingRights
	newCastleRights.remove(casltesToDisable)
	return
}

// canCastle checks if the king can castle on the position on the flag passed()
func (pos *Position) canCastle(side Color, castleFlag int) bool {
	castleChecks := [2][2]castlingRights{{K, Q}, {k, q}}
	if !pos.castling.castlingRights.canCastle(castleChecks[side][castleFlag-kingsideCastle]) {
		return false
	}

	rookBB := bitboardFromIndex(pos.castling.rooksStartSquare[side][castleFlag-kingsideCastle])
	kingBB := bitboardFromIndex(pos.castling.kingsStartSquare[side])

	kingToSq := bitboardFromIndex(pos.castling.kingsEndSquare[side][castleFlag-kingsideCastle])
	kingFromToPath := getRayPath(&kingBB, &kingToSq) | kingToSq

	rookEndBB := bitboardFromIndex(pos.castling.rooksEndSquare[side][castleFlag-kingsideCastle])
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
