package engine

import (
	"fmt"
	"strconv"
	"strings"
)

// Spsa seach parameters converted to variables, so it can be setted by setoption command
var (
	ReverseFutilityPruningMargin = 115
	AspirationWindowSize         = 25
	SEECaptureMargin             = 120
	SEEQuietMargin               = 80
	RFPImprovingMultiplier       = 65

	FutilityPruningMargin = [5]int{0, 120, 180, 280, 420}
	SEEPieceValues        = [6]int{10000, 900, 500, 300, 300, 100}
)

// Tunable is an strcut
type Tunable struct {
	Name     string
	Value    *int
	Def      int
	Min, Max int
	Step     int
}

var tunables []Tunable

// tune registers a parameter. The default is read from the variable itself.
func tune(name string, p *int, lo, hi int) {
	step := max(1, (hi-lo)/20)
	tunables = append(tunables, Tunable{Name: name, Value: p, Def: *p, Min: lo, Max: hi, Step: step})
}

func init() {
	tune("ReverseFutilityPruningMargin", &ReverseFutilityPruningMargin, 60, 200)
	tune("RFPImprovingMultiplier", &RFPImprovingMultiplier, 30, 110)
	tune("SEECaptureMargin", &SEECaptureMargin, 60, 200)
	tune("SEEQuietMargin", &SEEQuietMargin, 40, 140)
	tune("AspirationWindowSize", &AspirationWindowSize, 10, 60)

	tune("FutilityMargin1", &FutilityPruningMargin[1], 60, 200)
	tune("FutilityMargin2", &FutilityPruningMargin[2], 100, 300)
	tune("FutilityMargin3", &FutilityPruningMargin[3], 150, 450)
	tune("FutilityMargin4", &FutilityPruningMargin[4], 250, 650)

	tune("SEEQueen", &SEEPieceValues[1], 700, 1100)
	tune("SEERook", &SEEPieceValues[2], 350, 650)
	tune("SEEBishop", &SEEPieceValues[3], 220, 420)
	tune("SEEKnight", &SEEPieceValues[4], 220, 420)
}

// TunableOptionLines returns the "option name ..." lines for the uci command.
func TunableOptionLines() []string {
	lines := make([]string, 0, len(tunables))
	for _, t := range tunables {
		lines = append(lines, fmt.Sprintf("option name %s type spin default %d min %d max %d",
			t.Name, t.Def, t.Min, t.Max))
	}
	return lines
}

// SpsaInputLines returns the parameter list to paste into the OpenBench tune page.
func SpsaInputLines() []string {
	lines := make([]string, 0, len(tunables))
	for _, t := range tunables {
		lines = append(lines, fmt.Sprintf("%s, int, %d, %d, %d, %d, 0.002",
			t.Name, t.Def, t.Min, t.Max, t.Step))
	}
	return lines
}

// SetTunable sets a tunable by name. It returns false if the name is not a tunable.
func SetTunable(name, value string) bool {
	for _, t := range tunables {
		if strings.EqualFold(t.Name, name) {
			v, err := strconv.Atoi(value)
			if err != nil {
				return true // known option, bad value: ignore it
			}
			*t.Value = v
			return true
		}
	}
	return false
}
