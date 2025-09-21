package aconcagua

import "time"

const (
	DepthStrategy    = iota // fixed depth
	InfiniteStrategy        // Max depth
	MoveTimeStrategy        // Move Time
	TimeLeftStrategy        // wtime and btime passed
)

// TimeControl hanldes the time during search
// Strategy for time control during search
// Assuming a game will last around 50 moves (avegage is 40, but use a margin of 10)
// Moves 1 - 20 use 60% of time - 0.03 timeLeftInMiliseconds per move
// Moves 21 - 40 use 30% of time - Less pieces on the board search is faster - 0.015 timeLeftInMiliseconds per move
// Moves 41 - 50 use 10% of time -  0.01 timeLeftInMiliseconds per move
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
func (c *Clock) timeLeft(side Color) float32 {
	if side == 1 {
		return float32(c.btime + c.binc)
	}
	return float32(c.wtime + c.winc)
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
		timeLeft := clock.timeLeft(Color(side))
		if moveNumber <= 20 {
			return int(0.03 * timeLeft)
		}
		if moveNumber <= 40 {
			return int(0.015 * timeLeft)
		}

		return int(0.01 * timeLeft)
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
