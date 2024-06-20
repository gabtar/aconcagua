package aconcagua

// Color is an int referencing the white or black player in chess
type Color int

const (
	White Color = iota
	Black
)

// Opponent returns the Opponent color to the actual color
func (c Color) Opponent() Color {
	if c == White {
		return Black
	}
	return White
}
