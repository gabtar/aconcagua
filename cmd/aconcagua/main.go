package main

import (
	"github.com/gabtar/aconcagua/internal/engine"
	"github.com/gabtar/aconcagua/internal/uci"
	// "github.com/gabtar/aconcagua/internal/tuner"
)

func main() {
	eng := engine.NewEngine()
	uci := uci.NewUciProtocol(eng)
	uci.Start()

	// Use to run the tuner
	// dataset := tuner.LoadDataSet("./internal/tuner/training-set/lichess-big3-resolved.book", 7000000)
	// dataset := tuner.LoadDataSet("./internal/tuner/training-set/trainingdata.epd", 7878653) // Lichess + zurichess combined
	// params := tuner.GetEvaluationParams()
	// tuner.AdamTuner(params, &dataset, tuner.ScalingFactor, 300)

	// Find fixed magic numbers
	// engine.GenerateMagicNumbersForRooksAndBishops()
}
