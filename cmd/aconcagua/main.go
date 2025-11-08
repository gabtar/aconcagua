package main

import (
	"github.com/gabtar/aconcagua/internal/engine"
	"github.com/gabtar/aconcagua/internal/uci"
)

func main() {
	eng := engine.NewEngine()
	uci := uci.NewUciProtocol(eng)
	uci.Start()
}
