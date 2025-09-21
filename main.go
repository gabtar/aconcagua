package main

import (
	// "log"
	// "net/http"
	// _ "net/http/pprof"

	"github.com/gabtar/aconcagua/aconcagua"
	// "github.com/gabtar/aconcagua/tuner"
)

func main() {
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	eng := aconcagua.NewEngine()
	eng.StartUci()

	// dataSet := tuner.LoadDataSet("./tuner/training-set/quiet-labeled.epd")
	// params := tuner.GetEvaluationParams()
	// fmt.Println(params)

	// k := tuner.FindOptimalScalingFactor(dataSet, tuner.GetEvaluationParams())
	// fmt.Println("Best K", k)

	// tuner.Tuner(tuner.ScalingFactor, dataSet, 61)
}
