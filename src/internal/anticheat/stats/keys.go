package stats

// keys.go centralises the metric categories shared by collectors, detections,
// publishing and grading. Keys themselves remain string constants at their use
// sites; keeping them close to the measurement code makes each detection file
// readable in isolation.
const (
	killsCategory      = Category("kills")
	aimingCategory     = Category("aiming")
	reactionCategory   = Category("reaction")
	recoilCategory     = Category("recoil")
	behavioralCategory = Category("behavioral")

	scoreboardCategory = Category("scoreboard")
	sniperCategory     = Category("sniper")
	grenadeCategory    = Category("utility")

	gameInfoCategory  = Category("game_info")
	antiCheatCategory = Category("anti_cheat")
)
