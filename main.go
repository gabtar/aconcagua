package main

import (
	"github.com/gabtar/aconcagua/aconcagua"
	// "fmt"
	// "github.com/gabtar/aconcagua/tunner"
)

func main() {
	eng := aconcagua.NewEngine()
	eng.StartUci()

	// dataSet := tunner.LoadDataSet("quiet-labeled.epd")
	// fmt.Println(tunner.MeanSquareError(0.28, dataSet))

	// tunner.Tunner(0.28, dataSet, 41)
}
