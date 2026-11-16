package stats

// keys.go centralises the metric categories shared by collectors, detections,
// publishing and grading. Keys themselves remain string constants at their use
// sites; keeping them close to the measurement code makes each detection file
// readable in isolation.
const (
	channelCategoryKills      = Category("kills")
	channelCategoryAiming     = Category("aiming")
	channelCategoryReaction   = Category("reaction")
	channelCategoryRecoil     = Category("recoil")
	channelCategoryBehavioral = Category("behavioral")

	scoreboardCategory = Category("scoreboard")
	sniperCategory     = Category("sniper")
	grenadeCategory    = Category("utility")

	cheatscoreCategoryGameInfo  = Category("game_info")
	cheatscoreCategoryAntiCheat = Category("anti_cheat")
)
