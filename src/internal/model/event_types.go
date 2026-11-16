package model

// Timeline event type identifiers. The parser, renderer and report filters use
// these constants so adding a new event only requires one identifier plus its
// handler and presentation.
const (
	EventRoundStart    = "round_start"
	EventFreezeEnd     = "freeze_end"
	EventKill          = "kill"
	EventTeamKill      = "teamkill"
	EventSuicide       = "suicide"
	EventBombPlant     = "bomb_plant"
	EventBombDefuse    = "bomb_defuse"
	EventBombExplode   = "bomb_explode"
	EventBombPickup    = "bomb_pickup"
	EventBombDrop      = "bomb_drop"
	EventGrenade       = "grenade"
	EventGrenadeDamage = "grenade_damage"
	EventFlash         = "flash"
	EventMVP           = "mvp"
	EventRoundEnd      = "round_end"
	EventPlayerConnect = "player_connect"
	EventPlayerLeave   = "player_disconnect"
	EventTeamChange    = "team_change"
	EventTeamSwap      = "team_swap"
	EventHostagePickup = "hostage_pickup"
	EventHostageDrop   = "hostage_drop"
	EventHostageRescue = "hostage_rescue"
	EventHostageKilled = "hostage_killed"
)
