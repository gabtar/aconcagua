package main

import (
	// "fmt"

	// "github.com/gabtar/aconcagua/aconcagua"
	"github.com/gabtar/aconcagua/tuner"
)

func main() {
	// eng := aconcagua.NewEngine()
	// eng.StartUci()

	dataSet := tuner.LoadDataSet("./tuner/training-set/quiet-labeled.epd")
	// params := tuner.GetEvaluationParams()

	tuner.Tuner(1.2, dataSet, 51)
}
