package engine

type Variant int

const (
	Standard Variant = iota
	Chess960
	Atomic
	// TODO: add other specific variants
	// CrazyHouse
	// Antichess
)

// TODO:: create an interface that will implement all variants for make/unmake move/ evaluate position and other methods like generate moves etc...
// Think of a better name for the interface.
// Use main methods used in search and quiescent
type IPosition interface {
	Evaluate() int
	MakeMove(move *Move)
	UnmakeMove(move *Move)
	GenerateCaptures(ml *MoveList, pd *PositionData)
	GenerateNonCaptures(ml *MoveList, pd *PositionData)
	GetPositionData() PositionData
	IsLegal(move Move) bool // should depend on variant. This one should be used in move generator
}

// StandardPosition is a Standard chess position
// TODO: already covered in Position may need to refactor to implement the interface acordingly
type StandardPosition struct {
	pos Position
}

// Chess960Position is a Chess960/Fischer Random Chess position
// TODO: Now Position handles both standard and chess960. Later will be refactored to use different types for each variant
type Chess960Position struct {
	pos    Position
	castle castlingRights
}
