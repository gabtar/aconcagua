package main

import (
	"github.com/gabtar/aconcagua/internal/engine"
)

func main() {
	eng := engine.NewEngine()
	eng.StartUci()
}
