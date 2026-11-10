package mcp_test

import (
	"testing"
	"time"
)

// summaryOut is the part of a weekly summary the test reads.
type summaryOut struct {
	WeekStart string `json:"week_start"`
	Total     struct {
		Minutes int `json:"minutes"`
	} `json:"total"`
	Days []struct {
		Weekday string `json:"weekday"`
		Minutes int    `json:"minutes"`
	} `json:"days"`
	ByPerson []struct{ Name string } `json:"by_person"`
}

// TestMCPWeeklySummary adds up a week in the user's timezone, splitting an entry over
// midnight, and widens to the team with everyone.
func TestMCPWeeklySummary(t *testing.T) {
	f := newFixture(t)
	paris, _ := time.LoadLocation("Europe/Paris")
	f.entry(t, "owner", time.Date(2026, 11, 9, 9, 0, 0, 0, paris), 2*time.Hour)
	f.entry(t, "owner", time.Date(2026, 11, 10, 23, 0, 0, 0, paris), 2*time.Hour)
	f.entry(t, "teammate", time.Date(2026, 11, 11, 14, 0, 0, 0, paris), time.Hour)
	f.entry(t, "owner", time.Date(2026, 11, 16, 9, 0, 0, 0, paris), time.Hour)
	session := f.connect(t, "owner")
	var mine, team summaryOut
	args := map[string]any{"week_of": "2026-11-11", "timezone": "Europe/Paris"}
	call(t, session, "sablier_weekly_summary", args, &mine)
	if mine.WeekStart != "2026-11-09" || mine.Total.Minutes != 240 || len(mine.ByPerson) != 0 {
		t.Fatalf("want 240 own minutes from Monday 9th, got %+v", mine)
	}
	if mine.Days[0].Minutes != 120 || mine.Days[1].Minutes != 60 || mine.Days[2].Minutes != 60 || mine.Days[2].Weekday != "Wednesday" {
		t.Fatalf("days must split the entry over midnight, got %+v", mine.Days)
	}
	args["everyone"] = true
	call(t, session, "sablier_weekly_summary", args, &team)
	if team.Total.Minutes != 300 || len(team.ByPerson) != 2 {
		t.Fatalf("everyone adds the teammate's hour, got %+v", team)
	}
}
