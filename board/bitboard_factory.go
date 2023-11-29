package board

// bitboardFromCoordinate is a factory that returns a bitboard from a string coordinate
func bitboardFromCoordinate(coordinate string) (bitboard Bitboard) {
	return (bitboard + 1) << coordinateToSquareNumber[coordinate]
}

// bitboardFromCoordinates is a factory that returns a bitboard from an array of string coordinates
func bitboardFromCoordinates(coordinates []string) (bitboard Bitboard) {
	for _, coordinate := range coordinates {
		fileNumber := int(coordinate[0]) - 96
		rankNumber := int(coordinate[1]) - 48
		squareNumber := (fileNumber - 1) + 8*(rankNumber-1)

		// displaces 1 bit to the coordinate passed
		bitboard |= 0b1 << squareNumber
	}
	return
}

// bitboardFromIndex is a factory that returns a bitboard from an index square
func bitboardFromIndex(index int) (bitboard Bitboard) {
	// NOTE: Since bitscan cannot be used with empty sets i use this guard clause to
	// ensure returning a valid bitboard for the engine
	if index > 63 || index < 0 {
		bitboard = Bitboard(0)
	} else {
		bitboard = Bitboard(0b1 << index)
	}
	return
}
