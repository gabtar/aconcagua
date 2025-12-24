package main

import (
	"github.com/gabtar/aconcagua/internal/engine"
	"github.com/gabtar/aconcagua/internal/uci"
	// "fmt"
	// "github.com/gabtar/aconcagua/internal/tuner"
)

func main() {
	eng := engine.NewEngine()
	uci := uci.NewUciProtocol(eng)
	uci.Start()

	// dataset := tuner.LoadDataSet("./internal/tuner/training-set/quiet-labeled.epd")

	// dataset := tuner.LoadDataSet("./internal/tuner/training-set/lichess-big3-resolved.book", 5000000)
	// params := tuner.GetEvaluationParams()
	//
	// tuner.AdamTuner(params, &dataset, tuner.ScalingFactor, 1500)

	// sf := tuner.FindOptimalScalingFactor(dataset, params)
	// fmt.Println(sf)
}
