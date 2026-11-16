package html

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/Kandru/cs2-demo-analyzer/internal/model"
	"github.com/Kandru/cs2-demo-analyzer/internal/report"
)

func testDocument(m *model.Match, version string) *report.Document {
	if m == nil {
		m = testMatch()
	}
	return report.New(m, report.Header{
		Format:          report.HeaderFormat,
		ToolVersion:     version,
		SchemaVersion:   model.SchemaVersion,
		Options:         report.Options{Cheat: true, Aim: true}.String(),
		DemoSize:        1024,
		DemoModUnixNano: time.Date(2026, 11, 13, 20, 0, 0, 0, time.UTC).UnixNano(),
		GeneratedAt:     time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC),
	})
}

func testMatch() *model.Match {
	m := &model.Match{
		SchemaVersion: model.SchemaVersion,
		Map:           "de_mirage",
		Mode:          "Competitive",
		Duration:      1800,
		TickRate:      64,
		FileTime:      time.Date(2026, 11, 13, 20, 0, 0, 0, time.UTC),
		ParserVersion: "test",
		Summary: &model.Summary{
			WinnerName: "Alpha", Score: [2]int{13, 7}, Rounds: 1, RegulationRounds: 24,
			FlaggedPlayers: 1, TopFraggerName: "kalle", TopFraggerKills: 25,
			BestAimName: "kalle", BestAimRating: 72, BestAimGrade: "B+",
		},
		Teams: []*model.Team{
			{ID: 2, Name: "Alpha", StartSide: model.SideT, FinalSide: model.SideCT, Score: 13, PlayerIDs: []string{"1"}},
			{ID: 3, Name: "Bravo", StartSide: model.SideCT, FinalSide: model.SideT, Score: 7, PlayerIDs: []string{"2"}},
		},
		Players: []*model.Player{
			{
				ID: "1", SteamID64: 76561198000000001, Name: `<b>kalle</b>`, TeamID: 2, Slug: "kalle-000001", RoundsPlayed: 20,
				Stats: model.PlayerStats{Kills: 25, Deaths: 10, Assists: 5, Damage: 2000, Rounds: 20, HeadshotKills: 12, MVPs: 4,
					ADR: 100, KD: 2.5, HSPercent: 48, KASTPercent: 75, Shots: 200, Hits: 60},
				Aim: &model.AimStats{Rating: 72, Grade: "B+", Metrics: []model.AimMetric{
					{ID: "damage_speed", Label: "Damage speed", Display: "420 ms", Samples: 12, Score: 70, Weight: 20, Tooltip: "x"},
					{ID: "reaction", Label: "Reaction", Display: "380 ms", Samples: 12, Score: 60, Weight: 20, Tooltip: "x"},
					{ID: "preaim", Label: "Preaim", Display: "8.0°", Samples: 12, Score: 80, Weight: 15, Tooltip: "x"},
				}},
				Cheat: &model.CheatReport{Available: true, Likelihood: 62, Flagged: true, Narrative: "Snap and pre-aim signals are elevated.",
					Channels:    []model.CheatChannel{{ID: "hs", Label: "Headshot rate", Score: 0, Confidence: 0.9, Zone: "clean", Raw: 48, Unit: "%", Samples: 20, HasData: true}},
					Adjustments: []model.Adjustment{{Label: "Evidence stacking boost", Detail: "two channels"}},
					Moments:     []model.CheatMoment{{Kind: "snap", Label: "Snap", Rank: 2, ValueText: "2.40 °/ms", Tick: 120, Seconds: 1.875, Round: 1, Value: 2.4, Unit: "°/ms", Opponent: "victim", Weapon: "AK-47"}}},
				Weapons:    map[string]*model.WeaponStats{"AK-47": {Name: "AK-47", Kills: 15, Shots: 150, Hits: 45, Accuracy: 30}},
				RoundStats: map[int]*model.RoundStat{1: {Side: model.SideT, Kills: 2, Damage: 200, EquipValue: 4000, Survived: true}},
				SideStats:  map[model.Side]*model.PlayerStats{model.SideT: {Rounds: 10, Kills: 10, Deaths: 5, Assists: 2, Damage: 900, ADR: 90, HSPercent: 40}},
			},
		},
		Rounds: []*model.Round{{
			Number: 1, StartTick: 100, EndTick: 500, StartOffset: 10, Duration: 90, Winner: model.SideT, WinnerTeamID: 2,
			Reason: "Bomb exploded", StartScore: [2]int{0, 0}, EndScore: [2]int{1, 0},
			Events:  []*model.Event{{Type: "kill", Offset: 20, ActorID: "1", ActorName: "kalle", VictimName: "victim", Weapon: "AK-47", Headshot: true}},
			Players: map[string]*model.RoundStat{"1": {Side: model.SideT, Kills: 1}},
		}},
		Warmup: &model.Warmup{Duration: 30, Events: []*model.Event{{Type: "kill", Offset: 5, ActorName: "kalle", VictimName: "victim", Weapon: "USP-S"}}},
	}
	return m
}

func TestCheatMomentsMergeByTickAndFeedWatchList(t *testing.T) {
	m := testMatch()
	m.Players[0].Cheat.Moments = append(m.Players[0].Cheat.Moments,
		model.CheatMoment{Kind: "ttd_sub100", Label: "Sub-100 ms", Rank: 3, ValueText: "90 ms", Tick: 120, Value: 90, Unit: "ms", Opponent: "victim", Weapon: "AK-47"},
		model.CheatMoment{Kind: "preaim", Label: "Pre-aim", Rank: 3, ValueText: "3.2°", Tick: 300, Value: 3.2, Unit: "°", Opponent: "other", Weapon: "AK-47"},
	)
	v := newView(testDocument(m, "test"), Options{})
	if len(v.Cheaters) != 1 {
		t.Fatalf("got %d watch-list players, want 1", len(v.Cheaters))
	}
	pv := v.Players[0]
	if pv.MomentCount != 2 || len(pv.Moments) != 2 {
		t.Fatalf("moments not merged by tick: count=%d len=%d", pv.MomentCount, len(pv.Moments))
	}
	if len(pv.Moments[0].Reasons) != 2 {
		t.Fatalf("same-tick reasons not merged: %+v", pv.Moments[0].Reasons)
	}
	if pv.Moments[0].Class != "cue-strong" {
		t.Fatalf("merged cue class: got %q, want cue-strong", pv.Moments[0].Class)
	}
	if pv.Moments[1].RoundAnchor != "round-100" {
		t.Fatalf("round anchor: got %q, want round-100", pv.Moments[1].RoundAnchor)
	}
}

func TestRenderWritesSelfDescribingHeader(t *testing.T) {
	doc := testDocument(nil, "26.11.1")
	var buf bytes.Buffer
	if err := Render(&buf, doc, Options{}); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	if !strings.HasPrefix(page, "<!doctype html>\n"+report.HeaderPrefix) {
		t.Fatalf("report does not start with the metadata comment: %q", page[:120])
	}
	if strings.Count(page, "<!doctype html>") != 1 {
		t.Fatal("report must contain exactly one doctype")
	}
	h, err := report.ReadHeaderFrom(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	if !h.Matches(doc.Header) {
		t.Fatalf("rendered header mismatch: %+v", h)
	}
	if !strings.Contains(page, "generated") || !strings.Contains(page, "26.11.1") {
		t.Fatal("report footer lost generation metadata")
	}
}

func TestRenderHeaderLinksToIndexWhenPresent(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, testDocument(nil, "test"), Options{IndexHref: IndexFileName}); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	for _, want := range []string{
		`class="brand" href="index.html"`,
		`aria-label="Back to report index"`,
		`id="search-index"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in report header", want)
		}
	}
	for _, unwanted := range []string{`class="nav-back"`, "All reports"} {
		if strings.Contains(page, unwanted) {
			t.Errorf("legacy header item %q is still rendered", unwanted)
		}
	}
}

func TestRenderHeaderWithoutIndexUsesOverview(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, testDocument(nil, "test"), Options{}); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	if !strings.Contains(page, `class="brand" href="#overview"`) {
		t.Fatal("brand should fall back to #overview without an index")
	}
	if strings.Contains(page, `class="nav-back"`) || strings.Contains(page, "All reports") {
		t.Fatal("index link should not render when no index is written")
	}
}

func TestRenderIncludesSearchableData(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, testDocument(nil, "test"), Options{}); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	for _, want := range []string{
		`id="topsearch"`,
		`id="report-search"`,
		`id="search-results"`,
		`role="combobox"`,
		`id="search-index"`,
		`data-target="player-kalle-000001"`,
		`data-search="1 &lt;b&gt;kalle&lt;/b&gt; kalle-000001 alpha`,
		`data-search="round 1 round 1 bomb exploded`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in rendered report", want)
		}
	}
	if strings.Contains(page, "search-status") || strings.Contains(page, "search-hidden") {
		t.Fatal("legacy live-search markup is still rendered")
	}
}

func TestRenderContainsCoreSectionsAndEscapesNames(t *testing.T) {
	m := testMatch()
	m.Players[0].Name = `<script>alert(1)</script>`
	doc := testDocument(m, "26.11.1")
	var buf bytes.Buffer
	if err := Render(&buf, doc, Options{DemoDownload: &DemoLink{Href: "match.dem", Name: `match"demo.dem`, Size: "12.0 MiB"}}); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	for _, want := range []string{"DEMO ANALYZER", "Overview", "Watchlist", "Timeline", "Players", "Aim", "Cheat analysis", "Warmup", "de_mirage", "26.11.1", "When to look", "demo_gototick 120", "tick 120", "Round history", "Download demo", "match.dem", "12.0 MiB"} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in rendered report", want)
		}
	}
	for _, unwanted := range []string{"Cheat watch", "Select a card", "Open report", "All reports"} {
		if strings.Contains(page, unwanted) {
			t.Errorf("legacy UI text %q is still rendered", unwanted)
		}
	}
	if !strings.Contains(page, `download="match&#34;demo.dem"`) {
		t.Fatal("demo download filename was not escaped")
	}
	if strings.Contains(page, "<script>alert(1)</script>") {
		t.Fatal("player name was not escaped")
	}
	if !strings.Contains(page, "&lt;script&gt;") {
		t.Fatal("escaped player name missing")
	}
	if strings.Contains(page, "{{") {
		t.Fatal("unrendered template action")
	}
}

func TestTimelineOrdersAbsoluteTimesAndLifecycleSections(t *testing.T) {
	m := testMatch()
	m.Warmup = &model.Warmup{StartOffset: 0, EndOffset: 10, Duration: 10, Events: []*model.Event{{Type: "kill", Offset: 5, ActorName: "kalle", VictimName: "victim", Weapon: "USP-S"}}}
	m.Rounds[0].StartOffset = 10
	m.Rounds[0].EndOffset = 100
	m.Rounds[0].Duration = 90
	m.Rounds[0].Events = []*model.Event{
		{Type: "kill", Offset: 20, ActorName: "kalle", VictimName: "victim", Weapon: "AK-47"},
		{Type: "round_end", Offset: 90, Text: "Bomb exploded", Score: [2]int{1, 0}},
	}
	m.Events = []*model.Event{
		{Type: "player_connect", Offset: 0, ActorName: "kalle"},
		{Type: "team_change", Offset: 5.5, ActorName: "kalle", OldSide: model.SideSpectator, NewSide: model.SideT},
		{Type: "team_swap", Offset: 105, Text: "Alpha now CT - Bravo now T"},
		{Type: "player_disconnect", Offset: 120, ActorName: "kalle"},
	}
	v := newView(testDocument(m, "test"), Options{})
	if len(v.Timeline) < 3 {
		t.Fatalf("timeline sections: got %d, want at least 3", len(v.Timeline))
	}
	if v.Timeline[0].Kind != "warmup" || v.Timeline[1].Kind != "round" || v.Timeline[2].Kind != "gap" {
		t.Fatalf("section order: %q %q %q", v.Timeline[0].Kind, v.Timeline[1].Kind, v.Timeline[2].Kind)
	}
	round := v.Timeline[1]
	if len(round.Rows) != 2 {
		t.Fatalf("round rows: got %d, want 2", len(round.Rows))
	}
	if round.Rows[0].AbsoluteTime != "0:30" || round.Rows[1].AbsoluteTime != "1:40" {
		t.Fatalf("absolute times: %q %q", round.Rows[0].AbsoluteTime, round.Rows[1].AbsoluteTime)
	}
	gap := v.Timeline[2]
	if len(gap.Rows) != 2 {
		t.Fatalf("gap rows: got %d, want 2", len(gap.Rows))
	}
	if !strings.Contains(gap.Rows[0].Parts[0].Text, "now CT") {
		t.Fatalf("team swap text missing: %+v", gap.Rows[0].Parts)
	}
}

func TestTimelineFallsBackToEconomyForTeamSwap(t *testing.T) {
	m := testMatch()
	m.Events = nil
	m.Rounds = []*model.Round{
		{Number: 1, StartTick: 100, EndTick: 500, StartOffset: 10, EndOffset: 100, Duration: 90, Winner: model.SideT, EndScore: [2]int{1, 0}, Economy: []*model.TeamEconomy{{TeamID: 2, Side: model.SideT}, {TeamID: 3, Side: model.SideCT}}},
		{Number: 2, StartTick: 600, EndTick: 900, StartOffset: 110, EndOffset: 180, Duration: 70, Winner: model.SideT, EndScore: [2]int{2, 0}, Economy: []*model.TeamEconomy{{TeamID: 2, Side: model.SideCT}, {TeamID: 3, Side: model.SideT}}},
	}
	v := newView(testDocument(m, "test"), Options{})
	found := false
	for _, section := range v.Timeline {
		for _, row := range section.Rows {
			if row.Type == "team_swap" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("fallback team swap event missing")
	}
}

func TestRenderTimelineLifecycleRowsAreEscaped(t *testing.T) {
	m := testMatch()
	m.Events = []*model.Event{
		{Type: "player_connect", Offset: 1, ActorName: `<script>connect</script>`},
		{Type: "team_change", Offset: 2, ActorName: "kalle", OldSide: model.SideSpectator, NewSide: model.SideCT},
		{Type: "team_swap", Offset: 3, Text: "Teams switched sides"},
	}
	var buf bytes.Buffer
	if err := Render(&buf, testDocument(m, "test"), Options{}); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	for _, want := range []string{
		`data-type="player_connect"`,
		`&lt;script&gt;connect&lt;/script&gt; connected`,
		`joined Counter-Terrorists`,
		`data-type="team_swap"`,
		`Teams switched sides`,
		`0:01`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in timeline", want)
		}
	}
	if strings.Contains(page, "<script>connect</script>") {
		t.Fatal("lifecycle player name was not escaped")
	}
}

func TestRenderMetadataFooterAndRemovedDisclaimer(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, testDocument(nil, "test"), Options{IndexHref: IndexFileName}); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	for _, want := range []string{
		`<meta name="author" content="cs2-demo-analyzer from kandru.net">`,
		`href="https://github.com/Kandru/cs2-demo-analyzer" target="_blank" rel="noopener noreferrer"`,
		`href="https://kandru.net" target="_blank" rel="noopener noreferrer"`,
		`>kandru.net</a>`,
		`rel="icon"`,
		`apple-touch-icon`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in report", want)
		}
	}
	if !strings.Contains(page, `name="description" content="`) || !strings.Contains(page, `by cs2-demo-analyzer from kandru.net">`) {
		t.Fatal("description metadata is missing the byline")
	}
	if strings.Contains(page, "Demo content is untrusted input") {
		t.Fatal("removed footer disclaimer is still present")
	}
}

func TestIndexMetadataAndSharedFooter(t *testing.T) {
	v := &IndexView{
		Title:       "Match reports",
		Version:     "test",
		GeneratedAt: "2026.11.16 12:11",
		Count:       1,
		Groups: []IndexServerGroup{{
			Name:    "Server",
			Entries: []IndexEntry{{Href: "match.html", Name: "Mirage", Played: "2026-11-15 12:00", Demo: &DemoLink{Href: "match.dem", Name: "match.dem", Size: "1.0 MiB"}}},
		}},
	}
	var buf bytes.Buffer
	if err := RenderIndex(&buf, v); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	for _, want := range []string{
		`<meta name="author" content="cs2-demo-analyzer from kandru.net">`,
		`href="https://github.com/Kandru/cs2-demo-analyzer" target="_blank" rel="noopener noreferrer"`,
		`href="https://kandru.net" target="_blank" rel="noopener noreferrer"`,
		`rel="icon"`,
		"CS2 Demo Analyzer",
		"vtest",
		"Download demo",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in index", want)
		}
	}
	for _, unwanted := range []string{"reports-title", "Select a card", "Open report", "Demo not linked"} {
		if strings.Contains(page, unwanted) {
			t.Errorf("legacy index markup %q is still rendered", unwanted)
		}
	}
}

func TestDisplayTeamNameNormalisesLegacyFallbacks(t *testing.T) {
	cases := []struct {
		name string
		team *model.Team
		want string
	}{
		{"empty", &model.Team{StartSide: model.SideCT}, "CT"},
		{"legacy A", &model.Team{Name: "Team A", StartSide: model.SideT}, "T"},
		{"legacy B", &model.Team{Name: "team b", StartSide: model.SideCT}, "CT"},
		{"real", &model.Team{Name: "Alpha", StartSide: model.SideT}, "Alpha"},
		{"unknown side", &model.Team{Name: "Team A"}, "Team A"},
		{"nil", nil, "Team"},
	}
	for _, tc := range cases {
		if got := displayTeamName(tc.team); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestRenderNormalisesLegacyTeamNamesAndScores(t *testing.T) {
	m := testMatch()
	m.Teams[0].Name = "Team A"
	m.Teams[0].ClanName = "Team A"
	m.Teams[1].Name = "Team B"
	m.Teams[1].ClanName = "Team B"
	m.Summary.WinnerTeamID = 2
	m.Summary.WinnerName = "T"
	var buf bytes.Buffer
	if err := Render(&buf, testDocument(m, "test"), Options{}); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	for _, want := range []string{
		`<span class="win">13</span>`,
		`<span class="loss">7</span>`,
		`class="team-score win"`,
		`class="team-score loss"`,
		`<span class="hero-team-name">T</span>`,
		`<span class="hero-team-name">CT</span>`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in normalised report", want)
		}
	}
	for _, unwanted := range []string{"Team A", "Team B", "match drawn", `hero-winner">Draw`} {
		if strings.Contains(page, unwanted) {
			t.Errorf("legacy score/team text %q is still rendered", unwanted)
		}
	}
}

func TestRenderDetailedTimelineEventsAndBuys(t *testing.T) {
	m := testMatch()
	m.Rounds[0].Players["1"].Bought = []string{"AK-47", "Kevlar Vest", "Smoke Grenade"}
	m.Rounds[0].Events = append(m.Rounds[0].Events,
		&model.Event{Type: "flash", Offset: 21, ActorID: "1", ActorName: "kalle", VictimID: "2", VictimName: "victim", Amount: 2.4},
		&model.Event{Type: "grenade_damage", Offset: 22, ActorID: "1", ActorName: "kalle", VictimID: "2", VictimName: "victim", Weapon: "HE Grenade", Amount: 43},
		&model.Event{Type: "bomb_pickup", Offset: 23, ActorID: "1", ActorName: "kalle"},
		&model.Event{Type: "bomb_drop", Offset: 24, ActorID: "1", ActorName: "kalle"},
		&model.Event{Type: "hostage_rescue", Offset: 25, ActorID: "1", ActorName: "kalle"},
	)
	var buf bytes.Buffer
	if err := Render(&buf, testDocument(m, "test"), Options{}); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	for _, want := range []string{
		"2.4s blind",
		"43 damage",
		"picked up the bomb",
		"dropped the bomb",
		"rescued a hostage",
		"AK-47",
		"Kevlar Vest",
		"Smoke Grenade",
		`class="buy-list"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in detailed timeline", want)
		}
	}
}

func TestRenderKeepsSearchIndexEscaped(t *testing.T) {
	m := testMatch()
	m.Players[0].Name = `<img src=x onerror=alert(1)>`
	m.Players[0].Cheat = nil
	var buf bytes.Buffer
	if err := Render(&buf, testDocument(m, "test"), Options{}); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	if strings.Contains(page, `<img src=x onerror=alert(1)>`) {
		t.Fatal("search index did not escape player name")
	}
	if !strings.Contains(page, "&lt;img src=x onerror=alert(1)&gt;") {
		t.Fatal("escaped search label missing")
	}
}

func TestRenderTimelineSaysFreezeTimeEnded(t *testing.T) {
	m := testMatch()
	m.Rounds[0].Events = append(m.Rounds[0].Events, &model.Event{Type: "freeze_end", Offset: 1})
	var buf bytes.Buffer
	if err := Render(&buf, testDocument(m, "test"), Options{}); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	if !strings.Contains(page, "Freeze time ended") || strings.Contains(page, "weapons are live") {
		t.Fatal("freeze-time label was not updated")
	}
}

func TestIndexOmitsDownloadWhenDisabled(t *testing.T) {
	v := &IndexView{
		Title:   "Match reports",
		Version: "test",
		Count:   1,
		Groups: []IndexServerGroup{{
			Name:    "Server",
			Entries: []IndexEntry{{Href: "match.html", Name: "Mirage", Played: "2026-11-15 12:00"}},
		}},
	}
	var buf bytes.Buffer
	if err := RenderIndex(&buf, v); err != nil {
		t.Fatal(err)
	}
	page := buf.String()
	if strings.Contains(page, "Download demo") || strings.Contains(page, "Demo not linked") {
		t.Fatal("index rendered a demo link or placeholder when no download is enabled")
	}
}

func TestTimelineWarmupOnlyDemo(t *testing.T) {
	m := &model.Match{
		SchemaVersion: model.SchemaVersion,
		Map:           "de_mirage",
		Mode:          "Competitive",
		Warmup: &model.Warmup{
			StartOffset: 0,
			Duration:    30,
			Events: []*model.Event{
				{Type: "kill", Offset: 5, ActorName: "kalle", VictimName: "victim"},
				{Type: "kill", Offset: 10, ActorName: "kalle", VictimName: "victim"},
			},
		},
		Events: []*model.Event{{Type: "player_connect", Offset: 1, ActorName: "kalle"}},
	}
	v := newView(testDocument(m, "test"), Options{})
	if len(v.Timeline) != 1 || v.Timeline[0].Kind != "warmup" {
		t.Fatalf("warmup-only timeline: %+v", v.Timeline)
	}
	if len(v.Timeline[0].Rows) != 3 {
		t.Fatalf("warmup rows: got %d, want 3", len(v.Timeline[0].Rows))
	}
	if v.Timeline[0].EndTime != "0:30" {
		t.Fatalf("warmup end time: %q", v.Timeline[0].EndTime)
	}
}
