package main

import (
	// "fmt"
	"github.com/gabtar/aconcagua/aconcagua"
	// "github.com/gabtar/aconcagua/tuner"
)

func main() {
	eng := aconcagua.NewEngine()
	eng.StartUci()

	// Tune Evaluation Params - Using zurich chess training set
	// dataSet := tuner.LoadDataSet("./tuner/training-set/quiet-labeled.epd")

	// k := tuner.FindOptimalScalingFactor(dataSet, tuner.GetEvaluationParams())
	// fmt.Println("Best K", k)

	// params := tuner.GetEvaluationParams()

	// tuner.AdamTuner(params, dataSet, k, 2000)
}
