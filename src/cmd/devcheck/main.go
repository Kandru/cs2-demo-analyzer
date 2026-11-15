package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/Kandru/cs2-demo-analyzer/internal/model"
	"github.com/Kandru/cs2-demo-analyzer/internal/parser"
)

func main() {
	m, err := parser.Parse(context.Background(), os.Args[1], parser.Options{Logger: slog.Default()})
	if err != nil {
		fmt.Println("ERR", err)
	}
	if m == nil {
		return
	}
	fmt.Printf("map=%s mode=%s ticks=%d dur=%.1f teams=%d players=%d rounds=%d warmup=%v partial=%v warnings=%v\n", m.Map, m.Mode, m.Ticks, m.Duration, len(m.Teams), len(m.Players), len(m.Rounds), m.Warmup != nil, m.Partial, m.Warnings)
	for _, t := range m.Teams {
		fmt.Printf("TEAM %d %q side=%s final=%s score=%d\n", t.ID, t.Name, t.StartSide, t.FinalSide, t.Score)
	}
	for _, p := range m.Players {
		a, g := aimval(p)
		fmt.Printf("P %s %q team=%d k=%d d=%d a=%d dmg=%d adr=%.1f hs=%.1f aim=%.0f %s shots=%d hits=%d head=%d\n",
			p.ID, p.Name, p.TeamID, p.Stats.Kills, p.Stats.Deaths, p.Stats.Assists, p.Stats.Damage, p.Stats.ADR, p.Stats.HSPercent, a, g, p.Stats.Shots, p.Stats.Hits, p.Stats.HeadHits)
		if p.Aim != nil {
			for _, mt := range p.Aim.Metrics {
				fmt.Printf("   AIM %-14s %s score=%.0f n=%d\n", mt.ID, mt.Display, mt.Score, mt.Samples)
			}
		}
	}
	for _, r := range m.Rounds {
		fmt.Printf("R %d pre=%v win=%s %s score=%v->%v events=%d dur=%.0f\n", r.Number, r.Pregame, r.Winner, r.Reason, r.StartScore, r.EndScore, len(r.Events), r.Duration)
	}
}

func aimval(p *model.Player) (float64, string) {
	if p.Aim == nil {
		return 0, "-"
	}
	return p.Aim.Rating, p.Aim.Grade
}
