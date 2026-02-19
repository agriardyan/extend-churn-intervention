package handler

const (
	// Stat codes used by the game
	StatCodeRageQuit     = "rse-rage-quit"
	StatCodeMatchWins    = "rse-match-wins"
	StatCodeLosingStreak = "rse-current-losing-streak"

	// Challenge parameters
	ChallengeWinsNeeded   = 3
	ChallengeDurationDays = 7

	// Intervention parameters
	InterventionCooldownHours = 48
	RageQuitThreshold         = 3
	LosingStreakThreshold     = 5

	// Default item ID (fallback if env var not set)
	DefaultSpeedBoosterItemID = "speed_booster"
)
