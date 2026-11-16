package schemas_test

import (
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

func TestLastActivityFollowsTheHeartbeatChainOnly(t *testing.T) {
	orm := testdb.Migrated(t)
	start := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	if err := orm.Create(&schemas.User{ID: 1, Email: "me@example.test"}).Error; err != nil {
		t.Fatal(err)
	}
	for i, offset := range []time.Duration{10 * time.Minute, 20 * time.Minute, 3 * time.Hour} {
		beat := schemas.Heartbeat{UserID: 1, At: start.Add(offset), Source: "desktop", Entity: string(rune('a' + i))}
		if err := orm.Create(&beat).Error; err != nil {
			t.Fatal(err)
		}
	}
	checked := start.Add(4 * time.Hour)
	cases := []struct {
		name  string
		entry schemas.TimeEntry
		want  time.Time
	}{
		{"chain then gap", schemas.TimeEntry{UserID: 1, StartedAt: start}, start.Add(20 * time.Minute)},
		{"checkpoint after every beat", schemas.TimeEntry{UserID: 1, StartedAt: start, CheckedAt: &checked}, checked},
		{"no beats", schemas.TimeEntry{UserID: 2, StartedAt: start}, start},
	}
	for _, c := range cases {
		got, err := schemas.LastActivity(orm, c.entry, 15*time.Minute, start.Add(24*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		if !got.Equal(c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
