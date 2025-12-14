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
	strategy           int
	stop               bool
}

// Clock is the struct to store the white and black time, and increments
type Clock struct {
	wtime    int
	btime    int
	winc     int
	binc     int
	moveTime int
}

// timeLeft returns the time left for the side
func (c *Clock) timeLeft(side Color) (timeLeft float32, increment float32) {
	if side == 1 {
		return float32(c.btime), float32(c.binc)
	}
	return float32(c.wtime), float32(c.winc)
}

// Initialize initializes the TimeControl struct
func (tc *TimeControl) Initialize(strategy int, side int, moveNumber int, clock Clock) {
	tc.startTime = time.Now()
	tc.iterationStartTime = time.Now()
	tc.stop = false

	tc.stopAfter(tc.calculateSearchTime(strategy, side, moveNumber, clock))
}

// calculateSearchTime returns the search time in miliseconds
func (tc *TimeControl) calculateSearchTime(strategy int, side int, moveNumber int, clock Clock) int {
	if strategy == MoveTimeStrategy {
		return clock.moveTime
	}
	if strategy == TimeLeftStrategy {
		timeLeft, incr := clock.timeLeft(Color(side))
		if moveNumber <= 50 {
			return int(0.035*(1+float32(moveNumber)/100)*timeLeft) + int(incr)
		}
		return int(0.01*timeLeft) + int(incr)
	}
	return -1
}

// stopAfter stops the search after the given miliseconds
func (tc *TimeControl) stopAfter(miliseconds int) {
	if miliseconds == -1 {
		return
	}
	miliseconds = max(miliseconds, 50) // Give a safe margin

	go func() {
		time.Sleep(time.Duration(miliseconds) * time.Millisecond)
		tc.stop = true
	}()
}

// TimeStrategy returns the search strategy and the clock for a search
func TimeStrategy(params []string, depth int, wtime int, btime int, winc int, binc int, movetime int) (int, Clock) {
	if movetime != -1 {
		movetime, _ = strconv.Atoi(params[movetime+1])
		return MoveTimeStrategy, Clock{0, 0, 0, 0, movetime}
	}
	if wtime != -1 || btime != -1 {
		wtime, _ = strconv.Atoi(params[wtime+1])
		btime, _ = strconv.Atoi(params[btime+1])
		winc, _ = strconv.Atoi(params[winc+1])
		binc, _ = strconv.Atoi(params[binc+1])
		return TimeLeftStrategy, Clock{wtime, btime, winc, binc, 0}
	}
	if depth != -1 {
		return DepthStrategy, Clock{0, 0, 0, 0, 0}
	}
	return InfiniteStrategy, Clock{0, 0, 0, 0, 0}
}
