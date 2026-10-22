package uci

import "github.com/gabtar/aconcagua/internal/engine"

// SpsaCommandStruct prints the parameter list for the OpenBench tune page
type SpsaCommandStruct struct{}

func (c *SpsaCommandStruct) Execute(en *engine.Engine, stdout chan string, params ...string) {
	for _, line := range engine.SpsaInputLines() {
		stdout <- line
	}
}
