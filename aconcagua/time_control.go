package aconcagua

import "time"

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

// init initializes the TimeControl struct
func (tc *TimeControl) init(strategy int, side int, moveNumber int, clock Clock) {
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
		if moveNumber <= 25 { // For the first 25 moves takes around 70% of the total time
			return int(0.03*(1+float32(moveNumber)/25)*timeLeft) + int(incr*4/5)
		}
		if moveNumber <= 50 { // For moves between 25-50 takes around 15% of the total time
			return int(0.015*(1+float32(moveNumber-25)/50)*timeLeft) + int(incr*4/5)
		}

		// For moves >50 takes around 10% of the total time
		return int(0.01*timeLeft) + int(incr/2)
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
