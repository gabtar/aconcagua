package board

// Color is an int referencing the white or black player in chess
type Color int

const (
	White Color = iota
	Black
)

// opponent returns the opponent color to the actual color
func (c Color) opponent() Color {
	if c == White {
		return Black
	} else {
		return White
	}
}
