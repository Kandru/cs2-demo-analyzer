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
		DemoKey:         report.DemoKey("/demos/match.dem"),
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
		`class="nav-back" href="index.html"`,
		`aria-label="Back to report index"`,
		"All reports",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in report header", want)
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
		`id="search-status"`,
		`data-target="player-kalle-000001"`,
		`data-search="1 &lt;b&gt;kalle&lt;/b&gt; kalle-000001 alpha`,
		`data-search="round 1 round 1 bomb exploded`,
		"search-hidden",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in rendered report", want)
		}
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
	for _, want := range []string{"DEMO ANALYZER", "Scoreboard", "Rounds", "Players", "Aim", "Cheat analysis", "Warmup", "de_mirage", "26.11.1", "Cheat watch", "When to look", "demo_gototick 120", "tick 120", "Round history", "Download demo", "match.dem", "12.0 MiB"} {
		if !strings.Contains(page, want) {
			t.Errorf("missing %q in rendered report", want)
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
