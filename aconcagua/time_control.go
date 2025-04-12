package aconcagua

// TimeControl hanldes the time during search
// Strategy for time control during search
// Assuming a game will last around 50 moves (avegage is 40, but use a margin of 10)
// Moves 1 - 20 use 60% of time - 0.03 timeLeftInMiliseconds per move
// Moves 21 - 40 use 30% of time - Less pieces on the board search is faster - 0.015 timeLeftInMiliseconds per move
// Moves 41 - 50 use 10% of time -  0.01 timeLeftInMiliseconds per move
type TimeControl struct {
	timeLeftInMiliseconds   int
	searchTimeInMiliseconds int
	stop                    bool
}

// setSearchTime sets the search time in miliseconds
func (tc *TimeControl) setSearchTime(moveNumber int) {
	time := 0.0
	if moveNumber <= 20 {
		time = 0.03 * float64(tc.timeLeftInMiliseconds)
	} else if moveNumber <= 40 {
		time = 0.015 * float64(tc.timeLeftInMiliseconds)
	} else if moveNumber <= 50 {
		time = 0.01 * float64(tc.timeLeftInMiliseconds)
	}
	tc.searchTimeInMiliseconds = int(time)
}
