package board

// castling represents the castling rights available in the position
// Represented in a binary of 4 bits where 0 is no castling available and 1 is available
// 0000: "-" NO CASTLE
// 1000: "K" WHITE
// 0100: "Q" WHITE
// 0010: "k" BLACK
// 0001: "q" BLACK
// and all the combinations of them
type castling int

const (
	SHORT_CASTLE_WHITE = castling(0b1000)
	LONG_CASTLE_WHITE  = castling(0b0100)
	SHORT_CASTLE_BLACK = castling(0b0010)
	LONG_CASTLE_BLACK  = castling(0b0001)
)

// castleMap matchs a castle string to a castle type representation
var castleReference = map[rune]castling{
	'K': SHORT_CASTLE_WHITE,
	'Q': LONG_CASTLE_WHITE,
	'k': SHORT_CASTLE_BLACK,
	'q': LONG_CASTLE_BLACK,
}

func (c *castling) toFen() (castlingRights string) {
	if *c&SHORT_CASTLE_WHITE > 0 {
		castlingRights += "K"
	}
	if *c&LONG_CASTLE_WHITE > 0 {
		castlingRights += "Q"
	}
	if *c&SHORT_CASTLE_BLACK > 0 {
		castlingRights += "k"
	}
	if *c&LONG_CASTLE_BLACK > 0 {
		castlingRights += "q"
	}

	if *c == 0 {
		castlingRights += "-"
	}
	return
}

func (c *castling) fromFen(castleFen string) {
	if castleFen == "-" {
		*c = castling(0)
		return
	}

	for _, castleType := range castleFen {
		*c |= castleReference[castleType]
	}
}

func (c *castling) add(castle castling) {
	*c |= castle
}

func (c *castling) remove(castle castling) {
	*c &^= castle
}

func (c *castling) canCastle(to castling) bool {
	return *c&to > 0
}
