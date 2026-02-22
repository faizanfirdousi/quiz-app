package game

// CalculateScore implements Kahoot-style scoring formula.
// Full points for instant answers, linear decay based on time taken.
// Returns 0 for incorrect answers.
func CalculateScore(isCorrect bool, timeTakenMs int64, timeLimitMs int64, basePoints int) int {
	if !isCorrect {
		return 0
	}

	if timeLimitMs <= 0 {
		return basePoints
	}

	// Clamp timeTaken to timeLimitMs
	if timeTakenMs < 0 {
		timeTakenMs = 0
	}
	if timeTakenMs > timeLimitMs {
		timeTakenMs = timeLimitMs
	}

	// Full points for first half of time, then linear decay
	timeRatio := float64(timeTakenMs) / float64(timeLimitMs)
	bonus := int(float64(basePoints) * 0.5 * (1 - timeRatio))

	return basePoints + bonus
}
