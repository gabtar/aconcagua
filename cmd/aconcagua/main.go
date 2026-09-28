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
	// filename := "./internal/tuner/training-set/trainingdata.epd"
	// totalSamples := 7878653
	// tuner.AdamTuner(filename, totalSamples, tuner.ScalingFactor, 5)
	// sf := tuner.FindOptimalScalingFactor(dataset, params)
	// fmt.Println("Optimal Scaling Factor: ", sf)

	// Find fixed magic numbers
	// engine.GenerateMagicNumbersForRooksAndBishops()
}
