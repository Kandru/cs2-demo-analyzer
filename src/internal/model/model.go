// Package model contains the parser-independent domain model for a CS2 match.
//
// Nothing in this package may import demoinfocs or any presentation package.
// Parsers build the model, analysis enriches it, reporters render it.
package model

import "time"

// SchemaVersion is bumped whenever the JSON/HTML data contract changes in a
// backward-incompatible way. The manifest uses it to invalidate cached reports.
const SchemaVersion = 1

// Side is a counter-strike team side.
type Side string

const (
	SideCT         Side = "CT"
	SideT          Side = "T"
	SideSpectator  Side = "SPECTATOR"
	SideUnassigned Side = "UNASSIGNED"
	SideUnknown    Side = ""
)

// Team is one of the two participating teams. Team.ID is the stable TeamState
// entity ID, which does not change when the team switches sides at halftime.
type Team struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	ClanName  string   `json:"clanName,omitempty"`
	Flag      string   `json:"flag,omitempty"`
	StartSide Side     `json:"startSide"`
	FinalSide Side     `json:"finalSide"`
	Score     int      `json:"score"`
	PlayerIDs []string `json:"playerIds"`
}

// Match is the root object produced by the parser.
type Match struct {
	SchemaVersion int       `json:"schemaVersion"`
	ID            string    `json:"id"`
	DemoPath      string    `json:"demoPath"`
	DemoName      string    `json:"demoName"`
	DemoSize      int64     `json:"demoSize"`
	DemoHash      string    `json:"demoHash,omitempty"`
	DemoStamp     string    `json:"demoStamp,omitempty"`
	ServerName    string    `json:"serverName,omitempty"`
	Map           string    `json:"map"`
	Mode          string    `json:"mode"`
	TickRate      float64   `json:"tickRate"`
	Ticks         int       `json:"ticks"`
	Duration      float64   `json:"durationSeconds"`
	FileTime      time.Time `json:"fileTime"`
	AnalyzedAt    time.Time `json:"analyzedAt"`
	Teams         []*Team   `json:"teams"`
	Players       []*Player `json:"players"`
	Rounds        []*Round  `json:"rounds"`
	Warmup        *Warmup   `json:"warmup,omitempty"`
	Summary       *Summary  `json:"summary,omitempty"`
	Warnings      []string  `json:"warnings,omitempty"`
	Partial       bool      `json:"partial,omitempty"`
	ParseError    string    `json:"parseError,omitempty"`
	ParserVersion string    `json:"parserVersion,omitempty"`
	ToolVersion   string    `json:"toolVersion,omitempty"`
}

// Warmup contains kill/event activity before the first official round.
type Warmup struct {
	StartOffset float64  `json:"startOffsetSeconds"`
	EndOffset   float64  `json:"endOffsetSeconds"`
	Duration    float64  `json:"durationSeconds"`
	Events      []*Event `json:"events"`
	PlayerIDs   []string `json:"playerIds,omitempty"`
}

// Round is one official round (including overtime rounds).
type Round struct {
	Number       int                   `json:"number"`
	StartTick    int                   `json:"startTick"`
	EndTick      int                   `json:"endTick"`
	StartOffset  float64               `json:"startOffsetSeconds"`
	EndOffset    float64               `json:"endOffsetSeconds"`
	Duration     float64               `json:"durationSeconds"`
	Winner       Side                  `json:"winner"`
	WinnerTeamID int                   `json:"winnerTeamId"`
	Reason       string                `json:"reason"`
	Message      string                `json:"message,omitempty"`
	StartScore   [2]int                `json:"startScore"`
	EndScore     [2]int                `json:"endScore"`
	Overtime     bool                  `json:"overtime"`
	Pregame      bool                  `json:"pregame,omitempty"`
	Economy      []*TeamEconomy        `json:"economy,omitempty"`
	Events       []*Event              `json:"events"`
	Players      map[string]*RoundStat `json:"players"`
	Warnings     []string              `json:"warnings,omitempty"`
}

// TeamEconomy is the freeze-time-end economy snapshot for one team.
type TeamEconomy struct {
	TeamID         int    `json:"teamId"`
	Side           Side   `json:"side"`
	EquipmentValue int    `json:"equipmentValue"`
	Money          int    `json:"money"`
	BuyType        string `json:"buyType"`
}

// RoundStat is a player's contribution to a single round.
type RoundStat struct {
	Side          Side `json:"side"`
	Kills         int  `json:"kills"`
	Deaths        int  `json:"deaths"`
	Assists       int  `json:"assists"`
	Damage        int  `json:"damage"`
	Headshots     int  `json:"headshotKills"`
	FirstKill     bool `json:"firstKill,omitempty"`
	FirstDeath    bool `json:"firstDeath,omitempty"`
	TradeKill     bool `json:"tradeKill,omitempty"`
	Traded        bool `json:"traded,omitempty"`
	KAST          bool `json:"kast"`
	Survived      bool `json:"survived"`
	Clutch        int  `json:"clutch,omitempty"` // 1 = 1v1 won, 2 = 1v2, ...
	EquipValue    int  `json:"equipValue,omitempty"`
	Spent         int  `json:"spent,omitempty"`
	Plants        int  `json:"plants,omitempty"`
	Defuses       int  `json:"defuses,omitempty"`
	UtilityDamage int  `json:"utilityDamage,omitempty"`
	FlashAssists  int  `json:"flashAssists,omitempty"`
	Result        Side `json:"result,omitempty"` // winning side
}

// Player is one participant, keyed by SteamID64 when available.
type Player struct {
	ID           string                  `json:"id"`
	SteamID64    uint64                  `json:"steamId64,omitempty"`
	Name         string                  `json:"name"`
	Aliases      []string                `json:"aliases,omitempty"`
	ClanTag      string                  `json:"clanTag,omitempty"`
	TeamID       int                     `json:"teamId"`
	IsBot        bool                    `json:"isBot,omitempty"`
	Slug         string                  `json:"slug"`
	RoundsPlayed int                     `json:"roundsPlayed"`
	Stats        PlayerStats             `json:"stats"`
	SideStats    map[Side]*PlayerStats   `json:"sideStats,omitempty"`
	RoundStats   map[int]*RoundStat      `json:"roundStats,omitempty"`
	Weapons      map[string]*WeaponStats `json:"weapons,omitempty"`
	Aim          *AimStats               `json:"aim,omitempty"`
	Cheat        *CheatReport            `json:"cheat,omitempty"`
}

// PlayerStats is a match aggregate.
type PlayerStats struct {
	Kills           int     `json:"kills"`
	Deaths          int     `json:"deaths"`
	Assists         int     `json:"assists"`
	HeadshotKills   int     `json:"headshotKills"`
	Damage          int     `json:"damage"`
	MVPs            int     `json:"mvps"`
	Score           int     `json:"score"`
	FirstKills      int     `json:"firstKills"`
	FirstDeaths     int     `json:"firstDeaths"`
	TradeKills      int     `json:"tradeKills"`
	TradedDeaths    int     `json:"tradedDeaths"`
	KASTRounds      int     `json:"kastRounds"`
	Plants          int     `json:"plants"`
	Defuses         int     `json:"defuses"`
	Clutches        int     `json:"clutches"`
	UtilityDamage   int     `json:"utilityDamage"`
	FlashAssists    int     `json:"flashAssists"`
	Rounds          int     `json:"rounds"`
	Shots           int     `json:"shots,omitempty"`
	Hits            int     `json:"hits,omitempty"`
	HeadHits        int     `json:"headHits,omitempty"`
	DamageTaken     int     `json:"damageTaken,omitempty"`
	EquipValueSum   int     `json:"equipValueSum,omitempty"`
	MoneySpent      int     `json:"moneySpent,omitempty"`
	ADR             float64 `json:"adr"`
	KD              float64 `json:"kd"`
	KDR             float64 `json:"kdr"`
	HSPercent       float64 `json:"hsPercent"`
	KASTPercent     float64 `json:"kastPercent"`
	AccuracyPct     float64 `json:"accuracyPercent,omitempty"`
	HeadAccuracyPct float64 `json:"headAccuracyPercent,omitempty"`
}

// WeaponStats aggregates per-weapon usage.
type WeaponStats struct {
	Name      string  `json:"name"`
	Kills     int     `json:"kills"`
	Headshots int     `json:"headshotKills"`
	Damage    int     `json:"damage"`
	Shots     int     `json:"shots,omitempty"`
	Hits      int     `json:"hits,omitempty"`
	HeadHits  int     `json:"headHits,omitempty"`
	Accuracy  float64 `json:"accuracyPercent,omitempty"`
}

// Event is one timeline entry in a round or the warmup.
type Event struct {
	Seq           int     `json:"seq"`
	Tick          int     `json:"tick"`
	Offset        float64 `json:"offsetSeconds"`
	Type          string  `json:"type"`
	ActorID       string  `json:"actorId,omitempty"`
	ActorName     string  `json:"actorName,omitempty"`
	VictimID      string  `json:"victimId,omitempty"`
	VictimName    string  `json:"victimName,omitempty"`
	AssisterID    string  `json:"assisterId,omitempty"`
	AssisterName  string  `json:"assisterName,omitempty"`
	Weapon        string  `json:"weapon,omitempty"`
	Headshot      bool    `json:"headshot,omitempty"`
	Wallbang      bool    `json:"wallbang,omitempty"`
	Noscope       bool    `json:"noscope,omitempty"`
	ThroughSmoke  bool    `json:"throughSmoke,omitempty"`
	AssistedFlash bool    `json:"assistedFlash,omitempty"`
	Site          string  `json:"site,omitempty"`
	Grenade       string  `json:"grenade,omitempty"`
	TeamKill      bool    `json:"teamKill,omitempty"`
	Suicide       bool    `json:"suicide,omitempty"`
	Text          string  `json:"text,omitempty"`
	Score         [2]int  `json:"score,omitempty"`
}

// Summary contains match-level derived values used by all reporters.
type Summary struct {
	WinnerTeamID     int         `json:"winnerTeamId"`
	WinnerName       string      `json:"winnerName"`
	Score            [2]int      `json:"score"`
	Rounds           int         `json:"rounds"`
	RegulationRounds int         `json:"regulationRounds"`
	FlaggedPlayers   int         `json:"flaggedPlayers"`
	TopFraggerID     string      `json:"topFraggerId"`
	TopFraggerName   string      `json:"topFraggerName"`
	TopFraggerKills  int         `json:"topFraggerKills"`
	BestAimID        string      `json:"bestAimId"`
	BestAimName      string      `json:"bestAimName"`
	BestAimRating    float64     `json:"bestAimRating"`
	BestAimGrade     string      `json:"bestAimGrade,omitempty"`
	TeamStats        []TeamStats `json:"teamStats,omitempty"`
}

// TeamStats is a match aggregate per team.
type TeamStats struct {
	TeamID    int     `json:"teamId"`
	Name      string  `json:"name"`
	Score     int     `json:"score"`
	Kills     int     `json:"kills"`
	Deaths    int     `json:"deaths"`
	Assists   int     `json:"assists"`
	Damage    int     `json:"damage"`
	ADR       float64 `json:"adr"`
	HSPercent float64 `json:"hsPercent"`
	AimRating float64 `json:"aimRating"`
}

// AimStats is the Leetify-style aim profile for one player.
type AimStats struct {
	Rating     float64     `json:"rating"`
	Grade      string      `json:"grade"`
	Confidence float64     `json:"confidence"`
	Metrics    []AimMetric `json:"metrics"`
}

// AimMetric is one tile of the aim profile.
type AimMetric struct {
	ID        string  `json:"id"`
	Label     string  `json:"label"`
	Value     float64 `json:"value"`
	Display   string  `json:"display"`
	Unit      string  `json:"unit"`
	Score     float64 `json:"score"`
	Weight    float64 `json:"weight"`
	Samples   int     `json:"samples"`
	LowSample bool    `json:"lowSample,omitempty"`
	Tooltip   string  `json:"tooltip"`
}

// CheatReport is the vendored anti-cheat detector's verdict for one player.
type CheatReport struct {
	Available     bool           `json:"available"`
	Error         string         `json:"error,omitempty"`
	Likelihood    float64        `json:"likelihood"`
	Flagged       bool           `json:"flagged"`
	FlagThreshold float64        `json:"flagThreshold"`
	OverallGrade  string         `json:"overallGrade,omitempty"`
	Grades        []Grade        `json:"grades,omitempty"`
	Narrative     string         `json:"narrative,omitempty"`
	Channels      []CheatChannel `json:"channels,omitempty"`
	Adjustments   []Adjustment   `json:"adjustments,omitempty"`
}

// Grade is a category skill grade from the detector.
type Grade struct {
	Category string `json:"category"`
	Label    string `json:"label"`
	Grade    string `json:"grade"`
}

// CheatChannel is one transparency channel of the detector.
type CheatChannel struct {
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	Score       float64 `json:"score"`
	Confidence  float64 `json:"confidence"`
	Zone        string  `json:"zone"`
	Raw         float64 `json:"raw"`
	Unit        string  `json:"unit"`
	Samples     int     `json:"samples"`
	Weight      float64 `json:"weight"`
	HasData     bool    `json:"hasData"`
	Description string  `json:"description,omitempty"`
}

// Adjustment is a detector boost/discount/override that shaped the verdict.
type Adjustment struct {
	Label  string `json:"label"`
	Detail string `json:"detail,omitempty"`
}

// PlayerByID returns the player with the given stable ID.
func (m *Match) PlayerByID(id string) *Player {
	for _, p := range m.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// TeamByID returns the team with the given stable ID.
func (m *Match) TeamByID(id int) *Team {
	for _, t := range m.Teams {
		if t.ID == id {
			return t
		}
	}
	return nil
}

// RoundByNumber returns the round with the given label number.
func (m *Match) RoundByNumber(n int) *Round {
	for _, r := range m.Rounds {
		if r.Number == n {
			return r
		}
	}
	return nil
}
