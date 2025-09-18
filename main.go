package main

import (
	// "github.com/gabtar/aconcagua/aconcagua"
	"github.com/gabtar/aconcagua/tuner"
)

func main() {
	// eng := aconcagua.NewEngine()
	// eng.StartUci()

	dataSet := tuner.LoadDataSet("./tuner/training-set/quiet-labeled.epd")
	// params := tuner.GetEvaluationParams()
	// fmt.Println(params)

	// k := tuner.FindOptimalScalingFactor(dataSet, tuner.GetEvaluationParams())
	// fmt.Println("Best K", k)

	tuner.Tuner(tuner.ScalingFactor, dataSet, 10)
}
