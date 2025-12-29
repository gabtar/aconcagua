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

	// dataset := tuner.LoadDataSet("./internal/tuner/training-set/lichess-big3-resolved.book", 3000000)
	// params := tuner.GetEvaluationParams()
	//

	// tuner.AdamTuner(params, &dataset, tuner.ScalingFactor, 1000)
}
