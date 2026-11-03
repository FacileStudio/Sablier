package schemas

import (
	"testing"
	"time"
)

func TestAgentFromName(t *testing.T) {
	got := AgentFromName("claude-code")
	if got.Key != "claude-code" || got.Name != "Claude Code" || got.AvatarURL != "https://blobatar.dev/avatar/Claude%20Code?gen=2" {
		t.Fatalf("got %+v", got)
	}
	if empty := AgentFromName(" "); empty.Key != "agent" || empty.Name != "Agent" {
		t.Fatalf("empty name: %+v", empty)
	}
}

func TestBackfillAgentEntriesIsIdempotent(t *testing.T) {
	orm := openTestDatabase(t)
	if err := Migrate(orm, ""); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	owner := User{Email: "sara@example.com", Name: "Sara"}
	if err := orm.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	stopped := time.Now().UTC()
	ids := []string{"fac_1", "fac_2", "fac_3"}
	descriptions := []string{"mycelium: claude@lucy branch=main", "mycelium: claude@nova", "manual work"}
	for i := range ids {
		entry := TimeEntry{UserID: owner.ID, FacileID: &ids[i], LegacyDescription: descriptions[i], StartedAt: stopped.Add(-time.Hour), StoppedAt: &stopped}
		if err := orm.Create(&entry).Error; err != nil {
			t.Fatal(err)
		}
	}

	for run := 0; run < 2; run++ {
		if err := Migrate(orm, ""); err != nil {
			t.Fatalf("migrate run %d: %v", run, err)
		}
	}

	var agents []User
	if err := orm.Where("kind = ?", UserKindAgent).Find(&agents).Error; err != nil {
		t.Fatal(err)
	}
	if len(agents) != 1 || agents[0].Name != "Claude" || *agents[0].OwnerUserID != owner.ID {
		t.Fatalf("want one Claude agent owned by Sara, got %+v", agents)
	}
	var moved, kept int64
	orm.Model(&TimeEntry{}).Where("user_id = ?", agents[0].ID).Count(&moved)
	orm.Model(&TimeEntry{}).Where("user_id = ?", owner.ID).Count(&kept)
	if moved != 2 || kept != 1 {
		t.Fatalf("moved=%d kept=%d, want 2 and 1", moved, kept)
	}
}
