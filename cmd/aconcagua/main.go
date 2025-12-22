package main

import (
	// "github.com/gabtar/aconcagua/internal/engine"
	// "github.com/gabtar/aconcagua/internal/uci"
	"fmt"

	"github.com/gabtar/aconcagua/internal/tuner"
)

func main() {
	// eng := engine.NewEngine()
	// uci := uci.NewUciProtocol(eng)
	// uci.Start()

	// dataset := tuner.LoadDataSet("./internal/tuner/training-set/quiet-labeled.epd")

	// NOTE: Use only 4 million first samples to not excceed memory limits
	dataset := tuner.LoadDataSet("./internal/tuner/training-set/lichess-big3-resolved.book", 4000000)
	params := tuner.GetEvaluationParams()

	sf := tuner.FindOptimalScalingFactor(dataset, params)
	fmt.Println(sf)
}
