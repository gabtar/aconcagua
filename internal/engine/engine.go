package engine

// Engine represents a chess engine
type Engine struct {
	Pos         Position
	Search      Search
	TimeControl TimeControl
	OpeningBook PolyglotBook
	Options     Options
}

// Options are various engine options that can be set
type Options struct {
	UseOpeningBook bool
	Chess960       bool
}

// NewEngine returns a new Engine instance
func NewEngine() *Engine {
	pos := NewPosition()
	pos.LoadFromFenString(StartingFenString)
	return &Engine{
		Pos:         *pos,
		Search:      *NewSearch(),
		OpeningBook: PolyglotBook{},
		Options: Options{
			UseOpeningBook: false,
			Chess960:       false,
		},
	}
}
