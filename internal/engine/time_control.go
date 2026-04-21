package engine

import (
	"strconv"
	"time"
)

const (
	DepthStrategy    = iota // Fixed depth search
	InfiniteStrategy        // Max depth search
	MoveTimeStrategy        // Fixed move time
	TimeLeftStrategy        // Tournament time control play
)

// TimeControl hanldes the time during search
type TimeControl struct {
	startTime          time.Time
	iterationStartTime time.Time
	limits             TimeLimits
	strategy           int
	stop               bool
}

// TimeLimits stores the soft and hard time limits in miliseconds
type TimeLimits struct {
	softLimit int // Optimal time
	hardLimit int // Max time allowed
}

// Clock is the struct to store the white and black time, and increments
type Clock struct {
	wtime     int
	btime     int
	winc      int
	binc      int
	moveTime  int
	movesToGo int
}

// timeLeft returns the time left for the side
func (c *Clock) timeLeft(side Color) (timeLeft float64, increment float64) {
	if side == 1 {
		return float64(c.btime), float64(c.binc)
	}
	return float64(c.wtime), float64(c.winc)
}

// Initialize initializes the TimeControl struct
func (tc *TimeControl) Initialize(strategy int, side int, moveNumber int, clock Clock) {
	tc.startTime = time.Now()
	tc.iterationStartTime = time.Now()
	tc.stop = false

	tc.setupLimits(strategy, side, moveNumber, clock)
}

// setupLimits sets the limits for the search
func (tc *TimeControl) setupLimits(strategy int, side int, moveNumber int, clock Clock) {
	soft, hard := -1, -1
	switch strategy {
	case MoveTimeStrategy:
		soft = clock.moveTime
		hard = clock.moveTime
	case TimeLeftStrategy:
		timeLeft, incr := clock.timeLeft(Color(side))
		if clock.movesToGo == -1 {
			clock.movesToGo = tc.estimatedMovesToGo(moveNumber)
		}

		maxTime := int(timeLeft/float64(clock.movesToGo)) + int(incr)
		soft, hard = defineLimits(maxTime, moveNumber, int(timeLeft))
	}

	tc.limits.softLimit = soft
	tc.limits.hardLimit = hard
}

// defineLimits returns the limits for the search
func defineLimits(maxTime int, moveNumber int, timeLeft int) (softLimit int, hardLimit int) {
	if moveNumber >= 40 || timeLeft < 5000 {
		return maxTime / 2, maxTime
	}
	return maxTime / 2, maxTime * 5
}

// shouldStop returns true if the search should stop
func (tc *TimeControl) shouldStop() bool {
	if tc.limits.hardLimit == -1 { // Avoid stop when using Infinite/depth strategy
		return false
	}

	// If we reached hard limit, stop inmediately
	if int(time.Since(tc.startTime).Milliseconds()) >= tc.limits.hardLimit {
		tc.stop = true
		return true
	}

	return tc.stop
}

// shouldStopEarly returns true if the search should stop
func (tc *TimeControl) shouldStopEarly(stable bool) bool {
	if tc.shouldStop() {
		return true
	}

	if tc.limits.softLimit < 0 {
		return false // Infinite/depth strategy
	}

	elapsed := int(time.Since(tc.startTime).Milliseconds())
	if elapsed >= tc.limits.softLimit && stable {
		tc.stop = true
		return true
	}

	return tc.stop
}

// extendTime extends the search time by the factor
func (tc *TimeControl) extendTime(factor float64) {
	newSoft := int(float64(tc.limits.softLimit) * factor)
	tc.limits.softLimit = min(newSoft, tc.limits.hardLimit*4/5) // cap at 80%
}

// TimeStrategy returns the search strategy and the clock for a search
func TimeStrategy(params []string, depth int, wtime int, btime int, winc int, binc int, movetime int, movesToGo int) (int, Clock) {
	if movetime != -1 {
		movetime, _ = strconv.Atoi(params[movetime+1])
		return MoveTimeStrategy, Clock{0, 0, 0, 0, movetime, 0}
	}
	if wtime != -1 || btime != -1 {
		wtime, _ = strconv.Atoi(params[wtime+1])
		btime, _ = strconv.Atoi(params[btime+1])
		winc, _ = strconv.Atoi(params[winc+1])
		binc, _ = strconv.Atoi(params[binc+1])
		if movesToGo != -1 {
			movesToGo, _ = strconv.Atoi(params[movesToGo+1])
		}
		return TimeLeftStrategy, Clock{wtime, btime, winc, binc, 0, movesToGo}
	}
	if depth != -1 {
		return DepthStrategy, Clock{0, 0, 0, 0, 0, 0}
	}
	return InfiniteStrategy, Clock{0, 0, 0, 0, 0, 0}
}

// estimatedMovesToGo returns the an approximate moves to go for a given move number
func (tc *TimeControl) estimatedMovesToGo(moveNumber int) int {
	if moveNumber <= 20 {
		return 60 - moveNumber
	}
	if moveNumber < 40 {
		return 40 - (moveNumber-20)/3
	}
	if moveNumber < 60 {
		return 35 - (moveNumber-40)/4
	}
	return 30
}
