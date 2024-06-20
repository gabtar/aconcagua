// evaluation contains all functions related to position evaluation in the engine
package evaluation

import (
	"github.com/gabtar/aconcagua/aconcagua"
)

func Eval(pos *aconcagua.Position) int {
	// Only middlegame evaluation for now

	// TODO: implement mate during search instead of during evaluation
	if pos.Checkmate(pos.Turn) {
		return -100000
	} else if pos.Checkmate(pos.Turn.Opponent()) {
		return 100000
	}

	mgScore := [2]int{0, 0}
	egScore := [2]int{0, 0}
	phase := 0
	phaseInc := [12]int{0, 9, 5, 3, 3, 0, 0, 9, 5, 3, 3, 0}

	for p, bb := range pos.Bitboards {
		color := 0
		if p > 5 {
			color = 1
		}

		for bb > 0 {
			sq := aconcagua.Bsf(bb.NextBit())
			mgScore[color] += middlegamePiecesScore[p][sq]
			egScore[color] += endgamePiecesScore[p][sq]
			phase += phaseInc[p]

		}
	}

	mgPhase := phase
	if mgPhase > 62 {
		mgPhase = 62 // case of an early promotion
	}
	egPhase := 62 - mgPhase

	turn := pos.Turn
	opponent := aconcagua.White
	if opponent == turn {
		opponent = aconcagua.Black
	}

	mg := mgScore[turn] - mgScore[opponent]
	eg := egScore[turn] - egScore[opponent]

	return (mg*mgPhase + eg*egPhase) / 62
}
